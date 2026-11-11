# Promise lifecycle investigation and fix plan

Issue 1 in FINDINGS.md was confirmed by executable reproductions and is now fixed.
The promise retains its result and closes a completion channel to notify waiters.
The lifecycle regression tests pass with the race detector.

Run the active regression suite with:

```sh
go test -race ./locate -run '^TestPromise' -count=1 -timeout=30s
```

## Evidence

`locate/promise_lifecycle_test.go` contains regression assertions originally gated
behind `promise_knownbugs`. The tag has now been removed and the tests run normally.
Tests use isolated registries and no network.
`testing/synctest` detects blocked goroutines without sleeps or goroutine-count
heuristics. The module already requires Go 1.25.6.

The following results and reproduction commands record the behavior before the fix.

| Test | Original result |
| --- | --- |
| TestPromiseRepeatedAwait | Fails for success, lookup miss, and canceled search: second await returns an empty font and nil error. |
| TestPromiseConcurrentAwait | Fails: seven of eight waiters receive an empty font and nil error. |
| TestPromiseNoWaiter | Fails with a synctest deadlock: publisher is stuck sending at resolve.go:103. |
| TestPromiseCanceledWaiter | Fails with the same deadlock after a canceled waiter leaves. |
| TestPromiseWaitCancellationDoesNotCancelSearch | Passes: a later await receives the completed search result. |
| TestPromiseSearchCancellationReachesResolver | Passes: a cooperative resolver observes search cancellation. |

Run the non-deadlock cases together:

```sh
go test -tags promise_knownbugs ./locate -run '^TestPromise(RepeatedAwait|ConcurrentAwait|WaitCancellationDoesNotCancelSearch|SearchCancellationReachesResolver)$' -count=1 -timeout=20s
```

Run the deadlock cases separately because each currently panics the test process:

```sh
go test -tags promise_knownbugs ./locate -run '^TestPromiseNoWaiter$' -count=1 -timeout=20s
go test -tags promise_knownbugs ./locate -run '^TestPromiseCanceledWaiter$' -count=1 -timeout=20s
```

## Cause

Resolve launches a goroutine that sends exactly one result on an unbuffered
channel, then closes the channel. Publication therefore requires a waiter.
Once the single result is consumed, every subsequent receive reads the zero
value of fontPlusErr from the closed channel. Both the font and error are lost.

A one-element buffer fixes blocked publication but does not fix repeated or
concurrent awaits.

## Implemented contract

- A promise represents one search and retains its completed font/error pair.
- Every successful await of completion returns that same pair, including a
  fallback with a lookup error or a search cancellation error.
- Completion never requires a waiter. Multiple and repeated awaits are supported.
- FontWithContext's context controls only that caller's wait. Cancellation returns
  NullFont and the wait context's error, without consuming the eventual result or
  canceling the search or other waiters.
- Resolve's context controls the search and is passed to resolvers. Cancellation
  remains cooperative; a legacy adapter cannot interrupt its underlying I/O.
- Tie policy: an already completed result wins over wait cancellation.
  When cancellation wakes a pending waiter, recheck completion before returning
  the cancellation error. Completion after that check remains a concurrent event.
- Treat a nil wait context as context.Background, matching Resolve's handling.

## Implementation plan (completed)

1. Replace the consumable result channel with shared promise state containing
   a fontPlusErr and a completion channel of type chan struct{}.
2. Have the sole search goroutine assign the result once and close the completion
   channel. Store the state by pointer so all promise copies refer to it.
3. Await completion by receiving from the completion channel, then read the
   retained result. Closing the channel publishes the prior write to all waiters;
   no mutex is necessary for immutable completed state.
4. Implement the wait cancellation contract with a completion check before the
   select and another check in its cancellation branch. Normalize nil wait contexts.
5. Document repeatable awaits, separate search/wait contexts, tie behavior, and
   the limits of legacy provider cancellation. Keep public signatures unchanged.
6. Remove the known-bug build tag when the reproductions pass. Add focused tests
   for completed-result precedence, nil wait context, cancellation of one of
   several waiters, and a deadline on a pending wait using synctest virtual time.
7. Run the lifecycle tests with the race detector, then the offline full suite.
   Disable the Google integration test via an empty GOOGLE_FONTS_API_KEY for that
   offline run.

Do not expand this fix into provider I/O cancellation or the match/cache issues.
In particular, an uncooperative resolver can still keep its search goroutine
alive; the proposed change removes the separate leak during result publication.
