package modinfo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dexterp/ifaces/internal/resources/testtools/testpaths"
	"github.com/stretchr/testify/assert"
)

func srcFile() string {
	return `module github.com/author/mymodule

go 1.19`
}

func TestGetImport(t *testing.T) {
	p, err := GetImport([]byte(srcFile()), `/path/to/mod/go.mod`, `/path/to/mod/internal/read/read.go`)
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
	path, err := GetImport(nil, ``, srcpath)
	assert.NoError(t, err)
	assert.Equal(t, `github.com/author/mymodule/connect`, path)
}
