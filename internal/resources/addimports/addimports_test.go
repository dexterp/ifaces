package addimports

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddImports(t *testing.T) {
	src := `package pkg

import (
	_ "driver/sql"
	"io"
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
	expected := `package pkg

import (
	_ "driver/sql"
	"io"
	"os"
	driver "src.com/author/sql-driver"
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
		NewImport(``, `io`),
		NewImport(``, `os`),
		NewImport(`driver`, `src.com/author/sql-driver`),
	}

	out := &bytes.Buffer{}
	err := AddImports("srg.go", src, imports, out)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expected, out.String())
}
