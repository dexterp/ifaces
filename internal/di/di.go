package di

import (
	"github.com/dexterp/ifaces/internal/resources/cli"
	"github.com/dexterp/ifaces/internal/services/generator"
)

//
// Services injection
//

var (
	Args *cli.Args // Args command line options
)

func MakeIfaceGen() *generator.Generator {
	return generator.New(generator.Options{
		Comment: Args.Cmt,
		Iface:   Args.Iface,
		NoFDoc:  Args.NoFDoc,
		NoHdr:   Args.NoHdr,
		NoTDoc:  Args.NoTDoc,
		Pkg:     Args.Pkg,
		Post:    Args.Post,
		Pre:     Args.Pre,
		Wild:    Args.Wild,
	})
}
