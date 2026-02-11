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

func Find() locate.FontLocator {
	return func(descr fontfind.Descriptor) (fontfind.ScalableFont, error) {
		pattern := descr.Pattern
		style := descr.Style
		weight := descr.Weight
		return FindFallbackFont(pattern, style, weight)
	}
}

func FindFallbackFont(pattern string, style font.Style, weight font.Weight) (fontfind.ScalableFont, error) {
	fonts, _ := packaged.ReadDir("packaged")
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
		fname = f.Name()
	}
	var sFont fontfind.ScalableFont
	if fname == "" {
		return fontfind.NullFont, errors.New("font not found")
	}
	// font is packaged embedded font
	sFont.Name = fname
	sFont.Path = "packaged/" + fname
	sFont.FileSystem = packaged
	sFont.Style = style
	sFont.Weight = weight
	return sFont, nil
}
