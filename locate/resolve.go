package locate

import (
	"context"
	"fmt"

	"github.com/npillmayer/fontfind"
)

// notFound returns an application error for a missing resource.
func notFound(res string) error {
	return fmt.Errorf("font not found: %v", res)
}

type fontPlusErr struct {
	font fontfind.ScalableFont
	err  error
}

// TypefacePromise runs font searching asynchronously in the background.
// A call to `Typeface()` blocks until font loading is completed.
type TypefacePromise interface {
	Typeface() (fontfind.ScalableFont, error)
}

type fontLoader struct {
	await func(ctx context.Context) (fontfind.ScalableFont, error)
}

func (loader fontLoader) Typeface() (fontfind.ScalableFont, error) {
	return loader.await(context.Background())
}

// ResolveTypeface resolves a typefacee with given properties.
// It searches for fonts in the following order:
//
// ▪︎ Fonts packaged with the application binary
//
// ▪︎ System-fonts
//
// ▪︎ Google Fonts service (https://fonts.google.com/)
//
// ResolveTypeface will try to match style and weight requirements closely, but
// will load a font variant anyway if it matches approximately. If, for example,
// a system contains a font with weight 300, which would be considered a "light"
// variant, but no variant with weight 400 (normal), it will load the 300-variant.
//
// When looking for sytem-fonts, ResolveTypeface will use an existing fontconfig
// (https://www.freedesktop.org/wiki/Software/fontconfig/)
// installation, if present. fontconfig has to be configured in the global
// application setup by pointing to the absolute path of the `fc-list` binary.
// If fontconfig isn't installed or configured, then this step will silently be
// skipped and a file system scan of the sytem's fonts-folders will be done.
// (See also function `FindLocalFont`).
//
// A prerequisite to looking for Google fonts is a valid API-key (refer to
// https://developers.google.com/fonts/docs/developer_api). It has to be configured
// either in the application setup or as an environment variable GOOGLE_API_KEY.
// (See also function `FindGoogleFont`).
//
// If no suitable font can be found, an application-wide fallback font will be
// returned.
//
// Typefaces are not returned synchronously, but rather as a promise
// of kind TypefacePromise (async/await).
func ResolveTypeface(desc fontfind.Descriptor, resolvers ...FontLocator) TypefacePromise {
	// TODO include a context parameter
	ch := make(chan fontPlusErr)
	go func(ch chan<- fontPlusErr) {
		result := searchScalableFont(desc, resolvers)
		ch <- result
		close(ch)
	}(ch)
	loader := fontLoader{}
	loader.await = func(ctx context.Context) (fontfind.ScalableFont, error) {
		select {
		case <-ctx.Done():
			return fontfind.NullFont, ctx.Err()
		case r := <-ch:
			return r.font, r.err
		}
	}
	return loader
}

func searchScalableFont(desc fontfind.Descriptor, resolvers []FontLocator) (result fontPlusErr) {
	// name := fontregistry.NormalizeFontname(desc.Pattern, desc.Style, desc.Weight)
	// if t, err := fontregistry.GlobalRegistry().Typeface(name); err == nil {
	// 	result.font = t
	// 	return
	// }
	for _, resolver := range resolvers {
		if f, err := resolver(desc); err == nil {
			result.font = f
			result.err = err
			return
		}
	}
	// if f != nil { // if found, enter into font registry
	// 	f.Fontname = name
	// 	fontregistry.GlobalRegistry().StoreFont(name, f)
	// 	result.font, result.err = fontregistry.GlobalRegistry().Typeface(name, size)
	// 	result.desc.Family = name
	// 	//fontfind.GlobalRegistry().DebugList()
	// } else { // use fallback font
	// 	result.font, _ = fontregistry.GlobalRegistry().Typeface("fallback", size)
	// 	result.desc.Family = "fallback"
	// }
	return result
}
