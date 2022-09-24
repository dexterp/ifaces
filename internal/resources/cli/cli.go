package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/docopt/docopt-go"
)

var usage = `Usage:
  ifaces type [<out>] [(-f <src>)...] [-a] [-p <pkg>] [-i <iface>] [-t <type>] [-s] [--pre <pre>] [--post <post>] [--no-tdoc] [--no-fdoc] [-c <cmt>] [--print]
  ifaces recv [<out>] [(-f <src>)...] [-a] [-p <pkg>] [-i <iface>] [-t <type>] [-r <func>] [--pre <pre>] [--post <post>] [--no-tdoc] [--no-fdoc] [-c <cmt>] [--print]

Options:
  type            Generate interfaces for either all structs, a set of matching
                  types or the first type found after a go:generate command in a
                  Go source file.
  recv            Generate interface for an individual receiver from the command
                  line or the first receiver found after a go:generate command
                  in a Go source file.
  -h              Show this screen.
  <out>           Output file. Truncated unless -a is set. 
  -a              Add to output file instead of truncating.
  -c <cmt>        Comment at top of output file. [default: DO NOT EDIT. GENERATED BY ifaces]   
  -f <src>        Source file to scan.
  -i <iface>      Optional interface type name. If omitted the type name is used
                  with a prefix and/or suffix added.
  --no-fdoc       Do not copy function docs to the interface function type.
  --no-tdoc       Do not copy doc from the source type to the interface type.
  -p <pkg>        Package name. Defaults to the parent directory name.
  --post <post>   Add a suffix to interface type name.
  --pre <pre>     Add a prefix to interface type name.
  --print         Print generated source to stdout. The default when no output
                  file is provided.
  -r <func>       Receiver function match.
  -s              Generate an interface for all structs.
  --tdoc <tdoc>   Custom type document. Defaults to the document from the origin
                  source type.
  -t <type>       Generate interfaces for types that match a wildcard.
 `

func ParseArgs(argv []string, version string, stdout io.Writer, stderr io.Writer) (*Args, error) {
	var fnerr error
	fn := func(err error, usage string) {
		if len(argv) > 0 && (argv[0] == "-h" || argv[0] == "--help") {
			fmt.Fprintln(stdout, usage)
			os.Exit(0)
		} else {
			fnerr = err
			fmt.Fprint(stderr, err.Error())
			fmt.Fprintln(stderr, `invalid or incomplete options, see "ifaces -h" for cli options`)
		}
	}
	if fnerr != nil {
		return nil, fnerr
	}
	parser := &docopt.Parser{
		HelpHandler: fn,
	}
	args, err := parser.ParseArgs(usage, argv, version)
	if err != nil {
		return nil, err
	}
	config := &Args{}
	err = args.Bind(config)
	if err != nil {
		fmt.Fprintf(stderr, `error binding command arguments: %s`, err.Error())
	}
	return config, nil
}

type Args struct {
	SubType bool   `docopt:"type"`
	SubRecv bool   `docopt:"recv"`
	Out     string `docopt:"<out>"`

	Append    bool     `docopt:"-a"`
	Cmt       string   `docopt:"-c"`
	Iface     string   `docopt:"-i"`
	MatchFunc string   `docopt:"-r"`
	MatchType string   `docopt:"-t"`
	NoFDoc    bool     `docopt:"--no-fdoc"`
	NoTDoc    bool     `docopt:"--no-tdoc"`
	Pkg       string   `docopt:"-p"`
	Post      string   `docopt:"--post"`
	Pre       string   `docopt:"--pre"`
	Print     bool     `docopt:"--print"`
	Srcs      []string `docopt:"-f"`
	Struct    bool     `docopt:"-s"`
	TDoc      string   `docopt:"--tdoc"`
}
