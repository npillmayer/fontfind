package fontregistry

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/schuko/tracing"
)

type listTrace struct {
	tracing.Trace
	info func(string, ...interface{})
}

func (l listTrace) Infof(format string, args ...interface{}) { l.info(format, args...) }
func (l listTrace) GetTraceLevel() tracing.TraceLevel        { return tracing.LevelInfo }
func (l listTrace) SetTraceLevel(tracing.TraceLevel)         {}

func TestLogFontListConcurrent(t *testing.T) {
	r := New()
	var entries atomic.Int64
	log := listTrace{info: func(format string, args ...interface{}) {
		if strings.Contains(format, "typeface") {
			entries.Add(1)
		}
	}}
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range 100 {
			r.StoreFont(fmt.Sprint(i), fontfind.FallbackFont())
		}
	})
	wg.Go(func() {
		for range 100 {
			r.LogFontList(log)
		}
	})
	wg.Wait()
	r.LogFontList(log)
	if entries.Load() < 100 {
		t.Fatal("missing logged entries")
	}
}

func TestLogFontListSnapshot(t *testing.T) {
	r := New()
	f := fontfind.FallbackFont()
	r.StoreFont("original", f)
	var lines []string
	log := listTrace{info: func(format string, args ...interface{}) {
		// Logging callbacks may access the registry; no registry lock may be held.
		r.StoreFont("added during logging", f)
		lines = append(lines, fmt.Sprintf(format, args...))
	}}
	r.LogFontList(log)
	output := strings.Join(lines, "\n")
	if strings.Contains(output, "added during logging") || !strings.Contains(output, f.Path()) {
		t.Fatalf("expected original snapshot with actual path, got %s", output)
	}
}
