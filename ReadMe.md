# fontfind

`fontfind` is a Go package for discovering and loading scalable fonts from multiple sources.

The package is built around one practical goal: given a font request (`family pattern`, `style`, `weight`), return a usable font resource that can be loaded as bytes. Resolution is provider-based and can combine:

- embedded fallback fonts
- local/system fonts
- Google Fonts (with local caching)

The resolver pipeline is cache-backed (`fontregistry`) and designed so callers can still receive a fallback font object even when the requested font cannot be found.

## Installation

```bash
go get github.com/npillmayer/fontfind
```

## API Overview

### Core types (`package fontfind`)

- `Descriptor`: describes a requested font (`Pattern`, `Style`, `Weight`)
- `ScalableFont`: describes a resolved font variant and where to load it from
- `NullFont`: zero-value marker used for unresolved results
- `FallbackFont()`: returns packaged default fallback (`Go-Regular.otf`)

`ScalableFont` methods:

- `SetFS(fs fs.FS, path string)`
- `Path() string`
- `ReadFontData() ([]byte, error)`

### Resolution API (`package locate`)

- `ResolveFontLoc(desc, resolvers...) FontPromise`
- `ResolveFontLocWithContext(ctx, desc, resolvers...) FontPromise`

`FontPromise`:

- `Font() (fontfind.ScalableFont, error)`
- `FontWithContext(ctx) (fontfind.ScalableFont, error)`

Resolution behavior:

1. Normalize descriptor to registry key.
2. Try registry cache (`fontregistry.GlobalRegistry().GetFont`).
3. Run resolvers in provided order on cache miss.
4. Cache successful hits.
5. Return registry fallback font with an error if all resolvers fail.

### Resolver providers

- `locate/fallbackfont`: embedded packaged fonts (`Find`, `Default`)
- `locate/systemfont`: local/system lookup (`Find`, `FindLocalFont`)
- `locate/googlefont`: Google Fonts lookup + cache (`Find`, `FindWithIO`, `FindGoogleFont`)

## Example Applications

### 1. General app font resolution (system -> Google -> embedded fallback)

```go
package main

import (
	"fmt"
	"log"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/fontfind/locate"
	"github.com/npillmayer/fontfind/locate/fallbackfont"
	"github.com/npillmayer/fontfind/locate/googlefont"
	"github.com/npillmayer/fontfind/locate/systemfont"
	"golang.org/x/image/font"
)

func main() {
	desc := fontfind.Descriptor{
		Pattern: "Noto Sans",
		Style:   font.StyleNormal,
		Weight:  font.WeightNormal,
	}

	conf := googlefont.SimpleConfig("myapp") // uses GOOGLE_FONTS_API_KEY if needed
	system := systemfont.Find("myapp", nil)
	google := googlefont.Find(conf)
	fallback := fallbackfont.Find()

	promise := locate.ResolveFontLoc(desc, system, google, fallback)
	sf, err := promise.Font()
	if err != nil {
		// err may be non-nil while sf is still a usable fallback font
		log.Printf("font lookup degraded: %v", err)
	}

	data, err := sf.ReadFontData()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("resolved %s (%s), %d bytes\n", sf.Name, sf.Path(), len(data))
}
```

### 2. Timeout-aware async resolution

Use a context-aware resolver and wait with cancellation/deadline control.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/fontfind/locate"
	"golang.org/x/image/font"
)

func main() {
	desc := fontfind.Descriptor{
		Pattern: "Any",
		Style:   font.StyleNormal,
		Weight:  font.WeightNormal,
	}

	resolver := func(ctx context.Context, d fontfind.Descriptor) (fontfind.ScalableFont, error) {
		select {
		case <-ctx.Done():
			return fontfind.NullFont, ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return fontfind.FallbackFont(), nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	promise := locate.ResolveFontLocWithContext(ctx, desc, resolver)
	_, err := promise.FontWithContext(ctx)
	fmt.Println("result:", err) // context deadline exceeded
}
```

### 3. Embedded-only/offline deployments

For fully offline systems, use only the fallback resolver:

```go
desc := fontfind.Descriptor{Pattern: "Go", Style: font.StyleNormal, Weight: font.WeightNormal}
promise := locate.ResolveFontLoc(desc, fallbackfont.Find())
sf, err := promise.Font()
```

This keeps runtime dependencies minimal and still guarantees a packaged font.

## Notes

- Google Fonts access requires a valid API key (`GOOGLE_FONTS_API_KEY`) for live directory fetches.
- TTC (`*.ttc`) handling is intentionally limited at the moment.
- The registry caches results globally within the process.

## License

BSD 3-Clause. See `/Users/npi/prg/go/fontfind/LICENSE`.
