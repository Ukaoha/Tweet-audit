package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"tweet-audit/archive"
	"tweet-audit/audit"
	"tweet-audit/checkpoint"
	"tweet-audit/gemini"
	"tweet-audit/report"
)

func main() {
	dataDir := flag.String("data", "", "path to the archive's data/ directory")
	criteriaFile := flag.String("criteria", "criteria.json", "path to criteria JSON file")
	outputCSV := flag.String("output", "flagged-tweets.csv", "output CSV file path")
	checkpointFile := flag.String("checkpoint", "audit-checkpoint.json", "path to checkpoint file for resumable audits")
	dryRun := flag.Bool("dry-run", false, "print verdicts without writing CSV")
	limit := flag.Int("limit", 0, "limit audit to first N tweets (0 = all)")
	flag.Parse()

	if *dataDir == "" {
		log.Fatal("missing -data flag: path to the archive's data/ directory")
	}

	// Load account and tweets
	account, err := archive.LoadAccount(filepath.Join(*dataDir, "account.js"))
	if err != nil {
		log.Fatal(err)
	}

	tweets, err := archive.LoadTweets(filepath.Join(*dataDir, "tweets.js"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded %d tweets for @%s\n", len(tweets), account.Username)

	// Load criteria
	criteriaData, err := os.ReadFile(*criteriaFile)
	if err != nil {
		log.Fatalf("reading criteria file: %v", err)
	}

	var criteria audit.Criteria
	if err := json.Unmarshal(criteriaData, &criteria); err != nil {
		log.Fatalf("parsing criteria: %v", err)
	}

	// Get API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	// Create auditor
	client := gemini.NewClient(apiKey)
	auditor := audit.NewAuditor(client, criteria, account.Username)

	// Determine tweet slice
	ctx := context.Background()
	toAudit := len(tweets)
	if *limit > 0 && *limit < toAudit {
		toAudit = *limit
	}

	// Run audit with persistence
	verdicts := auditConcurrent(ctx, auditor, tweets[:toAudit], *checkpointFile)
	fmt.Printf("Audit complete: %d verdicts processed\n", len(verdicts))

	// Count flagged
	flaggedCount := 0
	for _, v := range verdicts {
		if v.Flag {
			flaggedCount++
		}
	}

	fmt.Printf("\nAudit complete: %d flagged (out of %d)\n", flaggedCount, len(verdicts))

	// Output
	if *dryRun {
		fmt.Println("\n--- Dry run: flagged tweets ---")
		for _, v := range verdicts {
			if v.Flag {
				fmt.Printf("%s | %s\n", v.URL, v.Reason)
			}
		}
	} else {
		if err := report.WriteCSV(*outputCSV, verdicts); err != nil {
			log.Fatalf("writing csv: %v", err)
		}
		fmt.Printf("CSV written to %s\n", *outputCSV)
	}
}

// auditConcurrent processes tweets concurrently using worker goroutines.
// It loads a checkpoint file to skip already-processed tweets and saves each
// verdict to disk immediately so interrupted runs can resume safely.
func auditConcurrent(ctx context.Context, auditor *audit.Auditor, tweets []archive.Tweet, checkpointPath string) []audit.Verdict {
	const numWorkers = 8
	const batchSize = 100

	// Load existing checkpoint (empty store if file doesn't exist)
	cp, err := checkpoint.Load(checkpointPath)
	if err != nil {
		log.Printf("warning: could not load checkpoint: %v (starting fresh)", err)
		cp, _ = checkpoint.Load(checkpointPath + ".reset")
	}

	skipped := cp.Len()
	if skipped > 0 {
		fmt.Printf("Resuming: skipping %d already-processed tweets\n", skipped)
	}

	// Filter out already-processed tweets
	var pending []archive.Tweet
	for _, tw := range tweets {
		if !cp.Has(tw.IDStr) {
			pending = append(pending, tw)
		}
	}

	fmt.Printf("Processing %d tweets (%d new after checkpoint)\n", len(tweets), len(pending))

	// Channels
	tweetChan := make(chan archive.Tweet, numWorkers)
	resultChan := make(chan audit.Verdict, numWorkers)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tw := range tweetChan {
				v, err := auditor.Audit(ctx, tw)
				if err != nil {
					log.Printf("error auditing tweet %s: %v", tw.IDStr, err)
					continue
				}
				resultChan <- v
			}
		}()
	}

	// Send pending tweets to workers
	go func() {
		for i, tw := range pending {
			if i%batchSize == 0 {
				fmt.Printf("Processing [%d/%d]\n", i, len(pending))
			}
			tweetChan <- tw
		}
		close(tweetChan)
	}()

	// Collect results; persist each verdict immediately for crash safety
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var mu sync.Mutex
	for v := range resultChan {
		mu.Lock()
		cp.Add(v)
		if err := cp.Save(); err != nil {
			log.Printf("warning: checkpoint save failed: %v", err)
		}
		mu.Unlock()
	}

	// Return all verdicts (previously stored + newly processed)
	return cp.All()
}
