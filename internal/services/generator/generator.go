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
	"github.com/dexterp/ifaces/internal/resources/modinfo"
	"github.com/dexterp/ifaces/internal/resources/parser"
	"github.com/dexterp/ifaces/internal/resources/print"
	"github.com/dexterp/ifaces/internal/resources/tdata"
	"github.com/dexterp/ifaces/internal/resources/types"
	"golang.org/x/tools/imports"
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
func (g Generator) Generate(srcs []*Src, current *bytes.Buffer, outfile string, output io.Writer) error {
	pkg, err := pckage(g.Pkg, outfile)
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
	finalSrc, err := imports.Process(outfile, importsOut.Bytes(), &imports.Options{
		TabIndent: true,
		TabWidth:  2,
		Fragment:  true,
		Comments:  true,
	})
	if err != nil {
		return err
	}
	_, err = output.Write(finalSrc)
	return err
}

func (g Generator) parse(srcs []*Src, current *bytes.Buffer, outfile string, pkg string) (data *tdata.TData, importsList []addimports.ImportIface, err error) {
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

func (g Generator) parseSrc(data *tdata.TData, importsList *[]addimports.ImportIface, srcs []*Src, pkg string) (err error) {
	foundFile := false
	for i := range srcs {
		pth := srcs[i].File
		src := srcs[i].Src
		p, err := parser.Parse(pth, src)
		if g.Print.HasWarnf("skipping \"%s\": %v\n", pth, err) {
			continue
		}
		foundFile = true
		if len(srcs) == 1 && data.Pkg == `` {
			data.Pkg = p.Package()
		}
		err = g.populateTypeInterfaces(data, srcs[i], p)
		if err != nil {
			return err
		}
		err = g.populateRecvInterfaces(data, srcs[i], p)
		if err != nil {
			return err
		}
		*importsList = append(*importsList, getImports(p.Imports(), srcs[i].File)...)
	}
	if !foundFile {
		return ErrorNoSourceFile
	}
	return nil
}

// parseTargetSrc scans any previously generated source before any additions.
func (g Generator) parseTargetSrc(data *tdata.TData, importsList *[]addimports.ImportIface, path string, src *bytes.Buffer) (err error) {
	if src.Len() == 0 {
		return nil
	}
	p, err := parser.Parse(path, src)
	if err != nil {
		return fmt.Errorf(`parse error: %w\n`, err)
	}
	for _, typ := range p.GetTypesByType(types.INTERFACE) {
		iface, finish := makeInterface(data, typ.Name(), typ.Doc(), false)
		methods := p.GetIfaceMethods(typ.Name())
		err = addIfaceMethods(iface, methods, g.NoFDoc)
		if err != nil {
			return err
		}
		err = finish()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g Generator) populateTypeInterfaces(data *tdata.TData, src *Src, p *parser.Parser) (err error) {
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
		recvs := p.GetRecvsByType(typ.Name())
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

func (g Generator) populateRecvInterfaces(data *tdata.TData, src *Src, p *parser.Parser) (err error) {
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
		typ := p.GetTypeByName(recv.TypeName())
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

func (g Generator) getRecvList(p *parser.Parser, src *Src) (r []parser.MethodIface) {
	if g.MatchFunc != `` {
		recvs := p.GetRecvsByName(g.MatchFunc)
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
	line := src.Line
	if len(r) == 0 && line > 0 {
		recv := p.GetRecvByLine(line)
		if recv != nil {
			r = append(r, recv)
		}
	}
	return
}

func (g Generator) getTypeList(p *parser.Parser, src *Src) (t []*parser.Type) {
	line := src.Line
	if g.Struct {
		t = append(t, p.GetTypesByType(types.STRUCT)...)
	}
	if g.MatchType != `` {
		t = append(t, p.GetTypeByPattern(g.MatchType)...)
	}
	if len(t) == 0 && line > 0 {
		typ := p.GetTypeByLine(line)
		if typ != nil {
			t = append(t, typ)
		}
	}
	return
}

type Src struct {
	File string
	Line int
	Src  any
}

func addIfaceMethods(iface *tdata.Interface, methods []parser.MethodIface, noFuncDoc bool) error {
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

func addRecvMethods(iface *tdata.Interface, recvs []parser.MethodIface, parsedPkg, targetPkg string, noFuncDoc bool) error {
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

func getImports(parsedImports []*parser.Import, path string) (imports []addimports.ImportIface) {
	for _, i := range parsedImports {
		imports = append(imports, i)
	}
	importPath, err := modinfo.GetImport(``, nil, path)
	if importPath != `` && err != nil {
		ip := addimports.NewImport(``, importPath)
		imports = append(imports, ip)
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

// pckage name of package in the output source.
func pckage(p, path string) (string, error) {
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
