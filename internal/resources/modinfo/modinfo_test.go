package modinfo

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dexterp/ifaces/internal/resources/testtools/testpaths"
	"github.com/stretchr/testify/assert"
)

func srcFile() string {
	return `module github.com/author/mymodule

go 1.19`
}

func TestGetImport(t *testing.T) {
	p, err := GetImport(`/path/to/mod/go.mod`, []byte(srcFile()), `/path/to/mod/internal/read/read.go`)
	assert.NoError(t, err)
	assert.Equal(t, `github.com/author/mymodule/internal/read`, p)
}

func TestGetImport_FromPath(t *testing.T) {
	// write go.mod
	tmpdir := testpaths.TempDir()
	p := filepath.Join(tmpdir, "test", "modinfo", "mymodule")
	err := os.MkdirAll(p, 0777)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	modpath := filepath.Join(p, `go.mod`)
	f, err := os.Create(modpath)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, err = f.WriteString(srcFile())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = f.Close()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	srcpath := filepath.Join(p, "connect", "connect.go")
	path, err := GetImport(``, nil, srcpath)
	assert.NoError(t, err)
	assert.Equal(t, `github.com/author/mymodule/connect`, path)
}

func TestModInfo_GetModVersion(t *testing.T) {
	gomodpath := filepath.Join(testpaths.RootDir(), `go.mod`)
	gomodbytes, err := os.ReadFile(gomodpath)
	assert.NoError(t, err)
	mod := "github.com/stretchr/testify"
	i, err := Load(gomodpath, gomodbytes)
	assert.NoError(t, err)
	v, err := i.GetVersion(mod)
	assert.NoError(t, err)
	assert.Regexp(t, `^v\d+\.\d+\.\d+`, v)
}

func TestGetModPath(t *testing.T) {
	mod := `github.com/stretchr/testify`
	_, filename, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	i, err := LoadFromParents(filename)
	assert.NoError(t, err)
	p, err := i.GetPath(mod)
	assert.NoError(t, err)
	assert.DirExists(t, p)
}

func TestGetModPath_Version(t *testing.T) {
	mod := `github.com/stretchr/testify`
	_, filename, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	i, err := LoadFromParents(filename)
	assert.NoError(t, err)
	v, err := i.GetVersion(mod)
	assert.NoError(t, err)
	p, err := i.GetPath(mod + `@` + v)
	assert.NoError(t, err)
	assert.DirExists(t, p)
}

func TestGetModPath_latest(t *testing.T) {
	mod := `github.com/stretchr/testify@latest`
	_, filename, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	i, err := LoadFromParents(filename)
	assert.NoError(t, err)
	p, err := i.GetPath(mod)
	assert.Error(t, err)
	assert.Equal(t, ErrLatestNotSupported, err)
	assert.Empty(t, p)
}
