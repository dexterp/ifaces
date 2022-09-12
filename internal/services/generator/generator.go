package generator

import (
	"bytes"
	_ "embed"
	"io"
	"text/template"

	"github.com/dexterp/ifaces/internal/resources/addimports"
	"github.com/dexterp/ifaces/internal/resources/parser"
	"github.com/dexterp/ifaces/internal/resources/tdata"
	"golang.org/x/tools/imports"
)

// Generator interface generator
type Generator struct {
	comment string
	iface   string
	noFDoc  bool
	noHdr   bool
	noTDoc  bool
	pkg     string
	post    string
	pre     string
	struc   bool
	wild    string
}

// Options
type Options struct {
	Comment string // Comment comment at the top of the file
	Iface   string // Iface explicitly set interface name
	NoFDoc  bool   // NoFDoc omit copying function documentation
	NoHdr   bool   // NoHdr omit writing header
	NoTDoc  bool   // NoTDoc omit copying type documentation
	Pkg     string // Pkg package name
	Post    string // Post postfix to interface name
	Pre     string // Pre prefix to interface name
	Struct  bool   // Struct generate an interface for all structs
	Wild    string // Wild wildcard match
}

// New generate IfaceGen
func New(opts Options) *Generator {
	return &Generator{
		comment: opts.Comment,
		iface:   opts.Iface,
		noFDoc:  opts.NoFDoc,
		noHdr:   opts.NoHdr,
		noTDoc:  opts.NoTDoc,
		pkg:     opts.Pkg,
		post:    opts.Post,
		pre:     opts.Pre,
		struc:   opts.Struct,
		wild:    opts.Wild,
	}
}

//go:embed head.gotmpl
var headtmpl string

//go:embed entity.gotmpl
var entitytmpl string

// Gen generate interfaces source code for the gen sub command.

// Head generate interfaces source code for the head sub command and write to
// output. Generate used with a go generator
func (g Generator) Head(inputfile string, inputsrc any, outfile string, current io.ReadWriter, output io.Writer) error {
	p, err := parser.Parse(inputfile, inputsrc)
	if err != nil {
		return err
	}
	var t *template.Template
	if g.noHdr {
		t, err = template.New("entity.gotmpl").Parse(entitytmpl)
	} else {
		t, err = template.New("head.gotmpl").Parse(headtmpl)
	}
	if err != nil {
		return err
	}
	pkg := g.pkg
	if pkg == `` {
		pkg = p.Package()
	}
	data := &tdata.TData{
		Comment: g.comment,
		NoFDoc:  g.noFDoc,
		NoTDoc:  g.noTDoc,
		Pkg:     pkg,
		Post:    g.post,
		Pre:     g.pre,
		Ifaces:  []*tdata.Interface{},
	}
	var types []*parser.Type
	if g.struc {
		types = append(types, p.TypeByType(parser.StructType)...)
	}
	if g.wild != `` {
		types = append(types, p.TypeByPattern(g.wild)...)
	}
	unique := map[string]any{}
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
		_, ok := unique[typ.Name()]
		if len(iface.Recvs) > 0 && !ok {
			data.Ifaces = append(data.Ifaces, iface)
			unique[typ.Name()] = struct{}{}
		}
	}
	err = t.Execute(current, data)
	if err != nil {
		return err
	}
	importsList := []addimports.Import{}
	for _, i := range p.Imports() {
		importsList = append(importsList, i)
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

// Entry generate interfaces source code for the item sub command and write to
// output
func (g Generator) Entry(inputfile string, inputsrc any, line int, outfile string, current io.ReadWriter, output io.Writer) error {
	p, err := parser.Parse(inputfile, inputsrc)
	if err != nil {
		return err
	}
	t, err := template.New("entity.gotmpl").Parse(entitytmpl)
	if err != nil {
		return err
	}
	data := &tdata.TData{
		Comment: g.comment,
		NoFDoc:  g.noFDoc,
		NoTDoc:  g.noTDoc,
		Post:    g.post,
		Pre:     g.pre,
		Ifaces:  []*tdata.Interface{},
	}
	decl := p.GetType(line)
	if decl == nil {
		return nil
	}
	recvs := p.GetRecvs(decl.Name())
	if len(recvs) == 0 {
		return nil
	}
	iface := &tdata.Interface{
		Type: &tdata.Type{
			NoTDoc: g.noTDoc,
			Pre:    g.pre,
			Post:   g.post,
			Decl:   decl},
	}
	for _, recv := range recvs {
		r := &tdata.Recv{
			NoFDoc: g.noFDoc,
			Recv:   recv,
		}
		iface.Recvs = append(iface.Recvs, r)
	}
	data.Ifaces = append(data.Ifaces, iface)
	err = t.Execute(current, data)
	if err != nil {
		return err
	}
	importsList := []addimports.Import{}
	for _, i := range p.Imports() {
		importsList = append(importsList, i)
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
