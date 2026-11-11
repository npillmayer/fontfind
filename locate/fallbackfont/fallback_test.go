package fallbackfont_test

import (
	"context"
	"testing"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/fontfind/fontregistry"
	"github.com/npillmayer/fontfind/locate"
	"github.com/npillmayer/fontfind/locate/fallbackfont"
	"golang.org/x/image/font"
)

func TestFindFallbackFont(t *testing.T) {
	for _, tc := range []struct {
		pattern string
		style   font.Style
		weight  font.Weight
		want    string
	}{
		{"Go-Regular", font.StyleNormal, font.WeightNormal, "Go-Regular.otf"},
		{"go-bold", font.StyleNormal, font.WeightBold, "Go-Bold.otf"},
		{"Go-Italic", font.StyleItalic, font.WeightNormal, "Go-Italic.otf"},
		{"Go-Bold-Italic", font.StyleItalic, font.WeightBold, "Go-Bold-Italic.otf"},
		{"nonexistent-embedded-font", font.StyleNormal, font.WeightNormal, ""},
		{"Go-Regular", font.StyleItalic, font.WeightNormal, ""},
		{"Go-Regular", font.StyleNormal, font.WeightBold, ""},
	} {
		t.Run(tc.pattern+"/"+tc.want, func(t *testing.T) {
			f, err := fallbackfont.FindFallbackFont(tc.pattern, tc.style, tc.weight)
			if tc.want == "" {
				if err == nil || f.Name != "" || f.Path() != "" {
					t.Fatalf("miss returned (%q, %v), want null font and error", f.Name, err)
				}
				return
			}
			if err != nil || f.Name != tc.want || f.Style != tc.style || f.Weight != tc.weight {
				t.Fatalf("match returned (%+v, %v), want %s", f, err, tc.want)
			}
			if data, err := f.ReadFontData(); err != nil || len(data) == 0 {
				t.Fatalf("matched font unreadable: %v", err)
			}
		})
	}
}

func TestEmbeddedMissPipeline(t *testing.T) {
	for _, nextProvider := range []bool{false, true} {
		t.Run(map[bool]string{false: "explicit fallback", true: "next provider"}[nextProvider], func(t *testing.T) {
			reg := fontregistry.New()
			desc := fontfind.Descriptor{Pattern: "nonexistent-embedded-font"}
			embedded := fallbackfont.Find()
			resolvers := []locate.FontLocatorWithContext{
				func(_ context.Context, d fontfind.Descriptor) (fontfind.ScalableFont, error) { return embedded(d) },
			}
			called := false
			if nextProvider {
				resolvers = append(resolvers, func(context.Context, fontfind.Descriptor) (fontfind.ScalableFont, error) {
					called = true
					return fontfind.FallbackFont(), nil
				})
			}
			f, err := locate.NewResolverPipeline(reg, resolvers...).Resolve(context.Background(), desc).Font()
			if f.Name != "Go-Regular.otf" || (err == nil) != nextProvider || called != nextProvider {
				t.Fatalf("pipeline returned (%q, %v), next provider called=%v", f.Name, err, called)
			}
			key := fontregistry.NormalizeFontname(desc.Pattern, desc.Style, desc.Weight)
			_, cacheErr := reg.GetFont(key)
			if (cacheErr == nil) != nextProvider {
				t.Fatalf("unexpected cache status: %v", cacheErr)
			}
		})
	}
}

func TestDefaultRemainsReadable(t *testing.T) {
	f, err := fallbackfont.Default()
	if err != nil || f.Name != "Go-Regular.otf" {
		t.Fatalf("default returned (%q, %v)", f.Name, err)
	}
	if data, err := f.ReadFontData(); err != nil || len(data) == 0 {
		t.Fatalf("default unreadable: %v", err)
	}
}
