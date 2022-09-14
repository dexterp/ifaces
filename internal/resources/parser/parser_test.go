package parser

import (
	"fmt"
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

//go:generator ifaces item -i DataIface

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
	if assert.Equal(t, 2, len(recvs)) {
		assert.Regexp(t, `Parse\(.*\)`, recvs[0].Signature())
		assert.Regexp(t, `Count\(\)`, recvs[1].Signature())
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
