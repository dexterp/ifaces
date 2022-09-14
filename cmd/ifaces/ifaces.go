package main

import (
	"bytes"
	"io"
	"os"

	"github.com/dexterp/ifaces/internal/di"
	"github.com/dexterp/ifaces/internal/resources/cli"
	"github.com/dexterp/ifaces/internal/resources/envs"
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
	if err != nil {
		r.print.Errorln(err.Error())
		os.Exit(-1)
	}
	var outfile io.Writer
	var closer func()
	if r.args.Print {
		outfile, closer = r.outWriter(r.args.Out, os.Stdout)
	} else {
		outfile, closer = r.outWriter(r.args.Out, nil)
	}
	defer closer()
	_, err = io.Copy(outfile, bufOutput)
	if err != nil {
		r.print.Errorf(`can not write to output: %s`, err.Error())
		os.Exit(-1)
	}
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
	for _, file := range r.args.Srcs {
		f, err := os.Open(file)
		if err != nil {
			r.print.Warnf(`skipping file %s: %s\n`, file, err.Error())
			continue
		}
		srcs = append(srcs, &generator.Src{
			File: file,
			Src:  f,
		})
	}
	if envs.Gofile() != `` || envs.Goline() > 0 {
		srcs = append(srcs, &generator.Src{
			File: envs.Gofile(),
			Line: envs.Goline(),
		})
	}
	if srcs == nil {
		r.print.Errorln(`unable to open any source files. exiting`)
		os.Exit(-1)
	}
	return srcs
}

// curGenSrc return the contents of any previously generated source file
func (r run) curGenSrc() *bytes.Buffer {
	cur := &bytes.Buffer{}
	if r.args.Out != `` && r.args.Append {
		curFile, err := os.Open(r.args.Out)
		if err != nil && !os.IsNotExist(err) {
			r.print.Errorf("error opening file %s: %s\n", r.args.Out, err.Error())
			os.Exit(-1)
		} else {
			_, err := io.Copy(cur, curFile)
			if err != nil {
				r.print.Errorf("error reading %s: %s\n", r.args.Out, err.Error())
				os.Exit(-1)
			}
		}
	}
	return cur
}

func (r run) outWriter(file string, output io.Writer) (io.Writer, func()) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		r.print.Errorf(`can not open file %s: %s`, file, err.Error())
		os.Exit(-1)
	}
	if output != nil {
		multi := io.MultiWriter(f, output)
		return multi, func() { f.Close() }
	}
	return f, func() { f.Close() }
}
