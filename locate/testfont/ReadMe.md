# testfont

`testfont` resolves fonts from the `testdata` directory, provided as an `embed.FS`.

Clients will want to create a `FontLocator` using the `Find` function:

```go
Find(embed.FS) locate.FontLocator
```

## Test Example

```go
//go:embed testdata/*
var testdata embed.FS

func TestAnythingWithFont(t *testing.T) {
	fontResolver := testfont.Find(testdata)
	font, err := locate.ResolveFontLoc(desc, fontResolver).Font()
	…
}
```
