package fontregistry

import (
	"testing"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
	"golang.org/x/image/font"
)

type sw struct {
	s font.Style
	w font.Weight
}

func TestGuess(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "resources")
	defer teardown()
	//
	for k, v := range map[string]sw{
		"fonts/Clarendon-bold.ttf":               {font.StyleNormal, font.WeightBold},
		"Microsoft/Gill Sans MT Bold Italic.ttf": {font.StyleItalic, font.WeightBold},
		"Cambria Math.ttf":                       {font.StyleNormal, font.WeightNormal},
	} {
		style, weight := fontfind.GuessStyleAndWeight(k)
		t.Logf("style = %d, weight = %d", style, weight)
		if style != v.s || weight != v.w {
			t.Errorf("expected different style or weight for %s", k)
		}
	}
}

func TestMatch(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "resources")
	defer teardown()
	//
	if !fontfind.Matches("fonts/Clarendon-bold.ttf",
		"clarendon", font.StyleNormal, font.WeightBold) {
		t.Errorf("expected match for Clarendon, haven't")
	}
	if !fontfind.Matches("Microsoft/Gill Sans MT Bold Italic.ttf",
		"gill sans", font.StyleItalic, font.WeightBold) {
		t.Errorf("expected match for Gill, haven't")
	}
	if !fontfind.Matches("Cambria Math.ttf",
		"cambria", font.StyleNormal, font.WeightNormal) {
		t.Errorf("expected match for Cambria Math, haven't")
	}
}

func TestNormalizeFont(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "resources")
	defer teardown()
	//
	n := NormalizeFontname("Clarendon", font.StyleItalic, font.WeightBold)
	if n == "fallback" || n != NormalizeFontname("Clarendon", font.StyleItalic, font.WeightBold) {
		t.Errorf("cache key must be deterministic and distinct from the fallback key")
	}
	for _, d := range []fontfind.Descriptor{
		{Pattern: "Clarendon", Style: font.StyleOblique, Weight: font.WeightBold},
		{Pattern: "Clarendon", Style: font.StyleItalic, Weight: font.WeightSemiBold},
		{Pattern: "Clarendon.ttf", Style: font.StyleItalic, Weight: font.WeightBold},
		{Pattern: "clarendon", Style: font.StyleItalic, Weight: font.WeightBold},
	} {
		if n == NormalizeFontname(d.Pattern, d.Style, d.Weight) {
			t.Errorf("distinct descriptor shares key: %+v", d)
		}
	}
	if NormalizeFontname("fallback", font.StyleNormal, font.WeightNormal) == fallbackFontKey {
		t.Error("family name collides with reserved fallback entry")
	}
}

func TestRegistryFallbackFont(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "resources")
	defer teardown()
	//
	fr := New()
	f, err := fr.FallbackFont()
	if err != nil {
		t.Fatal(err)
	}
	if f.Name != "Go-Regular.otf" {
		t.Fatalf("expected fallback font Go-Regular.otf, got %s", f.Name)
	}
}

func TestRegistryFontReturnsFallbackOnMiss(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "resources")
	defer teardown()
	//
	fr := New()
	f, err := fr.GetFont("font-not-in-registry")
	if err == nil {
		t.Fatal("expected miss error from registry lookup")
	}
	if f.Name != "Go-Regular.otf" {
		t.Fatalf("expected fallback font Go-Regular.otf, got %s", f.Name)
	}
}
