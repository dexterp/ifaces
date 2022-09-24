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

var ErrorNoSourceFile = errors.New(`no source files processed`)

// Generate generate interfaces source code for the gen sub command.
func (g Generator) Generate(srcs []*Src, current *bytes.Buffer, outfile string, output io.Writer) error {
	pkg, err := g.pckage(outfile)
	if err != nil {
		return err
	}
	data, importsList, err := g.parse(srcs, current, outfile, pkg)
	if err != nil {
		return err
	}
	templateOut := &bytes.Buffer{}
	err = g.applyTemplate(templateOut, data)
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
		Comment: g.comment,
		NoFDoc:  g.noFDoc,
		NoTDoc:  g.noTDoc,
		Pkg:     pkg,
		Post:    g.post,
		Pre:     g.pre,
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
		var name string
		ifaceDefined := g.iface != ``
		if ifaceDefined {
			name = g.pre + g.iface + g.post
		}
		types := g.getTypeList(p, srcs[i])
		for _, typ := range types {
			if !ifaceDefined {
				name = g.pre + typ.Name() + g.post
			}
			var iface *tdata.Interface
			add := false
			iface = data.Get(name)
			if iface == nil {
				iface = &tdata.Interface{
					Type: tdata.NewType(name, typ.Doc(), g.noTDoc),
				}
				add = true
			}
			if err != nil {
				return fmt.Errorf(`can not add interface: %w`, err)
			}
			err = g.addRecvMethods(iface, p, data.Pkg, typ.Name())
			if err != nil {
				return err
			}
			if !add {
				continue
			}
			if iface.Methods == nil {
				continue
			}
			err = data.Add(iface)
			if err != nil {
				return err
			}
		}
		*importsList = append(*importsList, g.getImportsList(p, srcs[i])...)
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
		var iface *tdata.Interface
		add := false
		iface = data.Get(typ.Name())
		if iface == nil {
			iface = &tdata.Interface{
				Type: tdata.NewType(typ.Name(), typ.Doc(), g.noTDoc),
			}
			add = true
		}
		if err != nil {
			return err
		}
		err = g.addIfaceMethods(iface, p, typ.Name())
		if err != nil {
			return err
		}
		if !add {
			continue
		}
		if iface.Methods == nil {
			continue
		}
		err = data.Add(iface)
		if err != nil {
			return err
		}
	}
	return nil
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

func (g Generator) getTypeList(p *parser.Parser, src *Src) (t []*parser.Type) {
	line := src.Line
	if g.struc {
		t = append(t, p.GetTypesByType(types.STRUCT)...)
	}
	if g.wild != `` {
		t = append(t, p.GetTypeByPattern(g.wild)...)
	}
	// TODO -Follow bug "types == nil" fails if "len(types) == 0"
	if len(t) == 0 && line > 0 {
		typ := p.GetTypeByLine(line)
		if typ != nil {
			t = append(t, typ)
		}
	}
	return
}

func (g Generator) addIfaceMethods(i *tdata.Interface, p *parser.Parser, typename string) error {
	methods := p.GetIfaceMethods(typename)
	for _, method := range methods {
		m := tdata.NewMethod(method.Name(), method.Signature(), method.Doc(), g.noFDoc)
		err := i.Add(m)
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

func (g Generator) addRecvMethods(i *tdata.Interface, p *parser.Parser, pkg, typename string) error {
	recvs := p.GetTypeRecvs(typename)
	for _, recv := range recvs {
		// skip generics
		if recv.UsesTypeParams() {
			continue
		}
		if pkg != p.Package() {
			recv.SetPkg(p.Package())
		}
		m := tdata.NewMethod(recv.Name(), recv.Signature(), recv.Doc(), g.noFDoc)
		err := i.Add(m)
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

func (g Generator) applyTemplate(current *bytes.Buffer, data *tdata.TData) error {
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
