package cli

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/dexterp/ifaces/internal/resources/testtools/testpaths"

	"github.com/stretchr/testify/assert"
)

var (
	prefix    = `PreIface`
	suffix    = `PostIface`
	pkg       = "mypkg"
	file      = "mypkg.go"
	matchType = "MyStruct"
	stdout    = &bytes.Buffer{}
	stderr    = &bytes.Buffer{}
)

func TestParseArgs_Struct_Manditory(t *testing.T) {
	cmd := []string{"ifaces", "struct"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdStruct)
}

func TestParseArgs_Struct_Optional(t *testing.T) {
	cmd := []string{"ifaces", "struct", "-e", prefix, "-s", suffix}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdStruct)
}

func TestParseArgs_Type_Manditory(t *testing.T) {
	cmd := []string{"ifaces", "type", "-i", "Iface"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdType)
	assert.Zero(t, len(args.Src))
	assert.Zero(t, stdout.String())
}

func TestParseArgs_Type_Optional(t *testing.T) {
	generatedsrc := filepath.Join(testpaths.TempDir(), pkg, file)
	pkg = `otherpkg`
	cmd := []string{"ifaces", "type", "-i", "Iface", "-o", generatedsrc, "-x", "github.com/stretchr/testify", "-p", pkg, "-a", "--nmethod", "-t", matchType, "--ntdoc", "--nfdoc", "-d"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	fmt.Print(stderr.String())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdType)
	assert.True(t, args.Append, `args.Append not set`)
	assert.Equal(t, pkg, args.Pkg, `args.Pkg incorrect`)
	assert.True(t, args.NoFDoc)
	assert.True(t, args.NoTDoc)
	assert.True(t, args.Print, `args.Print not set`)
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}

func TestParseArgs_Func_Manditory(t *testing.T) {
	cmd := []string{"ifaces", "func", "-i", "Iface"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdFunc)
	assert.Zero(t, len(args.Src))
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}

func TestParseArgs_Func_Optional(t *testing.T) {
	generatedsrc := filepath.Join(testpaths.TempDir(), pkg, file)
	pkg = `otherpkg`
	cmd := []string{"ifaces", "func", "-o", generatedsrc, "-f", "src.go", "-p", pkg, "-a", "-i", "Iface", "-m", "MyFunc", "-t", matchType, "--nfdoc", "-d"}
	fmt.Println(cmd)
	fmt.Print(usage(cmd[1:]))
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	fmt.Println(stderr.String())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdFunc)
	assert.NotZero(t, len(args.Src))
	assert.Equal(t, "MyFunc", args.MatchFunc)
	assert.True(t, args.Append, `args.Append not set`)
	assert.Equal(t, matchType, args.MatchType, `args.Wild incorrect`)
	assert.Equal(t, pkg, args.Pkg, `args.Pkg incorrect`)
	assert.True(t, args.NoFDoc)
	assert.True(t, args.Print, `args.Print not set`)
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}
