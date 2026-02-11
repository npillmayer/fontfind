package googlefont

import (
	"path"
	"testing"

	"github.com/npillmayer/schuko/schukonf/testconfig"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestCacheDownload(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse-test")
	defer teardown()
	//
	conf := testconfig.Conf{
		"app-key": "tyse-test",
	}
	//
	cachedir, err := cacheFontDirPath(conf, "A")
	if err != nil {
		t.Fatal(err)
	}
	err = downloadCachedFile(path.Join(cachedir, "test.svg"),
		"https://npillmayer.github.io/UAX/img/UAX-Logo-shadow.svg")
	if err != nil {
		t.Fatal(err)
	}
}
