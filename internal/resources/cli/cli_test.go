package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/dexterp/ifaces/internal/resources/testtools/testpaths"

	"github.com/stretchr/testify/assert"
)

var (
	pre       = `PreIface`
	post      = `PostIface`
	pkg       = "mypkg"
	file      = "mypkg.go"
	comment   = "DO NOT EDIT. This file has been generated"
	matchType = "*MyStruct*"
	stdout    = &bytes.Buffer{}
	stderr    = &bytes.Buffer{}
)

func TestParseArgs_Type_Manditory(t *testing.T) {
	cmd := []string{"ifaces", "type"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdType)
	assert.Zero(t, len(args.Srcs))
	assert.Equal(t, defaultCmt, args.Cmt)
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}

func TestParseArgs_Type_Optional(t *testing.T) {
	generatedsrc := filepath.Join(testpaths.TempDir(), pkg, file)
	pkg = `otherpkg`
	cmd := []string{"ifaces", "type", "-f", "src.go", generatedsrc, "-p", pkg, "-a", "--pre", pre, "--post", post, "-c", comment, "-t", matchType, "--no-fdoc", "--no-tdoc", "--print"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdType)
	assert.NotZero(t, len(args.Srcs))
	assert.True(t, args.Append, `args.Append not set`)
	assert.Equal(t, pre, args.Pre, `args.Pre incorrect`)
	assert.Equal(t, post, args.Post, `args.Post incorrect`)
	assert.Equal(t, comment, args.Cmt, `args.Cmt incorrect`)
	assert.Equal(t, matchType, args.MatchType, `args.Wild incorrect`)
	assert.Equal(t, pkg, args.Pkg, `args.Pkg incorrect`)
	assert.True(t, args.NoFDoc)
	assert.True(t, args.NoTDoc)
	assert.True(t, args.Print, `args.Print not set`)
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}

func TestParseArgs_Method_Manditory(t *testing.T) {
	cmd := []string{"ifaces", "method"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdMethod)
	assert.Zero(t, len(args.Srcs))
	assert.Equal(t, defaultCmt, args.Cmt)
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}

func TestParseArgs_Method_Optional(t *testing.T) {
	generatedsrc := filepath.Join(testpaths.TempDir(), pkg, file)
	pkg = `otherpkg`
	cmd := []string{"ifaces", "method", generatedsrc, "-f", "src.go", "-p", pkg, "-a", "-m", "MyFunc", "--pre", pre, "--post", post, "-c", comment, "-t", matchType, "--no-fdoc", "--no-tdoc", "--print"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.CmdMethod)
	assert.NotZero(t, len(args.Srcs))
	assert.Equal(t, "MyFunc", args.MatchFunc)
	assert.True(t, args.Append, `args.Append not set`)
	assert.Equal(t, pre, args.Pre, `args.Pre incorrect`)
	assert.Equal(t, post, args.Post, `args.Post incorrect`)
	assert.Equal(t, comment, args.Cmt, `args.Cmt incorrect`)
	assert.Equal(t, matchType, args.MatchType, `args.Wild incorrect`)
	assert.Equal(t, pkg, args.Pkg, `args.Pkg incorrect`)
	assert.True(t, args.NoFDoc)
	assert.True(t, args.NoTDoc)
	assert.True(t, args.Print, `args.Print not set`)
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}
