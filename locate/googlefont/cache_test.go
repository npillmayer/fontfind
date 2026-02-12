package googlefont

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/npillmayer/schuko/schukonf/testconfig"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestCacheDownload(t *testing.T) {
	const payload = "<svg xmlns='http://www.w3.org/2000/svg'></svg>\n"
	const url = "https://example.test/UAX-Logo-shadow.svg"
	prevTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != url {
			t.Fatalf("unexpected request URL %q", r.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(payload)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = prevTransport })

	conf := testconfig.Conf{
		"app-key":         "tyse-test",
		"fonts-cache-dir": t.TempDir(),
	}

	cachedir, err := cacheFontDirPath(conf, "A")
	if err != nil {
		t.Fatal(err)
	}
	dst := path.Join(cachedir, "test.svg")
	err = downloadCachedFile(dst, url)
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
