package di

import (
	"io"

	"github.com/dexterp/ifaces/internal/resources/cli"
	"github.com/dexterp/ifaces/internal/resources/print"
	"github.com/dexterp/ifaces/internal/services/generator"
)

//
// Services injection
//

var (
	Args   *cli.Args // Args command line options
	Stderr io.Writer
	Stdout io.Writer
	Level  int
)

func MakeIfaceGen() generator.GeneratorIface {
	return &generator.Generator{
		Type:      Args.CmdType,
		Method:    Args.CmdMethod,
		Comment:   Args.Cmt,
		Iface:     Args.Iface,
		MatchType: Args.MatchType,
		NoFDoc:    Args.NoFDoc,
		NoTDoc:    Args.NoTDoc,
		Pkg:       Args.Pkg,
		Post:      Args.Post,
		Pre:       Args.Pre,
		Print:     MakePrint(),
		Struct:    Args.Struct,
		TDoc:      Args.TDoc,
	}
}

//
// Resources Injection
//

var cachePrint *print.Print

func MakePrint() print.PrintIface {
	if cachePrint == nil {
		cachePrint = print.New(print.Options{
			Stderr: Stderr,
			Stdout: Stdout,
			Level:  Level,
		})
	}
	return cachePrint
}
