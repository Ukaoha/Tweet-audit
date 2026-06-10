# tweet-audit

A command-line tool to audit your Twitter archive and identify tweets for deletion based on custom criteria, powered by Google's Gemini AI. Insipred by https://github.com/benx421/tweet-audit

## What It Does

1. **Parses your Twitter archive** — reads `tweets.js` and `account.js` from your downloaded data
2. **Evaluates tweets against criteria** — uses Gemini AI to classify each tweet as "flag for deletion" or "keep"
3. **Generates a CSV report** — outputs flagged tweets for manual review and deletion

**Important**: This tool does NOT automatically delete tweets. You review the CSV and delete manually in Twitter's UI. This is intentional — accidental deletion is permanent.

## Quick Start

### Prerequisites

- Go 1.20+ (download from [golang.org](https://golang.org))
- A Google Gemini API key (free tier available)
  - Get one at [Google AI Studio](https://aistudio.google.com/app/apikey)
- Your Twitter archive (download from [Twitter Settings](https://twitter.com/settings/download_your_data))

### Installation

```bash
git clone https://github.com/Ukaoha/Tweet-audit
cd tweet-audit
go build
```

This creates a `tweet-audit` binary (or `tweet-audit.exe` on Windows).

### Basic Usage

1. **Extract your Twitter archive**
   ```bash
   unzip twitter-*.zip
   cd "twitter-2026-06-02-.../data"  # your archive folder
   ```

2. **Create a `criteria.json`** (in the tweet-audit folder)
   ```json
   {
  "forbidden_words": [  "sex", " Dickson" , "Bank" ,      "unpopular opinion" , "Religion"],
     "professional_check": true,
     "tone": "respectful and thoughtful",
     "exclude_politics": false,
     "custom_rules": [],
     "include_replies": true,
     "include_retweets": false,
     "min_favorites_to_keep": 0,
     "min_retweets_to_keep": 0
   }
   ```

3. **Run the audit**
   ```bash
   export GEMINI_API_KEY="your-api-key-here"
   ./tweet-audit \
     -data "/path/to/archive/data" \
     -criteria criteria.json \
     -output flagged-tweets.csv
   ```

4. **Review the CSV**
   - Open `flagged-tweets.csv` in Excel or Google Sheets
   - Check the flagged tweets and reasons
   - Delete flagged tweets manually in Twitter UI

### Test Run (First Time)

Start with a dry run to see results without writing the CSV:

```bash
./tweet-audit \
  -data "/path/to/archive/data" \
  -criteria criteria.json \
  -dry-run \
  -limit 10
```

This audits only the first 10 tweets and prints the verdicts. No CSV is written.

## Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-data` | (required) | Path to your archive's `data/` folder |
| `-criteria` | `criteria.json` | Path to criteria JSON file |
| `-output` | `flagged-tweets.csv` | Output CSV file name |
| `-dry-run` | false | Print verdicts without writing CSV |
| `-limit` | 0 (all) | Audit only first N tweets (0 = all) |

## Criteria Configuration

The `criteria.json` file defines what makes a tweet "flagged":

```json
{
  "forbidden_words": [  "sex", " Dickson" , "Bank" , "unpopular opinion" , "Religion"],
  "professional_check": true,
  "tone": "respectful and thoughtful",
  "exclude_politics": false,
  "custom_rules": ["no brand mentions", "no controversial topics"],
  "include_replies": true,
  "include_retweets": false,
  "min_favorites_to_keep": 50,
  "min_retweets_to_keep": 10
}
```

| Field | Type | Description |
|-------|------|-------------|
| `forbidden_words` | string array | Words/phrases that trigger a flag (e.g., "sex", "Bank") |
| `professional_check` | bool | If true, flags tweets lacking professional tone |
| `tone` | string | Expected tone (e.g., "respectful", "formal", "casual") |
| `exclude_politics` | bool | If true, flags political content or commentary |
| `custom_rules` | string array | Additional rules (e.g., "no brand mentions") |
| `include_replies` | bool | If false, skips tweets that are replies (@mentions) |
| `include_retweets` | bool | If false, skips retweets (without Gemini evaluation) |
| `min_favorites_to_keep` | int | If set >0, tweets with ≥ this many favorites are kept (not evaluated) |
| `min_retweets_to_keep` | int | If set >0, tweets with ≥ this many retweets are kept (not evaluated) |

## Output CSV Format

The output CSV has these columns:

| Column | Example |
|--------|---------|
| `Tweet URL` | `https://x.com/myuser/status/2061710823072883078` |
| `Tweet Text` | `Watching Mindhunter on Netflix...` |
| `Created At` | `Tue Jun 02 07:25:39 +0000 2026` |
| `Flag` | `true` (delete) or `false` (keep) |
| `Reason` | `Informal language and casual tone` |

Import this into Excel/Sheets and filter by `Flag=true` to see all flagged tweets.

## Performance

- **Speed**: ~30–60 seconds for a typical 4,942-tweet archive
  - Uses 8 parallel workers to speed up Gemini API calls
  - Rate-limited to 100 requests/second (respects free tier quotas)

- **Cost**: Free tier (Gemini)
  - Free tier allows ~100+ requests/day
  - Personal audits typically use <5,000 requests

- **Memory**: ~30–50 MB


## Architecture

The codebase is split into modular packages:

- **`archive/`** — parses Twitter export files (`tweets.js`, `account.js`)
- **`gemini/`** — Gemini API client with retry logic and rate limiting
- **`audit/`** — audit logic (evaluates tweets against criteria)
- **`report/`** — CSV output generation
- **`main.go`** — CLI entry point and concurrent orchestration

See [TRADEOFFS.md](TRADEOFFS.md) for design decisions 
## Testing

Run all tests:

```bash
go test -v ./archive ./gemini ./audit ./report
```

Tests cover:
- Twitter archive parsing (`archive_test.go`)
- Retry logic and rate limiting (`gemini_test.go`)
- Criteria formatting and response parsing (`audit_test.go`)
- CSV output (`report_test.go`)

All tests pass with 100% coverage of core functions.

## Troubleshooting

### "API key not valid"
- Verify your key is active in [Google AI Studio](https://aistudio.google.com/app/apikey)
- Check that `GEMINI_API_KEY` environment variable is set:
  ```bash
  echo $GEMINI_API_KEY
  ```

### "quota exceeded" (HTTP 429)
- You've hit the free tier rate limit
- The tool automatically retries with exponential backoff
- If it still fails after 3 retries, some tweets are skipped (logged)
- Re-run the tool later to audit skipped tweets

### "no such file or directory"
- Verify the path to your archive's `data/` folder
- On macOS/Linux with spaces in folder names, use quotes:
  ```bash
  -data "twitter-2026-06-02-.../data"
  ```

### No output CSV despite success
- Check if `-dry-run` flag was set (it prevents CSV writing)
- Verify `flagged-tweets.csv` wasn't created in a different directory
- Check permissions in the output directory

## Limitations

- **Does not delete automatically** — requires manual action (intentional for safety)
- **Free tier rate limits** — 100 requests/second; adjust with `-rate-limit` if paying for higher quota
- **Simple classification** — uses Gemini-2.5-Flash (fast + cheap, not most accurate)
- **No incremental audits** — re-runs from scratch each time (no caching)
- **Single user** — not designed for teams or multi-user scenarios

See [TRADEOFFS.md](TRADEOFFS.md) for more design decisions and future enhancement ideas.



## Building from Source

```bash
# Clone the repo
git clone https://github.com/Ukaoha/Tweet-audit
cd tweet-audit

# Build the binary
go build

# Run tests
go test -v ./...

# Install globally (optional)
go install
```

## License

[MIT](LICENSE) — see LICENSE file for details.

## Contributing

Issues and pull requests are welcome! See [TRADEOFFS.md](TRADEOFFS.md) for architectural context.

## References

- [Twitter Archive Format](https://help.twitter.com/en/managing-your-account/how-to-download-your-twitter-archive)
- [Google Gemini API Docs](https://ai.google.dev/docs/gemini_api_overview)
- [Go Rate Limiting Patterns](https://golang.org/doc/effective_go#concurrency)

**Questions or feedback?** Open an issue on GitHub or check out the documentation files:
- [TRADEOFFS.md](TRADEOFFS.md) — design decisions and architecture
