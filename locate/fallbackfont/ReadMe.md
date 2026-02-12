# fallbackfont

## Purpose

`fallbackfont` resolves fonts from embedded package assets.

It is the deterministic last-resort provider and contains the default packaged fallback (`Go-Regular.otf`).

## API

- `Find() locate.FontLocator`
- `Default() (fontfind.ScalableFont, error)`
- `FindFallbackFont(pattern, style, weight) (fontfind.ScalableFont, error)`

## Example Applications

### 1. Use as final resolver in a chain

```go
fallback := fallbackfont.Find()
sf, err := locate.ResolveFontLoc(desc, system, google, fallback).Font()
```

### 2. Access default fallback directly

```go
sf, err := fallbackfont.Default()
data, err := sf.ReadFontData()
```
