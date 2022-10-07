package paths

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dexterp/ifaces/internal/resources/envs"
	"github.com/stretchr/testify/assert"
)

func TestPathToImport_ModFile(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	p, err := PathToImport(file)
	assert.NoError(t, err)
	assert.Equal(t, `github.com/dexterp/ifaces/internal/resources/paths`, p)
}

func TestPathToImport_Go(t *testing.T) {
	file := filepath.Join(envs.Goroot(), `src`, `testing`, `testing.go`)
	i, err := PathToImport(file)
	assert.NoError(t, err)
	assert.Equal(t, `testing`, i)

}

func TestPathToImport_ModDir(t *testing.T) {
	m, err := filepath.Glob(filepath.Join(envs.Gopath(), `pkg`, `mod`, `github.com`, `stretchr`, `testify*`, `assert`))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NotZero(t, m) {
		t.FailNow()
	}
	p := m[0]
	i, err := PathToImport(p)
	assert.NoError(t, err)
	assert.Equal(t, `github.com/stretchr/testify/assert`, i)

}
