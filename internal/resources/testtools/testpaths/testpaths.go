package testpaths

import (
	"path"
	"path/filepath"
	"runtime"
)

// TempDir returns the path to the localized tmp/ directory.
func TempDir() (tempdir string) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Can not get filename")
	}
	tempdir, err := filepath.Abs(path.Join(path.Dir(filename), "..", "..", "..", "..", "tmp"))
	if err != nil {
		panic(err)
	}
	return tempdir
}
