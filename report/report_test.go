package report

import (
	"os"
	"strings"
	"testing"

	"tweet-audit/audit"
)

func TestWriteCSV(t *testing.T) {
	verdicts := []audit.Verdict{
		{
			TweetID:   "123",
			Username:  "testuser",
			URL:       "https://x.com/testuser/status/123",
			Text:      "Hello world",
			CreatedAt: "Mon Jan 01 12:00:00 +0000 2026",
			Flag:      true,
			Reason:    "Informal language",
		},
		{
			TweetID:   "456",
			Username:  "testuser",
			URL:       "https://x.com/testuser/status/456",
			Text:      "Professional update",
			CreatedAt: "Tue Jan 02 12:00:00 +0000 2026",
			Flag:      false,
			Reason:    "No issues",
		},
	}

	tmpfile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	err = WriteCSV(tmpfile.Name(), verdicts)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	csv := string(content)

	// Verify header
	if !strings.Contains(csv, "Tweet URL") {
		t.Error("CSV missing 'Tweet URL' header")
	}
	if !strings.Contains(csv, "Flag") {
		t.Error("CSV missing 'Flag' header")
	}
	if !strings.Contains(csv, "Reason") {
		t.Error("CSV missing 'Reason' header")
	}

	// Verify first verdict (flagged)
	if !strings.Contains(csv, "https://x.com/testuser/status/123") {
		t.Error("CSV missing first tweet URL")
	}
	if !strings.Contains(csv, "Hello world") {
		t.Error("CSV missing first tweet text")
	}
	if !strings.Contains(csv, "true") {
		t.Error("CSV missing flag=true for first tweet")
	}
	if !strings.Contains(csv, "Informal language") {
		t.Error("CSV missing reason for first tweet")
	}

	// Verify second verdict (not flagged)
	if !strings.Contains(csv, "https://x.com/testuser/status/456") {
		t.Error("CSV missing second tweet URL")
	}
	if !strings.Contains(csv, "Professional update") {
		t.Error("CSV missing second tweet text")
	}
	if !strings.Contains(csv, "false") {
		t.Error("CSV missing flag=false for second tweet")
	}
}

func TestWriteCSVEmpty(t *testing.T) {
	verdicts := []audit.Verdict{}

	tmpfile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	err = WriteCSV(tmpfile.Name(), verdicts)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	csv := string(content)

	// Should have header only
	if !strings.Contains(csv, "Tweet URL") {
		t.Error("CSV missing header even with no data")
	}

	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) != 1 {
		t.Errorf("empty verdicts CSV has %d lines, want 1 (header only)", len(lines))
	}
}

func TestWriteCSVWithSpecialChars(t *testing.T) {
	// Test that CSV escaping handles commas and quotes
	verdicts := []audit.Verdict{
		{
			TweetID:   "789",
			Username:  "user",
			URL:       "https://x.com/user/status/789",
			Text:      `Tweet with comma, and "quote"`,
			CreatedAt: "Wed Jan 03 12:00:00 +0000 2026",
			Flag:      true,
			Reason:    `Reason with, comma`,
		},
	}

	tmpfile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	err = WriteCSV(tmpfile.Name(), verdicts)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	csv := string(content)

	// The content should still be readable (csv.Writer handles escaping)
	if !strings.Contains(csv, "Tweet with comma") {
		t.Error("CSV missing tweet text with comma")
	}
}
