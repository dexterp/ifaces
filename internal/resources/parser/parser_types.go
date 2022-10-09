package parser

import (
	"github.com/dexterp/ifaces/internal/resources/typecheck"
)

// Comment generator comment
type Comment struct {
	File string // Filename
	Text string // Text the actual comment
	Line int    // Start Start line
}

// Import
type Import struct {
	File string
	Name string
	Path string
}

// Prefix
type Prefix struct {
	Text string
	File string
}

// Type type declaration
type Type struct {
	Doc  string
	File string // File originating file
	Line int
	Name string
	Type int
}

// Method receiver or interface method
type Method struct {
	Doc      string
	File     string // File originating file
	fn       *Func
	Line     int
	Name     string
	Prefixes []string
	Pkg      string
	TypeName string
	HasType  typecheck.HasType
}

// Signature return the function signature
func (i Method) Signature() string {
	s := i.fn.Package(i.Pkg).String()
	return s
}

// NeedsImport
func (i Method) NeedsImport() bool {
	//f := parsefunc.ToFuncDecl(i.Sig, i.Pkg, i.HasType)
	//return f.NeedsImport
	// TODO - Implement or refactor this method
	return false
}

// ImportPrefixes
func (i Method) ImportPrefixes() (prefixes []Prefix) {
	for p := range i.fn.Prefixes {
		prefixes = append(prefixes, Prefix{
			Text: p,
			File: i.File,
		})
	}
	return
}
