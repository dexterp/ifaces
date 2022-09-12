package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	src = `package parser

import (
	"io"
)

//` + `go:generate ifaces head /tmp/pkg/pkg_ifaces.go

//` + `go:generate ifaces entry /tmp/pkg/pkg_ifaces.go --post Parser 

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

`
	struct1 = `Parser`
	line1   = 12
)

func TestParser_GetType(t *testing.T) {
	p, err := Parse(`src.go`, []byte(src))
	if err != nil {
		t.Error(err)
	}
	typ := p.GetType(9)
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
	for _, recv := range recvs {
		r := recv.Signature()
		fmt.Println(r)
	}
}

func TestParser_Imports(t *testing.T) {
	p, err := Parse(`src.go`, []byte(src))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(p.Imports()), 1)
}
