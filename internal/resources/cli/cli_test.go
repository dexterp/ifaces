package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/dexterp/ifaces/internal/resources/testtools/testpaths"

	"github.com/stretchr/testify/assert"
)

var (
	pre            = `PreIface`
	post           = `PostIface`
	pkg            = "mypkg"
	file           = "mypkg.go"
	comment        = "DO NOT EDIT. This file has been generated"
	commentDefault = `DO NOT EDIT. GENERATED BY ifaces`
	wild           = "*MyStruct*"
	stdout         = &bytes.Buffer{}
	stderr         = &bytes.Buffer{}
)

func TestParseArgs_Gen_Manditory(t *testing.T) {
	cmd := []string{"ifaces"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, commentDefault, args.Cmt)
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}

func TestParseArgs_Gen_Optional(t *testing.T) {
	generatedsrc := filepath.Join(testpaths.TempDir(), pkg, file)
	pkg = `otherpkg`
	cmd := []string{"ifaces", generatedsrc, "-p", pkg, "-a", "--pre", pre, "--post", post, "-c", comment, "-m", wild, "--print"}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.Append, `args.Append not set`)
	assert.Equal(t, pre, args.Pre, `args.Pre incorrect`)
	assert.Equal(t, post, args.Post, `args.Post incorrect`)
	assert.Equal(t, comment, args.Cmt, `args.Cmt incorrect`)
	assert.Equal(t, wild, args.Match, `args.Wild incorrect`)
	assert.Equal(t, pkg, args.Pkg, `args.Pkg incorrect`)
	assert.True(t, args.Print, `args.Print not set`)
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}

func TestParseArgs_Gen_OptionalDynamicPkg(t *testing.T) {
	generatedsrc := filepath.Join(testpaths.TempDir(), pkg, file)
	cmd := []string{"ifaces", generatedsrc, "-a", "--pre", pre, "--post", post, "-c", comment, "-m", wild}
	args, err := ParseArgs(cmd[1:], ``, stdout, stderr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, args.Append, `args.Append not set`)
	assert.Equal(t, pre, args.Pre, `args.Pre incorrect`)
	assert.Equal(t, post, args.Post, `args.Post incorrect`)
	assert.Equal(t, comment, args.Cmt, `args.Cmt incorrect`)
	assert.Equal(t, wild, args.Match, `args.Wild incorrect`)
	assert.Equal(t, pkg, args.Pkg, `args.Pkg incorrect`)
	assert.Zero(t, stdout.String())
	assert.Zero(t, stderr.String())
}
