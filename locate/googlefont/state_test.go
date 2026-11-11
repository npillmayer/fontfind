package googlefont

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/schuko/schukonf/testconfig"
	"golang.org/x/image/font"
)

func TestLocatorDirectoryConcurrentAndIsolated(t *testing.T) {
	var wg sync.WaitGroup
	for _, family := range []string{"ClientOne", "ClientTwo"} {
		host := newFakeIO(t)
		host.webfontsJSON = []byte(`{"items":[{"family":"` + family + `","variants":["regular"]}]}`)
		// A deliberately unmatched request exercises concurrent directory access
		// without mixing in download/cache-file behavior.
		lookup := Find(testconfig.Conf{}, host)
		for range 8 {
			wg.Go(func() {
				if _, err := lookup(fontfind.Descriptor{Pattern: "absent"}); err == nil {
					t.Error("expected lookup miss")
				}
			})
		}
		wg.Wait()
		if len(host.requestedURL) != 1 {
			t.Fatalf("directory fetched %d times", len(host.requestedURL))
		}
		// The decoded state also belongs to this service, not another client.
		svc := newGoogleService(host)
		info, err := svc.matchGoogleFontInfo(testconfig.Conf{}, family, font.StyleNormal, font.WeightNormal)
		if err != nil || len(info) != 1 || info[0].Family != family {
			t.Fatalf("client directory mismatch: %v, %v", info, err)
		}
	}
}

type directoryTransport func(*http.Request) (*http.Response, error)

func (f directoryTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDirectLookupDoesNotRetainOtherClientFailure(t *testing.T) {
	t.Setenv("GOOGLE_FONTS_API_KEY", "")
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	var requests atomic.Int64
	http.DefaultTransport = directoryTransport(func(r *http.Request) (*http.Response, error) {
		requests.Add(1)
		if r.URL.Query().Get("key") != "fixture-key" {
			t.Error("unexpected API key")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"items":[]}`)), Header: make(http.Header)}, nil
	})
	_, err := FindGoogleFont(testconfig.Conf{}, "missing", font.StyleNormal, font.WeightNormal)
	if err == nil {
		t.Fatal("expected missing-key error")
	}
	_, err = FindGoogleFont(testconfig.Conf{"google-fonts-api-key": "fixture-key"}, "missing", font.StyleNormal, font.WeightNormal)
	if err == nil {
		t.Fatal("expected no matching font")
	}
	if requests.Load() != 1 {
		t.Fatalf("second client's directory was never requested: requests=%d", requests.Load())
	}
}
