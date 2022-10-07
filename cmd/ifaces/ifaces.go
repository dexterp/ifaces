package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/dexterp/ifaces/internal/di"
	"github.com/dexterp/ifaces/internal/resources/cli"
	"github.com/dexterp/ifaces/internal/resources/envs"
	"github.com/dexterp/ifaces/internal/resources/modinfo"
	"github.com/dexterp/ifaces/internal/resources/print"
	"github.com/dexterp/ifaces/internal/resources/srcio"
	"github.com/dexterp/ifaces/internal/resources/version"
	"github.com/dexterp/ifaces/internal/services/generate"
)

func main() {
	args := getArgs()
	di.Args = args
	di.Stderr = os.Stderr
	di.Stdout = os.Stdout
	r := &run{
		args:  args,
		gen:   di.MakeIfaceGen(),
		print: di.MakePrint(),
	}
	r.checkSrcs()
	r.runGen()
}

type run struct {
	args  *cli.Args
	gen   generate.GenerateIface
	print print.PrintIface
}

func (r run) checkSrcs() {
	if r.args.Src == `` && (envs.Gofile() == `` || envs.Goline() < 1) {
		r.print.Fatalln(`no source file provided. needs -f <file> option or to run as part of a go generator. exiting`)
	}
}

func (r run) runGen() {
	srcsList := r.srcList()
	curGenSrc := r.curGenSrc()
	bufOutput := &bytes.Buffer{}
	err := r.gen.Generate(srcsList, curGenSrc, r.args.Out, bufOutput)
	r.print.HasFatalln(err)
	var outfile io.Writer
	var closer func()
	if r.args.Print || r.args.Out == `` {
		outfile, closer = r.outWriter(r.args.Out, os.Stdout)
	} else {
		outfile, closer = r.outWriter(r.args.Out)
	}
	defer closer()
	_, err = io.Copy(outfile, bufOutput)
	r.print.HasFatalf(`can not write to output: %v`, err)
}

func getArgs() *cli.Args {
	args, err := cli.ParseArgs(os.Args[1:], version.Version, os.Stdout, os.Stderr)
	if args == nil && err == nil {
		os.Exit(0)
	} else if err != nil {
		os.Exit(127)
	}
	return args
}

func (r run) srcList() (srcs []srcio.Source) {
	added := map[string]any{}
	path := r.args.Src
	path = r.pathModulePrefix(path)
	path = r.goGeneratePath(&srcs, added, path)
	r.expandPath(&srcs, added, path)
	if srcs == nil {
		r.print.Fatalln(`no valid source files found`)
	}
	return
}

// curGenSrc return the contents of any previously generated source file
func (r run) curGenSrc() *bytes.Buffer {
	cur := &bytes.Buffer{}
	if r.args.Out != `` && r.args.Append {
		curFile, err := os.Open(r.args.Out)
		if err != nil && !os.IsNotExist(err) {
			r.print.Fatalf("error opening file %s: %v\n", r.args.Out, err)
		} else {
			_, err := io.Copy(cur, curFile)
			r.print.HasFatalf("error reading %s: %v\n", r.args.Out, err)
		}
	}
	return cur
}

func (r run) outWriter(file string, writers ...io.Writer) (io.Writer, func()) {
	var (
		closers []io.Closer
		wrtrs   []io.Writer
	)
	if file != `` {
		f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
		r.print.HasFatalf("can not open file %s: %\n", file, err)
		wrtrs = append(wrtrs, f)
		closers = append(closers, f)
	}
	if writers != nil {
		wrtrs = append(wrtrs, writers...)
	}
	multi := io.MultiWriter(wrtrs...)
	return multi, func() {
		for _, c := range closers {
			c.Close()
		}
	}
}

// pathModulePrefix prefix the module path to the path if it exists
func (r run) pathModulePrefix(path string) string {
	if path == `` {
		return path
	}
	// If Module option set, get the module path and add it as a prefix to path
	var modpath string
	if r.args.Module != `` {
		dir, err := os.Getwd()
		r.print.HasFatalln(err)
		mi, err := modinfo.LoadFromParents(dir)
		r.print.HasFatalf(`error loading go.mod file: %v`, err)
		modpath, err = mi.GetPath(r.args.Module)
		r.print.HasFatalf(`can not find module directory %s: %v`, r.args.Module, err)
	}
	if path != `` && modpath != `` {
		path = filepath.Join(modpath, path)
	}
	return path
}

// expandPath expands path to a directory
func (r run) expandPath(srcs *[]srcio.Source, added map[string]any, path string) {
	if path == `` {
		return
	}
	// Expand file path to all go source files in the directory
	info, err := os.Stat(path)
	if e := errors.Unwrap(err); e != nil {
		err = e
	}
	r.print.HasFatalf("%s: %v\n", path, err)

	path = filepath.Clean(path)
	var dir string
	if info.IsDir() {
		dir = path
	} else {
		if _, ok := added[path]; !ok {
			*srcs = append(*srcs, srcio.Source{
				File: path,
			})
			added[path] = true
		}
		dir = filepath.Dir(path)
	}
	matches, err := filepath.Glob(filepath.Join(dir, `*.go`))
	r.print.HasFatalf(`error reading directory %s:`, dir, err)
	for _, m := range matches {
		if _, ok := added[m]; ok {
			continue
		}
		*srcs = append(*srcs, srcio.Source{
			File: m,
			Src:  nil,
		})
		added[m] = true
	}
}

// goGeneratePath add go:generate environment variable to sources list
func (r run) goGeneratePath(srcs *[]srcio.Source, added map[string]any, path string) string {
	// Skip adding go:generate path if a path is already provided
	if path != `` {
		return path
	}
	if envs.Gofile() != `` || envs.Goline() > 0 {
		*srcs = append(*srcs, srcio.Source{
			File: envs.Gofile(),
			Line: envs.Goline(),
		})
		added[envs.Gofile()] = true
	}
	return envs.Gofile()
}
