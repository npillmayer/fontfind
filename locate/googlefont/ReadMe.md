# googlefont

## Purpose

`googlefont` resolves fonts via the Google Fonts directory API and caches downloaded files locally.

The package supports I/O abstraction so tests can run offline with fixture JSON and fake network responses.

## API

- `type IO` (env/http/fs abstraction)
- `Find(conf) locate.FontLocator`
- `FindWithIO(conf, hostio) locate.FontLocator`
- `FindGoogleFont(conf, pattern, style, weight) (fontfind.ScalableFont, error)`
- `ListGoogleFonts(conf, pattern)`
- `SimpleConfig(appkey) schuko.Configuration`

Configuration note:

- requires `google-fonts-api-key` in config or `GOOGLE_FONTS_API_KEY` in env for live API usage.

## Example Applications

### 1. Resolve and cache a Google font

```go
conf := googlefont.SimpleConfig("myapp")
resolver := googlefont.Find(conf)
sf, err := locate.ResolveFontLoc(desc, resolver).Font()
```

### 2. Test with offline fixture-driven I/O

```go
fake := newFakeIO(t) // implements googlefont.IO
resolver := googlefont.FindWithIO(conf, fake)
sf, err := resolver(desc)
```
