package addimports

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type imp struct {
	name string
	path string
}

func (i imp) Name() string {
	return i.name
}

func (i imp) Path() string {
	return i.path
}

func TestAddImports(t *testing.T) {
	src := `package pkg

type MyIO struct {
}

func (i MyIO) GetReader() io.Reader {
	return nil
}

func (i MyIO) GetStderr() *os.Stderr {
	return nil
}
`
	expected := `package pkg

import (
	"io"
	"os"
)

type MyIO struct {
}

func (i MyIO) GetReader() io.Reader {
	return nil
}

func (i MyIO) GetStderr() *os.Stderr {
	return nil
}
`
	imports := []Import{
		&imp{
			name: ``,
			path: `io`,
		},
		&imp{
			name: ``,
			path: `os`,
		},
	}

	out := &bytes.Buffer{}
	err := AddImports("srg.go", src, imports, out)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expected, out.String())
}
