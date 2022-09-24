package print

import (
	"fmt"
	"io"
)

const (
	DEBUG = iota + 1
	INFO
	WARN
	ERROR
)

//go:generate ifaces type print_iface.go --post Iface

// Print handle printing
type Print struct {
	stderr io.Writer
	stdout io.Writer
	lvl    int
}

// Options options for New
type Options struct {
	Stderr io.Writer
	Stdout io.Writer
	Level  int // Print level. 0 is default, less then 0 disables printing
}

// New return a Print type
func New(opts Options) *Print {
	var lvl int
	if opts.Level > 0 {
		lvl = opts.Level
	} else if opts.Level == 0 {
		lvl = WARN
	}
	return &Print{
		stderr: opts.Stderr,
		stdout: opts.Stdout,
		lvl:    lvl,
	}
}

// Level set level
func (p *Print) Level(lvl int) {
	p.lvl = lvl
}

// Errorf print error
func (p Print) Errorf(format string, a ...any) {
	if ERROR >= p.lvl {
		fmt.Fprintf(p.stderr, format, a...)
	}
}

// Errorln print error
func (p Print) Errorln(a ...any) {
	if ERROR >= p.lvl {
		fmt.Fprintln(p.stderr, a...)
	}
}

// Warn print warning
func (p Print) Warn(a ...any) {
	if WARN >= p.lvl {
		fmt.Fprintln(p.stderr, a...)
	}
}

// Warnf print warning
func (p Print) Warnf(format string, a ...any) {
	if WARN >= p.lvl {
		fmt.Fprintf(p.stderr, format, a...)
	}
}
