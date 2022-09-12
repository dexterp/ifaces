package print

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	bufStderr = &bytes.Buffer{}
	bufStdout = &bytes.Buffer{}
	print     = New(Options{
		Stderr: bufStderr,
		Stdout: bufStdout,
	})
)

func TestPrint_Errorf(t *testing.T) {
	print.Level(ERROR)
	print.Errorf("this is an %s\n", `error`)
	assert.Equal(t, "this is an error\n", bufStderr.String())
	print.Level(ERROR + 1)
	bufStderr.Reset()
	print.Errorf("this is an %s\n", `error`)
	assert.Equal(t, "", bufStderr.String())
}

func TestPrint_Errorln(t *testing.T) {
	print.Level(ERROR)
	print.Errorln(`this is an`, `error`)
	assert.Equal(t, "this is an error\n", bufStderr.String())
	print.Level(ERROR + 1)
	bufStderr.Reset()
	print.Errorln("this is an", `error`)
	assert.Equal(t, "", bufStderr.String())
}

func TestPrint_Warn(t *testing.T) {
	print.Level(WARN)
	print.Warn(`this`, `is`, `an`, `error`)
	assert.Equal(t, "this is an error\n", bufStderr.String())
	print.Level(WARN + 1)
	bufStderr.Reset()
	print.Warn(`this`, `is`, `an`, `error`)
	assert.Equal(t, "", bufStderr.String())
}

func TestPrint_Warnf(t *testing.T) {
	print.Level(WARN)
	print.Warnf("this is an %s\n", `error`)
	assert.Equal(t, "this is an error\n", bufStderr.String())
	print.Level(WARN + 1)
	bufStderr.Reset()
	print.Warnf("this is an %s\n", `error`)
	assert.Equal(t, "", bufStderr.String())
}
