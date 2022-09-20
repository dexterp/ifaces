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
	"unicode"

	"github.com/dexterp/ifaces/internal/resources/match"
)

const (
	Unknwn = iota
	StructType
)

var (
	regexSignature    = regexp.MustCompile(`^(?P<name>[A-Za-z][A-Za-z0-9_]+)(?:\[(?P<generics>[^\]]*)\])?\((?P<params>[^\)]+)?\)\s*\(?(?P<return>[^(?:$|\)]*)\)?`)
	regexTypeString   = `[A-Za-z][A-Za-z0-9]*`
	regexType         = regexp.MustCompile(`^(\.\.\.)?(\*)?(` + `(?:[a-z]+\.)?` + regexTypeString + `(?:\{\})?)$`)
	regexTypeMapOrGen = regexp.MustCompile(`^(` + regexTypeString + `)\[(\*?)(` + regexTypeString + `)\](.*)$`)
	regexTypeSlice    = regexp.MustCompile(`^(\*)?\[\](.*)$`)
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

// Parset parse method
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

// GetRecvs GetRecvs returns all the function interfaces for
func (p *Parser) GetRecvs(typ string) (recvs []*Recv) {
	recvs = []*Recv{}
	for _, recv := range p.funcRecvs {
		sig := recv.Signature()
		if recv.TypeName() == typ && match.Capitalized(sig) {
			recvs = append(recvs, recv)
		}
	}
	return recvs
}

// GetType GetType returns the type at or after line. returns nil if the end of
// file is reached or it encounters a iface generator comment.
func (p Parser) GetType(line int) *Type {
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

// TypeByPattern use pattern to match types and return a list of type decleartions
func (p Parser) TypeByPattern(pattern string) []*Type {
	ts := []*Type{}
	for _, t := range p.typeDelcs {
		if match.Match(t.Name(), pattern) {
			ts = append(ts, t)
		}
	}
	return ts
}

// TypeByType return type
func (p *Parser) TypeByType(ifacetype int) []*Type {
	ts := []*Type{}
	for _, t := range p.typeDelcs {
		switch t.Type() {
		case StructType:
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
	case *ast.StructType:
		return StructType
	}
	return 0
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

// Doc returns function documentation
func (r Recv) Doc() string {
	return strings.TrimSuffix(r.funcDecl.Doc.Text(), "\n")
}

// Signature return the function signature
func (r Recv) Signature() string {
	sig := r.signature()
	f := toFuncDecl(r.pkg, r.hasType, sig)
	return f.String()
}

func (r Recv) signature() string {
	buf := new(bytes.Buffer)
	printer.Fprint(buf, r.fset, r.funcDecl.Type)
	sig := strings.Replace(buf.String(), `func`, r.funcDecl.Name.String(), 1)
	sig = strings.TrimSuffix(sig, "\n")
	return sig
}

// UsesGenerics returns true if function declaration contains generic types
func (r Recv) UsesGenerics() bool {
	sig := r.signature()
	f := toFuncDecl(r.pkg, r.hasType, sig)
	return f.UsesGenerics()
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

func toFuncDecl(pkg string, hastype hasType, funcdecl string) *funcDecl {
	name, generics, params, returns := parseSig(funcdecl)
	decl := &funcDecl{
		name:     name,
		generics: parseGenerics(pkg, hastype, generics),
		params:   parseParams(pkg, hastype, params),
		returns:  parseParams(pkg, hastype, returns),
	}
	return decl
}

func parseSig(sig string) (name, generics, params, ret string) {
	mtchs := regexSignature.FindStringSubmatch(sig)
	if mtchs == nil {
		return ``, ``, ``, ``
	}
	return mtchs[1], mtchs[2], mtchs[3], mtchs[4]
}

func parseParams(pkg string, hastype hasType, paramsstr string) (params []*param) {
	for _, p := range strings.Split(paramsstr, `,`) {
		p = strings.TrimSpace(p)
		m := strings.Split(p, ` `)
		switch len(m) {
		case 1:
			params = append(params, &param{
				hastype: hastype,
				pkg:     pkg,
				name:    m[0],
			})
		case 2:
			params = append(params, &param{
				hastype:  hastype,
				pkg:      pkg,
				name:     m[0],
				typ:      parseTyp(pkg, hastype, m[1]),
				typMap:   parseTypMap(pkg, hastype, m[1]),
				typSlice: parseTypSlice(pkg, hastype, m[1]),
			})
		}
	}
	return params
}

func parseGenerics(pkg string, hastype hasType, gen string) (generics []*genrics) {
	if gen == `` {
		return generics
	}
	for _, g := range strings.Split(gen, `,`) {
		var (
			name, typStr string
			alternates   []*typ
		)

		genrune := []rune(strings.TrimSpace(g))
		for i, ch := range genrune {
			if unicode.IsSpace(ch) {
				name = string(genrune[:i])
				if len(genrune) > i+1 {
					typStr = string(genrune[i+1:])
				}
				break
			}
		}
		for _, t := range strings.Split(typStr, `|`) {
			t = strings.TrimSpace(t)
			curType := parseTyp(pkg, hastype, t)
			if curType != nil {
				alternates = append(alternates, curType)
			}
		}
		generics = append(generics, &genrics{
			hastype: hastype,
			pkg:     pkg,
			name:    name,
			typ:     alternates,
		})
	}
	return generics
}

func parseTyp(pkg string, hastype hasType, t string) *typ {
	mtch := regexType.FindStringSubmatch(t)
	if mtch == nil {
		return nil
	}
	if mtch[2] == `interface{}` {
		return &typ{
			iface: true,
		}
	}
	return &typ{
		pkg:      pkg,
		hastype:  hastype,
		variadic: mtch[1],
		ptr:      mtch[2],
		typ:      mtch[3],
	}
}

func parseTypMap(pkg string, hastype hasType, m string) *typMap {
	mtch := regexTypeMapOrGen.FindStringSubmatch(m)
	if mtch == nil {
		return nil
	}
	if mtch[1] != `map` {
		return nil
	}
	return &typMap{
		ptr:      mtch[2],
		key:      mtch[3],
		elmType:  parseTyp(pkg, hastype, mtch[4]),
		elmMap:   parseTypMap(pkg, hastype, mtch[4]),
		elmSlice: parseTypSlice(pkg, hastype, mtch[4]),
	}
}

func parseTypSlice(pkg string, hastype hasType, s string) *typSlice {
	mtch := regexTypeSlice.FindStringSubmatch(s)
	if mtch == nil {
		return nil
	}
	return &typSlice{
		ptr:      mtch[1],
		elmType:  parseTyp(pkg, hastype, mtch[2]),
		elmMap:   parseTypMap(pkg, hastype, mtch[2]),
		elmSlice: parseTypSlice(pkg, hastype, mtch[2]),
	}
}

type funcDecl struct {
	name     string
	generics []*genrics
	params   []*param
	returns  []*param
}

func (f funcDecl) UsesGenerics() bool {
	return f.generics != nil
}

func (f funcDecl) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString(f.name)

	if len(f.generics) > 0 {
		items := []string{}
		for _, g := range f.generics {
			items = append(items, g.String())
		}
		buf.WriteString(`[` + strings.Join(items, `, `) + `]`)
	}

	if len(f.params) == 0 {
		buf.WriteString(`()`)
	} else {
		l := []string{}
		for _, p := range f.params {
			l = append(l, p.string())
		}
		buf.WriteString(`(` + strings.Join(l, `, `) + `)`)
	}

	n := len(f.returns)
	if n == 1 {
		if strings.Contains(f.returns[0].string(), " ") {
			buf.WriteString(` (` + f.returns[0].string() + `)`)
		} else {
			buf.WriteString(` ` + f.returns[0].string())
		}
	} else if n > 0 {
		l := []string{}
		for _, p := range f.returns {
			l = append(l, p.string())
		}
		buf.WriteString(` (` + strings.Join(l, `, `) + `)`)
	}
	return buf.String()
}

type param struct {
	hastype  hasType
	pkg      string
	name     string
	typ      *typ
	typSlice *typSlice
	typMap   *typMap
}

func (p param) string() string {
	okName := p.name != ``
	okType := p.typ != nil || p.typMap != nil || p.typSlice != nil
	switch {
	case okName && okType:
		return p.name + " " + p.Type()
	case okName:
		return p.name
	case okType:
		return p.Type()
	}
	return ``
}

func (p param) Type() string {
	switch {
	case p.typ != nil:
		return p.typ.string()
	case p.typMap != nil:
		return p.typMap.string()
	case p.typSlice != nil:
		return p.typSlice.string()
	}
	return ``
}

type genrics struct {
	pkg     string
	hastype hasType
	name    string
	typ     []*typ
}

func (g genrics) Types() (types []string) {
	if g.typ == nil {
		return
	}
	for _, t := range g.typ {
		types = append(types, t.string())
	}
	return types
}

func (g genrics) String() string {
	return g.name + ` ` + strings.Join(g.Types(), ` | `)
}

type hasType interface {
	HasType(typ string) bool
}

type typ struct {
	pkg      string // original package name
	hastype  hasType
	variadic string // variadic paramater
	ptr      string // asterisk or empty
	typ      string // type name
	iface    bool
}

func (t typ) string() string {
	if t.iface {
		return `interface{}`
	}
	if t.pkg != `` && !strings.Contains(t.typ, `.`) && t.hastype.HasType(t.typ) {
		return t.variadic + t.ptr + t.pkg + `.` + t.typ
	}
	return t.variadic + t.ptr + t.typ
}

type typMap struct {
	ptr      string
	key      string
	elmType  *typ
	elmMap   *typMap
	elmSlice *typSlice
}

func (tm typMap) string() string {
	buf := &bytes.Buffer{}
	buf.WriteString(`map[` + tm.ptr + tm.key + `]`)
	switch {
	case tm.elmType != nil:
		buf.WriteString(tm.elmType.string())
	case tm.elmMap != nil:
		buf.WriteString(tm.elmMap.string())
	case tm.elmSlice != nil:
		buf.WriteString(tm.elmSlice.string())
	}
	return buf.String()
}

type typSlice struct {
	ptr      string
	elmType  *typ
	elmMap   *typMap
	elmSlice *typSlice
}

func (ts typSlice) string() string {
	buf := &bytes.Buffer{}
	buf.WriteString(ts.ptr + `[]`)
	switch {
	case ts.elmType != nil:
		buf.WriteString(ts.elmType.string())
	case ts.elmMap != nil:
		buf.WriteString(ts.elmMap.string())
	case ts.elmSlice != nil:
		buf.WriteString(ts.elmSlice.string())
	}
	return buf.String()
}
