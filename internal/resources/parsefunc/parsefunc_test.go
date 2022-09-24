package parsefunc

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToFuncDecl(t *testing.T) {
	inSig := makeFuncType(`Func`, ``, ``, ``)
	f := ToFuncDecl(``, &testHasType{}, inSig)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_1(t *testing.T) {
	inSig := makeFuncType(`Func`, `C comparable`, `str string`, `count uint`)
	f := ToFuncDecl(``, &testHasType{}, inSig)
	assert.Equal(t, inSig, f.String())
}

func TestToFuncDecl_2(t *testing.T) {
	inSig := makeFuncType(`Sort`, `K comparable, Y uint8 | uint16`, `i K, x Y`, `error`)
	f := ToFuncDecl(``, &testHasType{}, inSig)
	assert.Equal(t, inSig, f.String())
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

type testHasType struct {
}

func (ht testHasType) HasType(typ string) bool {
	return false
}
