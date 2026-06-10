# Tradeoffs & Design Decisions

This document summarizes the architectural choices for `tweet-audit`.

## 1. Language & Concurrency: Go
**Decision:** Used Go for the core implementation.
**Reasoning:** 
- **Concurrency primitives**: I used Go’s goroutines and channels because they  are significantly more efficient than OS threads or the `async/await` event loop  in JavaScript. With Concurrency I can manage thousands of concurrent tasks with minimal memory overhead (kilobytes instead of megabytes per task).
- **Type Safety**: Using Go ensures that criteria parsing and API response handling are verified at compile time, reducing the risk of runtime crashes compared to dynamic languages.
**Tradeoff:** Go has a stricter type system and more verbose error handling than Python, which slightly increased initial development time but ensures the tool is predictable and reliable under load.

## 2. Concurrency Strategy: Batching vs. Sequential vs. Full Async
**Decision:** Implementation uses a fixed worker pool to manage concurrent execution (Batching).
**Batching (Worker Pool)**: By using a defined number of parallel workers (default 8), I achieved a balanced throughput. I can saturate the permitted API rate limits (100 requests/minute on the free tier) consistently without overwhelming local resources or the remote server.
- **Sequential processing**: Auditing one tweet at a time is the safest but slowest approach. At an average of 1.5 seconds per Gemini API call, an archive of 5,000 tweets would take over 2 hours to finish.
- **Full Async (Unbounded)**: Spawning a goroutine for every single tweet would attempt to launch 5,000+ network requests simultaneously. This would immediately trigger the Gemini API's rate limits (429 errors) and potentially cause local "too many open files" errors on the host machine.
**Tradeoff:** The order of results is not preserved—tweet #100 might be processed and recorded before tweet #1. However, since the final output is a CSV for manual review, the original chronological order is secondary to speed.
# 3. Resilience: Resumable Audits (Persistence)
**Decision:** Implement a checkpoint-based persistence layer using an atomic JSON store.
**Reasoning:** Gemini API calls are the most time-consuming and quota-heavy part of the process. If the program crashes after 4,000 tweets due to a network failure, the user shouldn't have to spend more quota re-auditing them.
**Tradeoff:** Writing to disk after every tweet adds a small amount of latency, but we ensure data integrity by using the `os.Rename` atomic swap pattern to prevent file corruption.

## 4. API & Model: Gemini Flash
**Decision:** Base the auditing engine on Google Gemini 2.5 Flash via a raw HTTP client.
**Reasoning:** Flash is 90% cheaper and significantly faster than the "Pro" or "Ultra" models. For the task of binary classification (flag/keep based on text), Flash's intelligence is more than sufficient.
**Tradeoff:** Flash might miss extremely subtle sarcasm that a larger model would catch, but the dry-run and CSV-review steps provide a safety net for the user.
