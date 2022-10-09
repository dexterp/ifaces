package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecvToFuncDecl(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`Func`, ``, ``)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ParamsGroup(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ParamsGroup`, `a, b string`, ``)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ParamsChan(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ParamsChan`, `c chan string`, ``)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ParamsChanRecv(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ParamsChan`, `c *<-chan string`, ``)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ParamsChanSend(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ParamsChan`, `c ...*chan<- string`, ``)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ParamsMap(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ParamsMap`, `d map[string]interface{}`, ``)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ParamsMapMap(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ParamsMapMap`, `d map[string]map[string]string`, ``)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ParamsMapSlice(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ParamsMapSlice`, `d ...*map[string][]string`, ``)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ReturnsGroupSlice(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ReturnsGroupSlice`, `in ...string`, `a, b, c []string`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ReturnsSliceMap(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ReturnsSliceMap`, ``, `*[]map[string]MyType`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ReturnsSliceSlice(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ReturnsSliceSlice`, ``, `[][]MyType`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestRecvToFunc_ReturnsPkgParam(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`ReturnsPkgParam`, ``, `pkg.MyType`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	f := RecvToFunc(astFuncDecl, hasTypeMock(``))
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_AddPkg(t *testing.T) {
	astFuncDecl, inSig, err := makeFuncType(`AddPkg`, `a []*MyType, b ...[]map[*MyType]*NoType`, `MyType`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	expected := `AddPkg(a []*pkg.MyType, b ...[]map[*pkg.MyType]*NoType) pkg.MyType`
	pkg := `pkg`
	typ := `MyType`
	f := RecvToFunc(astFuncDecl, hasTypeMock(typ))
	assert.Equal(t, expected, f.Package(pkg).String())
	assert.NotEqual(t, inSig, f.Package(pkg).String())
}

func makeFuncType(inFunc, inParams, inRet string) (*ast.FuncDecl, string, error) {
	inRetWrap := inRet
	if strings.Contains(inRetWrap, `,`) || strings.Contains(inRetWrap, " ") {
		inRetWrap = `(` + inRetWrap + `)`
	}
	if inRetWrap != `` {
		inRetWrap = ` ` + inRetWrap
	}
	sigIface := fmt.Sprintf(`%s(%s)%s`, inFunc, inParams, inRetWrap)
	sigRecv := fmt.Sprintf(`func (t MyType) %s(%s)%s`, inFunc, inParams, inRetWrap)
	src := fmt.Sprintf(`package pkg

type MyType struct {
}

%s {
}
`, sigRecv)
	fset := token.NewFileSet()
	n, err := parser.ParseFile(fset, `pkg.go`, src, 0)
	if err != nil {
		return nil, ``, err
	}
	var astFuncDecl *ast.FuncDecl
	ast.Inspect(n, func(c ast.Node) bool {
		x, ok := c.(*ast.FuncDecl)
		if ok {
			astFuncDecl = x
			return false
		}
		return true
	})
	if astFuncDecl == nil {
		return nil, ``, errors.New(`unable to parse for *ast.FuncDecl`)
	}
	return astFuncDecl, sigIface, nil
}

func hasTypeMock(typ string) func(string) bool {
	return func(chk string) bool {
		return typ == chk
	}
}
