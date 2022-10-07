package parser

import (
	"github.com/dexterp/ifaces/internal/resources/parser/parsefunc"
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
	name string
	path string
}

func (i Import) Name() string {
	return i.name
}

func (i Import) Path() string {
	return i.path
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
	Doc       string
	File      string // File originating file
	Line      int
	Name      string
	Prefixes  []string
	Pkg       string
	signature string
	TypeName  string
	hasType   typecheck.HasType
}

// Signature return the function signature
func (i Method) Signature() string {
	f := parsefunc.ToFuncDecl(i.signature, i.Pkg, i.hasType)
	return f.String()
}

// NeedsImport
func (i Method) NeedsImport() bool {
	f := parsefunc.ToFuncDecl(i.signature, i.Pkg, i.hasType)
	return f.NeedsImport
}

// ImportPrefixes
func (i Method) ImportPrefixes() (prefixes []Prefix) {
	f := parsefunc.ToFuncDecl(i.signature, i.Pkg, i.hasType)
	for p := range f.Prefixes {
		prefixes = append(prefixes, Prefix{
			Text: p,
			File: i.File,
		})
	}
	return
}

// UsesTypeParams returns true if function declaration contains type parmeters
func (i Method) UsesTypeParams() bool {
	f := parsefunc.ToFuncDecl(i.signature, i.Pkg, i.hasType)
	return f.UsesTypeParams()
}
