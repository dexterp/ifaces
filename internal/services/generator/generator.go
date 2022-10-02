package generator

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"text/template"

	"github.com/dexterp/ifaces/internal/resources/addimports"
	"github.com/dexterp/ifaces/internal/resources/match"
	"github.com/dexterp/ifaces/internal/resources/parser"
	"github.com/dexterp/ifaces/internal/resources/pathtoimport"
	"github.com/dexterp/ifaces/internal/resources/print"
	"github.com/dexterp/ifaces/internal/resources/source"
	"github.com/dexterp/ifaces/internal/resources/srcformat"
	"github.com/dexterp/ifaces/internal/resources/tdata"
	"github.com/dexterp/ifaces/internal/resources/types"
)

//go:generate ifaces type generator_iface.go --post Iface

// Generator interface generator
type Generator struct {
	Type      bool             // Type type subcommand
	Method    bool             // Method method sub command
	Comment   string           // Comment comment at the top of the file
	Iface     string           // Iface explicitly set interface name
	Module    string           // Module name of module to scan instead of scanning the file system
	NoFDoc    bool             // NoFDoc omit copying function documentation
	NoTDoc    bool             // NoTDoc omit copying type documentation
	Pkg       string           // Pkg package name
	Post      string           // Post postfix to interface name
	Pre       string           // Pre prefix to interface name
	Print     print.PrintIface // Print handler
	Struct    bool             // Struct generate an interface for all structs
	TDoc      string           // TDoc type document
	MatchType string           // MatchType match types
	MatchFunc string           // MatchFunc match receivers
}

//go:embed generate.gotmpl
var gentmpl string

var ErrorNoSourceFile = errors.New(`no source files processed`)

// Generate generate interfaces source code for the gen sub command.
func (g Generator) Generate(srcs []*source.Source, current *bytes.Buffer, outfile string, output io.Writer) error {
	pkg, err := setOutputPackage(g.Pkg, outfile)
	if err != nil {
		return err
	}
	data, importsList, err := g.parse(srcs, current, outfile, pkg)
	if err != nil {
		return err
	}
	templateOut := &bytes.Buffer{}
	err = applyTemplate(templateOut, data)
	if err != nil {
		return err
	}
	importsOut := &bytes.Buffer{}
	err = addimports.AddImports(outfile, templateOut, importsList, importsOut)
	if err != nil {
		return err
	}
	finalSrc := &bytes.Buffer{}
	err = srcformat.Format(outfile, importsOut, finalSrc)
	if err != nil {
		return err
	}
	_, err = io.Copy(output, finalSrc)
	return err
}

func (g Generator) parse(srcs []*source.Source, current *bytes.Buffer, outfile string, pkg string) (data *tdata.TData, importsList []addimports.ImportIface, err error) {
	data = &tdata.TData{
		Comment: g.Comment,
		NoFDoc:  g.NoFDoc,
		NoTDoc:  g.NoTDoc,
		Pkg:     pkg,
		Post:    g.Post,
		Pre:     g.Pre,
	}
	err = g.parseTargetSrc(data, &importsList, outfile, current)
	if err != nil {
		return nil, nil, err
	}
	err = g.parseSrc(data, &importsList, srcs, pkg)
	if err != nil {
		return nil, nil, err
	}
	return
}

func (g Generator) parseSrc(data *tdata.TData, importsList *[]addimports.ImportIface, srcs []*source.Source, pkg string) (err error) {
	p, err := parser.ParseFiles(srcs)
	if err != nil {
		return err
	}
	goGenerateSrc := firstWithLine(srcs...)
	err = g.populateTypeInterfaces(data, goGenerateSrc, p)
	if err != nil {
		return err
	}
	err = g.populateRecvInterfaces(data, goGenerateSrc, p)
	if err != nil {
		return err
	}
	*importsList = append(*importsList, getImportsAddPath(p.Imports(), srcs[0].File)...)
	return nil
}

// parseTargetSrc scans any previously generated source before any additions.
func (g Generator) parseTargetSrc(data *tdata.TData, importsList *[]addimports.ImportIface, path string, src *bytes.Buffer) (err error) {
	if src.Len() == 0 {
		return nil
	}
	p, err := parser.Parse(path, src, 0)
	if err != nil {
		return fmt.Errorf(`parse error: %w\n`, err)
	}
	for _, typ := range p.Query().GetTypesByType(types.INTERFACE) {
		iface, finish := makeInterface(data, typ.Name(), typ.Doc(), false)
		methods := p.Query().GetIfaceMethods(typ.Name())
		err = addIfaceMethods(iface, methods, g.NoFDoc)
		if err != nil {
			return err
		}
		err = finish()
		if err != nil {
			return err
		}
	}
	*importsList = append(*importsList, getImports(p.Imports())...)
	return nil
}

func (g Generator) populateTypeInterfaces(data *tdata.TData, src *source.Source, p *parser.Parser) (err error) {
	if !g.Type {
		return
	}
	var name string
	ifaceDefined := g.Iface != ``
	if ifaceDefined {
		name = g.Pre + g.Iface + g.Post
	}
	types := g.getTypeList(p, src)
	for _, typ := range types {
		if !ifaceDefined {
			name = g.Pre + typ.Name() + g.Post
		}
		doc := g.TDoc
		if doc == `` {
			doc = typ.Doc()
		}
		iface, finish := makeInterface(data, name, doc, g.NoTDoc)
		recvs := p.Query().GetRecvsByType(typ.Name())
		err = addRecvMethods(iface, recvs, p.Package(), data.Pkg, g.NoFDoc)
		if err != nil {
			return err
		}
		if iface.Methods == nil {
			continue
		}
		err = finish()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g Generator) populateRecvInterfaces(data *tdata.TData, src *source.Source, p *parser.Parser) (err error) {
	if !g.Method {
		return
	}
	var name string
	ifaceDefined := g.Iface != ``
	if ifaceDefined {
		name = g.Pre + g.Iface + g.Post
	}
	recvs := g.getRecvList(p, src)
	for _, recv := range recvs {
		typ := p.Query().GetTypeByName(recv.TypeName())
		if !ifaceDefined {
			name = g.Pre + typ.Name() + g.Post
		}
		doc := g.TDoc
		if doc == `` {
			doc = g.TDoc
		}
		iface, finish := makeInterface(data, name, doc, g.NoTDoc)
		m := tdata.NewMethod(recv.Name(), recv.Signature(), recv.Doc(), g.NoFDoc)
		err := iface.Add(m)
		if err != nil && err != tdata.ErrorDuplicateMethod {
			return err
		}
		err = finish()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g Generator) getRecvList(p *parser.Parser, src *source.Source) (r []*parser.Method) {
	if g.MatchFunc != `` {
		recvs := p.Query().GetRecvsByName(g.MatchFunc)
		if g.MatchType != `` {
			for _, recv := range recvs {
				if match.Match(recv.TypeName(), g.MatchType) && match.Capitalized(recv.TypeName()) {
					r = append(r, recv)
				}
			}
		} else {
			r = recvs
		}
	}
	if src != nil {
		file := src.File
		line := src.Line
		if len(r) == 0 && line > 0 {
			recv := p.Query().GetRecvByLine(file, line)
			if recv != nil {
				r = append(r, recv)
			}
		}
	}
	return
}

func (g Generator) getTypeList(p *parser.Parser, src *source.Source) (t []*parser.Type) {
	if g.Struct {
		t = append(t, p.Query().GetTypesByType(types.STRUCT)...)
	}
	if g.MatchType != `` {
		t = append(t, p.Query().GetTypeByPattern(g.MatchType)...)
	}
	if src != nil {
		file := src.File
		line := src.Line
		if len(t) == 0 && line > 0 {
			typ := p.Query().GetTypeByLine(file, line)
			if typ != nil {
				t = append(t, typ)
			}
		}
	}
	return
}

func addIfaceMethods(iface *tdata.Interface, methods []*parser.Method, noFuncDoc bool) error {
	for _, method := range methods {
		m := tdata.NewMethod(method.Name(), method.Signature(), method.Doc(), noFuncDoc)
		err := iface.Add(m)
		if err != nil {
			if err == tdata.ErrorDuplicateMethod {
				continue
			} else {
				return fmt.Errorf(`can not add method: %w`, err)
			}
		}
	}
	return nil
}

func addRecvMethods(iface *tdata.Interface, recvs []*parser.Method, parsedPkg, targetPkg string, noFuncDoc bool) error {
	for _, recv := range recvs {
		if recv.UsesTypeParams() {
			continue
		}
		if targetPkg != parsedPkg {
			recv.SetPkg(parsedPkg)
		}
		m := tdata.NewMethod(recv.Name(), recv.Signature(), recv.Doc(), noFuncDoc)
		err := iface.Add(m)
		if err != nil {
			if err == tdata.ErrorDuplicateMethod {
				continue
			} else {
				return fmt.Errorf(`can not add method: %w`, err)
			}
		}
	}
	return nil
}

func applyTemplate(current *bytes.Buffer, data *tdata.TData) error {
	t, err := template.New("generate.gotmpl").Parse(gentmpl)
	if err != nil {
		return err
	}
	err = t.Execute(current, data)
	if err != nil {
		return err
	}
	return nil
}

func getImportsAddPath(parsedImports []*parser.Import, path string) (imports []addimports.ImportIface) {
	for _, i := range parsedImports {
		if i.Name() == `_` || i.Name() == `.` {
			continue
		}
		imports = append(imports, i)
	}

	importPath, err := pathtoimport.PathToImport(path)
	if importPath != `` && err != nil {
		ip := addimports.NewImport(``, importPath)
		imports = append(imports, ip)
	}
	return
}

func getImports(parsedImports []*parser.Import) (imports []addimports.ImportIface) {
	for _, i := range parsedImports {
		if i.Name() == `_` || i.Name() == `.` {
			continue
		}
		imports = append(imports, i)
	}
	return
}

func makeInterface(data *tdata.TData, name, doc string, noTDoc bool) (iface *tdata.Interface, finish func() error) {
	finish = func() error { return nil }
	iface = data.Get(name)
	if iface == nil {
		iface = &tdata.Interface{
			Type: tdata.NewType(name, doc, noTDoc),
		}
		finish = func() error {
			return data.Add(iface)
		}
	}
	return
}

// setOutputPackage name of package in the output source.
func setOutputPackage(p, path string) (string, error) {
	if p != `` {
		return p, nil
	}
	if path == `` {
		return p, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return ``, err
	}
	d := filepath.Dir(abs)
	p = filepath.Base(d)
	return p, nil
}

func firstWithLine(srcs ...*source.Source) *source.Source {
	for _, src := range srcs {
		if src.Line > 0 {
			return src
		}
	}
	return nil
}
