package googlefont

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/npillmayer/schuko/schukonf/testconfig"
)

func TestCacheDownload(t *testing.T) {
	hostio := newFakeIO(t)
	const payload = "<svg xmlns='http://www.w3.org/2000/svg'></svg>\n"
	hostio.fontBytes = []byte(payload)
	const url = "https://example.test/UAX-Logo-shadow.svg"

	conf := testconfig.Conf{
		"app-key":         "tyse-test",
		"fonts-cache-dir": t.TempDir(),
	}

	cachedir, err := cacheFontDirPath(hostio, conf, "A")
	if err != nil {
		t.Fatal(err)
	}
	dst := path.Join(cachedir, "test.svg")
	err = downloadCachedFile(hostio, dst, url)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte(payload)) {
		t.Fatalf("cached file differs from response payload")
	}
}
