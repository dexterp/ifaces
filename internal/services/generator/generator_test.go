package generator

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	src1 = `package somepkg

import (
	"io"
	"os"
)

//

// MyStruct type document
type MyStruct struct {
}

// Get func doc
func (m MyStruct) Get() string {
	return ""
}

// Set func doc
func (m *MyStruct) Set(item string) {
}

//` + `go:generate ifaces /tmp/test_ifaces.go -a

// SomeStruct type document
type SomeStruct struct {
}

// Scan scan func document
func (m SomeStruct) Scan(in io.Reader) error {
	fmt.Fprintf(os.Stderr, "debug print\n")
	return nil
}

// Delete delete func document
func (m SomeStruct) Delete() error {
}

// IgnoreStruct type document
type IgnoreStruct struct {
} 

// Connect connect func document
func (m *IgnoreStruct) Connect(connetstr string) {
}
`
	pre     = `Pre`
	post    = `Post`
	comment = `DO NOT EDIT`
	pkg     = `mypkg`
	wild    = `My*`
)

func TestGenerator_Generate(t *testing.T) {
	gen := New(Options{
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Match:   wild,
	})
	outfile := "src1.go"
	srcs := []*Src{
		{
			File: outfile,
			Src:  src1,
		},
	}
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	err := gen.Generate(srcs, in, outfile, out)
	assert.NoError(t, err)
	expected := fmt.Sprintf(`// DO NOT EDIT

package %s

// PreMyStructPost type document
type PreMyStructPost interface {
	// Get func doc
	Get() string
	// Set func doc
	Set(item string)
}
`, pkg)
	assert.Equal(t, expected, out.String())
}

func TestGenerator_Gen_Struct(t *testing.T) {
	gen := New(Options{
		Comment: comment,
		Pkg:     pkg,
		Post:    post,
		Pre:     pre,
		Struct:  true,
		Match:   wild,
	})
	outfile := `test_ifaces.go`
	srcs := []*Src{
		{
			File: outfile,
			Src:  src1,
		},
	}
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	err := gen.Generate(srcs, in, "test_ifaces.go", out)
	assert.NoError(t, err)
	expected := fmt.Sprintf(`// DO NOT EDIT

package %s

import (
	"io"
)

// PreMyStructPost type document
type PreMyStructPost interface {
	// Get func doc
	Get() string
	// Set func doc
	Set(item string)
}

// PreSomeStructPost type document
type PreSomeStructPost interface {
	// Scan scan func document
	Scan(in io.Reader) error
	// Delete delete func document
	Delete() error
}

// PreIgnoreStructPost type document
type PreIgnoreStructPost interface {
	// Connect connect func document
	Connect(connetstr string)
}
`, pkg)
	assert.Equal(t, expected, out.String())
}

func TestGenerator_Gen_NoTypeDoc(t *testing.T) {
	gen := New(Options{
		NoTDoc:  true,
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Match:   wild,
	})
	outfile := `test_ifaces.go`
	srcs := []*Src{
		{
			File: outfile,
			Src:  src1,
		},
	}
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	err := gen.Generate(srcs, in, "test_ifaces.go", out)
	assert.NoError(t, err)
	expected := fmt.Sprintf(`// DO NOT EDIT

package %s

type PreMyStructPost interface {
	// Get func doc
	Get() string
	// Set func doc
	Set(item string)
}
`, pkg)
	assert.Equal(t, expected, out.String())
}

func TestGenerator_Gen_NoFuncDoc(t *testing.T) {
	gen := New(Options{
		NoFDoc:  true,
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Match:   wild,
	})
	outfile := `test_ifaces.go`
	srcs := []*Src{
		{
			File: outfile,
			Src:  src1,
		},
	}
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	err := gen.Generate(srcs, in, "test_ifaces.go", out)
	assert.NoError(t, err)
	expected := fmt.Sprintf(`// DO NOT EDIT

package %s

// PreMyStructPost type document
type PreMyStructPost interface {
	Get() string
	Set(item string)
}
`, pkg)
	assert.Equal(t, expected, out.String())
}

func TestGenerator_Gen_Entry(t *testing.T) {
	gen := New(Options{
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
	})

	curSrc := fmt.Sprintf(`// DO NOT EDIT

package %s

import (
	"os"
)

// Iface
type Iface interface {
	SetStderr(stderr *os.Stderr)
}
`, pkg)

	expected := fmt.Sprintf(`// DO NOT EDIT

package %s

import (
	"io"
	"os"
)

// Iface
type Iface interface {
	SetStderr(stderr *os.Stderr)
}

// PreSomeStructPost type document
type PreSomeStructPost interface {
	// Scan scan func document
	Scan(in io.Reader) error
	// Delete delete func document
	Delete() error
}
`, pkg)
	outfile := `test_ifaces.go`
	srcs := []*Src{
		{
			File: outfile,
			Src:  src1,
			Line: 23,
		},
	}
	in := &bytes.Buffer{}
	in.WriteString(curSrc)
	out := &bytes.Buffer{}
	err := gen.Generate(srcs, in, "test_ifaces.go", out)
	assert.NoError(t, err)
	assert.Equal(t, expected, out.String())
}

func TestGenerator_Gen_Entry_NoTypeDoc(t *testing.T) {
	gen := New(Options{
		NoTDoc:  true,
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
	})

	curSrc := fmt.Sprintf(`// DO NOT EDIT

package %s

import (
	"os"
)

// Iface
type Iface interface {
	SetStderr(stderr *os.Stderr)
}
`, pkg)
	expected := fmt.Sprintf(`// DO NOT EDIT

package %s

import (
	"io"
	"os"
)

// Iface
type Iface interface {
	SetStderr(stderr *os.Stderr)
}

type PreSomeStructPost interface {
	// Scan scan func document
	Scan(in io.Reader) error
	// Delete delete func document
	Delete() error
}
`, pkg)
	outfile := `test_ifaces.go`
	srcs := []*Src{
		{
			File: outfile,
			Src:  src1,
			Line: 23,
		},
	}
	in := &bytes.Buffer{}
	in.WriteString(curSrc)
	out := &bytes.Buffer{}
	err := gen.Generate(srcs, in, "test_ifaces.go", out)
	assert.NoError(t, err)
	assert.Equal(t, expected, out.String())
}

func TestGenerator_Gen_Entry_NoFuncDoc(t *testing.T) {
	gen := New(Options{
		NoFDoc:  true,
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
	})

	curSrc := fmt.Sprintf(`// DO NOT EDIT

package %s

import (
	"os"
)

// Iface
type Iface interface {
	SetStderr(stderr *os.Stderr)
}
`, pkg)

	expected := fmt.Sprintf(`// DO NOT EDIT

package %s

import (
	"io"
	"os"
)

// Iface
type Iface interface {
	SetStderr(stderr *os.Stderr)
}

// PreSomeStructPost type document
type PreSomeStructPost interface {
	Scan(in io.Reader) error
	Delete() error
}
`, pkg)
	srcfile := `test_ifaces.go`
	srcs := []*Src{
		{
			File: srcfile,
			Src:  src1,
			Line: 23,
		},
	}
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	in.WriteString(curSrc)
	err := gen.Generate(srcs, in, "test_ifaces.go", out)
	assert.NoError(t, err)
	assert.Equal(t, expected, out.String())
}
