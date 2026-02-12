# AGENTS.md

Guidance for human and AI contributors working in this repository.

Last updated: February 12, 2026.

## Project Scope

`fontfind` is a Go library for locating and selecting scalable fonts by descriptor:

- family/pattern
- style
- weight

It currently resolves fonts through provider-style locators:

- embedded fallback fonts
- system fonts
- Google Fonts

Primary goal: make resolution reliable and deterministic across environments before adding major new features.

## Repository Map

- `font.go`: core types and package-level utilities
- `match.go`: filename/variant matching heuristics and confidence model
- `locate/resolve.go`: resolver orchestration (async promise pattern)
- `locate/fallbackfont`: embedded packaged font resolver
- `locate/systemfont`: local/system resolver and fontconfig-list parsing
- `locate/googlefont`: Google Fonts directory lookup and cache/download logic
- `fontregistry`: global registry/cache abstraction
- `doc/PURPOSE.md`: current project brief and backlog context

## Current Reality (Important)

Treat this repo as early-stage and under refactor.

Known issues at scan time:

- `go test ./...` is not fully green
- `locate/res_test.go` has a compile-blocking unused variable
- Google tests require live network and `GOOGLE_FONTS_API_KEY`
- `fontfind.FallbackFont()` still panics
- TTC (`*.ttc`) support is incomplete

Do not assume polished behavior from comments or historical API names.

## Engineering Priorities

When proposing or implementing changes, prioritize in this order:

1. Test/build stability in local CI-like runs
2. Removal of panic paths in public flow
3. Clear, consistent configuration contract
4. Deterministic matching quality improvements
5. Feature expansion (for example deeper TTC handling)

## Expected Workflow

For non-trivial changes:

1. Read `doc/PURPOSE.md` first
2. Identify whether change is bugfix, cleanup, or feature
3. Add or adjust tests first when practical
4. Make smallest safe code change
5. Run relevant tests
6. Update docs if behavior/config changed

## Testing Guidance

Baseline command:

```bash
go test ./...
```

Because some tests depend on network/API keys:

- prefer unit tests that run offline by default
- gate integration tests behind explicit conditions (env vars or tags)
- avoid introducing new tests that require network access by default

If a change touches matching/resolution logic, add coverage in the closest package tests.

## Configuration Rules

Google Fonts key naming is currently inconsistent across messages/comments.
When editing related code or docs:

- standardize on one environment variable name
- keep error messages, comments, and tests aligned
- document the final contract clearly

## API and Compatibility Notes

- Public types/functions are in active flux; still, avoid unnecessary breaking changes.
- If behavior changes, prefer additive changes or clear migration notes.
- Keep resolver order explicit in code and docs.

## Code Style

- Follow idiomatic Go (`gofmt`, small focused functions, clear error wrapping).
- Keep dependency additions minimal and justified.
- Prefer explicit errors over silent fallback unless fallback is part of designed behavior.
- Avoid hidden global side effects beyond existing singleton patterns.

## Definition of Done (Per Change)

A change is done when all are true:

- code compiles
- relevant tests pass
- no new panic path in normal execution flow
- docs/comments updated for behavior or config changes
- change scope is minimal and reviewable

## Starter Backlog

If you need a next task, pick one:

1. Fix compile failure in `locate/res_test.go`
2. Implement non-panicking fallback in `font.go` (`FallbackFont`)
3. Unify Google API key contract in code/tests/docs
4. Improve variant selection in Google resolver (replace first-variant shortcut)
5. Add a top-level `README.md` with usage and resolver/config examples
