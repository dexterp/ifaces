package generator

import (
	"bytes"
	_ "embed"
	"errors"
	"io"
	"path/filepath"
	"text/template"

	"github.com/dexterp/ifaces/internal/resources/addimports"
	"github.com/dexterp/ifaces/internal/resources/modinfo"
	"github.com/dexterp/ifaces/internal/resources/parser"
	"github.com/dexterp/ifaces/internal/resources/print"
	"github.com/dexterp/ifaces/internal/resources/tdata"
	"golang.org/x/tools/imports"
)

//go:generate ifaces generator_iface.go --post Iface

// Generator interface generator
type Generator struct {
	comment string
	iface   string
	noFDoc  bool
	noTDoc  bool
	pkg     string
	post    string
	pre     string
	print   print.PrintIface
	struc   bool
	wild    string
}

// Options
type Options struct {
	Comment string           // Comment comment at the top of the file
	Iface   string           // Iface explicitly set interface name
	NoFDoc  bool             // NoFDoc omit copying function documentation
	NoTDoc  bool             // NoTDoc omit copying type documentation
	Pkg     string           // Pkg package name
	Post    string           // Post postfix to interface name
	Pre     string           // Pre prefix to interface name
	Print   print.PrintIface // Print handler
	Struct  bool             // Struct generate an interface for all structs
	Match   string           // Wild wildcard match
}

// New generate IfaceGen
func New(opts Options) *Generator {
	return &Generator{
		comment: opts.Comment,
		iface:   opts.Iface,
		noFDoc:  opts.NoFDoc,
		noTDoc:  opts.NoTDoc,
		pkg:     opts.Pkg,
		post:    opts.Post,
		pre:     opts.Pre,
		print:   opts.Print,
		struc:   opts.Struct,
		wild:    opts.Match,
	}
}

//go:embed generate.gotmpl
var gentmpl string

//go:embed entity.gotmpl
var entrytmpl string

var ErrorNoSourceFile = errors.New(`no source files processed`)

// Generate generate interfaces source code for the gen sub command.
func (g Generator) Generate(srcs []*Src, current *bytes.Buffer, outfile string, output io.Writer) error {
	pkg, err := g.pckage(outfile)
	if err != nil {
		return err
	}
	data, importsList, err := g.genData(srcs, pkg)
	if err != nil {
		return err
	}
	err = g.applyTemplate(current, data)
	if err != nil {
		return err
	}
	importsOut := &bytes.Buffer{}
	err = addimports.AddImports(outfile, current, importsList, importsOut)
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

func (g Generator) genData(srcs []*Src, pkg string) (*tdata.TData, []addimports.ImportIface, error) {
	data := &tdata.TData{
		Comment: g.comment,
		NoFDoc:  g.noFDoc,
		NoTDoc:  g.noTDoc,
		Pkg:     pkg,
		Post:    g.post,
		Pre:     g.pre,
		Ifaces:  []*tdata.Interface{},
	}
	importsList := []addimports.ImportIface{}
	uniqueType := map[string]any{}
	foundFile := false
	for i := range srcs {
		pth := srcs[i].File
		src := srcs[i].Src
		p, err := parser.Parse(pth, src)
		if err != nil {
			if errors.Unwrap(err) != nil {
				err = errors.Unwrap(err)
			}
			g.print.Warnf("skipping \"%s\": %s\n", pth, err.Error())
			continue
		}
		foundFile = true
		if len(srcs) == 1 && data.Pkg == `` {
			data.Pkg = p.Package()
		}
		types := g.getTypeList(p, srcs[i])
		for _, typ := range types {
			iface := &tdata.Interface{
				Type: &tdata.Type{
					NoTDoc: g.noTDoc,
					Iface:  g.iface,
					Pre:    g.pre,
					Post:   g.post,
					Decl:   typ,
				},
				Recvs: g.getRecvs(p, data.Pkg, typ.Name()),
			}
			_, ok := uniqueType[typ.Name()]
			if len(iface.Recvs) > 0 && !ok {
				data.Ifaces = append(data.Ifaces, iface)
				uniqueType[typ.Name()] = struct{}{}
			}
		}
		importsList = append(importsList, g.getImportsList(p, srcs[i])...)
	}
	if !foundFile {
		return nil, nil, ErrorNoSourceFile
	}
	return data, importsList, nil
}

func (g Generator) getImportsList(p *parser.Parser, src *Src) (importsList []addimports.ImportIface) {
	pth := src.File
	for _, i := range p.Imports() {
		importsList = append(importsList, i)
	}
	importpath, err := modinfo.GetImport(nil, ``, pth)
	if importpath != `` && err != nil {
		ip := addimports.NewImport(``, importpath)
		importsList = append(importsList, ip)
	}
	return
}

func (g Generator) getTypeList(p *parser.Parser, src *Src) (types []*parser.Type) {
	line := src.Line
	if g.struc {
		types = append(types, p.TypeByType(parser.StructType)...)
	}
	if g.wild != `` {
		types = append(types, p.TypeByPattern(g.wild)...)
	}
	// TODO -Follow bug "types == nil" fails if "len(types) == 0"
	if len(types) == 0 && line > 0 {
		typ := p.GetType(line)
		if typ != nil {
			types = append(types, typ)
		}
	}
	return
}

func (g Generator) getRecvs(p *parser.Parser, pkg, typename string) (tdatarecv []*tdata.Recv) {
	recvs := p.GetRecvs(typename)
	for _, recv := range recvs {
		// skip generics
		if recv.UsesGenerics() {
			continue
		}
		if pkg != p.Package() {
			recv.SetPkg(p.Package())
		}
		r := &tdata.Recv{
			NoFDoc: g.noFDoc,
			Recv:   recv,
		}
		tdatarecv = append(tdatarecv, r)
	}
	return
}

func (g Generator) applyTemplate(current *bytes.Buffer, data *tdata.TData) error {
	t, err := template.New("entity.gotmpl").Parse(entrytmpl)
	if current.Bytes() == nil {
		t, err = template.New("head.gotmpl").Parse(gentmpl)
	}
	if err != nil {
		return err
	}
	err = t.Execute(current, data)
	if err != nil {
		return err
	}
	return nil
}

// pckage name of package in the output source.
func (g Generator) pckage(out string) (string, error) {
	if g.pkg != `` {
		return g.pkg, nil
	}
	if out == `` {
		return g.pkg, nil
	}
	abs, err := filepath.Abs(out)
	if err != nil {
		return ``, err
	}
	d := filepath.Dir(abs)
	newpkg := filepath.Base(d)
	return newpkg, nil
}

type Src struct {
	File string
	Line int
	Src  any
}
