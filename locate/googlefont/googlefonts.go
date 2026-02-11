package googlefont

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/npillmayer/fontloading"
	"github.com/npillmayer/schuko"
	"github.com/npillmayer/schuko/tracing"
	font "golang.org/x/image/font"
)

// GoogleFontInfo describes a font entry in the Google Font Service.
type GoogleFontInfo struct {
	fontloading.FontVariantsLocation
	Version string            `json:"version"`
	Subsets []string          `json:"subsets"`
	Files   map[string]string `json:"files"`
}

type googleFontsList struct {
	Items []GoogleFontInfo `json:"items"`
}

var loadGoogleFontsDir sync.Once
var googleFontsDirectory googleFontsList
var googleFontsLoadError error
var googleFontsAPI string = `https://www.googleapis.com/webfonts/v1/webfonts?`

func setupGoogleFontsDirectory(conf schuko.Configuration) (err error) {
	loadGoogleFontsDir.Do(func() {
		tracer().Infof("setting up Google Fonts service directory")
		apikey := conf.GetString("google-fonts-api-key")
		if apikey == "" {
			if apikey = os.Getenv("GOOGLE_FONTS_API_KEY"); apikey == "" {
				err = errors.New("Google fonts API key not set")
				tracer().Errorf(err.Error())
				googleFontsLoadError = fmt.Errorf(`Google Fonts API-key must be set in global configuration or as GOOGLE_API_KEY in environment;
      please refer to https://developers.google.com/fonts/docs/developer_api`)
				return
			}
		}
		values := url.Values{
			"sort": []string{"alpha"},
			"key":  []string{apikey},
		}
		var resp *http.Response
		resp, err = http.Get(googleFontsAPI + values.Encode())
		if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
			tracer().Errorf("Google Fonts API request not OK, error = %v", err)
			googleFontsLoadError = fmt.Errorf("could not get fonts-diretory from Google font service")
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			tracer().Errorf("Google Fonts API request not OK, status = %d", resp.StatusCode)
			googleFontsLoadError = fmt.Errorf("could not get fonts-diretory from Google font service")
			return
		}
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&googleFontsDirectory)
		if err != nil {
			err = fmt.Errorf("could not decode fonts-list from Google font service")
			return
		}
		tracer().Infof("transfered list of %d fonts from Google Fonts service",
			len(googleFontsDirectory.Items))
	})
	return
}

func FindGoogleFont(conf schuko.Configuration, pattern string, style font.Style, weight font.Weight) (
	fontloading.ScalableFont, error) {
	//
	fiList, err := matchGoogleFontInfo(conf, pattern, style, weight)
	if err != nil {
		return fontloading.NullFont, err
	}
	if len(fiList) == 0 {
		return fontloading.NullFont, fmt.Errorf("no matching Google font found")
	}
	fi := fiList[0]
	variant := fi.Variants[0] // TODO find variant with highest confidence
	cachedir, name, err := cacheGoogleFont(conf, fi, variant)
	if err != nil {
		return fontloading.NullFont, err
	}
	fsys := os.DirFS(cachedir)
	return fontloading.ScalableFont{
		Name:       name,
		Style:      style,
		Weight:     weight,
		FileSystem: fsys,
		Path:       name,
	}, nil
}

// FindGoogleFont scans the Google Font Service for fonts matching `pattern` and
// having a given style and weight.
//
// Will include all fonts with a match-confidence greater than `font.LowConfidence`.
//
// A prerequisite to looking for Google fonts is a valid API-key (refer to
// https://developers.google.com/fonts/docs/developer_api). It has to be configured
// either in the application setup or as an environment variable GOOGLE_API_KEY.
func matchGoogleFontInfo(conf schuko.Configuration, pattern string, style font.Style, weight font.Weight) (
	[]GoogleFontInfo, error) {
	//
	var fiList []GoogleFontInfo
	if err := setupGoogleFontsDirectory(conf); err != nil {
		return fiList, err
	}
	r, err := regexp.Compile(strings.ToLower(pattern))
	if err != nil {
		return fiList, fmt.Errorf("cannot match Google font: invalid font name pattern: %v", err)
	}
	tracer().Debugf("trying to match (%s)", strings.ToLower(pattern))
	for _, finfo := range googleFontsDirectory.Items {
		//trace().Debugf("testing (%s)", strings.ToLower(finfo.Family))
		if r.MatchString(strings.ToLower(finfo.Family)) {
			tracer().Debugf("Google font name matches pattern: %s", finfo.Family)
			_, _, confidence := fontloading.ClosestMatch([]fontloading.FontVariantsLocation{finfo.FontVariantsLocation}, pattern,
				style, weight)
			if confidence > fontloading.LowConfidence {
				fiList = append(fiList, finfo)
				break
			}
		}
	}
	if len(fiList) == 0 {
		return fiList, errors.New("no Google font matches pattern")
	}
	tracer().Debugf("found Google font: %v", fiList[0])
	return fiList, nil
}

// ---------------------------------------------------------------------------

// cacheGoogleFont loads a font described by fi with a given variant.
// The loaded font is cached in the user's cache directory.
func cacheGoogleFont(conf schuko.Configuration, fi GoogleFontInfo, variant string) (
	cachedir, name string, err error) {
	//
	var fileurl string
	for _, v := range fi.Variants {
		if v == variant {
			fileurl = fi.Files[v]
		}
	}
	if fileurl == "" {
		return "", "", fmt.Errorf("no variant equals %s, cannot cache %s", variant, fi.Family)
	}
	letter := strings.ToUpper(fi.Family[:1])
	cachedir, err = cacheFontDirPath(conf, letter)
	if err != nil {
		return "", "", err
	}
	ext := path.Ext(fileurl)
	name = fi.Family + "-" + variant + ext
	filepath := path.Join(cachedir, name)
	tracer().Infof("caching font %s as %s", fi.Family, filepath)
	if _, err := os.Stat(filepath); err == nil {
		tracer().Infof("font already cached: %s", filepath)
	} else {
		err = downloadCachedFile(filepath, fileurl)
	}
	return
}

// ---------------------------------------------------------------------------

// ListGoogleFonts produces a listing of available fonts from the Google webfont
// service, with font-family names matching a given pattern.
// Output goes into the trace file with log-level info.
//
// If not aleady done, the list of available fonts will be downloaded from Google.
func ListGoogleFonts(conf schuko.Configuration, pattern string) {
	level := tracer().GetTraceLevel()
	tracer().SetTraceLevel(tracing.LevelInfo)
	if err := setupGoogleFontsDirectory(conf); err != nil {
		tracer().Errorf("unable to list Google fonts: %v", err)
	} else {
		listGoogleFonts(googleFontsDirectory, pattern)
	}
	tracer().SetTraceLevel(level)
}

func listGoogleFonts(list googleFontsList, pattern string) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		tracer().Errorf("cannot list Google fonts: invalid pattern: %v", err)
	}
	tracer().Infof("%d fonts in Google font list", len(list.Items))
	tracer().Infof("======================================")
	for i, finfo := range list.Items {
		if r.MatchString(finfo.Family) {
			tracer().Infof("[%4d] %-20s: %s", i, finfo.Family, finfo.Version)
			tracer().Infof("       subsets: %v", finfo.Subsets)
			for k, v := range finfo.Files {
				tracer().Infof("       - %-18s: %s", k, v[len(v)-4:])
			}
		}
	}
}
