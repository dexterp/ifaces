package parser

import (
	"bytes"
	"go/ast"
	"strings"

	"github.com/dexterp/ifaces/internal/resources/stringx"
	"github.com/dexterp/ifaces/internal/resources/typecheck"
)

func RecvToFunc(f *ast.FuncDecl, hasType typecheck.HasType) *Func {
	p := &funcparse{
		hasType: hasType,
	}
	return p.recvMethod(f)
}

func IfaceToFunc(f *ast.Field, hasType typecheck.HasType) *Func {
	p := &funcparse{
		hasType: hasType,
	}
	return p.ifaceMethod(f)
}

type Func struct {
	Prefixes map[string]any
	pkg      string
	name     string
	params   []*param
	results  []*param
}

// Package set package name
func (f *Func) Package(pkg string) *Func {
	f.pkg = pkg
	return f
}

func (f *Func) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString(f.name)
	buf.WriteString(f.stringParams())
	buf.WriteString(f.stringReturns())
	return buf.String()
}

func (f *Func) stringParams() string {
	buf := &bytes.Buffer{}
	if len(f.params) == 0 {
		buf.WriteString(`()`)
	} else {
		l := []string{}
		for _, p := range f.params {
			l = append(l, p.string())
		}
		buf.WriteString(`(` + strings.Join(l, `, `) + `)`)
	}
	return buf.String()
}

func (f *Func) stringReturns() string {
	buf := &bytes.Buffer{}
	switch len(f.results) {
	case 0:
		return ``
	case 1:
		if strings.Contains(f.results[0].string(), " ") {
			buf.WriteString(` (` + f.results[0].string() + `)`)
		} else {
			buf.WriteString(` ` + f.results[0].string())
		}
	default:
		l := []string{}
		for _, p := range f.results {
			l = append(l, p.string())
		}
		buf.WriteString(` (` + strings.Join(l, `, `) + `)`)
	}
	return buf.String()
}

type funcparse struct {
	hasType   typecheck.HasType
	pkgIfNone *string
	prefixes  map[string]any
}

func (p funcparse) recvMethod(f *ast.FuncDecl) *Func {
	if f.Recv == nil {
		return nil
	}
	var (
		name = f.Name.String()
	)
	fn := &Func{
		name:     name,
		Prefixes: map[string]any{},
	}
	p.pkgIfNone = &fn.pkg
	p.prefixes = fn.Prefixes

	if f.Type.Params != nil {
		fn.params = p.params(f.Type.Params.List)
	}
	if f.Type.Results != nil {
		fn.results = p.params(f.Type.Results.List)
	}
	return fn
}

func (p funcparse) ifaceMethod(f *ast.Field) *Func {
	if f.Names == nil {
		return nil
	}
	funcName := f.Names[0].Name
	ft, ok := f.Type.(*ast.FuncType)
	if !ok {
		return nil
	}
	fn := &Func{
		name:     funcName,
		Prefixes: map[string]any{},
	}
	p.pkgIfNone = &fn.pkg
	p.prefixes = fn.Prefixes
	if ft.Params != nil {
		fn.params = p.params(ft.Params.List)
	}
	if ft.Results != nil {
		fn.results = p.params(ft.Results.List)
	}
	return fn
}

func (p *funcparse) parseExpr(n ast.Expr) typeExpr {
	if t := p.typChan(n, ``, ``); t != nil {
		return t
	}
	if t := p.typeInterface(n, ``, ``); t != nil {
		return t
	}
	if t := p.typ(n, ``, ``, ``); t != nil {
		return t
	}
	if t := p.typSlice(n, ``, ``); t != nil {
		return t
	}
	return p.typMap(n, ``, ``)
}

func (p *funcparse) params(fields []*ast.Field) (prms []*param) {
	for _, f := range fields {
		prm := &param{}
		for _, i := range f.Names {
			prm = &param{
				name: i.Name,
			}
			prms = append(prms, prm)
		}
		if f.Type != nil {
			prm.typ = p.parseExpr(f.Type)
		}
		if len(f.Names) == 0 {
			prms = append(prms, prm)
		}
	}
	return
}

func (p *funcparse) typ(expr ast.Expr, ellip, star, pkg string) *typ {
	switch v := expr.(type) {
	case *ast.Ellipsis:
		return p.typ(v.Elt, `...`, star, pkg)
	case *ast.StarExpr:
		return p.typ(v.X, ellip, `*`, pkg)
	case *ast.SelectorExpr:
		return p.typ(v.Sel, ellip, star, v.X.(*ast.Ident).Name)
	case *ast.Ident:
		if pkg != `` {
			p.prefixes[pkg] = struct{}{}
		}
		return &typ{
			ellipsis:  ellip,
			hasType:   p.hasType,
			star:      star,
			pkgIfNone: p.pkgIfNone,
			pkg:       pkg,
			name:      v.Name,
		}
	}
	return nil
}

func (p *funcparse) typChan(expr ast.Expr, ellip, star string) *typChan {
	switch v := expr.(type) {
	case *ast.Ellipsis:
		return p.typChan(v.Elt, `...`, star)
	case *ast.StarExpr:
		return p.typChan(v.X, ellip, `*`)
	case *ast.ChanType:
		c := &typChan{
			ellipsis: ellip,
			star:     star,
			typ:      p.parseExpr(v.Value),
		}
		if v.Dir == ast.RECV {
			c.recv = `<-`
		} else if v.Dir == ast.SEND {
			c.send = `<-`
		}
		return c
	}
	return nil
}

func (p *funcparse) typeInterface(expr ast.Expr, ellip, star string) *typInterface {
	switch v := expr.(type) {
	case *ast.Ellipsis:
		return p.typeInterface(v.Elt, `...`, star)
	case *ast.StarExpr:
		return p.typeInterface(v.X, ellip, `*`)
	case *ast.InterfaceType:
		return &typInterface{
			ellipsis: ellip,
			star:     star,
		}
	}
	return nil
}

func (p *funcparse) typMap(expr ast.Expr, ellip, star string) *typMap {
	fn := p.typMap
	_ = fn
	switch v := expr.(type) {
	case *ast.Ellipsis:
		return p.typMap(v.Elt, `...`, star)
	case *ast.StarExpr:
		return p.typMap(v.X, ellip, `*`)
	case *ast.MapType:
		return &typMap{
			ellipsis: ellip,
			star:     star,
			key:      p.parseExpr(v.Key),
			typ:      p.parseExpr(v.Value),
		}
	}
	return nil
}

func (p *funcparse) typSlice(expr ast.Expr, ellip, star string) *typSlice {
	switch v := expr.(type) {
	case *ast.Ellipsis:
		return p.typSlice(v.Elt, `...`, star)
	case *ast.StarExpr:
		return p.typSlice(v.X, ellip, `*`)
	case *ast.ArrayType:
		return &typSlice{
			ellipsis: ellip,
			star:     star,
			typ:      p.parseExpr(v.Elt),
		}
	}
	return nil
}

type param struct {
	name string
	typ  typeExpr
}

func (p param) string() string {
	if p.typ != nil {
		return strings.Join(stringx.NotEmpty(p.name, p.typ.string()), ` `)
	}
	return p.name
}

type typ struct {
	ellipsis  string // ellipsis expression, `...` if set or empty
	hasType   typecheck.HasType
	star      string  // star expression, `*` if set or empty
	pkgIfNone *string // set the package name if empty
	pkg       string  // package to which the type belongs
	name      string  // type name
}

func (t typ) string() string {
	pkg := t.pkg
	if pkg != `` {
		pkg = pkg + `.`
	} else if *t.pkgIfNone != `` && t.hasType(t.name) {
		pkg = *t.pkgIfNone + `.`
	}
	return t.ellipsis + t.star + pkg + t.name
}

type typChan struct {
	ellipsis string // ellipsis expression, `...` if set or empty
	star     string // star expression, `*`  if setor empty
	recv     string
	send     string
	typ      typeExpr
}

func (t typChan) string() string {
	return t.ellipsis + t.star + t.recv + `chan` + t.send + ` ` + t.typ.string()
}

type typInterface struct {
	ellipsis string // ellipsis expression, `...` if set or empty
	star     string // star expression, `*` if set or empty
}

func (t typInterface) string() string {
	return t.ellipsis + t.star + `interface{}`
}

type typMap struct {
	ellipsis string // ellipsis expression, `...` if set or empty
	star     string // star expression, `*` if set or empty
	key      typeExpr
	typ      typeExpr
}

func (t typMap) string() string {
	return t.ellipsis + t.star + `map[` + t.key.string() + `]` + t.typ.string()
}

type typSlice struct {
	ellipsis string // ellipsis expression, `...` if set or empty
	star     string // star expression, `*` if set or empty
	typ      typeExpr
}

func (t typSlice) string() string {
	return t.ellipsis + t.star + `[]` + t.typ.string()
}

type typeExpr interface {
	string() string
}
