package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/dexterp/ifaces/internal/di"
	"github.com/dexterp/ifaces/internal/resources/cli"
	"github.com/dexterp/ifaces/internal/resources/version"
)

func main() {
	args := getArgs()
	if args.Gen {

	}

	file, line := getFileLine(args)
	current := currentSrc(args)
	src, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %s\n", file, err.Error())
		os.Exit(-1)
	}
	var outfile io.Writer
	var closer func()
	if args.Print {
		outfile, closer = outWriter(args.Out, os.Stdout)
	} else {
		outfile, closer = outWriter(args.Out, nil)
	}
	defer closer()
	di.Args = args
	gen := di.MakeIfaceGen()
	switch {
	case args.Head:
		err = gen.Head(file, src, args.Out, current, outfile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
	case args.Entry:
		err = gen.Entry(file, src, line, args.Out, current, outfile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
	default:
		fmt.Fprintln(os.Stderr, `unrecognized sub command see "ifaces -h" for command line options`)
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

func currentSrc(args *cli.Args) *bytes.Buffer {
	// current source
	cur := &bytes.Buffer{}
	if args.Entry || args.Append {
		curFile, err := os.Open(args.Out)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error opening file %s: %s\n", args.Out, err.Error())
			os.Exit(-1)
		}
		if curFile != nil {
			_, err := io.Copy(cur, curFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading %s: %s\n", args.Out, err.Error())
				os.Exit(-1)
			}
		}
	}
	return cur
}

func getFileLine(args *cli.Args) (string, int) {
	file := os.Getenv("GOFILE")
	errMsg := `ifaces (head|entity) sub commands should be run as go generators. exiting`
	if file == `` && (args.Head || args.Entry) {
		fmt.Fprintln(os.Stderr, errMsg)
		os.Exit(-1)
	}
	lStr := os.Getenv("GOLINE")
	if lStr == `` && (args.Head || args.Entry) {
		fmt.Fprintln(os.Stderr, errMsg)
		os.Exit(-1)
	}
	line, err := strconv.Atoi(lStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errMsg)
		os.Exit(-1)
	}
	return file, line
}

func outWriter(file string, output io.Writer) (io.Writer, func()) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, `can not open file %s: %s`, file, err.Error())
		os.Exit(-1)
	}
	if output != nil {
		multi := io.MultiWriter(f, output)
		return multi, func() { f.Close() }
	}
	return f, func() { f.Close() }
}
