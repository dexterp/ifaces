package print

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Level int

const (
	DEBUG = iota + 1
	INFO
	WARN
	ERROR
)

type Exit int

const (
	EXIT = iota + 1
	PANIC
)

var mu = &sync.Mutex{}

//go:generate ifaces type print_iface.go --post Iface

// Print handle printing
type Print struct {
	stderr io.Writer
	stdout io.Writer
	lvl    Level
	exit   Exit
}

// Options options for New
type Options struct {
	Stderr io.Writer
	Stdout io.Writer
	Level  Level // Print level. 0 is default, less then 0 disables printing
	Exit   Exit  // Exit type. EXIT is
}

// New return a Print type
func New(opts Options) *Print {
	var lvl Level
	if opts.Level > 0 {
		lvl = opts.Level
	} else if opts.Level == 0 {
		lvl = WARN
	}
	return &Print{
		stderr: opts.Stderr,
		stdout: opts.Stdout,
		lvl:    lvl,
		exit:   opts.Exit,
	}
}

// Level set level
func (p *Print) Level(lvl Level) {
	p.lvl = lvl
}

// Errorf print error
func (p Print) Errorf(format string, a ...any) {
	if ERROR >= p.lvl {
		mu.Lock()
		defer mu.Unlock()
		fmt.Fprintf(p.stderr, format, a...)
	}
}

// HasErrorf same as Errorf function but only prints if a holds an error value.
func (p Print) HasErrorf(format string, a ...any) bool {
	return hasErr(a, func() {
		p.Errorf(format, a...)
	})
}

// Errorln print error
func (p Print) Errorln(a ...any) {
	if ERROR >= p.lvl {
		mu.Lock()
		defer mu.Unlock()
		fmt.Fprintln(p.stderr, a...)
	}
}

// HasErrorln same as Errorln function but only prints if a holds an error value.
func (p Print) HasErrorln(a ...any) bool {
	return hasErr(a, func() {
		p.Errorln(a...)
	})
}

// Warnln print warning
func (p Print) Warnln(a ...any) {
	if WARN >= p.lvl {
		mu.Lock()
		defer mu.Unlock()
		fmt.Fprintln(p.stderr, a...)
	}
}

// HasWarn same as Warn function but only prints if a holds an error value.
func (p Print) HasWarnln(a ...any) bool {
	return hasErr(a, func() {
		p.Warnln(a...)
	})
}

// Warnf print warning
func (p Print) Warnf(format string, a ...any) {
	if WARN >= p.lvl {
		mu.Lock()
		defer mu.Unlock()
		fmt.Fprintf(p.stderr, format, a...)
	}
}

// HasWarnf same as Warn but only prints if a holds an error value.
func (p Print) HasWarnf(format string, a ...any) bool {
	return hasErr(a, func() {
		p.Warnf(format, a...)
	})
}

// Fatalln print a message then exit or panic if Exit level is set to PANIC. See
// the New function and Options struct to set the exit type.
func (p Print) Fatalln(a ...any) {
	fmt.Fprintln(p.stderr, a...)
	p.callExit(-1)
}

// HasFatalln same as Fatal but only prints an error and exits or panics if a
// holds an error value.
func (p Print) HasFatalln(a ...any) {
	hasErr(a, func() {
		p.Fatalln(a...)
	})
}

// Fatalf print a formatted message then exit or panic if exit level is set to
// PANIC. See the New function and Options struct to set the exit type.
func (p Print) Fatalf(format string, a ...any) {
	fmt.Fprintf(p.stderr, format, a...)
	p.callExit(-1)
}

// HasFatal same as Fatal but only prints an error and exits or panics if a
// holds an error value.
func (p Print) HasFatalf(format string, a ...any) {
	hasErr(a, func() {
		p.Fatalf(format, a...)
	})
}

func (p Print) callExit(code int) {
	if p.exit == PANIC {
		panic(fmt.Errorf(`exit code %d`, code))
	}
	os.Exit(code)
}

func hasErr(v []any, f func()) bool {
	hasError := false
	for _, elm := range v {
		if _, ok := elm.(error); ok {
			hasError = true
			f()
			break
		}
	}
	return hasError
}
