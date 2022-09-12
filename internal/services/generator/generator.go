package generator

import (
	"bytes"
	_ "embed"
	"io"
	"text/template"

	"github.com/dexterp/ifaces/internal/resources/addimports"
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

// Generate generate interfaces source code for the gen sub command.
func (g Generator) Generate(srcs []*Src, current *bytes.Buffer, outfile string, output io.Writer) error {
	data, importsList, err := g.genData(srcs)
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

func (g Generator) genData(srcs []*Src) (*tdata.TData, []addimports.Import, error) {
	data := &tdata.TData{
		Comment: g.comment,
		NoFDoc:  g.noFDoc,
		NoTDoc:  g.noTDoc,
		Pkg:     g.pkg,
		Post:    g.post,
		Pre:     g.pre,
		Ifaces:  []*tdata.Interface{},
	}
	importsList := []addimports.Import{}
	uniqueType := map[string]any{}
	for i := range srcs {
		types := []*parser.Type{}
		pth := srcs[i].File
		line := srcs[i].Line
		src := srcs[i].Src
		p, err := parser.Parse(pth, src)
		if err != nil {
			g.print.Warnf(`skipping "%s" due to error: %s`, pth, err)
			continue
		}
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
		for _, i := range p.Imports() {
			importsList = append(importsList, i)
		}
		for _, typ := range types {
			iface := &tdata.Interface{
				Type: &tdata.Type{
					NoTDoc: g.noTDoc,
					Iface:  g.iface,
					Pre:    g.pre,
					Post:   g.post,
					Decl:   typ},
			}
			recvs := p.GetRecvs(typ.Name())
			for _, recv := range recvs {
				r := &tdata.Recv{
					NoFDoc: g.noFDoc,
					Recv:   recv,
				}
				iface.Recvs = append(iface.Recvs, r)
			}
			_, ok := uniqueType[typ.Name()]
			if len(iface.Recvs) > 0 && !ok {
				data.Ifaces = append(data.Ifaces, iface)
				uniqueType[typ.Name()] = struct{}{}
			}
		}
	}
	return data, importsList, nil
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

type Src struct {
	File string
	Line int
	Src  any
}
