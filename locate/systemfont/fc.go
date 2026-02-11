package systemfont

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/npillmayer/fontloading"
	"github.com/npillmayer/schuko"
	xfont "golang.org/x/image/font"
)

func findFontConfigBinary(conf schuko.Configuration) (path string, err error) {
	path = conf.GetString("fontconfig")
	if path == "" {
		tracer().Infof("fontconfig not configured: key 'fontconfig' should point location of 'fc-list' binary")
		err = errors.New("fontconfig not configured")
	}
	return
}

// findFontListConfig will create a sub-filesystem for the user's configuration directory,
// suffixed with "<appkey>/fontconfig".
func findFontListConfigDir(appkey string) (fs.FS, error) {
	if appkey == "" {
		return nil, errors.New("missing app-key for font list config search")
	}
	uconfdir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("cannot open user configuration directory: %w", err)
	}
	fontListConfigDir := path.Join(uconfdir, appkey)
	return fs.Sub(os.DirFS(fontListConfigDir), "fontconfig")
}

func findFontList(appkey string) ([]byte, error) {
	const listfile = "fontconfig.txt"
	configFS, err := findFontListConfigDir(appkey)
	if err != nil {
		return nil, err
	}
	if readFS, ok := configFS.(fs.ReadFileFS); ok {
		// Fast file reading within the sandboxed config area
		return readFS.ReadFile(listfile)
	}
	// else do it the traditional way
	file, err := configFS.Open(listfile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

var noFonts = []fontloading.FontVariantsLocation{}

// loadFontConfigList searches the user's configuration directory for a font list file,
// then reads the file and parses it into a list of font variants.
// This list of font variants is then stored globally.
func loadFontConfigList(appkey string) ([]fontloading.FontVariantsLocation, bool) {
	fclist, err := findFontList(appkey)
	if err != nil {
		return noFonts, false
	}
	r := bytes.NewReader(fclist)
	scanner := bufio.NewScanner(r)
	ttc := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}
		fontpath := strings.TrimSpace(fields[0])
		fontname := strings.TrimSpace(fields[1])
		fontname = strings.TrimPrefix(fontname, ".")
		fontvari := strings.ToLower(fields[2])
		if strings.HasSuffix(fontpath, ".ttc") {
			ttc++
			continue
		}
		desc := fontloading.FontVariantsLocation{
			Family: fontname,
			Path:   fontpath,
		}
		if strings.Contains(fontvari, "regular") {
			desc.Variants = []string{"regular"}
		} else if strings.Contains(fontvari, "text") {
			desc.Variants = []string{"regular"}
		} else if strings.Contains(fontvari, "light") {
			desc.Variants = []string{"light"}
		} else if strings.Contains(fontvari, "italic") {
			desc.Variants = []string{"italic"}
		} else if strings.Contains(fontvari, "bold") {
			desc.Variants = []string{"bold"}
		} else if strings.Contains(fontvari, "black") {
			desc.Variants = []string{"bold"}
		}
		fontConfigDescriptors = append(fontConfigDescriptors, desc)
	}
	if err = scanner.Err(); err != nil {
		err = fmt.Errorf("encountered a problem during reading of fontconfig font list: %s", fclist)
		return fontConfigDescriptors, false
	}
	if ttc > 0 {
		tracer().Infof("skipping %d platform fonts: TTC not yet supported", ttc)
	}
	return fontConfigDescriptors, true
}

var loadFontConfigListTask sync.Once
var loadedFontConfigListOK bool
var fontConfigDescriptors []fontloading.FontVariantsLocation

// findFontConfigFont searches for a locally installed font variant using the fontconfig
// system (https://www.freedesktop.org/wiki/Software/fontconfig/).
// However, we need some preparation from the user to de-couple from the
// fontconfig library.
func findFontConfigFont(appkey string, pattern string, style xfont.Style, weight xfont.Weight) (
	desc fontloading.FontVariantsLocation, variant string) {
	//
	loadFontConfigListTask.Do(func() {
		_, loadedFontConfigListOK = loadFontConfigList(appkey)
		tracer().Infof("loaded fontconfig list")
	})
	if !loadedFontConfigListOK {
		return
	}
	var confidence fontloading.MatchConfidence
	desc, variant, confidence = fontloading.ClosestMatch(fontConfigDescriptors, pattern, style, weight)
	tracer().Debugf("closest fontconfig match confidence for %s|%s= %d", desc.Family, variant, confidence)
	if confidence > fontloading.LowConfidence {
		return
	}
	return fontloading.FontVariantsLocation{}, ""
}
