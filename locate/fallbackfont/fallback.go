package fallbackfont

import (
	"embed"
	"errors"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/fontfind/locate"
	"github.com/npillmayer/schuko/tracing"
	"golang.org/x/image/font"
)

// tracer writes to trace with key 'tyse.font'
func tracer() tracing.Trace {
	return tracing.Select("tyse.font")
}

//go:embed packaged/*
var packaged embed.FS

const defaultFallbackFilename = "Go-Regular.otf"

// Find creates a locator that resolves fonts from the embedded fallback set.
// A missing match returns an error so subsequent locators can be tried.
func Find() locate.FontLocator {
	return func(descr fontfind.Descriptor) (fontfind.ScalableFont, error) {
		pattern := descr.Pattern
		style := descr.Style
		weight := descr.Weight
		return FindFallbackFont(pattern, style, weight)
	}
}

// Default returns the default packaged fallback font.
func Default() (fontfind.ScalableFont, error) {
	// Ensure packaged default exists in embedded resources.
	if _, err := packaged.Open("packaged/" + defaultFallbackFilename); err != nil {
		return fontfind.NullFont, err
	}
	path := "packaged/" + defaultFallbackFilename
	sfnt := fontfind.ScalableFont{
		Name:   defaultFallbackFilename,
		Style:  font.StyleNormal,
		Weight: font.WeightNormal,
	}
	sfnt.SetFS(packaged, path)
	return sfnt, nil
}

// FindFallbackFont looks up a matching font in embedded fallback resources.
// It matches a case-insensitive filename substring and inferred style and weight.
// If no match exists, it returns NullFont and an error. Use Default to request
// the packaged default without matching a name.
func FindFallbackFont(pattern string, style font.Style, weight font.Weight) (fontfind.ScalableFont, error) {
	fonts, err := packaged.ReadDir("packaged")
	if err != nil {
		return fontfind.NullFont, err
	}
	var fname string // path to embedded font, if any
	for _, f := range fonts {
		if f.IsDir() {
			continue
		}
		if fontfind.Matches(f.Name(), pattern, style, weight) {
			tracer().Debugf("found embedded font file %s", f.Name())
			fname = f.Name()
			break
		}
	}
	var sFont fontfind.ScalableFont
	if fname == "" {
		return fontfind.NullFont, errors.New("font not found")
	}
	// font is packaged embedded font
	sFont.Name = fname
	sFont.Style = style
	sFont.Weight = weight
	sFont.SetFS(packaged, "packaged/"+fname)
	return sFont, nil
}
