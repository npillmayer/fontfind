# fontregistry

## Purpose

`fontregistry` provides the in-process cache for resolved fonts.

It stores `fontfind.ScalableFont` values under normalized keys and also manages a cached fallback font entry.

## API

- `type Registry`
- `NewRegistry() *Registry`
- `GlobalRegistry() *Registry`
- `(*Registry).StoreFont(normalizedName, font)`
- `(*Registry).GetFont(normalizedName) (font, error)`
- `(*Registry).FallbackFont() (font, error)`
- `NormalizeFontname(name, style, weight) string`

Behavior note:

`LogFontList` logs a snapshot taken under the registry lock, releasing the lock
before invoking the logger. It prints actual font paths and respects the logger's
configured level. A logger shared between goroutines must support concurrent use.

- `GetFont` returns a non-nil error on cache miss, but still returns fallback when available.

Cache keys preserve the exact request pattern, style, and weight. Case,
whitespace, dots, and extensions remain significant because custom providers
may distinguish them. Treat `NormalizeFontname` results as opaque: its former
lossy keys must be regenerated, including keys in any client-persisted cache.
The reserved `"fallback"` entry is separate from a request for a family with
that name.

A registry belongs to a resolution policy. Create a fresh registry when provider
order, provider configuration, or fallback policy changes. Sharing one registry
across different policies allows an earlier cache hit to bypass the new policy.
- Clients may create their own registry instances for isolated caching. Additionally, a global registry is provided for convenience.

## Example Applications

### 1. Cache a resolved font

```go
key := fontregistry.NormalizeFontname("Noto Sans", font.StyleNormal, font.WeightNormal)
fontregistry.GlobalRegistry().StoreFont(key, sf)
```

### 2. Read from cache with fallback semantics

```go
key := fontregistry.NormalizeFontname("NoSuch", font.StyleNormal, font.WeightNormal)
sf, err := fontregistry.GlobalRegistry().GetFont(key)
// err != nil means cache miss; sf may still be fallback.
```
