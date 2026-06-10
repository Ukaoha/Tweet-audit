# Tradeoffs & Design Decisions

This document summarizes the architectural choices for `tweet-audit`.

## 1. Language & Concurrency: Go
**Decision:** Used Go for the core implementation.
**Reasoning:** 
- **Concurrency primitives**: I used Go’s goroutines and channels because they are significantly more efficient than OS threads or the `async/await` event loop in JavaScript. I can manage thousands of concurrent tasks with minimal memory overhead (kilobytes instead of megabytes per task).
- **Static Binary**: Go compiles all dependencies into a single executable. This is a massive win for a CLI tool, as users don't need to install or manage a runtime environment (like a Python interpreter or Node.js) to run the audit.
**Tradeoff:** Go has a stricter type system and more verbose error handling than Python, which slightly increased development time but ensures the tool is predictable and reliable under load.

## 2. Concurrency Strategy: Batching vs. Sequential vs. Full Async
**Decision:** Implementation uses a fixed worker pool for Batching.
**Reasoning:** 
- **Sequential processing**: Auditing one tweet at a time is the safest but slowest approach. At 1.5 seconds per API call, 5,000 tweets would take over 2 hours.
- **Full Async (Unbounded)**: Spawning a goroutine for every single tweet would trigger Gemini's 429 rate limits and potentially cause local "too many open files" errors.
- **Batching (Worker Pool)**: By using 8 parallel workers, I achieved balanced throughput. I can saturate the permitted API rate limits (100 requests/minute) consistently without overwhelming local resources.
**Tradeoff:** Result ordering is not preserved in the CSV, but chronological order is secondary to speed for this task.

## 3. Resilience: Resumable Audits (Persistence)
**Decision:** Implement a checkpoint-based persistence layer using an atomic JSON store.
**Reasoning:** Gemini API calls are the most time-consuming part of the process. Saving progress to `audit-checkpoint.json` ensures a network failure doesn't require spending more quota re-auditing already-processed tweets.
**Tradeoff:** Writing to disk frequently adds small latency, but using the `os.Rename` atomic swap pattern prevents file corruption during crashes.

## 4. Architecture: Separation of Concerns
**Decision:** Split the project into modular packages (`archive`, `gemini`, `audit`, `checkpoint`) instead of a single `main.go`.
**Reasoning:** 
- **Maintainability**: If I decide to swap Gemini for OpenAI or Claude, I only need to update the `gemini` package. The rest of the application remains untouched.
- **Independent Testing**: Each package has its own unit tests. I can verify the archive parsing logic or criteria formatting without ever needing to make a live network call.
**Tradeoff:** Slightly more initial boilerplate and "folder sprawl" compared to a monolithic 1,000-line `main.go` file.

## 5. Error Handling: Resilience vs. Fail-Fast
**Decision:** Exponential Backoff for retries; Log-and-Continue for the overall process.
**Reasoning:** 
- **Exponential Backoff**: When the API returns 429 (Rate Limit) or transient 5xx errors, I retry with increasing delays (500ms, 1s, 2s).
- **Log-and-Continue**: If a tweet fails after all retries, I log it and move to the next. This ensures an audit isn't ruined by a single failure.
**Tradeoff:** Users must check logs for skipped tweets, but they get nearly-complete results.

#
