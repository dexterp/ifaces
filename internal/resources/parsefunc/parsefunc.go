package parsefunc

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

var (
	regexSignature    = regexp.MustCompile(`^(?P<name>[A-Za-z][A-Za-z0-9_]+)(?:\[(?P<generics>[^\]]*)\])?\((?P<params>[^\)]+)?\)\s*\(?(?P<return>[^(?:$|\)]*)\)?`)
	regexTypeString   = `[A-Za-z][A-Za-z0-9]*`
	regexType         = regexp.MustCompile(`^(\.\.\.)?(\*)?(` + `(?:[a-z]+\.)?` + regexTypeString + `(?:\{\})?)$`)
	regexTypeMapOrGen = regexp.MustCompile(`^(` + regexTypeString + `)\[((?:\*?)` + regexTypeString + `)\](.*)$`)
	regexTypeSlice    = regexp.MustCompile(`^(\*)?\[\](.*)$`)
)

func ToFuncDecl(pkg string, has hasType, method string) *FuncDecl {
	name, typeParams, params, returns := parseFuncDecl(method)
	decl := &FuncDecl{
		name:       name,
		typeParams: parseTypeParams(pkg, has, typeParams),
		params:     parseParams(pkg, has, params),
		returns:    parseParams(pkg, has, returns),
	}
	return decl
}

func parseFuncDecl(sig string) (name, generics, params, ret string) {
	mtchs := regexSignature.FindStringSubmatch(sig)
	if mtchs == nil {
		return ``, ``, ``, ``
	}
	return mtchs[1], mtchs[2], mtchs[3], mtchs[4]
}

func parseParams(pkg string, has hasType, paramsstr string) (params []*param) {
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
					hastype: has,
					pkg:     pkg,
					name:    pair[0],
				})
			case 2:
				params = append(params, &param{
					hastype:  has,
					pkg:      pkg,
					name:     pair[0],
					typ:      parseTyp(pkg, has, pair[1]),
					typMap:   parseTypMap(pkg, has, pair[1]),
					typSlice: parseTypSlice(pkg, has, pair[1]),
				})
			}
		} else {
			params = append(params, &param{
				hastype:  has,
				pkg:      pkg,
				typ:      parseTyp(pkg, has, pair[0]),
				typMap:   parseTypMap(pkg, has, pair[0]),
				typSlice: parseTypSlice(pkg, has, pair[0]),
			})
		}
	}
	return params
}

func parseTypeParams(pkg string, hastype hasType, gen string) (generics []*typeParams) {
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
			curType := parseTyp(pkg, hastype, t)
			if curType != nil {
				alternates = append(alternates, curType)
			}
		}
		generics = append(generics, &typeParams{
			hastype: hastype,
			pkg:     pkg,
			name:    name,
			typ:     alternates,
		})
	}
	return generics
}

func parseTyp(pkg string, hastype hasType, t string) *typ {
	mtch := regexType.FindStringSubmatch(t)
	if mtch == nil {
		return nil
	}
	if mtch[3] == `interface{}` {
		return &typ{
			iface: true,
		}
	}
	return &typ{
		pkg:      pkg,
		hastype:  hastype,
		variadic: mtch[1],
		ptr:      mtch[2],
		typ:      mtch[3],
	}
}

func parseTypMap(pkg string, hastype hasType, m string) *typMap {
	mtch := regexTypeMapOrGen.FindStringSubmatch(m)
	if mtch == nil {
		return nil
	}
	if mtch[1] != `map` {
		return nil
	}
	return &typMap{
		key:      parseTyp(pkg, hastype, mtch[2]),
		elmType:  parseTyp(pkg, hastype, mtch[3]),
		elmMap:   parseTypMap(pkg, hastype, mtch[3]),
		elmSlice: parseTypSlice(pkg, hastype, mtch[3]),
	}
}

func parseTypSlice(pkg string, hastype hasType, s string) *typSlice {
	mtch := regexTypeSlice.FindStringSubmatch(s)
	if mtch == nil {
		return nil
	}
	return &typSlice{
		ptr:      mtch[1],
		elmType:  parseTyp(pkg, hastype, mtch[2]),
		elmMap:   parseTypMap(pkg, hastype, mtch[2]),
		elmSlice: parseTypSlice(pkg, hastype, mtch[2]),
	}
}

type FuncDecl struct {
	name       string
	typeParams []*typeParams
	params     []*param
	returns    []*param
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
	hastype  hasType
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
	hastype hasType
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

type hasType interface {
	HasType(typ string) bool
}

type typ struct {
	pkg      string // original package name
	hastype  hasType
	variadic string // variadic paramater
	ptr      string // asterisk or empty
	typ      string // type name
	iface    bool
}

func (t typ) string() string {
	if t.iface {
		return `interface{}`
	}
	if t.pkg != `` && !strings.Contains(t.typ, `.`) && t.hastype.HasType(t.typ) {
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
