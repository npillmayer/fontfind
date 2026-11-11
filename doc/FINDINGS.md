# Findings and Follow-ups

Updated November 11, 2026. Confirm source-inspection findings with reproductions
before claiming fixes. Keep general fixes in fontfind, not application or TUI
workarounds.

## 1. Promise Lifecycle: Fixed

The original unbuffered result channel blocked publication without a waiter.
Repeated or concurrent awaits could return an empty font and nil error.

Promises now retain their result and close a completion channel to notify all
waiters. Completion requires no receiver. Repeated and concurrent awaits return
the same font/error pair. Canceling a wait affects only that waiter; search
cancellation is separate and cooperative. An observed completed result takes
precedence over wait cancellation, and nil wait contexts are supported.

Regressions cover abandoned promises, canceled waiters, repeated and concurrent
awaits, deadlines, and search cancellation. See
[promise_lifecycle_test.go](../locate/promise_lifecycle_test.go) and
[PROMISE-LIFECYCLE.md](PROMISE-LIFECYCLE.md).

Legacy adapters still do not make underlying I/O cancellable; see issue 4.

## 2. Truthful Match and Fallback Results: Partly Fixed

**Fixed:** the embedded provider previously returned its last candidate even if
nothing matched. Missing names and mismatched variants now return NullFont and
an error. The pipeline can try later providers or return its registry fallback
with a lookup error, without caching that fallback as a successful request match.
Explicit default access through `fallbackfont.Default()` and
`fontfind.FallbackFont()` remains available.

Tests cover matching variants, misses, pipeline continuation, cache behavior,
and readable defaults in [fallback_test.go](../locate/fallbackfont/fallback_test.go).

**Deferred by agreement:** richer resolution results. System directory searches
still ignore requested style/weight, returned metadata can repeat the request,
and the pipeline discards individual provider errors. A future compatible API
should expose match quality, actual resource metadata, provider provenance, and
failure reasons. Do not infer an exact match from nil error or from
`ScalableFont.Style`/`Weight` alone.

## 3. Cache Keys and Policy Isolation: Addressed

Old normalization merged styles and weights, stripped dotted suffixes, folded
case, trimmed whitespace, and replaced spaces with underscores. It could also
collide with the reserved `"fallback"` entry.

Keys now preserve the exact pattern and numeric style/weight in an unambiguous
representation separate from the reserved fallback key. Treat keys as opaque;
regenerate persisted keys produced by the old format. The public
`NormalizeFontname` signature is unchanged.

Regressions verify 600 distinct descriptors and subsequent cache hits, plus
separate registries with different provider orders and fallbacks. See
[cache_identity_test.go](../locate/cache_identity_test.go).

**Policy contract:** cache hits bypass providers. Create a fresh registry when
provider order, configuration, or fallback policy changes. Keys do not encode
policy identity. Global convenience functions share a registry; clients with
incompatible policies should use separate pipelines and registries.

## 4. Resource and Provider I/O: Partly Fixed

### Completed Concurrency Fixes

- Registry logging now snapshots under lock, releases the lock before logger
  callbacks, prints actual paths, and respects configured trace levels. A race
  test reproduced the previous concurrent map access.
- System-font locators own their font-list initialization state. Concurrent tests
  reproduced the old first-client-wins contamination.
- Direct Google lookups initialize independently and no longer inherit another
  client's directory initialization failure. A fixture test reproduced a missing
  API key poisoning later direct lookups.
- Reusable `Find` locators retain their own initialization caches. Direct lookup
  calls initialize on each call. Google listing no longer changes trace levels.

Tests: [log_test.go](../fontregistry/log_test.go),
[systemfont/state_test.go](../locate/systemfont/state_test.go), and
[googlefont/state_test.go](../locate/googlefont/state_test.go).

Create new locators to reload configuration or retry cached initialization
failures. Keep configuration immutable during use. Shared custom I/O and loggers
must support concurrent use. Separate registries and locators do not isolate
shared disk-cache paths or all host filesystem operations.

### Remaining Work

Estimates include implementation and regression tests in focused development
days. These are planning ranges, not commitments. Cancellation scope and
cross-process coordination are the main uncertainties.

| Follow-up | Risk | Effort | Proposed scope |
| --- | --- | --- | --- |
| Bounded reads | Low | 0.5-1 day | Add a caller-specified limit and clear size-limit error; retain ReadFontData compatibility and bound allocation during reading. |
| Cancellable provider I/O | Medium | 1.5-3 days | Add context-aware providers and HTTP requests, cancellation checks during traversal/copying, and compatibility adapters. |
| Shared download-cache coordination | Medium | 1-2 days | Atomically publish complete downloads and coordinate same-destination downloads within the process. |

Allow roughly **4-7 focused development days** for the combined work and
integration. Strict cross-process coordination adds approximately **1-2 days**,
with portability and stale-lock concerns.

Bounded-read tests should cover exact limits, oversized inputs, unknown or
misleading sizes, read errors, and resource closure. Checking length after
`ReadFontData` does not bound allocation. Applying limits to Google directory
JSON and downloads is additional scope requiring separate defaults.

Cancellation remains cooperative at supported boundaries. HTTP requests can be
canceled, but filesystem calls and the external system-font search library may
not be interruptible while running. Adapting legacy providers does not make their
I/O cancellable.

Download correctness and deduplication are separate goals. Temporary files and
atomic publication protect readers from partial destinations; per-destination
coordination avoids duplicate downloads. The injectable I/O API needs additional
operations and failure/cleanup tests. Neither atomic publication nor download
coordination has been implemented yet.

Recommended sequence:

1. Bounded reads.
2. Atomic cache publication to protect correctness on failure or cancellation.
3. Context-aware provider I/O.
4. Download deduplication if needed; strict cross-process locking is optional scope.

## Validation

The full offline suite passed with the race detector after the concurrency fixes:

```sh
GOOGLE_FONTS_API_KEY= go test -race ./... -count=1 -timeout=60s
```

An empty API key skips the live Google integration test. This covers exercised
paths, not cancellation or download-coordination guarantees for remaining work.
