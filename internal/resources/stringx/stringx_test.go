package stringx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExPkgPath(t *testing.T) {
	assert.Equal(t, `pkg`, ExPkgPath(`github.com/author/pkg-go`))
	assert.Equal(t, `subpkg`, ExPkgPath(`github.com/author/pkg-go/subpkg`))
}

func TestIsIdent(t *testing.T) {
	assert.True(t, IsIdent(`abc123`))
	assert.False(t, IsIdent(`123abc`))
	assert.False(t, IsIdent(``))
}

func TestIsPkg(t *testing.T) {
	assert.True(t, IsPkg(`abc123`))
	assert.False(t, IsPkg(`123abc`))
	assert.False(t, IsPkg(``))
	assert.False(t, IsPkg(`ABC123`))
}

func TestNotEmpty(t *testing.T) {
	assert.Equal(t, []string{`a`, `b`}, NotEmpty(``, `a`, ``, `b`))
}

func TestStripVersion(t *testing.T) {
	assert.Equal(t, `github.com/author/pkg-go`, StripVersion(`github.com/author/pkg-go@v1.0.0`))
	assert.Equal(t, `github.com/author/pkg-go/pkg`, StripVersion(`github.com/author/pkg-go@v1.0.0/pkg`))
}
