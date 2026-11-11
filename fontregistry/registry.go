package fontregistry

import (
	"fmt"
	"sync"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/schuko/tracing"
	xfont "golang.org/x/image/font"
)

// Registry caches resolved scalable fonts by normalized name.
type Registry struct {
	sync.Mutex
	fonts map[string]fontfind.ScalableFont
}

var globalFontRegistry *Registry

var globalRegistryCreation sync.Once

// GlobalRegistry returns the application-wide font registry singleton.
func GlobalRegistry() *Registry {
	globalRegistryCreation.Do(func() {
		globalFontRegistry = New()
	})
	return globalFontRegistry
}

// New creates an empty font registry.
func New() *Registry {
	fr := &Registry{
		fonts: make(map[string]fontfind.ScalableFont),
	}
	return fr
}

const fallbackFontKey = "fallback"

// StoreFont pushes a font into the registry if it isn't contained yet.
//
// The font will be stored using the normalized font name as a key. If this
// key is already associated with a font, that font will not be overridden.
func (fr *Registry) StoreFont(normalizedName string, f fontfind.ScalableFont) {
	if f.Name == "" {
		tracer().Errorf("registry cannot store null font")
		return
	}
	fr.Lock()
	defer fr.Unlock()
	//style, weight := GuessStyleAndWeight(f.Fontname)
	//fname := NormalizeFontname(f.Fontname, style, weight)
	if _, ok := fr.fonts[normalizedName]; !ok {
		tracer().Debugf("registry stores font %s as %s", f.Name, normalizedName)
		fr.fonts[normalizedName] = f
	}
}

// GetFont returns a cached font by normalized name.
//
// On a cache miss, GetFont returns the registry fallback font together with
// a non-nil error describing the miss.
func (fr *Registry) GetFont(normalizedName string) (fontfind.ScalableFont, error) {
	//
	tracer().Debugf("registry searches for font %s", normalizedName)
	fr.Lock()
	if t, ok := fr.fonts[normalizedName]; ok {
		fr.Unlock()
		tracer().Infof("registry found font %s", normalizedName)
		return t, nil
	}
	fr.Unlock()
	tracer().Infof("registry does not contain font %s", normalizedName)
	missErr := fmt.Errorf("font %s not found in registry", normalizedName)
	f, fallbackErr := fr.FallbackFont()
	if fallbackErr != nil {
		return fontfind.NullFont, fmt.Errorf("%w; fallback failed: %v", missErr, fallbackErr)
	}
	return f, missErr
}

// FallbackFont returns the default fallback font from registry cache.
// If absent, it will load and cache the packaged fallback under key "fallback".
func (fr *Registry) FallbackFont() (fontfind.ScalableFont, error) {
	fr.Lock()
	if t, ok := fr.fonts[fallbackFontKey]; ok {
		fr.Unlock()
		return t, nil
	}
	fr.Unlock()

	f := fontfind.FallbackFont()
	fr.Lock()
	defer fr.Unlock()
	// Another goroutine may have inserted fallback while we were loading.
	if t, ok := fr.fonts[fallbackFontKey]; ok {
		return t, nil
	}
	tracer().Infof("font registry caches fallback font %s", fallbackFontKey)
	fr.fonts[fallbackFontKey] = f
	return f, nil
}

// LogFontList logs a snapshot of the registry at Info level, respecting the
// supplied tracer's configured level. The tracer must support concurrent use
// if callers share it between goroutines.
func (fr *Registry) LogFontList(tracer tracing.Trace) {
	fr.Lock()
	fonts := make(map[string]fontfind.ScalableFont, len(fr.fonts))
	for k, v := range fr.fonts {
		fonts[k] = v
	}
	fr.Unlock()
	tracer.Infof("--- registered fonts ---")
	for k, v := range fonts {
		tracer.Infof("typeface [%s] = %s @ %s", k, v.Name, v.Path())
	}
	tracer.Infof("------------------------")
}

// NormalizeFontname returns an opaque cache key for an exact request descriptor.
// It preserves the pattern verbatim and the numeric style and weight. Providers
// may interpret patterns differently, so case, whitespace, punctuation, and font
// extensions must not be normalized here. The key cannot equal the reserved
// "fallback" key. Callers should generate keys with this function rather than
// depend on their representation; keys from the former lossy format must be rebuilt.
func NormalizeFontname(fname string, style xfont.Style, weight xfont.Weight) string {
	return fmt.Sprintf("font-request:v1:%q:%d:%d", fname, style, weight)
}
