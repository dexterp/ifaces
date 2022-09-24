package parser

import (
	"bufio"
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/dexterp/ifaces/internal/resources/match"
	"github.com/dexterp/ifaces/internal/resources/parsefunc"
	"github.com/dexterp/ifaces/internal/resources/types"
)

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
		path:      path,
		src:       srcByte,
		fset:      fset,
		astFile:   node,
		genCmts:   []*GeneratorCmt{},
		funcRecvs: []*Recv{},
		typeDelcs: []*Type{},
	}
	err = p.parse()
	if err != nil {
		return nil, err
	}
	return p, nil
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
		case io.Reader:
			b, err := io.ReadAll(s)
			return &b, err
		case func() (*[]byte, error):
			return s()
		}
		return nil, errors.New("invalid source")
	}
	b, err := os.ReadFile(filename)
	return &b, err
}

// Parser parse method
type Parser struct {
	astFile   *ast.File
	fset      *token.FileSet
	path      string
	src       *[]byte
	genCmts   []*GeneratorCmt
	funcRecvs []*Recv
	typeDelcs []*Type
}

// parse
func (p *Parser) parse() error {
	err := p.parseGeneratorCmts()
	if err != nil {
		return err
	}
	p.parseAstFile()
	return nil
}

// parseGeneratorCmts
func (p *Parser) parseGeneratorCmts() error {
	re := regexp.MustCompile(`^//go:generate\s*ifaces\W`)
	buf := *p.src
	var line int
	for {
		advance, token, err := bufio.ScanLines(buf, true)
		if err != nil && err != io.EOF {
			return err
		}
		if advance == 0 {
			break
		}
		line++
		if re.Match(token) {
			p.genCmts = append(p.genCmts, &GeneratorCmt{
				Text: string(token),
				Line: line,
			})
		}
		if advance <= len(buf) {
			buf = buf[advance:]
		}
	}
	return nil
}

// parseAstFile
func (p *Parser) parseAstFile() {
	ast.Inspect(p.astFile, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncDecl:
			if v.Recv != nil {
				p.funcRecvs = append(p.funcRecvs, &Recv{
					hasType:  p,
					funcDecl: v,
					fset:     p.fset,
					src:      p.src,
				})
			}
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if ok {
					p.typeDelcs = append(p.typeDelcs, &Type{
						fset:     p.fset,
						genDecl:  v,
						typeSpec: ts,
					})
				}
			}
		}
		return true
	})
}

// GetIfaceMethods returns all the methods of an interface
func (p *Parser) GetIfaceMethods(iface string) (methods []*IfaceMethod) {
	for _, d := range p.typeDelcs {
		t, ok := d.typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			continue
		}

		name := d.typeSpec.Name.String()
		if name != iface {
			continue
		}

		for _, astField := range t.Methods.List {
			methods = append(methods, &IfaceMethod{
				typeName: name,
				pkg:      p.Package(),
				hasType:  p,
				fset:     p.fset,
				astField: astField,
			})
		}
	}
	return
}

// GetTypeRecvs returns all the function interfaces for
func (p *Parser) GetTypeRecvs(typ string) (recvs []*Recv) {
	for _, recv := range p.funcRecvs {
		sig := recv.Signature()
		if recv.TypeName() == typ && match.Capitalized(sig) {
			recvs = append(recvs, recv)
		}
	}
	return recvs
}

// GetTypeByLine GetTypeByLine returns the type at or after line. returns nil if the end of
// file is reached or it encounters a iface generator comment.
func (p Parser) GetTypeByLine(line int) *Type {
	var end int
	for _, c := range p.genCmts {
		if c.Line > line {
			end = c.Line
		}
	}
	for _, t := range p.typeDelcs {
		if end == 0 && t.Line() >= line {
			return t
		} else if t.Line() >= line && t.Line() < end {
			return t
		}
	}
	return nil
}

// HasType check if parsed source has type
func (p Parser) HasType(typ string) bool {
	for _, t := range p.typeDelcs {
		if typ == t.Name() {
			return true
		}
	}
	return false
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

// GetTypeByPattern use pattern to match types and return a list of type declarations
func (p Parser) GetTypeByPattern(pattern string) (ts []*Type) {
	for _, t := range p.typeDelcs {
		if match.Match(t.Name(), pattern) {
			ts = append(ts, t)
		}
	}
	return ts
}

// GetTypesByType return type
func (p *Parser) GetTypesByType(typ int) (ts []*Type) {
	for _, t := range p.typeDelcs {
		if t.Type() == typ {
			ts = append(ts, t)
		}
	}
	return ts
}

// GeneratorCmt
type GeneratorCmt struct {
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
	fset     *token.FileSet
	genDecl  *ast.GenDecl
	typeSpec *ast.TypeSpec
}

// Line line number
func (r Type) Line() int {
	return r.fset.Position(r.typeSpec.Pos()).Line
}

// Doc type documentation
func (r Type) Doc() string {
	return strings.TrimSuffix(r.genDecl.Doc.Text(), "\n")
}

// Name type name
func (r Type) Name() string {
	return strings.TrimSuffix(r.typeSpec.Name.Name, "\n")
}

// Type get type of
func (r Type) Type() int {
	switch r.typeSpec.Type.(type) {
	case *ast.InterfaceType:
		return types.INTERFACE
	case *ast.StructType:
		return types.STRUCT
	}
	return types.UNKNOWN
}

type IfaceMethod struct {
	typeName string
	pkg      string
	hasType  hasType
	fset     *token.FileSet
	astField *ast.Field
}

func (i *IfaceMethod) SetPkg(pkg string) {
	i.pkg = pkg
}

func (i IfaceMethod) Line() int {
	return i.fset.Position(i.astField.Pos()).Line
}

func (i IfaceMethod) TypeName() string {
	return i.typeName
}

// Name method name
func (i IfaceMethod) Name() string {
	return i.astField.Names[0].String()
}

// Doc returns function documentation
func (i IfaceMethod) Doc() string {
	return i.astField.Doc.Text()
}

// Signature return the function signature
func (i IfaceMethod) Signature() string {
	sig := signature(i.fset, i.astField.Names[0].String(), i.astField.Type)
	f := parsefunc.ToFuncDecl(i.pkg, i.hasType, sig)
	return f.String()
}

// UsesTypeParams returns true if function declaration contains type parameters
func (i IfaceMethod) UsesTypeParams() bool {
	sig := signature(i.fset, i.astField.Names[0].String(), i.astField.Type)
	f := parsefunc.ToFuncDecl(i.pkg, i.hasType, sig)
	return f.UsesTypeParams()
}

// Recv function receiver
type Recv struct {
	pkg      string
	hasType  hasType
	fset     *token.FileSet
	funcDecl *ast.FuncDecl
	src      *[]byte
}

// SetPkg set the package
func (r *Recv) SetPkg(pkg string) {
	r.pkg = pkg
}

// Line line number
func (r Recv) Line() int {
	return r.fset.Position(r.funcDecl.Pos()).Line
}

// TypeName typename
func (r Recv) TypeName() string {
	return strings.TrimSuffix(getReceiverTypeName(r.funcDecl, r.src), "\n")
}

// Name function name
func (r Recv) Name() string {
	return r.funcDecl.Name.String()
}

// Doc returns function documentation
func (r Recv) Doc() string {
	return strings.TrimSuffix(r.funcDecl.Doc.Text(), "\n")
}

// Signature return the function signature
func (r Recv) Signature() string {
	sig := signature(r.fset, r.funcDecl.Name.String(), r.funcDecl.Type)
	f := parsefunc.ToFuncDecl(r.pkg, r.hasType, sig)
	return f.String()
}

func signature(fset *token.FileSet, funcName string, n ast.Node) string {
	buf := new(bytes.Buffer)
	printer.Fprint(buf, fset, n)
	sig := strings.Replace(buf.String(), `func`, funcName, 1)
	sig = strings.TrimSuffix(sig, "\n")
	return sig
}

// UsesTypeParams returns true if function declaration contains type parameters
func (r Recv) UsesTypeParams() bool {
	sig := signature(r.fset, r.funcDecl.Name.String(), r.funcDecl.Type)
	f := parsefunc.ToFuncDecl(r.pkg, r.hasType, sig)
	return f.UsesTypeParams()
}

func getReceiverTypeName(fd *ast.FuncDecl, src *[]byte) string {
	rev := fd.Recv
	if rev == nil {
		return ``
	}
	typ := fd.Recv.List[0].Type
	return typeName(typ, *src)
}

func typeName(n ast.Node, src []byte) string {
	stype := string(src[n.Pos()-1 : n.End()-1])
	if len(stype) > 0 && stype[0] == '*' {
		stype = stype[1:]
	}
	return stype
}

type hasType interface {
	HasType(typ string) bool
}
