package parsefunc

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testType string

func TestToFuncDecl(t *testing.T) {
	inSig := makeFuncType(`Func`, ``, ``, ``)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_TypeParams(t *testing.T) {
	inSig := makeFuncType(`Func`, `C comparable`, `str string`, `count uint`)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_TypeParamsWithAlternates(t *testing.T) {
	inSig := makeFuncType(`Sort`, `K comparable, Y uint8 | uint16`, `i K, x Y`, `error`)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.True(t, f.UsesTypeParams())
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_ParamsGroup(t *testing.T) {
	inSig := makeFuncType(`ParamsGroup`, ``, `a, b string`, ``)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_ParamsMap(t *testing.T) {
	inSig := makeFuncType(`ParamsMap`, ``, `d map[string]interface{}`, ``)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_ParamsMapMap(t *testing.T) {
	inSig := makeFuncType(`ParamsMapMap`, ``, `d map[string]map[string]string`, ``)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_ParamsMapSlice(t *testing.T) {
	inSig := makeFuncType(`ParamsMapSlice`, ``, `d map[string][]string`, ``)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_ReturnsGroupSlice(t *testing.T) {
	inSig := makeFuncType(`ReturnsGroupSlice`, ``, `in string`, `a, b, c []string`)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_ReturnsSliceMap(t *testing.T) {
	inSig := makeFuncType(`ReturnsSliceMap`, ``, ``, `[]map[string]MyType`)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_ReturnsSliceSlice(t *testing.T) {
	inSig := makeFuncType(`ReturnsSliceSlice`, ``, ``, `[][]MyType`)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_ReturnsPkgParam(t *testing.T) {
	inSig := makeFuncType(`ReturnsPkgParam`, ``, ``, `pkg.MyType`)
	f := ToFuncDecl(inSig, ``, testHasType)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_AddPkg(t *testing.T) {
	inSig := `AddPkg(a []*MyType, b []map[MyType]*NoType) MyType`
	expected := `AddPkg(a []*pkg.MyType, b []map[pkg.MyType]*NoType) pkg.MyType`
	pkg := `pkg`
	testType = `MyType`
	f := ToFuncDecl(inSig, pkg, testHasType)
	assert.Equal(t, expected, f.String())
}

func TestParseSig(t *testing.T) {
	inFunc := `Connect`
	inParams := `host string, port string`
	inRets := `*Conn, error`
	inSig := fmt.Sprintf(`%s(%s)%s`, inFunc, inParams, inRets)
	outFunc, outGenerics, outParams, outRets := parseFuncDecl(inSig)
	assert.Equal(t, inFunc, outFunc)
	assert.Equal(t, ``, outGenerics)
	assert.Equal(t, inParams, outParams)
	assert.Equal(t, inRets, outRets)
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
	if inRetWrap != `` {
		inRetWrap = ` ` + inRetWrap
	}
	return fmt.Sprintf(`%s%s(%s)%s`, inFunc, inGenericsWrap, inParams, inRetWrap)
}

func testHasType(typ string) bool {
	return typ == testType
}
