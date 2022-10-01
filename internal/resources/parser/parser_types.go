package parser

import (
	"github.com/dexterp/ifaces/internal/resources/parsefunc"
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
	file string
	name string
	path string
}

func (i Import) Name() string {
	return i.name
}

func (i Import) Path() string {
	return i.path
}

// Type type declaration
type Type struct {
	doc  string
	file string
	line int
	name string
	typ  int
}

// File originating file
func (r Type) File() string {
	return r.file
}

// Line line number
func (r Type) Line() int {
	return r.line
}

// Doc type documentation
func (r Type) Doc() string {
	return r.doc
}

// Name type name
func (r Type) Name() string {
	return r.name
}

// Type get type of
func (r Type) Type() int {
	return r.typ
}

// Method receiver or interface method
type Method struct {
	doc       string
	file      string
	line      int
	name      string
	pkg       string
	signature string
	typeName  string
	hasType   typecheck.HasType
}

// SetPkg set the prefix when exporting to a new package. E.G. MyType will be
// converted to pkg.MyType.
func (i *Method) SetPkg(pkg string) {
	i.pkg = pkg
}

// File originating file
func (i Method) File() string {
	return i.file
}

// Line return line number in source code
func (i Method) Line() int {
	return i.line
}

// TypeName return the type name or receiver name for this method
func (i Method) TypeName() string {
	return i.typeName
}

// Name method name
func (i Method) Name() string {
	return i.name
}

// Doc method documentation
func (i Method) Doc() string {
	return i.doc
}

// Signature return the function signature
func (i Method) Signature() string {
	f := parsefunc.ToFuncDecl(i.pkg, i.hasType, i.signature)
	return f.String()
}

// UsesTypeParams returns true if function declaration contains type parameters
func (i Method) UsesTypeParams() bool {
	f := parsefunc.ToFuncDecl(i.pkg, i.hasType, i.signature)
	return f.UsesTypeParams()
}
