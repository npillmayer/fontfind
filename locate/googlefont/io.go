package googlefont

import (
	"io"
	"io/fs"
	"net/http"
	"os"
)

// IO abstracts host environment access for Google-font lookup and caching.
// It allows tests to replace OS and network interactions with deterministic fakes.
type IO interface {
	Getenv(string) string
	HTTPGet(string) (*http.Response, error)
	UserCacheDir() (string, error)
	DirFS(string) fs.FS
	Stat(string) (os.FileInfo, error)
	MkdirAll(string, fs.FileMode) error
	Create(string) (io.WriteCloser, error)
}

type systemIO struct{}

func (systemIO) Getenv(k string) string {
	return os.Getenv(k)
}

func (systemIO) HTTPGet(u string) (*http.Response, error) {
	return http.Get(u)
}

func (systemIO) UserCacheDir() (string, error) {
	return os.UserCacheDir()
}

func (systemIO) DirFS(path string) fs.FS {
	return os.DirFS(path)
}

func (systemIO) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (systemIO) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (systemIO) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}
