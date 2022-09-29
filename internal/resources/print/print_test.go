package print

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	bufStderr = &bytes.Buffer{}
	bufStdout = &bytes.Buffer{}
	print     = New(Options{
		Stderr: bufStderr,
		Stdout: bufStdout,
		Exit:   PANIC,
	})
)

func TestPrint_Errorf(t *testing.T) {
	bufStderr.Reset()
	print.Level(ERROR)
	print.Errorf("this is an %s\n", `error`)
	assert.Equal(t, "this is an error\n", bufStderr.String())
	print.Level(ERROR + 1)
	bufStderr.Reset()
	print.Errorf("this is an %s\n", `error`)
	assert.Equal(t, "", bufStderr.String())
}

func TestPrint_HasErrorf(t *testing.T) {
	bufStderr.Reset()
	print.Level(ERROR)
	assert.False(t, print.HasErrorf("should not print: %v", `error`))
	assert.Zero(t, bufStderr.String())
	bufStderr.Reset()
	assert.True(t, print.HasErrorf("should print: %v", errors.New(`error`)))
	assert.Contains(t, bufStderr.String(), `error`)
	fmt.Println(bufStderr.String())
}

func TestPrint_Errorln(t *testing.T) {
	bufStderr.Reset()
	print.Level(ERROR)
	print.Errorln(`this is an`, `error`)
	assert.Equal(t, "this is an error\n", bufStderr.String())
	print.Level(ERROR + 1)
	bufStderr.Reset()
	print.Errorln("this is an", `error`)
	assert.Equal(t, "", bufStderr.String())
}

func TestPrint_HasErrorln(t *testing.T) {
	bufStderr.Reset()
	print.Level(ERROR)
	assert.False(t, print.HasErrorln(`should not print: `, `error`))
	assert.Zero(t, bufStderr.String())
	bufStderr.Reset()
	assert.True(t, print.HasErrorln(`should print: `, errors.New(`error`)))
	assert.Contains(t, bufStderr.String(), `error`)
	fmt.Println(bufStderr.String())
}

func TestPrint_Warnln(t *testing.T) {
	bufStderr.Reset()
	print.Level(WARN)
	print.Warnln(`this`, `is`, `an`, `error`)
	assert.Equal(t, "this is an error\n", bufStderr.String())
	print.Level(WARN + 1)
	bufStderr.Reset()
	print.Warnln(`this`, `is`, `an`, `error`)
	assert.Equal(t, "", bufStderr.String())
}

func TestPrint_HasWarnln(t *testing.T) {
	bufStderr.Reset()
	print.Level(WARN)
	assert.False(t, print.HasWarnln("should not print", `error`))
	assert.Zero(t, bufStderr.String())
	bufStderr.Reset()
	assert.True(t, print.HasWarnln("should print", errors.New(`error`)))
	assert.Contains(t, bufStderr.String(), `error`)
}

func TestPrint_Warnf(t *testing.T) {
	bufStderr.Reset()
	print.Level(WARN)
	print.Warnf("this is an %s\n", `error`)
	assert.Equal(t, "this is an error\n", bufStderr.String())
	print.Level(WARN + 1)
	bufStderr.Reset()
	print.Warnf("this is an %s\n", errors.New(`error`))
	assert.Zero(t, bufStderr.String())
}

func TestPrint_HasWarnf(t *testing.T) {
	bufStderr.Reset()
	print.Level(WARN)
	assert.False(t, print.HasWarnf("should not print: %v", `error`))
	assert.Zero(t, bufStderr.String())
	bufStderr.Reset()
	assert.True(t, print.HasWarnf("should print: %f", errors.New(`error`)))
	assert.Contains(t, bufStderr.String(), `error`)
	fmt.Println(bufStderr.String())
}

func TestPrint_Fatalln(t *testing.T) {
	bufStderr.Reset()
	assert.Panics(t, func() {
		print.Fatalln(`this is the end`)
	})
	assert.Equal(t, "this is the end\n", bufStderr.String())
}

func TestPrint_HasFatalln(t *testing.T) {
	bufStderr.Reset()
	assert.NotPanics(t, func() {
		print.HasFatalln(`this is the end `, `error`)
	})
	assert.Zero(t, bufStderr.String())
	bufStderr.Reset()
	assert.Panics(t, func() {
		print.HasFatalln(`this is the end `, errors.New(`error`))
	})
	assert.Contains(t, bufStderr.String(), `error`)
}

func TestPrint_Fatalf(t *testing.T) {
	bufStderr.Reset()
	assert.Panics(t, func() {
		print.Fatalf(`this is the end: %v`, `error`)
	})
	assert.Contains(t, bufStderr.String(), `error`)
}

func TestPrint_HasFatalf(t *testing.T) {
	bufStderr.Reset()
	assert.NotPanics(t, func() {
		print.HasFatalf(`this is the end: %v`, `error`)
	})
	assert.Zero(t, bufStderr.String())
	bufStderr.Reset()
	assert.Panics(t, func() {
		print.HasFatalf(`this is the end: %v`, errors.New(`error`))
	})
	assert.Contains(t, bufStderr.String(), `error`)
}
