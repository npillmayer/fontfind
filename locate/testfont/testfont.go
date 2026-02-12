package testfont

import (
	"embed"
	"errors"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/fontfind/locate"
	"github.com/npillmayer/schuko/tracing"
	"golang.org/x/image/font"
)

// tracer writes to trace with key 'testfont'
func tracer() tracing.Trace {
	return tracing.Select("testfont")
}

// Find creates a locator that resolves fonts from the embedded fallback set.
func Find(testdata embed.FS) locate.FontLocator {
	return func(descr fontfind.Descriptor) (fontfind.ScalableFont, error) {
		pattern := descr.Pattern
		style := descr.Style
		weight := descr.Weight
		return findTestFont(testdata, pattern, style, weight)
	}
}

// findTestFont looks up a matching font in embedded fallback resources.
// If no match exists, it returns the first available packaged font.
func findTestFont(testdata embed.FS, pattern string, style font.Style, weight font.Weight) (
	fontfind.ScalableFont, error) {
	//
	fontDir, err := testdata.ReadDir("testdata")
	if err != nil {
		return fontfind.NullFont, err
	}
	var fname string // path to embedded font, if any
	for _, f := range fontDir {
		if f.IsDir() {
			continue
		}
		if fontfind.Matches(f.Name(), pattern, style, weight) {
			tracer().Debugf("found embedded font file %s", f.Name())
			fname = f.Name()
			break
		}
		fname = f.Name()
	}
	var sFont fontfind.ScalableFont
	if fname == "" {
		return fontfind.NullFont, errors.New("font not found")
	}
	// font is packaged embedded font from testdata directory
	sFont.Name = fname
	sFont.Style = style
	sFont.Weight = weight
	sFont.SetFS(testdata, "testdata/"+fname)
	return sFont, nil
}
