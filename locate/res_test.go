package locate_test

import (
	"testing"

	"github.com/npillmayer/fontloading"
	"github.com/npillmayer/fontloading/locate"
	"github.com/npillmayer/fontloading/locate/fallbackfont"
	"github.com/npillmayer/fontloading/locate/googlefont"
	"github.com/npillmayer/schuko/schukonf/testconfig"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
	"golang.org/x/image/font"
)

func TestLoadPackagedFont(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fontloading")
	defer teardown()
	//
	desc := fontloading.Descriptor{
		Pattern: "Go",
		Style:   font.StyleNormal,
		Weight:  font.WeightNormal,
	}
	fallback := fallbackfont.Find()
	loader := locate.ResolveTypeface(desc, fallback)
	//time.Sleep(500)
	f, err := loader.Typeface()
	if err != nil {
		t.Error(err)
	}
	if f.FileSystem == nil {
		t.Fatalf("font's filesystem is nil, cannot access font data")
	}
}

func TestResolveGoogleFont(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "resources")
	defer teardown()
	//
	conf := testconfig.Conf{
		"app-key": "tyse-test",
	}
	desc := fontloading.Descriptor{
		Pattern: "Antic",
		Style:   font.StyleNormal,
		Weight:  font.WeightNormal,
	}
	google := googlefont.Find(conf)
	loader := locate.ResolveTypeface(desc, google)
	//time.Sleep(500)
	f, err := loader.Typeface()
	if err != nil {
		t.Error(err)
	}
	if f.FileSystem == nil {
		t.Fatalf("font's filesystem is nil, cannot access font data")
	}
	//
	//fontregistry.GlobalRegistry().LogFontList()
}

var fclist = `
/System/Library/Fonts/Supplemental/NotoSansGothic-Regular.ttf: Noto Sans Gothic:style=Regular
/System/Library/Fonts/NotoSerifMyanmar.ttc: Noto Serif Myanmar,Noto Serif Myanmar Light:style=Light,Regular
/System/Library/Fonts/Supplemental/NotoSansCarian-Regular.ttf: Noto Sans Carian:style=Regular
/System/Library/Fonts/NotoSansMyanmar.ttc: Noto Sans Zawgyi:style=Regular
/System/Library/Fonts/Supplemental/NotoSansSylotiNagri-Regular.ttf: Noto Sans Syloti Nagri:style=Regular
/System/Library/Fonts/NotoNastaliq.ttc: Noto Nastaliq Urdu:style=Bold
/System/Library/Fonts/Supplemental/NotoSansCham-Regular.ttf: Noto Sans Cham:style=Regular
/System/Library/Fonts/NotoSansArmenian.ttc: Noto Sans Armenian:style=Bold
`

func TestFCFind(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "resources")
	defer teardown()
	//
	// TODO create test fs.FS
	conf := testconfig.Conf{
		"fontconfig": "/usr/local/bin/fc-list",
		"app-key":    "tyse-test",
	}
	_ = conf
	//
	// f, v := findFontConfigFont(conf, "new york", font.StyleItalic, font.WeightNormal)
	// t.Logf("found font = (%s) in variant (%s)", f.Family, v)
	// t.Logf("font file = %s", f.Path)
	// if f.Family != "New York" {
	// 	t.Errorf("expected to find font New York, found %v", f)
	// }
}
