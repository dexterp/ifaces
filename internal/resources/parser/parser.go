package parser

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dexterp/ifaces/internal/resources/source"
	"github.com/dexterp/ifaces/internal/resources/typecheck"
	"github.com/dexterp/ifaces/internal/resources/types"
)

var reComments = regexp.MustCompile(`^//go:generate .*ifaces\W`)

// ParserFiles parse files
type Parser struct {
	pkg          *string
	files        *[]*file
	comments     *[]*Comment
	recvMethods  *[]*Method
	ifaceMethods *[]*Method
	types        *[]*Type
	imports      *[]*Import
}

// Parse parse source were path is path to source and src is an array of bytes,
// string of the sources or nil
func Parse(path string, src any, line int) (*Parser, error) {
	s := ``
	p := &Parser{
		pkg:          &s,
		files:        &[]*file{},
		comments:     &[]*Comment{},
		ifaceMethods: &[]*Method{},
		recvMethods:  &[]*Method{},
		types:        &[]*Type{},
		imports:      &[]*Import{},
	}
	err := parseFile(p.pkg, p.files, p.imports, p.comments, p.types, p.recvMethods, p.ifaceMethods, path, src, line)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ParseFiles(srcs []*source.Source) (p *Parser, first error) {
	s := ``
	p = &Parser{
		pkg:          &s,
		files:        &[]*file{},
		comments:     &[]*Comment{},
		recvMethods:  &[]*Method{},
		ifaceMethods: &[]*Method{},
		types:        &[]*Type{},
		imports:      &[]*Import{},
	}
	for _, src := range srcs {
		first = parseFile(p.pkg, p.files, p.imports, p.comments, p.types, p.recvMethods, p.ifaceMethods, src.File, src.Src, src.Line)
		if first != nil {
			return nil, first
		}
	}
	return p, nil
}

// Imports list of imports
func (p Parser) Imports() (imports []*Import) {
	return *p.imports
}

// InterfaceMethods
func (p Parser) InterfaceMethods() []*Method {
	return *p.ifaceMethods
}

// Package package file
func (p Parser) Package() string {
	return *p.pkg
}

// Comments
func (p *Parser) Comments() []*Comment {
	return *p.comments
}

// Query get query struct
func (p *Parser) Query() *Query {
	return &Query{Parser: p}
}

// ReceiverMethods return a list of parsed interface methods
func (p *Parser) ReceiverMethods() []*Method {
	return *p.recvMethods
}

// Types return a list of parsed types
func (p *Parser) Types() []*Type {
	return *p.types
}

// hasTypeCheck returns a function to check if a type exists in the parsed source.
func hasTypeCheck(types *[]*Type) typecheck.HasType {
	return func(typ string) (found bool) {
		for _, t := range *types {
			if t.Name() == typ {
				found = true
			}
		}
		return found
	}
}

func parseAstFile(types *[]*Type, recvMethods *[]*Method, ifaceMethods *[]*Method, fset *token.FileSet, astFile *ast.File, file string) {
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncDecl:
			parseAstFuncDecl(recvMethods, fset, v, file, hasTypeCheck(types))
		case *ast.GenDecl:
			parseAstGenDecl(types, ifaceMethods, fset, v, file, hasTypeCheck(types))
		}
		return true
	})
}

func parseAstFuncDecl(recvMethods *[]*Method, fset *token.FileSet, astFuncDecl *ast.FuncDecl, file string, hasType typecheck.HasType) {
	parseReceiverMethods(recvMethods, fset, astFuncDecl, file, hasType)
}

func parseAstGenDecl(types *[]*Type, ifaceMethods *[]*Method, fset *token.FileSet, astGenDecl *ast.GenDecl, file string, hasType typecheck.HasType) {
	for _, astSpec := range astGenDecl.Specs {
		ts, ok := astSpec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		parseType(types, fset, astGenDecl, ts, file)
		switch v := ts.Type.(type) {
		case *ast.InterfaceType:
			for _, astField := range v.Methods.List {
				parseInterfaceMethod(ifaceMethods, fset, ts, astField, file, hasType)
			}
		}
	}
}

// parseComments
func parseComments(comments *[]*Comment, fset *token.FileSet, cgs []*ast.CommentGroup, file string) {
	for _, cg := range cgs {
		for _, c := range cg.List {
			if !reComments.MatchString(c.Text) {
				continue
			}
			*comments = append(*comments, &Comment{
				File: file,
				Line: fset.Position(c.Pos()).Line,
				Text: c.Text,
			})
		}
	}
}

func parseFile(pkg *string, files *[]*file, imports *[]*Import, comments *[]*Comment, types *[]*Type, recvMethods *[]*Method, ifaceMethods *[]*Method, path string, src any, line int) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return err
	}
	curFile := &file{
		fset: fset,
		file: f,
	}
	*files = append(*files, curFile)
	parseAstFile(types, recvMethods, ifaceMethods, fset, f, path)
	parseComments(comments, curFile.fset, curFile.file.Comments, path)
	parseImports(imports, f.Imports, path)
	if *pkg == `` {
		*pkg = f.Name.String()
	}
	return nil
}

func parseImports(imports *[]*Import, imp []*ast.ImportSpec, file string) {
	for _, i := range imp {
		name := ``
		path := ``
		if i.Name != nil {
			name = i.Name.String()
		}
		if i.Path != nil {
			path = strings.Trim(i.Path.Value, `"`)
		}
		*imports = append(*imports, &Import{
			file: file,
			name: name,
			path: path,
		})
	}
}

func parseInterfaceMethod(methods *[]*Method, fset *token.FileSet, ts *ast.TypeSpec, astField *ast.Field, file string, hasType typecheck.HasType) {
	*methods = append(*methods, &Method{
		doc:       astField.Doc.Text(),
		file:      filepath.Base(file),
		line:      fset.Position(astField.Pos()).Line,
		name:      astField.Names[0].String(),
		signature: signature(fset, astField.Names[0].String(), astField.Type),
		typeName:  ts.Name.String(),
		hasType:   hasType,
	})
}

func parseReceiverMethods(recvMethods *[]*Method, fset *token.FileSet, astFuncDecl *ast.FuncDecl, file string, hasType typecheck.HasType) {
	if astFuncDecl.Recv == nil {
		return
	}
	*recvMethods = append(*recvMethods, &Method{
		doc:       strings.TrimSuffix(astFuncDecl.Doc.Text(), "\n"),
		file:      filepath.Base(file),
		line:      fset.Position(astFuncDecl.Pos()).Line,
		name:      astFuncDecl.Name.String(),
		signature: signature(fset, astFuncDecl.Name.String(), astFuncDecl.Type),
		typeName:  parseReceiverMethodsTypeName(*astFuncDecl),
		hasType:   hasType,
	})
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

func parseType(types *[]*Type, fset *token.FileSet, astGenDecl *ast.GenDecl, astTypeSpec *ast.TypeSpec, file string) {
	*types = append(*types, &Type{
		doc:  strings.TrimSuffix(astGenDecl.Doc.Text(), "\n"),
		file: filepath.Base(file),
		line: fset.Position(astTypeSpec.Pos()).Line,
		name: strings.TrimSuffix(astTypeSpec.Name.String(), "\n"),
		typ:  parseTypeType(astTypeSpec),
	})
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

func signature(fset *token.FileSet, funcName string, n ast.Node) string {
	buf := new(bytes.Buffer)
	printer.Fprint(buf, fset, n)
	sig := strings.Replace(buf.String(), `func`, funcName, 1)
	sig = strings.TrimSuffix(sig, "\n")
	return sig
}

type file struct {
	fset *token.FileSet
	file *ast.File
}
