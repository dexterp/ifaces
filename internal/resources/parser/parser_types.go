package parser

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"

	"github.com/dexterp/ifaces/internal/resources/parsefunc"
	"github.com/dexterp/ifaces/internal/resources/typecheck"
	"github.com/dexterp/ifaces/internal/resources/types"
)

// Comment generator comment
type Comment struct {
	Text string // Text the actual comment
	Line int    // Start Start line
}

// Import
type Import struct {
	fset       *token.FileSet
	importSpec *ast.ImportSpec
}

func (i Import) Name() string {
	if i.importSpec.Name != nil {
		return i.importSpec.Name.Name
	}
	return ``
}

func (i Import) Path() string {
	if i.importSpec.Path != nil {
		return strings.Trim(i.importSpec.Path.Value, "\"")
	}
	return ``
}

// Type type declaration
type Type struct {
	doc  string
	line int
	name string
	typ  int
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

//go:generate ifaces type parser_ifaces.go --post Iface

// Method receiver or interface method
type Method struct {
	doc       string
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

func parseInterfaceMethod(fset *token.FileSet, ts *ast.TypeSpec, astField *ast.Field, hasType typecheck.HasType) *Method {
	return &Method{
		doc:       astField.Doc.Text(),
		line:      fset.Position(astField.Pos()).Line,
		name:      astField.Names[0].String(),
		signature: signature(fset, astField.Names[0].String(), astField.Type),
		typeName:  ts.Name.String(),
		hasType:   hasType,
	}
}

func parseType(fset *token.FileSet, astGenDecl *ast.GenDecl, astTypeSpec *ast.TypeSpec) *Type {
	return &Type{
		doc:  strings.TrimSuffix(astGenDecl.Doc.Text(), "\n"),
		line: fset.Position(astTypeSpec.Pos()).Line,
		name: strings.TrimSuffix(astTypeSpec.Name.String(), "\n"),
		typ:  parseTypeType(astTypeSpec),
	}
}

func parseTypeType(astTypeSpec *ast.TypeSpec) int {
	switch astTypeSpec.Type.(type) {
	case *ast.InterfaceType:
		return types.INTERFACE
	case *ast.StructType:
		return types.STRUCT
	}
	return types.UNKNOWN
}

func parseReceiverMethods(fset *token.FileSet, astFuncDecl *ast.FuncDecl, hasType typecheck.HasType) *Method {
	return &Method{
		doc:       strings.TrimSuffix(astFuncDecl.Doc.Text(), "\n"),
		line:      fset.Position(astFuncDecl.Pos()).Line,
		name:      astFuncDecl.Name.String(),
		signature: signature(fset, astFuncDecl.Name.String(), astFuncDecl.Type),
		typeName:  parseReceiverMethodsTypeName(*astFuncDecl),
		hasType:   hasType,
	}
}

func parseReceiverMethodsTypeName(astFuncDecl ast.FuncDecl) string {
	if len(astFuncDecl.Recv.List) != 1 {
		return ``
	}
	switch v := astFuncDecl.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if ident, ok := v.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return v.Name
	}
	return ``
}

func signature(fset *token.FileSet, funcName string, n ast.Node) string {
	buf := new(bytes.Buffer)
	printer.Fprint(buf, fset, n)
	sig := strings.Replace(buf.String(), `func`, funcName, 1)
	sig = strings.TrimSuffix(sig, "\n")
	return sig
}
