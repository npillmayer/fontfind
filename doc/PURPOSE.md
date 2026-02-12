# fontfind: Purpose, Status, and Direction

As of February 12, 2026.

## Purpose

`fontfind` is a Go library for discovering, matching, and loading scalable font files by descriptor:

- `pattern` (font family/name pattern)
- `style` (normal/italic/oblique)
- `weight` (light/normal/bold/etc.)

Core model types are in `font.go`:

- `Descriptor` describes what to search for.
- `ScalableFont` describes a found font with `fs.FS` + `Path` for reading bytes.

The package also includes:

- matching heuristics from filenames and variant names (`match.go`)
- a font registry/singleton cache abstraction (`fontregistry/registry.go`)
- resolver pipeline orchestration with async promise-like API (`locate/resolve.go`)

## Resolver Architecture

Typeface resolution is provider-based. Current resolver packages:

- `locate/fallbackfont`: embedded packaged fonts (`go:embed`)
- `locate/systemfont`: local/system font lookup
- `locate/googlefont`: Google Fonts API lookup + local caching

`ResolveTypeface` accepts one or more `FontLocator` functions and returns the first successful match.

## Current Status

This repository is early-stage and actively being refactored.

Evidence:

- very small history (4 commits total)
- commits are recent (February 10-11, 2026)
- no tags/releases yet
- no top-level `README.md`

Recent commit messages:

- `Initial commit`
- `First refactoring`
- `Renamed package`

## Build/Test Health Snapshot

From `go test ./...` on February 12, 2026 (after test refactoring):

- root package: no tests
- `fontregistry`: tests pass
- `locate`: tests pass
- `locate/googlefont`: tests pass
- `locate/systemfont`: no tests
- `locate/fallbackfont`: no tests

Test behavior notes:

- Google API integration tests are skipped if `GOOGLE_FONTS_API_KEY` is not set.
- Cache download test is now an offline unit test (mocked HTTP transport, no real network).

## Recent Progress

Completed in current refactor cycle:

- Added and wired `systemfont.IO` abstraction for fontconfig list lookup path.
- Reworked `TestFCFind` to use fixture-driven resolver flow via `systemfont.Find(..., testIO)`.
- Fixed system-font path handling bug to use descriptor path for loaded font.
- Converted Google API-key tests from fail-fast to skip-if-missing-key behavior.
- Refactored cache download test to deterministic offline unit-test style.
- Standardized API-key contract on `GOOGLE_FONTS_API_KEY` in code/comments/messages.

## Known Gaps and Risks

The codebase explicitly marks several unfinished areas:

- TTC support not implemented (`*.ttc` skipped in comments and parser behavior)
- `fontfind.FallbackFont()` currently `panic("fallback fonts not yet implemented")`
- `locate` package comment calls current implementation a stand-in/quick hack
- Google variant selection currently picks first variant (`fi.Variants[0]`) instead of best match
- async resolve API has TODO for context/cancellation integration
- `systemfont` fontconfig loader uses package-global cache state (`sync.Once` + globals), which may complicate test isolation for future cases

Quality and consistency issues observed:

- typo/terminology/documentation drift in comments and package text (expected in early refactor phase)

## Implementation Notes (What Already Works)

- Embedded fallback fonts are packaged and can resolve basic matches.
- System font path can use a pre-generated fontconfig list, then fallback to `go-findfont`.
- Google Fonts lookup and caching pipeline is implemented (with API-key requirement for live integration tests).
- Matching confidence model (`No/Low/High/Perfect`) exists and is used by local/google matching paths.

## Suggested Direction (Starting Backlog)

1. Remove runtime panic path:
   - implement `fontfind.FallbackFont()` via `locate/fallbackfont`, or remove/deprecate API until ready
2. Improve match quality:
   - replace `fi.Variants[0]` with confidence-based variant selection
3. Registry integration pass:
   - reconnect `ResolveTypeface` search flow with `fontregistry` cache (currently mostly commented out)
4. Documentation baseline:
   - add `README.md` with usage examples, resolver order, config keys, and test modes
5. Complete TTC story: (NOT A NEAR-TERM GOAL)
   - decide parse/load strategy for `.ttc` and implement at least first readable variant support
6. Keep integration boundaries explicit:
   - maintain skip/gate behavior for tests requiring live network/API keys

## Practical Project Direction

The current direction appears to be:

- modular provider-based font resolution
- pragmatic fallback-first behavior
- gradual migration/refactor from earlier resource/location code layout

Near-term value is highest in reliability and contract clarity (tests, fallback behavior, config), then feature depth (TTC and richer matching).
