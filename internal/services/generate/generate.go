package generate

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/dexterp/ifaces/internal/resources/addimports"
	"github.com/dexterp/ifaces/internal/resources/cond"
	"github.com/dexterp/ifaces/internal/resources/match"
	"github.com/dexterp/ifaces/internal/resources/parser"
	"github.com/dexterp/ifaces/internal/resources/paths"
	"github.com/dexterp/ifaces/internal/resources/print"
	"github.com/dexterp/ifaces/internal/resources/srcformat"
	"github.com/dexterp/ifaces/internal/resources/srcio"
	"github.com/dexterp/ifaces/internal/resources/tdata"
	"github.com/dexterp/ifaces/internal/resources/types"
)

//go:generate ifaces type -o generate_iface.go -i GenerateIface

// Generate interface generator
type Generate struct {
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
	targets   map[string]*target
}

//go:embed generate.gotmpl
var gentmpl string

var ErrorNoSourceFile = errors.New(`no source files processed`)

var reMatchValidPackage = regexp.MustCompile(`^[a-z][a-z0-9]+$`)

var reMatchPackagePath = regexp.MustCompile(`(?:^|/)([a-z0-9]+)[^/]*$`)

// Generate generate interfaces source code for the gen sub command.
func (g *Generate) Generate(srcs []srcio.Source, current *bytes.Buffer, outfile string, output io.Writer) error {
	g.init()
	pkg, err := g.setOutputPackage(g.Pkg, outfile)
	if err != nil {
		return err
	}
	err = g.parse(srcs, current, outfile, pkg)
	if err != nil {
		return err
	}

	// TODO - move this exported import to parser
	var importValue *parser.Import
	if srcs != nil {
		i, err := paths.PathToImport(srcs[0].File)
		if err == nil {
			importValue = &parser.Import{
				Path: i,
			}
		}
	}
	for _, t := range g.targets {
		templateOut := &bytes.Buffer{}
		err = applyTemplate(templateOut, t.tdata)
		if err != nil {
			return err
		}
		importsOut := &bytes.Buffer{}
		importsList := []addimports.Import{}
		if t.exported && importValue != nil {
			importsList = append(importsList, addimports.NewImport(importValue.Name, importValue.Path))
		}
		for i := range t.imports {
			importsList = append(importsList, addimports.NewImport(i.Name, i.Path))
		}
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
	}
	return err
}

func (g *Generate) init() {
	g.targets = map[string]*target{}
}

func (g *Generate) parse(srcs []srcio.Source, current *bytes.Buffer, outfile string, pkg string) (err error) {
	err = g.parseTargetSrc(outfile, current)
	if err != nil {
		return err
	}
	err = g.parseSrc(srcs, pkg)
	if err != nil {
		return err
	}
	return nil
}

func (g *Generate) parseSrc(srcs []srcio.Source, pkg string) (err error) {
	p, err := parser.ParseFiles(srcs)
	if err != nil {
		return err
	}
	for _, t := range g.targets {
		t.tdata.Pkg = cond.First(pkg, p.Package).(string)
		goGenerateSrc := firstWithLine(srcs...)
		err = g.populateTypeInterfaces(t, goGenerateSrc, p)
		if err != nil {
			return err
		}
		err = g.populateRecvInterfaces(t, goGenerateSrc, p)
		if err != nil {
			return err
		}
	}
	return nil
}

// parseTargetSrc scans any previously generated source before any additions.
func (g *Generate) parseTargetSrc(path string, src *bytes.Buffer) (err error) {
	// No need to parse source if file is empty or dose not exists
	if src.Len() == 0 {
		g.getOrMakeTarget(path, nil)
		return nil
	}
	p, err := parser.Parse(path, src, 0)
	if err != nil {
		return fmt.Errorf(`error parsing target source: %w`, err)
	}
	q := parser.NewQuery(p)
	t := g.getOrMakeTarget(path, p.Imports)
	for _, typ := range q.GetTypesByType(types.INTERFACE) {
		iface, finish := makeInterface(t.tdata, typ.Name, typ.Doc, false)
		methods := q.GetIfaceMethods(typ.Name)
		err = g.addIfaceMethods(iface, methods)
		if err != nil {
			return err
		}
		err = finish()
		if err != nil {
			return err
		}
	}
	return
}

func (g *Generate) populateTypeInterfaces(t *target, src *srcio.Source, p *parser.Parser) (err error) {
	if !g.Type && !g.Struct {
		return
	}
	var name string
	ifaceDefined := g.Iface != ``
	if ifaceDefined {
		name = g.Pre + g.Iface + g.Post
	}
	types := g.getTypeList(p, src)
	q := parser.NewQuery(p)
	for _, typ := range types {
		if !ifaceDefined {
			name = g.Pre + typ.Name + g.Post
		}
		doc := g.TDoc
		if doc == `` {
			doc = typ.Doc
		}
		iface, finish := makeInterface(t.tdata, name, doc, g.NoTDoc)
		recvs := &[]*parser.Method{}
		*recvs = q.GetRecvsByType(typ.Name)
		addPackage(recvs, p.Package, t.tdata.Pkg)
		err = addRecvMethods(iface, recvs, p.Package, t.tdata.Pkg, g.NoFDoc)
		if err != nil {
			return err
		}
		g.addPrefixImports(p.Imports, *recvs)
		if isExported(*recvs) {
			t.exported = true
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

func (g *Generate) populateRecvInterfaces(t *target, src *srcio.Source, p *parser.Parser) (err error) {
	data := t.tdata
	if !g.Method {
		return
	}
	var name string
	ifaceDefined := g.Iface != ``
	if ifaceDefined {
		name = g.Pre + g.Iface + g.Post
	}
	recvs := g.getRecvList(src, p)
	q := parser.NewQuery(p)
	for _, recv := range recvs {
		typ := q.GetTypeByName(recv.TypeName)
		if !ifaceDefined {
			name = g.Pre + typ.Name + g.Post
		}
		doc := g.TDoc
		if doc == `` {
			doc = g.TDoc
		}
		iface, finish := makeInterface(data, name, doc, g.NoTDoc)
		m := tdata.NewMethod(recv.Name, recv.Signature(), recv.Doc, g.NoFDoc)
		err := iface.Add(m)
		if err != nil && err != tdata.ErrorDuplicateMethod {
			return err
		}
		err = finish()
		if err != nil {
			return err
		}
	}
	g.addPrefixImports(p.Imports, recvs)
	if isExported(recvs) {
		t.exported = true
	}
	return nil
}

func (g *Generate) getOrMakeTarget(file string, imports []*parser.Import) *target {
	if t, ok := g.targets[file]; ok {
		return t
	}

	// Output target
	t := &target{
		file:    file,
		imports: map[*parser.Import]any{},
		tdata: &tdata.TData{
			Comment: g.Comment,
			NoFDoc:  g.NoFDoc,
			NoTDoc:  g.NoTDoc,
			Post:    g.Post,
			Pre:     g.Pre,
		},
	}
	for _, i := range imports {
		t.imports[i] = true
	}
	g.targets[file] = t
	return t
}

func (g Generate) getRecvList(src *srcio.Source, p *parser.Parser) (r []*parser.Method) {
	q := parser.NewQuery(p)
	if g.MatchFunc != `` {
		recvs := q.GetRecvsByName(g.MatchFunc)
		if g.MatchType != `` {
			for _, recv := range recvs {
				if match.Match(recv.TypeName, g.MatchType) && match.Capitalized(recv.TypeName) {
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
			recv := q.GetRecvByLine(file, line)
			if recv != nil {
				r = append(r, recv)
			}
		}
	}
	return
}

func (g Generate) getTypeList(p *parser.Parser, src *srcio.Source) (t []parser.Type) {
	q := parser.NewQuery(p)
	if g.Struct {
		t = append(t, q.GetTypesByType(types.STRUCT)...)
	}
	if g.MatchType != `` {
		t = append(t, q.GetTypeByPattern(g.MatchType)...)
	}
	if src != nil {
		file := src.File
		line := src.Line
		if len(t) == 0 && line > 0 {
			typ := q.GetTypeByLine(file, line)
			if typ != nil {
				t = append(t, *typ)
			}
		}
	}
	return
}

// setOutputPackage name of package in the output source.
func (g Generate) setOutputPackage(pkgCli, path string) (string, error) {
	if pkgCli != `` {
		return pkgCli, nil
	} else if path == `` {
		return ``, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return ``, err
	}
	d := filepath.Dir(abs)
	pkgCli = filepath.Base(d)
	m := reMatchValidPackage.FindAllString(pkgCli, 1)
	if m == nil {
		g.Print.HasFatalf(`invalid package name "%s" determined from path %s`, path, pkgCli)
	}
	pkgCli = m[0]
	return pkgCli, nil
}

func (g *Generate) addIfaceMethods(iface *tdata.Interface, methods []*parser.Method) error {
	for _, method := range methods {
		m := tdata.NewMethod(method.Name, method.Signature(), method.Doc, g.NoFDoc)
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

func addPackage(recvs *[]*parser.Method, parsedPkg, targetPkg string) {
	for _, recv := range *recvs {
		if targetPkg != parsedPkg {
			recv.Pkg = parsedPkg
		}
	}
}

func addRecvMethods(iface *tdata.Interface, recvs *[]*parser.Method, parsedPkg, targetPkg string, noFuncDoc bool) error {
	for _, recv := range *recvs {
		if targetPkg != parsedPkg {
			recv.Pkg = parsedPkg
		}
		m := tdata.NewMethod(recv.Name, recv.Signature(), recv.Doc, noFuncDoc)
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

func (g *Generate) addPrefixImports(parsed []*parser.Import, recvs []*parser.Method) {
	if parsed == nil || recvs == nil {
		return
	}
	for _, t := range g.targets {
		for _, r := range recvs {
			if r.Prefixes == nil {
				continue
			}
			for _, pi := range parsed {
				if cond.EqualAnyString(pi.Name, `_`, `.`) {
					continue
				} else if pi.Name != `` && cond.EqualAnyString(pi.Name, r.Prefixes...) {
					t.imports[pi] = struct{}{}
					continue
				}
				m := reMatchPackagePath.FindAllString(pi.Path, 1)
				if m != nil && cond.EqualAnyString(m[0], r.Prefixes...) {
					t.imports[pi] = struct{}{}
					continue
				}
			}
		}
	}
}

func firstWithLine(srcs ...srcio.Source) *srcio.Source {
	for _, src := range srcs {
		if src.Line > 0 {
			return &src
		}
	}
	return nil
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

func isExported(recvs []*parser.Method) bool {
	for _, r := range recvs {
		if r.NeedsImport() {
			return true
		}
	}
	return false
}
