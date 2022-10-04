package di

import (
	"io"

	"github.com/dexterp/ifaces/internal/resources/cli"
	"github.com/dexterp/ifaces/internal/resources/print"
	"github.com/dexterp/ifaces/internal/services/generate"
)

//
// Services injection
//

var (
	Args   *cli.Args // Args command line options
	Stderr io.Writer
	Stdout io.Writer
	Level  print.Level
)

func MakeIfaceGen() generate.GenerateIface {
	return &generate.Generate{
		Type:      Args.CmdType,
		Method:    Args.CmdFunc,
		Comment:   Args.Cmt,
		Iface:     Args.Iface,
		MatchFunc: Args.MatchFunc,
		MatchType: Args.MatchType,
		Module:    Args.Module,
		NoFDoc:    Args.NoFDoc,
		NoTDoc:    Args.NoTDoc,
		Pkg:       Args.Pkg,
		Post:      Args.Post,
		Pre:       Args.Pre,
		Print:     MakePrint(),
		Struct:    Args.CmdStruct,
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
