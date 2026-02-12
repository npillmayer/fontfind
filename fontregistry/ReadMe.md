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

- `GetFont` returns a non-nil error on cache miss, but still returns fallback when available.
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
