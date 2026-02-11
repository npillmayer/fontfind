package systemfont

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/flopp/go-findfont"
	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/fontfind/locate"
	"github.com/npillmayer/schuko/tracing"
	"golang.org/x/image/font"
)

// tracer writes to trace with key 'tyse.font'
func tracer() tracing.Trace {
	return tracing.Select("tyse.font")
}

func Find(appkey string) locate.FontLocator {
	return func(descr fontfind.Descriptor) (fontfind.ScalableFont, error) {
		pattern := descr.Pattern
		style := descr.Style
		weight := descr.Weight
		return FindLocalFont(appkey, pattern, style, weight)
	}
}

// FindLocalFont searches for a locally installed font variant.
//
// If present and configured, FindLocalFont will be using the fontconfig
// system (https://www.freedesktop.org/wiki/Software/fontconfig/).
//
// If fontconfig is not configured, FindLocalFont will fall back to scanning
// the system's fonts-folders (OS dependent).
func FindLocalFont(appkey string, pattern string, style font.Style, weight font.Weight) (fontfind.ScalableFont, error) {
	//
	variants, v := findFontConfigFont(appkey, pattern, style, weight)
	if variants.Family != "" {
		if fsys, path, err := pointSubFS(v); err == nil {
			return fontfind.ScalableFont{
				Name:       pattern,
				Weight:     weight,
				Style:      style,
				FileSystem: fsys,
				Path:       path,
			}, nil
		}
		return fontfind.NullFont, errors.New("path error with fontconfig file path")
	}
	if loadedFontConfigListOK { // fontconfig is active, but didn't find a font
		// therefore don't do a file system scan
		return fontfind.NullFont, errors.New("no such font")
	}
	// otherwise fontconfig is not active => scan file system
	fpath, err := findfont.Find(pattern) // go-findfont lib does not accept style & weight
	if err == nil && fpath != "" {
		tracer().Debugf("%s is a system font: %s", pattern, fpath)
		if fsys, path, err := pointSubFS(fpath); err == nil {
			return fontfind.ScalableFont{
				Name:       pattern,
				Weight:     weight,
				Style:      style,
				FileSystem: fsys,
				Path:       path,
			}, nil
		}
		return fontfind.NullFont, errors.New("path error with system font file path")
	}
	return fontfind.NullFont, errors.New("no such font")
}

func pointSubFS(fontpath string) (fs.FS, string, error) {
	d, f := filepath.Split(fontpath)
	return os.DirFS(d), f, nil
}
