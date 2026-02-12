# fontfind: Purpose, Status, and Direction

As of February 12, 2026.

## Purpose

`fontfind` is a Go library for discovering and loading scalable fonts from multiple sources based on a descriptor:

- `Pattern` (family/name pattern)
- `Style`
- `Weight`

Core model:

- `fontfind.Descriptor` identifies requested font properties.
- `fontfind.ScalableFont` represents a located font resource.

The `ScalableFont` model is now resource-centric:

- metadata fields: `Name`, `Style`, `Weight`
- storage is encapsulated (`fileSystem`, `path` are internal)
- public access is via methods:
  - `SetFS(fs fs.FS, path string)`
  - `Path() string`
  - `ReadFontData() ([]byte, error)`

## Current API Shape

Resolver orchestration is in package `locate`:

- `ResolveFontLoc(desc, resolvers...)` returns async `FontPromise`
- `ResolveFontLocWithContext(ctx, desc, resolvers...)` adds cancellation/deadline support
- `FontPromise` methods:
  - `Font()`
  - `FontWithContext(ctx)`

Locator function types:

- `FontLocator` (legacy/simple)
- `FontLocatorWithContext` (context-aware variant)

Registry/cache is in package `fontregistry`:

- `StoreFont(...)`
- `GetFont(...)`
- `FallbackFont()`
- `NormalizeFontname(...)`

## Resolver Architecture

Typeface resolution is provider-based:

- `locate/fallbackfont`: embedded packaged fonts (includes default Go fallback font)
- `locate/systemfont`: local/system font lookup (fontconfig list + filesystem fallback)
- `locate/googlefont`: Google Fonts directory lookup + local cache/download

`ResolveFontLoc*` now uses registry cache lookup/store again:

1. normalize descriptor -> cache key
2. try `fontregistry.GetFont`
3. on miss, query resolvers in order
4. store successful result in registry
5. if unresolved, return fallback font plus lookup error

## Fallback Behavior

Fallback behavior is explicit and deterministic:

- `fontfind.FallbackFont()` returns embedded `Go-Regular.otf`
- `fontregistry.FallbackFont()` caches fallback under key `"fallback"`
- `fallbackfont.Default()` returns packaged default fallback as a regular resolver utility

Goal: callers can always receive a usable font object even when lookup fails.

## Build and Test Status

Current local snapshot:

- `go test ./...` passes
- `fontregistry`, `locate`, `locate/googlefont` have active tests
- `locate/fallbackfont` and `locate/systemfont` currently have no direct package test files

Testing approach highlights:

- Google-font tests are offline by default
- network interactions are simulated via injected I/O
- Google directory data is fixture-driven (`locate/googlefont/testdata/webfonts.json`)
- cache download behavior is unit-tested including non-200 error handling

## Recent Refactor Outcomes

- `ScalableFont` storage access moved behind methods (encapsulation of FS/path)
- resolver API renamed (`ResolveTypeface*` -> `ResolveFontLoc*`)
- context-aware resolve variant added
- registry method names unified around `Font` terminology (`StoreFont`, `GetFont`, `FallbackFont`)
- Google variant selection changed from first entry to confidence-based choice
- resolver cache flow reconnected to registry

## Known Gaps and Risks

- TTC (`*.ttc`) handling is still intentionally incomplete.
- `locate` package header still describes implementation as transitional/quick-hack.
- both `systemfont` and `googlefont` contain singleton/global cache state (`sync.Once`), which may affect strict test isolation in expanded suites.
- current context cancellation in resolve flow is strong at orchestration level; individual legacy `FontLocator` implementations still rely on adapter behavior and may not always interrupt mid-operation immediately.

## Suggested Direction (Next Work)

1. Documentation baseline:
   - add top-level `README.md` with current API naming and examples (`ResolveFontLoc*`, `FontPromise`, `ScalableFont` methods).
2. Tighten context semantics:
   - progressively migrate resolver implementations to native `FontLocatorWithContext`.
3. Expand deterministic tests:
   - add focused tests around singleton state/reset expectations.
4. TTC roadmap (not near-term):
   - define parsing/loading approach and compatibility constraints.
5. Terminology cleanup:
   - fix stale/legacy comments and naming drift in package docs.

## Practical Direction

The repository direction is now clearer than before:

- provider-based resolution
- deterministic fallback guarantees
- registry-backed caching
- testability through I/O abstraction and fixtures

Near-term value is highest in docs and API clarity, then deeper capability work (true resolver-level cancellation, TTC support).
