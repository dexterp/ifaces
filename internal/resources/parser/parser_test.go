package parser

import (
	"fmt"
	"testing"

	"github.com/dexterp/ifaces/internal/resources/types"
	"github.com/stretchr/testify/assert"
)

func varLine() int {
	return 12
}

func varPkg() string {
	return `mypkg`
}

func varSrc() string {
	return fmt.Sprintf(`package %s

import (
	"io"
)

//`+`go:generate ifaces type /tmp/pkg/pkg_ifaces.go

//`+`go:generate ifaces type /tmp/pkg/pkg_ifaces.go --post Parser 

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

`, varPkg())
}

func varStruct1() string {
	return `Parser`
}

func varIfaceSrc() string {
	return fmt.Sprintf(`// DO NOT EDIT. Generated by ifaces

package %s

type PrintIface interface {
	// Errorf error print formatter, writes to stderr
	Errorf(format string, a ...any)
	// WarnF warning print formatter, writes to stderr
	Warnf(format string, a ...any)
}
	`, varPkg())
}

func TestParser_GetIfaceMethods(t *testing.T) {
	p, err := Parse(`src_ifaces.go`, []byte(varIfaceSrc()))
	assert.NoError(t, err)
	ifaces := p.GetTypesByType(types.INTERFACE)
	if !assert.Len(t, ifaces, 1) {
		t.FailNow()
	}
	name := ifaces[0].Name()
	assert.Equal(t, `PrintIface`, name)
	methods := p.GetIfaceMethods(name)
	if !assert.Len(t, methods, 2) {
		t.FailNow()
	}
	assert.Equal(t, `Errorf(format string, a ...any)`, methods[0].Signature())
	assert.Equal(t, `Warnf(format string, a ...any)`, methods[1].Signature())
}

func TestParser_GetTypeByLine(t *testing.T) {
	p, err := Parse(`src.go`, []byte(varSrc()))
	if err != nil {
		t.Error(err)
	}
	typ := p.GetTypeByLine(7)
	assert.Nil(t, typ)
	typ = p.GetTypeByLine(9)
	if !assert.NotNil(t, typ) {
		t.FailNow()
	}
	assert.NotEmpty(t, typ.Doc(), `type document is empty`)
	assert.Equal(t, varStruct1(), typ.Name(), `invalid type name`)
	assert.Equal(t, varLine(), typ.Line(), `wrong line number`)
	assert.Equal(t, types.STRUCT, typ.Type(), `wrong type`)
}

func TestParser_GetTypeRecvs(t *testing.T) {
	p, err := Parse(`src.go`, []byte(varSrc()))
	if err != nil {
		t.Error(err)
	}
	recvs := p.GetTypeRecvs(varStruct1())
	if assert.Equal(t, 3, len(recvs)) {
		assert.Regexp(t, `Parse\(.*\)`, recvs[0].Signature())
		assert.Regexp(t, `Count\(\)`, recvs[1].Signature())
		assert.Regexp(t, `Add\(.*\)`, recvs[2].Signature())
	}
}

func TestParser_Imports(t *testing.T) {
	p, err := Parse(`src.go`, []byte(varSrc()))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(p.Imports()), 1)
}

func TestParser_Package(t *testing.T) {
	p, err := Parse(`src.go`, []byte(varSrc()))
	assert.NoError(t, err)
	assert.Equal(t, varPkg(), p.Package())
}

func TestParser_parseGeneratorCmts(t *testing.T) {
	p, err := Parse(`src.go`, []byte(varSrc()))
	assert.NoError(t, err)
	if assert.Equal(t, 2, len(p.genCmts)) {
		assert.Equal(t, 7, p.genCmts[0].Line)
		assert.Equal(t, 9, p.genCmts[1].Line)
	}
}
