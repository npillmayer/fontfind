package fontloading

import (
	"path"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/image/font"
)

// Descriptor represents all the known variants of a font.
type FontVariantsLocation struct {
	Family   string   `json:"family"`
	Variants []string `json:"variants"`
	Path     string   // only used if just a single variant
}

// Matches returns true if a font's filename contains pattern and indicators
// for a given style and weight.
func Matches(fontfilename, pattern string, style font.Style, weight font.Weight) bool {
	basename := path.Base(fontfilename)
	basename = basename[:len(basename)-len(path.Ext(basename))]
	basename = strings.ToLower(basename)
	tracer().Debugf("basename of font = %s", basename)
	if !strings.Contains(basename, strings.ToLower(pattern)) {
		return false
	}
	s, w := GuessStyleAndWeight(basename)
	if s == style && w == weight {
		return true
	}
	return false
}

// MatchConfidence is a type for expressing the confidence level of font matching.
type MatchConfidence int

const (
	NoConfidence      MatchConfidence = 0
	LowConfidence     MatchConfidence = 2
	HighConfidence    MatchConfidence = 3
	PerfectConfidence MatchConfidence = 4
)

// ClosestMatch scans a list of font desriptors and returns the closest match
// for a given set of parameters.
//
// If no variant matches, returns `NoConfidence`.
func ClosestMatch(fdescs []FontVariantsLocation, pattern string, style font.Style,
	weight font.Weight) (match FontVariantsLocation, variant string, confidence MatchConfidence) {
	//
	r, err := regexp.Compile(strings.ToLower(pattern))
	if err != nil {
		tracer().Errorf("invalid font name pattern")
		return
	}
	for _, fdesc := range fdescs {
		//trace().Debugf("trying to match %s", strings.ToLower(fdesc.Family))
		if !r.MatchString(strings.ToLower(fdesc.Family)) {
			continue
		}
		for _, v := range fdesc.Variants {
			s := MatchStyle(v, style)
			w := MatchWeight(v, weight)
			if (s+w)/2 > confidence {
				//trace().Debugf("variant %+v match confidence = %d + %d", v, s, w)
				confidence = (s + w) / 2
				variant = v
				match = fdesc
			}
		}
	}
	return
}

// ---------------------------------------------------------------------------

// GuessStyleAndWeight trys to guess a font's style and weight from the
// font's file name.
func GuessStyleAndWeight(fontfilename string) (font.Style, font.Weight) {
	fontfilename = path.Base(fontfilename)
	ext := path.Ext(fontfilename)
	fontfilename = strings.ToLower(fontfilename[:len(fontfilename)-len(ext)])
	s := strings.Split(fontfilename, "-")
	if len(s) > 1 {
		switch s[len(s)-1] {
		case "light", "xlight":
			return font.StyleNormal, font.WeightLight
		case "normal", "medium", "regular", "r":
			return font.StyleNormal, font.WeightNormal
		case "bold", "b":
			return font.StyleNormal, font.WeightBold
		case "xbold", "black":
			return font.StyleNormal, font.WeightExtraBold
		}
	}
	style, weight := font.StyleNormal, font.WeightNormal
	if strings.Contains(fontfilename, "italic") {
		style = font.StyleItalic
	}
	if strings.Contains(fontfilename, "light") {
		weight = font.WeightLight
	}
	if strings.Contains(fontfilename, "bold") {
		weight = font.WeightBold
	}
	return style, weight
}

// MatchStyle trys to match a font-variant to a given style.
func MatchStyle(variantName string, style font.Style) MatchConfidence {
	variantName = strings.ToLower(variantName)
	switch style {
	case font.StyleNormal:
		switch variantName {
		case "regular", "400":
			return PerfectConfidence
		case "100", "200", "300", "500":
			return HighConfidence
		}
		return NoConfidence
	case font.StyleItalic:
		if strings.Contains(variantName, "italic") {
			return PerfectConfidence
		}
		if strings.Contains(variantName, "obliq") {
			return HighConfidence
		}
		return NoConfidence
	case font.StyleOblique:
		if strings.Contains(variantName, "obliq") {
			return PerfectConfidence
		}
		if strings.Contains(variantName, "italic") {
			return HighConfidence
		}
		return NoConfidence
	}
	return NoConfidence
}

// MatchWeight trys to match a font-variant to a given weight.
func MatchWeight(variantName string, weight font.Weight) MatchConfidence {
	/* from https://pkg.go.dev/golang.org/x/image/font
	WeightThin       Weight = -3 // CSS font-weight value 100.
	WeightExtraLight Weight = -2 // CSS font-weight value 200.
	WeightLight      Weight = -1 // CSS font-weight value 300.
	WeightNormal     Weight = +0 // CSS font-weight value 400.
	WeightMedium     Weight = +1 // CSS font-weight value 500.
	WeightSemiBold   Weight = +2 // CSS font-weight value 600.
	WeightBold       Weight = +3 // CSS font-weight value 700.
	WeightExtraBold  Weight = +4 // CSS font-weight value 800.
	WeightBlack      Weight = +5 // CSS font-weight value 900.
	*/
	if strconv.Itoa(int(weight)+4*100) == variantName {
		return PerfectConfidence
	}
	switch variantName {
	case "regular", "400", "italic", "oblique", "normal", "text":
		switch weight {
		case font.WeightNormal, font.WeightMedium:
			return PerfectConfidence
		case font.WeightThin, font.WeightExtraLight, font.WeightLight:
			return LowConfidence
		}
		return NoConfidence
	case "100", "200", "300":
		switch weight {
		case font.WeightThin, font.WeightExtraLight, font.WeightLight:
			return PerfectConfidence
		case font.WeightNormal, font.WeightMedium:
			return LowConfidence
		}
		return NoConfidence
	case "500":
		switch weight {
		case font.WeightMedium:
			return PerfectConfidence
		case font.WeightSemiBold:
			return HighConfidence
		case font.WeightNormal, font.WeightBold:
			return LowConfidence
		}
		return NoConfidence
	case "bold", "700":
		switch weight {
		case font.WeightBold:
			return PerfectConfidence
		case font.WeightSemiBold, font.WeightExtraBold:
			return HighConfidence
		}
		return NoConfidence
	case "extrabold", "600", "800", "900":
		switch weight {
		case font.WeightSemiBold:
			return LowConfidence
		case font.WeightBold:
			return HighConfidence
		}
		return NoConfidence
	}
	return NoConfidence
}
