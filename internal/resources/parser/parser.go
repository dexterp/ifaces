package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dexterp/ifaces/internal/resources/cond"
	"github.com/dexterp/ifaces/internal/resources/srcio"
	"github.com/dexterp/ifaces/internal/resources/typecheck"
	"github.com/dexterp/ifaces/internal/resources/types"
)

var reComments = regexp.MustCompile(`^//go:generate .*ifaces\W`)

// Parser represents a source file with the extracted package name, comments
// matching go:generate statements, interface methods, receiver methods and type
// definitions.
type Parser struct {
	// Package package name
	Package string

	// Comment comments which match the prefix "//go:generate ifaces"
	Comments []Comment

	// InterfaceMethods interface methods
	InterfaceMethods []*Method

	// Imports imports
	Imports []*Import

	// ReceiverMethods receiver methods
	ReceiverMethods []*Method

	// Types types within this source
	Types []Type
}

// Parse parses an individual file represented by path and src. src is the
// contents of the go source file which can be a, string, byte array or
// io.Reader. If src is nil then the source is read from file. Line represents
// the line number in the file where the go:generate comment is located
// otherwise it should be set to 0 or less.
func Parse(file string, src any, line int) (*Parser, error) {
	p := &parse{}
	err := p.parse(file, src, line)
	if err != nil {
		return nil, err
	}
	return &Parser{
		Package:          p.pkg,
		Comments:         p.comments,
		InterfaceMethods: p.ifaceMethods,
		Imports:          p.imports,
		ReceiverMethods:  p.recvMethods,
		Types:            p.types,
	}, nil
}

// ParseFiles creates the same output as Parse but parses multiple files.
func ParseFiles(srcs []srcio.Source) (*Parser, error) {
	p := &parse{}
	for _, src := range srcs {
		first := p.parse(src.File, src.Src, src.Line)
		if first != nil {
			return nil, first
		}
	}
	return &Parser{
		Package:          p.pkg,
		Comments:         p.comments,
		InterfaceMethods: p.ifaceMethods,
		Imports:          p.imports,
		ReceiverMethods:  p.recvMethods,
		Types:            p.types,
	}, nil
}

type parse struct {
	pkg          string
	comments     []Comment
	ifaceMethods []*Method
	imports      []*Import
	recvMethods  []*Method
	types        []Type
}

// hasTypeCheck returns a function to check if a type exists in the parsed source.
func (p *parse) hasTypeCheck() typecheck.HasType {
	return func(typ string) (found bool) {
		for _, t := range p.types {
			if t.Name == typ {
				found = true
			}
		}
		return found
	}
}

func (p *parse) parse(path string, src any, line int) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return err
	}
	p.parseAstFile(fset, f, path)
	p.parseComments(fset, f.Comments, path)
	p.parseImports(f.Imports, path)
	p.pkg = cond.First(p.pkg, f.Name.String()).(string)
	return nil
}

func (p *parse) parseAstFile(fset *token.FileSet, astFile *ast.File, file string) {
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncDecl:
			p.parseAstFuncDecl(fset, v, file)
		case *ast.GenDecl:
			p.parseAstGenDecl(fset, v, file)
		}
		return true
	})
}

func (p *parse) parseAstFuncDecl(fset *token.FileSet, astFuncDecl *ast.FuncDecl, file string) {
	p.parseReceiverMethods(fset, astFuncDecl, file)
}

func (p *parse) parseAstGenDecl(fset *token.FileSet, astGenDecl *ast.GenDecl, file string) {
	for _, astSpec := range astGenDecl.Specs {
		ts, ok := astSpec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		p.parseType(fset, astGenDecl, ts, file)
		switch v := ts.Type.(type) {
		case *ast.InterfaceType:
			for _, astField := range v.Methods.List {
				p.parseInterfaceMethod(fset, ts, astField, file)
			}
		}
	}
}

// parseComments
func (p *parse) parseComments(fset *token.FileSet, cgs []*ast.CommentGroup, file string) {
	for _, cg := range cgs {
		for _, c := range cg.List {
			if !reComments.MatchString(c.Text) {
				continue
			}
			p.comments = append(p.comments, Comment{
				File: file,
				Line: fset.Position(c.Pos()).Line,
				Text: c.Text,
			})
		}
	}
}

func (p *parse) parseImports(imp []*ast.ImportSpec, file string) {
	for _, i := range imp {
		name := ``
		path := ``
		if i.Name != nil {
			name = i.Name.String()
		}
		if i.Path != nil {
			path = strings.Trim(i.Path.Value, `"`)
		}
		p.imports = append(p.imports, &Import{
			File: file,
			Name: name,
			Path: path,
		})
	}
}

func (p *parse) parseInterfaceMethod(fset *token.FileSet, ts *ast.TypeSpec, astField *ast.Field, file string) {
	fn := IfaceToFunc(astField, p.hasTypeCheck())
	p.ifaceMethods = append(p.ifaceMethods, &Method{
		Doc:      astField.Doc.Text(),
		File:     filepath.Base(file),
		Line:     fset.Position(astField.Pos()).Line,
		Name:     astField.Names[0].String(),
		fn:       fn,
		TypeName: ts.Name.String(),
		HasType:  p.hasTypeCheck(),
	})
}

func (p *parse) parseReceiverMethods(fset *token.FileSet, astFuncDecl *ast.FuncDecl, file string) {
	if astFuncDecl.Recv == nil {
		return
	}
	fn := RecvToFunc(astFuncDecl, p.hasTypeCheck())
	p.recvMethods = append(p.recvMethods, &Method{
		Doc:      strings.TrimSuffix(astFuncDecl.Doc.Text(), "\n"),
		File:     filepath.Base(file),
		Line:     fset.Position(astFuncDecl.Pos()).Line,
		Name:     astFuncDecl.Name.String(),
		Prefixes: parseSigPrefixes(fn),
		fn:       fn,
		TypeName: parseReceiverMethodsTypeName(*astFuncDecl),
		HasType:  p.hasTypeCheck(),
	})
}

func parseSigPrefixes(fn *Func) (prefixes []string) {
	for p := range fn.Prefixes {
		prefixes = append(prefixes, p)
	}
	return
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

func (p *parse) parseType(fset *token.FileSet, astGenDecl *ast.GenDecl, astTypeSpec *ast.TypeSpec, file string) {
	p.types = append(p.types, Type{
		Doc:  strings.TrimSuffix(astGenDecl.Doc.Text(), "\n"),
		File: filepath.Base(file),
		Line: fset.Position(astTypeSpec.Pos()).Line,
		Name: strings.TrimSuffix(astTypeSpec.Name.String(), "\n"),
		Type: parseTypeType(astTypeSpec),
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
