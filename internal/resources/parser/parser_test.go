package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	src = fmt.Sprintf(`package %s

import (
	"io"
)

//`+`go:generate ifaces /tmp/pkg/pkg_ifaces.go

//`+`go:generate ifaces /tmp/pkg/pkg_ifaces.go --post Parser 

// Parser Parser parses data
type Parser struct {
}

// Parse
func (p *Parser) Parse(wtr io.Writer) {
}

// Count
func (p *Parser) Count() int {
	return 0
}

// Add
func (p *Parser) Add(d1 *Data, d2 Data, d3 *pkg.Data, d4 pkg.Data) error {
}

//go:generator ifaces -i DataIface

// Data data information
type Data struct {
}

// Scan 
func (d *Data) Scan(item string) bool {
	return false
}

`, pkg)
	pkg     = `mypkg`
	struct1 = `Parser`
	line1   = 12
)

func TestParser_GetType(t *testing.T) {
	p, err := Parse(`src.go`, []byte(src))
	if err != nil {
		t.Error(err)
	}
	typ := p.GetType(7)
	assert.Nil(t, typ)
	typ = p.GetType(9)
	if !assert.NotNil(t, typ) {
		t.FailNow()
	}
	assert.NotEmpty(t, typ.Doc(), `type document is empty`)
	assert.Equal(t, struct1, typ.Name(), `invalid type name`)
	assert.Equal(t, line1, typ.Line(), `wrong line number`)
	assert.Equal(t, StructType, typ.Type(), `wrong type`)
}

func TestParser_GetRecvs(t *testing.T) {
	p, err := Parse(`src.go`, []byte(src))
	if err != nil {
		t.Error(err)
	}
	recvs := p.GetRecvs(struct1)
	if assert.Equal(t, 3, len(recvs)) {
		assert.Regexp(t, `Parse\(.*\)`, recvs[0].Signature())
		assert.Regexp(t, `Count\(\)`, recvs[1].Signature())
		assert.Regexp(t, `Add\(.*\)`, recvs[2].Signature())
	}
}

func TestParser_Imports(t *testing.T) {
	p, err := Parse(`src.go`, []byte(src))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(p.Imports()), 1)
}

func TestParser_Package(t *testing.T) {
	p, err := Parse(`src.go`, []byte(src))
	assert.NoError(t, err)
	assert.Equal(t, pkg, p.Package())
}

func TestParser_parseGeneratorCmts(t *testing.T) {
	p, err := Parse(`src.go`, []byte(src))
	assert.NoError(t, err)
	if assert.Equal(t, 2, len(p.genCmts)) {
		assert.Equal(t, 7, p.genCmts[0].Line)
		assert.Equal(t, 9, p.genCmts[1].Line)
	}
}

func TestParseSig(t *testing.T) {
	inFunc := `Connect`
	inParams := `host string, port string`
	inRets := `*Conn, error`
	inSig := fmt.Sprintf(`%s(%s)%s`, inFunc, inParams, inRets)
	outFunc, outGenerics, outParams, outRets := parseSig(inSig)
	assert.Equal(t, inFunc, outFunc)
	assert.Equal(t, ``, outGenerics)
	assert.Equal(t, inParams, outParams)
	assert.Equal(t, inRets, outRets)
}

func TestToFuncDecl(t *testing.T) {
	inSig := makeFuncType(`Func`, ``, ``, ``)
	f := toFuncDecl(``, &testHasType{}, inSig)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_1(t *testing.T) {
	inSig := makeFuncType(`Func`, `C comparable`, `str string`, `count uint`)
	f := toFuncDecl(``, &testHasType{}, inSig)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_2(t *testing.T) {
	inSig := makeFuncType(`Sort`, `K comparable, Y uint8 | uint16`, `i K, x Y`, `error`)
	f := toFuncDecl(``, &testHasType{}, inSig)
	assert.Equal(t, inSig, f.String())
}

func TestRegexSignature(t *testing.T) {
	inFunc := `Func`
	inGenerics := ``
	inParams := `count int`
	inRet := `int, error`
	inRetWrap := inRet
	if strings.Contains(inRetWrap, `,`) {
		inRetWrap = `(` + inRetWrap + `)`
	}
	sig := fmt.Sprintf(`%s%s(%s) %s`, inFunc, inGenerics, inParams, inRetWrap)
	matches := regexSignature.FindStringSubmatch(sig)
	if assert.Equal(t, 5, len(matches)) {
		assert.Equal(t, inFunc, matches[1])
		assert.Equal(t, inGenerics, matches[2])
		assert.Equal(t, inParams, matches[3])
		assert.Equal(t, inRet, matches[4])
	}
}

func makeFuncType(inFunc, inGenerics, inParams, inRet string) string {
	inRetWrap := inRet
	inGenericsWrap := inGenerics
	if inGenericsWrap != `` {
		inGenericsWrap = `[` + inGenericsWrap + `]`
	}
	if strings.Contains(inRetWrap, `,`) || strings.Contains(inRetWrap, " ") {
		inRetWrap = `(` + inRetWrap + `)`
	}
	return fmt.Sprintf(`%s%s(%s) %s`, inFunc, inGenericsWrap, inParams, inRetWrap)
}

type testHasType struct {
}

func (ht testHasType) HasType(typ string) bool {
	return false
}
