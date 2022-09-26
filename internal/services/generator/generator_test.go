package generator

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	src1 = `package originpkg

import (
	"io"
	"os"
)

//

// MyStruct type document
type MyStruct struct {
}

// Get func doc
func (m MyStruct) Get() (item string) {
	return ""
}

// Set func doc
func (m *MyStruct) Set(item string) {
}

//` + `go:generate ifaces /tmp/test_ifaces.go

// SomeStruct type document
type SomeStruct struct {
}

// AddData add data doc
func (m *SomeStruct) AddData(d ...Data) {
}

// Add add data
func (m *SomeStruct) Add(d ...any) {
}

// Collate collate data
func (m *SomeStruct) Collate(in []*Data) error {
	return nil
}

// Scan scan func document
func (m SomeStruct) Scan(in io.Reader) error {
	fmt.Fprintf(os.Stderr, "debug print\n")
	return nil
}

// ScanMap scan map func
func (m SomeStruct) ScanMap(in map[string]string) error {
	return nil
}

// ScanMapMap scan map func
func (m SomeStruct) ScanMapMap(in map[string]map[string]string) error {
	return nil
}

// ScanMapSlice scan map func
func (m SomeStruct) ScanMapSlice(in map[string][]string) error {
	return nil
}

// ScanSlice scan slice func
func (m SomeStruct) ScanSlice(in []string) error {
	return nil
}

// ScanSliceMap scan slice map func
func (m SomeStruct) ScanSliceMap(in []map[string]string) error {
	return nil
}

// ScanSliceSlice scan slice map func
func (m SomeStruct) ScanSliceSlice(in [][]interface{}) error {
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

// Data
type Data struct {
}
`
	src2 = `package originpkg

type Service struct {
}

func (s Service) Download(wrtr io.Writer) error {
	return nil
}

//` + `go:generate ifaces recv service_ifaces.go --post Iface

func (s *Service) Upload(rdr io.Reader) error {
	return nil
}

func (s *Service) Shutdown() error {
}

//` + `go:generate ifaces recv service_ifaces.go --post Iface

func (s *Service) Status() int {
}

`
	pre       = `Pre`
	post      = `Post`
	comment   = `DO NOT EDIT`
	pkg       = `mypkg`
	matchType = `My*`
)

func TestGenerator_Generate(t *testing.T) {
	gen := New(Options{
		Type:      true,
		Pre:       pre,
		Post:      post,
		Comment:   comment,
		Pkg:       pkg,
		MatchType: matchType,
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
	Get() (item string)
	// Set func doc
	Set(item string)
}
`, pkg)
	assert.Equal(t, expected, out.String())
	fmt.Println(out.String())
}

func TestGenerator_Type_Struct(t *testing.T) {
	gen := New(Options{
		Type:    true,
		Comment: comment,
		Pkg:     pkg,
		Post:    post,
		Pre:     pre,
		Struct:  true,
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
	Get() (item string)
	// Set func doc
	Set(item string)
}

// PreSomeStructPost type document
type PreSomeStructPost interface {
	// AddData add data doc
	AddData(d ...originpkg.Data)
	// Add add data
	Add(d ...any)
	// Collate collate data
	Collate(in []*originpkg.Data) error
	// Scan scan func document
	Scan(in io.Reader) error
	// ScanMap scan map func
	ScanMap(in map[string]string) error
	// ScanMapMap scan map func
	ScanMapMap(in map[string]map[string]string) error
	// ScanMapSlice scan map func
	ScanMapSlice(in map[string][]string) error
	// ScanSlice scan slice func
	ScanSlice(in []string) error
	// ScanSliceMap scan slice map func
	ScanSliceMap(in []map[string]string) error
	// ScanSliceSlice scan slice map func
	ScanSliceSlice(in [][]interface{}) error
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
	fmt.Println(out.String())
}

func TestGenerator_Type_Match_NoTypeDoc(t *testing.T) {
	gen := New(Options{
		Type:      true,
		NoTDoc:    true,
		Pre:       pre,
		Post:      post,
		Comment:   comment,
		Pkg:       pkg,
		MatchType: matchType,
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
	Get() (item string)
	// Set func doc
	Set(item string)
}
`, pkg)
	assert.Equal(t, expected, out.String())
}

func TestGenerator_Type_Match_NoFuncDoc(t *testing.T) {
	gen := New(Options{
		Type:      true,
		NoFDoc:    true,
		Pre:       pre,
		Post:      post,
		Comment:   comment,
		Pkg:       pkg,
		MatchType: matchType,
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
	Get() (item string)
	Set(item string)
}
`, pkg)
	assert.Equal(t, expected, out.String())
}

func TestGenerator_Type_Entry(t *testing.T) {
	gen := New(Options{
		Type:    true,
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
	// AddData add data doc
	AddData(d ...originpkg.Data)
	// Add add data
	Add(d ...any)
	// Collate collate data
	Collate(in []*originpkg.Data) error
	// Scan scan func document
	Scan(in io.Reader) error
	// ScanMap scan map func
	ScanMap(in map[string]string) error
	// ScanMapMap scan map func
	ScanMapMap(in map[string]map[string]string) error
	// ScanMapSlice scan map func
	ScanMapSlice(in map[string][]string) error
	// ScanSlice scan slice func
	ScanSlice(in []string) error
	// ScanSliceMap scan slice map func
	ScanSliceMap(in []map[string]string) error
	// ScanSliceSlice scan slice map func
	ScanSliceSlice(in [][]interface{}) error
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
	err := gen.Generate(srcs, in, "src_ifaces.go", out)
	assert.NoError(t, err)
	assert.Equal(t, expected, out.String())
}

func TestGenerator_Type_Entry_NoTypeDoc(t *testing.T) {
	gen := New(Options{
		Type:    true,
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
	// AddData add data doc
	AddData(d ...originpkg.Data)
	// Add add data
	Add(d ...any)
	// Collate collate data
	Collate(in []*originpkg.Data) error
	// Scan scan func document
	Scan(in io.Reader) error
	// ScanMap scan map func
	ScanMap(in map[string]string) error
	// ScanMapMap scan map func
	ScanMapMap(in map[string]map[string]string) error
	// ScanMapSlice scan map func
	ScanMapSlice(in map[string][]string) error
	// ScanSlice scan slice func
	ScanSlice(in []string) error
	// ScanSliceMap scan slice map func
	ScanSliceMap(in []map[string]string) error
	// ScanSliceSlice scan slice map func
	ScanSliceSlice(in [][]interface{}) error
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

func TestGenerator_Type_Entry_NoFuncDoc(t *testing.T) {
	gen := New(Options{
		Type:    true,
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
	AddData(d ...originpkg.Data)
	Add(d ...any)
	Collate(in []*originpkg.Data) error
	Scan(in io.Reader) error
	ScanMap(in map[string]string) error
	ScanMapMap(in map[string]map[string]string) error
	ScanMapSlice(in map[string][]string) error
	ScanSlice(in []string) error
	ScanSliceMap(in []map[string]string) error
	ScanSliceSlice(in [][]interface{}) error
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

func TestGenerator_Recv(t *testing.T) {
	gen := New(Options{
		Comment: comment,
		Pkg:     pkg,
		Post:    `Iface`,
		Method:  true,
	})
	srcfile := `test_ifaces.go`
	srcs := []*Src{
		{
			File: srcfile,
			Line: 10,
			Src:  src2,
		},
	}
	expected := `// DO NOT EDIT

package mypkg

import "io"

type ServiceIface interface {
	Upload(rdr io.Reader) error
}
`
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	err := gen.Generate(srcs, in, srcfile, out)
	assert.NoError(t, err)
	assert.Equal(t, expected, out.String())

	gen = New(Options{
		Comment: comment,
		Pkg:     pkg,
		Post:    `Iface`,
		Method:  true,
	})
	srcfile = `test_ifaces.go`
	srcs = []*Src{
		{
			File: srcfile,
			Src:  src2,
			Line: 19,
		},
	}
	current := expected
	expected = `// DO NOT EDIT

package mypkg

import "io"

type ServiceIface interface {
	Upload(rdr io.Reader) error
	Status() int
}
`

	in = &bytes.Buffer{}
	out = &bytes.Buffer{}
	in.WriteString(current)
	err = gen.Generate(srcs, in, srcfile, out)
	assert.NoError(t, err)
	assert.Equal(t, expected, out.String())
}
