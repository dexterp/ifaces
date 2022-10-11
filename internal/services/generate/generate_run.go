package generate

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/dexterp/ifaces/internal/resources/cond"
	"github.com/dexterp/ifaces/internal/resources/parser"
	"github.com/dexterp/ifaces/internal/resources/print"
	"github.com/dexterp/ifaces/internal/resources/srcio"
	"github.com/dexterp/ifaces/internal/resources/stringx"
	"github.com/dexterp/ifaces/internal/resources/tdata"
	"github.com/dexterp/ifaces/internal/resources/types"
	"github.com/jinzhu/copier"
)

var (
	ErrRecvNotFound = errors.New(`could not match receiver method`)
	ErrTypeNotFound = errors.New(`could not match type`)
)

type Options struct {
	Print     print.PrintIface
	NoFuncDoc bool
	NoTypeDoc bool
}

// Run return run struct
func (o Options) Run() *Run {
	return &Run{
		Options: o,
	}
}

type Run struct {
	Options Options
	srcs    []srcio.Source
	dests   []srcio.Destination
}

func (r *Run) Src(srcs []srcio.Source, dests []srcio.Destination) *Run {
	r.srcs = srcs
	r.dests = dests
	return r
}

// Type generate code by parsing types
func (r Run) Type(typ, iface, pkg, tdoc string) error {
	x := run{
		srcs:  r.srcs,
		dests: r.dests,
	}
	err := copier.Copy(x, r.Options)
	if err != nil {
		return err
	}
	p := typeParams{
		typ:   typ,
		iface: iface,
		pkg:   pkg,
		tdoc:  tdoc,
	}
	return x.typ(p)
}

// Recv generate code by parsing receivers type and method.
func (r Run) Recv(typ, method, iface, pkg, mdoc string) error {
	x := run{
		srcs:  r.srcs,
		dests: r.dests,
	}
	err := copier.Copy(x, r.Options)
	if err != nil {
		return err
	}
	p := recvParams{
		typ:    typ,
		method: method,
		iface:  iface,
		pkg:    pkg,
		mdoc:   mdoc,
	}
	return x.recv(p)
}

type run struct {
	srcs      []srcio.Source      `copier:"-"`
	dests     []srcio.Destination `copier:"-"`
	Print     print.PrintIface    `copier:"must"`
	targets   map[string]*target
	NoTypeDoc bool
	NoFuncDoc bool
}

func (r *run) typ(params typeParams) error {
	r.targets = r.makeTargets(params.pkg)
	err := r.parseTargets()
	if err != nil {
		return err
	}
	err = r.parseTypes(params)
	if err != nil {
		return err
	}
	return nil
}

func (r *run) recv(params recvParams) error {
	r.targets = r.makeTargets(params.pkg)
	err := r.parseTargets()
	if err != nil {
		return err
	}
	err = r.parseRecvs(params)
	if err != nil {
		return err
	}
	return nil
}

func (r *run) addPrefixImports(parsed []*parser.Import, recvs ...*parser.Method) {
	if parsed == nil || recvs == nil {
		return
	}
	for _, t := range r.targets {
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
				m := stringx.ExPkgPath(pi.Path)
				if m != `` && cond.EqualAnyString(m, r.Prefixes...) {
					t.imports[pi] = struct{}{}
					continue
				}
			}
		}
	}
}

func (g run) appendRecvsByLine(r []*parser.Method, file string, line int, q *parser.Query) []*parser.Method {
	if line > 0 {
		recv := q.GetRecvByLine(file, line)
		if recv != nil {
			r = append(r, recv)
		}
	}
	return r
}

func (g run) appendRecvsByTypeMethod(r []*parser.Method, typ, method string, q *parser.Query) []*parser.Method {
	if typ != `` {
		return r
	}
	recv := q.GetRecvByTypeMethod(typ, method)
	if recv != nil {
		r = append(r, recv)
	}
	return r
}

func (r run) appendTypeByLine(t []parser.Type, file string, line int, q *parser.Query) []parser.Type {
	typ := q.GetTypeByLine(file, line)
	if typ != nil {
		t = append(t, *typ)
	}
	return t
}

func (r run) appendTypeByName(t []parser.Type, typ string, q *parser.Query) []parser.Type {
	if x := q.GetTypeByName(typ); x != nil {
		return append(t, *x)
	}
	return t
}

// adds interface methods to iface
func (r *run) interfaceMethods(iface *tdata.Interface, methods []*parser.Method) error {
	for _, method := range methods {
		m := tdata.NewMethod(method.Name, method.Signature(), method.Doc, r.NoFuncDoc)
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

func (r run) makeTargets(pkg string) (targets map[string]*target) {
	targets = map[string]*target{}
	for _, d := range r.dests {
		t := &target{
			file:   d.File,
			output: d.Output,
		}

		// set package for target file
		if pkg == `` {
			abs, err := filepath.Abs(d.File)
			r.Print.HasFatalf(`unable to get absolute path for %s: %w`, d.File, err)
			dir := filepath.Dir(abs)
			pkg = filepath.Base(dir)
			m := stringx.ExPkg(pkg)
			if m == `` {
				r.Print.HasFatalf(`invalid package name "%s" determined from path %s`, d.File, pkg)
			}

		}
		t.pkg = pkg
		targets[d.File] = t
	}
	return
}

func (r *run) parseRecvs(params recvParams) (err error) {
	p, err := parser.ParseFiles(r.srcs)
	if err != nil {
		return err
	}
	q := parser.NewQuery(p)
	src := firstWithLine(r.srcs...)

	// Get receivers by type and method, if provided else get by line number
	// which is supplied by go generate
	var recv *parser.Method
	if params.typ != `` {
		recv = q.GetRecvByTypeMethod(params.typ, params.method)
	} else if src != nil && src.Line > 0 {
		recv = q.GetRecvByLine(src.File, src.Line)
	}

	if recv == nil {
		return ErrRecvNotFound
	}

	for _, t := range r.targets {
		t.tdata.Pkg = cond.First(params.pkg, p.Package).(string)
		data := t.tdata
		// Set interface name
		var name string
		typ := q.GetTypeByName(recv.TypeName)
		if params.iface == `` {
			name = typ.Name
		} else {
			name = params.iface
		}

		// Set type document
		var tdoc string
		if params.tdoc == `` {
			tdoc = params.tdoc
		} else {
			tdoc = typ.Doc
		}

		// Create interface
		iface, finish := makeInterface(data, name, tdoc, r.NoTypeDoc)
		m := tdata.NewMethod(recv.Name, recv.Signature(), recv.Doc, r.NoFuncDoc)
		err := iface.Add(m)
		if err != nil && err != tdata.ErrorDuplicateMethod {
			return err
		}
		err = finish()
		if err != nil {
			return err
		}
		r.addPrefixImports(p.Imports, recv)
		if isExported(recv) {
			t.exported = true
		}
	}
	return nil
}

// parseTargets scans any previously generated source before any additions.
func (r *run) parseTargets() (err error) {
	for _, t := range r.targets {
		src := t.src
		path := t.file
		p, err := parser.Parse(path, src, 0)
		if err != nil {
			return fmt.Errorf(`error parsing target source: %w`, err)
		}
		q := parser.NewQuery(p)
		for _, i := range p.Imports {
			t.imports[i] = struct{}{}
		}
		for _, typ := range q.GetTypesByType(types.INTERFACE) {
			methods := q.GetIfaceMethods(typ.Name)
			iface, finish := makeInterface(t.tdata, typ.Name, typ.Doc, false)
			err = r.interfaceMethods(iface, methods)
			if err != nil {
				return err
			}
			err = finish()
			if err != nil {
				return err
			}
		}
	}
	return
}

func (r *run) parseTypes(params typeParams) (err error) {
	p, err := parser.ParseFiles(r.srcs)
	if err != nil {
		return err
	}
	q := parser.NewQuery(p)
	src := firstWithLine(r.srcs...)

	// Get type by by name, if provided else get by line number which is
	// supplied by go generate
	var typ *parser.Type
	if params.typ != `` {
		typ = q.GetTypeByName(params.typ)
	} else if src != nil && src.Line > 0 {
		typ = q.GetTypeByLine(src.File, src.Line)
	}

	if typ == nil {
		return ErrTypeNotFound
	}

	for _, t := range r.targets {
		t.tdata.Pkg = cond.First(params.pkg, p.Package).(string)
		var tdoc string
		if params.tdoc != `` {
			tdoc = params.tdoc
		} else {
			tdoc = typ.Doc
		}
		iface, finish := makeInterface(t.tdata, params.iface, tdoc, r.NoTypeDoc)

		recvs := q.GetRecvsByType(typ.Name)
		addPackage(&recvs, p.Package, t.tdata.Pkg)
		err = addRecvMethods(iface, &recvs, p.Package, t.tdata.Pkg, r.NoFuncDoc)
		if err != nil {
			return err
		}
		r.addPrefixImports(p.Imports, recvs...)
		if isExported(recvs...) {
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
	return
}

type recvParams struct {
	typ    string
	method string
	iface  string
	pkg    string
	tdoc   string
	mdoc   string
}

type typeParams struct {
	typ   string
	iface string
	pkg   string
	tdoc  string
}
