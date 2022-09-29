package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/dexterp/ifaces/internal/di"
	"github.com/dexterp/ifaces/internal/resources/cli"
	"github.com/dexterp/ifaces/internal/resources/envs"
	"github.com/dexterp/ifaces/internal/resources/modinfo"
	"github.com/dexterp/ifaces/internal/resources/print"
	"github.com/dexterp/ifaces/internal/resources/version"
	"github.com/dexterp/ifaces/internal/services/generator"
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
	gen   generator.GeneratorIface
	print print.PrintIface
}

func (r run) checkSrcs() {
	if len(r.args.Srcs) == 0 && (envs.Gofile() == `` || envs.Goline() < 1) {
		r.print.Errorln(`no source file provided. needs -f <file> option or to run as part of a go generator. exiting`)
		os.Exit(-1)
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
		os.Exit(-1)
	}
	return args
}

func (r run) srcList() (srcs []*generator.Src) {
	var modpath string
	if r.args.Module != `` {
		dir, err := os.Getwd()
		r.print.HasFatalf(`can not determine working directory: %v`, err)
		mi, err := modinfo.LoadFromParents(dir)
		r.print.HasFatalf(`error loading go module: %v`, err)
		modpath, err = mi.GetPath(r.args.Module)
		r.print.HasFatalf(`can not find module directory %s: %s`, r.args.Module, err.Error())
	}
	for _, file := range r.args.Srcs {
		if modpath != `` {
			file = filepath.Join(modpath, file)
		}
		_, err := os.Stat(file)
		if r.print.HasWarnf(`skipping %s: %v`, file, err) {
			continue
		}
		srcs = append(srcs, &generator.Src{
			File: file,
			Src:  nil,
		})
	}
	if envs.Gofile() != `` || envs.Goline() > 0 {
		srcs = append(srcs, &generator.Src{
			File: envs.Gofile(),
			Line: envs.Goline(),
		})
	}
	if srcs == nil {
		r.print.Fatalln(`unable to open any source files. exiting`)
	}
	return srcs
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
