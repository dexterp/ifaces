package parser

import (
	"bufio"
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"

	"github.com/dexterp/ifaces/internal/resources/typecheck"
)

// Parser parse method
type Parser struct {
	astFile          *ast.File
	fset             *token.FileSet
	comments         []*Comment
	ReceiverMethods  *[]MethodIface
	InterfaceMethods *[]MethodIface
	Types            *[]*Type
}

// Parse parse source were path is path to source and src is an array of bytes,
// string of the sources or nil
func Parse(path string, src any) (*Parser, error) {
	srcByte, err := readSource(path, src)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	p := &Parser{
		fset:             fset,
		astFile:          node,
		comments:         []*Comment{},
		InterfaceMethods: &[]MethodIface{},
		ReceiverMethods:  &[]MethodIface{},
		Types:            &[]*Type{},
	}
	err = p.parseComments(srcByte)
	if err != nil {
		return nil, err
	}
	p.parseAstFile()
	return p, nil
}

// parseComments
func (p *Parser) parseComments(src *[]byte) error {
	re := regexp.MustCompile(`^//go:generate.*ifaces\W`)
	buf := bytes.NewBuffer(*src)
	scanner := bufio.NewScanner(buf)
	var line int
	for scanner.Scan() {
		line++
		b := scanner.Bytes()
		if re.Match(b) {
			p.comments = append(p.comments, &Comment{
				Text: string(b),
				Line: line,
			})
		}
	}
	return nil
}

func (p *Parser) parseAstFile() {
	ast.Inspect(p.astFile, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncDecl:
			if v.Recv != nil {
				*p.ReceiverMethods = append(*p.ReceiverMethods, parseReceiverMethods(p.fset, v, p.HasTypeCheck()))
			}
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if ok {
					p.parseAstTypeSpec(ts)
					*p.Types = append(*p.Types, parseType(p.fset, v, ts))
				}
			}
		}
		return true
	})
}

func (p *Parser) parseAstTypeSpec(ts *ast.TypeSpec) {
	switch v := ts.Type.(type) {
	case *ast.InterfaceType:
		for _, astField := range v.Methods.List {
			*p.InterfaceMethods = append(*p.InterfaceMethods, parseInterfaceMethod(p.fset, ts, astField, p.HasTypeCheck()))
		}
	}
}

// HasTypeCheck returns a function to check if a type exists in the parsed source.
func (p Parser) HasTypeCheck() typecheck.HasType {
	return func(typ string) (found bool) {
		for _, t := range *p.Types {
			if t.Name() == typ {
				found = true
			}
		}
		return found
	}
}

// Imports list of imports
func (p Parser) Imports() (imports []*Import) {
	imports = []*Import{}
	for _, spec := range p.astFile.Imports {
		if spec.Name != nil && spec.Name.String() == `_` {
			continue
		}
		imports = append(imports, &Import{
			fset:       p.fset,
			importSpec: spec,
		})
	}
	return imports
}

// Package package file
func (p Parser) Package() string {
	return p.astFile.Name.String()
}

// If src != nil, readSource converts src to a []byte if possible;
// otherwise it returns an error. If src == nil, readSource returns
// the result of reading the file specified by filename.
func readSource(filename string, src any) (*[]byte, error) {
	var b []byte
	if src != nil {
		switch s := src.(type) {
		case string:
			b = []byte(s)
			return &b, nil
		case []byte:
			return &s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				b = s.Bytes()
				return &b, nil
			}
		}
		return nil, errors.New("invalid source")
	}
	b, err := os.ReadFile(filename)
	return &b, err
}
