package srcformat

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"io"

	"golang.org/x/tools/imports"
)

func Format(filename string, src any, out io.Writer) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return err
	}
	formatNodeOut := &bytes.Buffer{}
	err = format.Node(formatNodeOut, fset, f)
	if err != nil {
		return err
	}
	importsProcessOut, err := imports.Process(filename, formatNodeOut.Bytes(), &imports.Options{
		TabIndent: true,
		TabWidth:  2,
		Fragment:  false,
		Comments:  true,
	})
	if err != nil {
		return err
	}
	_, err = out.Write(importsProcessOut)
	return err
}
