# fallbackfont

## Purpose

`fallbackfont` resolves fonts from embedded package assets. As we are in a Go eco-sytem,
these are the
[Go fonts](https://go.dev/blog/go-fonts),
packaged and embedded (in OTF format).

It also contains the default packaged fallback (`Go-Regular.otf`).
`Find` and `FindFallbackFont` search for a matching filename, style, and weight;
they return an error when no match exists. This lets the pipeline continue to
later providers. If all providers miss, the pipeline returns its registry's
fallback together with a lookup error, without caching it as a successful match.

Use `Default()` or `fontfind.FallbackFont()` to request the packaged default
directly, without a name search.

## API

- `Find() locate.FontLocator`
- `Default() (fontfind.ScalableFont, error)`
- `FindFallbackFont(pattern, style, weight) (fontfind.ScalableFont, error)`

## Example Applications

### 1. Use as final resolver in a chain

```go
fallbackSearcher := fallbackfont.Find()
sf, err := locate.ResolveFontLoc(desc, system, google, fallbackSearcher).Font()
```

### 2. Access default fallback directly

```go
sf, err := fallbackfont.Default()
data, err := sf.ReadFontData()
```
