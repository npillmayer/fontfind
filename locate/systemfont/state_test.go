package systemfont

import (
	"fmt"
	"io"
	"io/fs"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"github.com/npillmayer/fontfind"
)

type fixtureIO struct {
	fsys  fs.FS
	reads atomic.Int64
}

func (f *fixtureIO) UserConfigDir() (string, error)      { return "/config", nil }
func (f *fixtureIO) DirFS(string) fs.FS                  { f.reads.Add(1); return f.fsys }
func (f *fixtureIO) ReadAll(r io.Reader) ([]byte, error) { return io.ReadAll(r) }
func fixture(name string) *fixtureIO {
	return &fixtureIO{fsys: fstest.MapFS{"fontconfig/fontlist.txt": &fstest.MapFile{Data: []byte(fmt.Sprintf("/fonts/%s.ttf: Probe:style=Regular\n", name))}}}
}

func TestLocatorStateIsolation(t *testing.T) {
	for _, direct := range []bool{false, true} {
		t.Run(fmt.Sprint(direct), func(t *testing.T) {
			var wg sync.WaitGroup
			for i := range 8 {
				wg.Go(func() {
					name := fmt.Sprintf("client%d", i)
					host := fixture(name)
					lookup := Find("app", host)
					for range 4 {
						d := fontfind.Descriptor{Pattern: "Probe"}
						f, err := lookup(d)
						if direct {
							f, err = FindLocalFont("app", host, d.Pattern, d.Style, d.Weight)
						}
						if err != nil || f.Path() != name+".ttf" {
							t.Errorf("client %s got %q, %v", name, f.Path(), err)
						}
					}
					if !direct && host.reads.Load() != 1 {
						t.Errorf("directory loaded %d times", host.reads.Load())
					}
				})
			}
			wg.Wait()
		})
	}
}

func TestSharedLocatorConcurrentInitialization(t *testing.T) {
	host := fixture("shared")
	lookup := Find("app", host)
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			f, err := lookup(fontfind.Descriptor{Pattern: "Probe"})
			if err != nil || f.Path() != "shared.ttf" {
				t.Errorf("shared lookup: %q, %v", f.Path(), err)
			}
		})
	}
	wg.Wait()
	if host.reads.Load() != 1 {
		t.Fatalf("loaded directory %d times", host.reads.Load())
	}
}
