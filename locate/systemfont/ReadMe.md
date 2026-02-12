# systemfont

## Purpose

`systemfont` resolves fonts from local machine sources.

It prefers a fontconfig list (`fontlist.txt` under the app config area) and falls back to platform directory scanning.

## API

- `type IO` (injectable host I/O for tests)
- `Find(appkey, io) locate.FontLocator`
- `FindLocalFont(appkey, io, pattern, style, weight) (fontfind.ScalableFont, error)`

`appkey` determines where fontconfig list data is looked up.

## Example Applications

### 1. Standard local lookup

```go
system := systemfont.Find("myapp", nil)
sf, err := system(fontfind.Descriptor{
	Pattern: "Noto Sans",
	Style:   font.StyleNormal,
	Weight:  font.WeightNormal,
})
```

### 2. Fixture-based lookup in tests

```go
mockIO := newTestIO() // implements systemfont.IO
system := systemfont.Find("tyse-test", mockIO)
sf, err := locate.ResolveFontLoc(desc, system).Font()
```
