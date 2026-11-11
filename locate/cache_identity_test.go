package locate_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/fontfind/fontregistry"
	"github.com/npillmayer/fontfind/locate"
	"golang.org/x/image/font"
)

func TestPipelinePreservesRequestIdentity(t *testing.T) {
	reg := fontregistry.New()
	// Initialize the reserved fallback entry before looking up a family named fallback.
	if _, err := reg.FallbackFont(); err != nil {
		t.Fatal(err)
	}
	calls := 0
	resolver := func(_ context.Context, d fontfind.Descriptor) (fontfind.ScalableFont, error) {
		calls++
		f := fontfind.FallbackFont()
		f.Name = fmt.Sprintf("%q/%d/%d", d.Pattern, d.Style, d.Weight)
		return f, nil
	}
	pipeline := locate.NewResolverPipeline(reg, resolver)
	patterns := []string{"Family", "family", " Family", "Family ", "Family Name", "Family_Name", "Family.ttf", "Family.otf", "Family.*", "Family-italic", "Family-bold", "fallback", "", "a\"/style=1/weight=3", "a\x00b"}
	styles := []font.Style{font.StyleNormal, font.StyleItalic, font.StyleOblique, font.Style(99)}
	weights := []font.Weight{font.WeightThin, font.WeightExtraLight, font.WeightLight, font.WeightNormal, font.WeightMedium, font.WeightSemiBold, font.WeightBold, font.WeightExtraBold, font.WeightBlack, font.Weight(99)}
	for pass := range 2 {
		for _, pattern := range patterns {
			for _, style := range styles {
				for _, weight := range weights {
					d := fontfind.Descriptor{Pattern: pattern, Style: style, Weight: weight}
					f, err := pipeline.Resolve(context.Background(), d).Font()
					want := fmt.Sprintf("%q/%d/%d", pattern, style, weight)
					if err != nil || f.Name != want {
						t.Fatalf("pass %d, request %+v: got (%q, %v), want %q", pass, d, f.Name, err, want)
					}
				}
			}
		}
		if want := len(patterns) * len(styles) * len(weights); calls != want {
			t.Fatalf("pass %d: resolver called %d times, want %d", pass, calls, want)
		}
	}
}

func TestPipelineFreshRegistryForPolicyChange(t *testing.T) {
	provider := func(name string) locate.FontLocatorWithContext {
		return func(context.Context, fontfind.Descriptor) (fontfind.ScalableFont, error) {
			f := fontfind.FallbackFont()
			f.Name = name
			return f, nil
		}
	}
	a, b := provider("provider A"), provider("provider B")
	regA, regB := fontregistry.New(), fontregistry.New()
	customFallback := fontfind.FallbackFont()
	customFallback.Name = "client B fallback"
	regB.StoreFont("fallback", customFallback)
	desc := fontfind.Descriptor{Pattern: "same request"}
	for _, tc := range []struct {
		pipeline locate.ResolverPipeline
		want     string
	}{
		{locate.NewResolverPipeline(regA, a, b), "provider A"},
		{locate.NewResolverPipeline(regB, b, a), "provider B"},
		{locate.NewResolverPipeline(regA, a, b), "provider A"},
	} {
		f, err := tc.pipeline.Resolve(context.Background(), desc).Font()
		if err != nil || f.Name != tc.want {
			t.Fatalf("policy result (%q, %v), want %q", f.Name, err, tc.want)
		}
	}
	for _, tc := range []struct {
		reg  *fontregistry.Registry
		want string
	}{
		{regA, "Go-Regular.otf"}, {regB, "client B fallback"},
	} {
		f, err := locate.NewResolverPipeline(tc.reg).Resolve(context.Background(), fontfind.Descriptor{Pattern: "missing"}).Font()
		if err == nil || f.Name != tc.want {
			t.Fatalf("fallback result (%q, %v), want %q with miss error", f.Name, err, tc.want)
		}
	}
}
