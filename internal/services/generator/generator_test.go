package generator

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	src = `package somepkg

import (
	"io"
	"os"
)

//` + `go:generate ifaces head /tmp/test_ifaces.go

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

//` + `go:generate ifaces entry /tmp/test_ifaces.go

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

func TestGenerator_Head(t *testing.T) {
	gen := New(Options{
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Wild:    wild,
	})
	cur := &bytes.Buffer{}
	buf := &bytes.Buffer{}
	err := gen.Head("mysrc.go", src, "test_ifaces.go", cur, buf)
	if err != nil {
		t.Error(err)
	}
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
	assert.Equal(t, expected, buf.String())
}

func TestGenerator_Head_Struct(t *testing.T) {
	gen := New(Options{
		Comment: comment,
		Pkg:     pkg,
		Post:    post,
		Pre:     pre,
		Struct:  true,
		Wild:    wild,
	})
	cur := &bytes.Buffer{}
	out := &bytes.Buffer{}
	err := gen.Head("mysrc.go", src, "test_ifaces.go", cur, out)
	if err != nil {
		t.Error(err)
	}
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

func TestGenerator_Head_NoHdr(t *testing.T) {
	gen := New(Options{
		NoHdr:   true,
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Wild:    wild,
	})
	cur := &bytes.Buffer{}
	cur.WriteString(fmt.Sprintf(`// DO NOT EDIT

package %s

`, pkg))

	out := &bytes.Buffer{}
	err := gen.Head("mysrc.go", src, "test_ifaces.go", cur, out)
	if err != nil {
		t.Error(err)
	}
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

func TestGenerator_Head_NoTypeDoc(t *testing.T) {
	gen := New(Options{
		NoTDoc:  true,
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Wild:    wild,
	})
	cur := &bytes.Buffer{}
	buf := &bytes.Buffer{}
	err := gen.Head("mysrc.go", src, "test_ifaces.go", cur, buf)
	if err != nil {
		t.Error(err)
	}
	expected := fmt.Sprintf(`// DO NOT EDIT

package %s

type PreMyStructPost interface {
	// Get func doc
	Get() string
	// Set func doc
	Set(item string)
}
`, pkg)
	assert.Equal(t, expected, buf.String())
}

func TestGenerator_Head_NoFuncDoc(t *testing.T) {
	gen := New(Options{
		NoFDoc:  true,
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Wild:    wild,
	})
	cur := &bytes.Buffer{}
	output := &bytes.Buffer{}
	err := gen.Head("mysrc.go", src, "test_ifaces.go", cur, output)
	if err != nil {
		t.Error(err)
	}
	expected := fmt.Sprintf(`// DO NOT EDIT

package %s

// PreMyStructPost type document
type PreMyStructPost interface {
	Get() string
	Set(item string)
}
`, pkg)
	assert.Equal(t, expected, output.String())
}

func TestGenerator_Item(t *testing.T) {
	gen := New(Options{
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Wild:    wild,
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

	curTarget := &bytes.Buffer{}
	newTarget := &bytes.Buffer{}

	curTarget.WriteString(curSrc)
	err := gen.Entry("mysrc.go", src, 23, "test_ifaces.go", curTarget, newTarget)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expected, newTarget.String())
}

func TestGenerator_Item_NoTypeDoc(t *testing.T) {
	gen := New(Options{
		NoTDoc:  true,
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Wild:    wild,
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

	cur := &bytes.Buffer{}
	buf := &bytes.Buffer{}

	cur.WriteString(curSrc)

	err := gen.Entry("mysrc.go", src, 23, "test_ifaces.go", cur, buf)
	if err != nil {
		t.Error(err)
	}
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
	assert.Equal(t, expected, buf.String())
}

func TestGenerator_Item_NoFuncDoc(t *testing.T) {
	gen := New(Options{
		NoFDoc:  true,
		Pre:     pre,
		Post:    post,
		Comment: comment,
		Pkg:     pkg,
		Wild:    wild,
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

	cur := &bytes.Buffer{}
	buf := &bytes.Buffer{}

	cur.WriteString(curSrc)
	err := gen.Entry("mysrc.go", src, 23, "test_ifaces.go", cur, buf)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expected, buf.String())
}
