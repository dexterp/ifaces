package parsefunc

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"

	"github.com/dexterp/ifaces/internal/resources/typecheck"
)

var (
	regexSignature    = regexp.MustCompile(`^(?P<name>[A-Za-z][A-Za-z0-9_]+)(?:\[(?P<generics>[^\]]*)\])?\((?P<params>[^\)]+)?\)\s*\(?(?P<return>[^(?:$|\)]*)\)?`)
	regexTypeString   = `[A-Za-z][A-Za-z0-9]*`
	regexType         = regexp.MustCompile(`^(\.\.\.)?(\*)?(` + `(?:[a-z]+\.)?` + regexTypeString + `(?:\{\})?)$`)
	regexTypeMapOrGen = regexp.MustCompile(`^(` + regexTypeString + `)\[((?:\*?)` + regexTypeString + `)\](.*)$`)
	regexTypeSlice    = regexp.MustCompile(`^(\*)?\[\](.*)$`)
)

type FuncDecl struct {
	// NeedsImport if true then parsing has detected the need to add the current
	// package to the list of imports for the target file
	NeedsImport bool

	// Prefixes a list of prefix packages that were used by within this function
	// signature.
	Prefixes map[string]any

	hasType typecheck.HasType

	pkg        string
	name       string
	typeParams []*typeParams
	params     []*param
	returns    []*param
}

func ToFuncDecl(method, pkg string, hasType typecheck.HasType) *FuncDecl {
	name, typeParams, params, returns := parseFuncDecl(method)
	decl := &FuncDecl{
		Prefixes: map[string]any{},
		name:     name,
		hasType:  hasType,
		pkg:      pkg,
	}
	decl.typeParams = decl.parseTypeParams(typeParams)
	decl.params = decl.parseParams(params)
	decl.returns = decl.parseParams(returns)
	return decl
}

func parseFuncDecl(sig string) (name, generics, params, ret string) {
	mtchs := regexSignature.FindStringSubmatch(sig)
	if mtchs == nil {
		return ``, ``, ``, ``
	}
	return mtchs[1], mtchs[2], mtchs[3], mtchs[4]
}

func (f FuncDecl) parseParams(paramsstr string) (params []*param) {
	var (
		paramsList [][]string
		maxlen     int
	)
	if paramsstr == `` {
		return
	}
	for _, p := range strings.Split(paramsstr, `,`) {
		p = strings.TrimSpace(p)
		pair := strings.Split(p, ` `)
		paramsList = append(paramsList, pair)
		if len(pair) > maxlen {
			maxlen = len(pair)
		}
	}

	for _, pair := range paramsList {
		if maxlen == 2 { // named params
			switch len(pair) {
			case 1:
				params = append(params, &param{
					hastype: f.hasType,
					pkg:     f.pkg,
					name:    pair[0],
				})
			case 2:
				params = append(params, &param{
					hastype:  f.hasType,
					pkg:      f.pkg,
					name:     pair[0],
					typ:      f.parseTyp(pair[1]),
					typMap:   f.parseTypMap(pair[1]),
					typSlice: f.parseTypSlice(pair[1]),
				})
			}
		} else {
			params = append(params, &param{
				hastype:  f.hasType,
				pkg:      f.pkg,
				typ:      f.parseTyp(pair[0]),
				typMap:   f.parseTypMap(pair[0]),
				typSlice: f.parseTypSlice(pair[0]),
			})
		}
	}
	return params
}

func (f *FuncDecl) parseTypeParams(gen string) (generics []*typeParams) {
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
			curType := f.parseTyp(t)
			if curType != nil {
				alternates = append(alternates, curType)
			}
		}
		generics = append(generics, &typeParams{
			hastype: f.hasType,
			pkg:     f.pkg,
			name:    name,
			typ:     alternates,
		})
	}
	return generics
}

func (f *FuncDecl) parseTyp(t string) *typ {
	mtch := regexType.FindStringSubmatch(t)
	if mtch == nil {
		return nil
	}
	if mtch[3] == `interface{}` {
		return &typ{
			iface: true,
		}
	}
	if strings.Contains(mtch[3], `.`) {
		typSlc := strings.Split(mtch[3], `.`)
		if len(typSlc) == 2 {
			f.Prefixes[typSlc[0]] = mtch[3]
		}
	} else if f.hasType(mtch[3]) {
		f.NeedsImport = true
	}
	return &typ{
		pkg:      f.pkg,
		hastype:  f.hasType,
		variadic: mtch[1],
		ptr:      mtch[2],
		typ:      mtch[3],
	}
}

func (f *FuncDecl) parseTypMap(m string) *typMap {
	mtch := regexTypeMapOrGen.FindStringSubmatch(m)
	if mtch == nil {
		return nil
	}
	if mtch[1] != `map` {
		return nil
	}
	return &typMap{
		key:      f.parseTyp(mtch[2]),
		elmType:  f.parseTyp(mtch[3]),
		elmMap:   f.parseTypMap(mtch[3]),
		elmSlice: f.parseTypSlice(mtch[3]),
	}
}

func (f *FuncDecl) parseTypSlice(s string) *typSlice {
	mtch := regexTypeSlice.FindStringSubmatch(s)
	if mtch == nil {
		return nil
	}
	return &typSlice{
		ptr:      mtch[1],
		elmType:  f.parseTyp(mtch[2]),
		elmMap:   f.parseTypMap(mtch[2]),
		elmSlice: f.parseTypSlice(mtch[2]),
	}
}

func (f FuncDecl) UsesTypeParams() bool {
	return f.typeParams != nil
}

func (f FuncDecl) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString(f.name)
	buf.WriteString(f.stringTypeParams())
	buf.WriteString(f.stringParams())
	buf.WriteString(f.stringReturns())
	return buf.String()
}

func (f FuncDecl) stringTypeParams() string {
	buf := &bytes.Buffer{}
	if len(f.typeParams) > 0 {
		items := []string{}
		for _, g := range f.typeParams {
			items = append(items, g.String())
		}
		buf.WriteString(`[` + strings.Join(items, `, `) + `]`)
	}
	return buf.String()
}

func (f FuncDecl) stringParams() string {
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

func (f FuncDecl) stringReturns() string {
	buf := &bytes.Buffer{}
	switch len(f.returns) {
	case 0:
		return ``
	case 1:
		if strings.Contains(f.returns[0].string(), " ") {
			buf.WriteString(` (` + f.returns[0].string() + `)`)
		} else {
			buf.WriteString(` ` + f.returns[0].string())
		}
	default:
		l := []string{}
		for _, p := range f.returns {
			l = append(l, p.string())
		}
		buf.WriteString(` (` + strings.Join(l, `, `) + `)`)
	}
	return buf.String()
}

type param struct {
	hastype  typecheck.HasType
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

type typeParams struct {
	pkg     string
	hastype typecheck.HasType
	name    string
	typ     []*typ
}

func (g typeParams) Types() (types []string) {
	if g.typ == nil {
		return
	}
	for _, t := range g.typ {
		types = append(types, t.string())
	}
	return types
}

func (g typeParams) String() string {
	return g.name + ` ` + strings.Join(g.Types(), ` | `)
}

type typ struct {
	pkg      string // original package name
	hastype  typecheck.HasType
	variadic string // variadic paramater
	ptr      string // asterisk or empty
	typ      string // type name
	iface    bool
}

func (t typ) string() string {
	if t.iface {
		return `interface{}`
	}
	if t.pkg != `` && !strings.Contains(t.typ, `.`) && t.hastype(t.typ) {
		return t.variadic + t.ptr + t.pkg + `.` + t.typ
	}
	return t.variadic + t.ptr + t.typ
}

type typMap struct {
	key      *typ
	elmType  *typ
	elmMap   *typMap
	elmSlice *typSlice
}

func (tm typMap) string() string {
	buf := &bytes.Buffer{}
	buf.WriteString(`map[` + tm.key.string() + `]`)
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
