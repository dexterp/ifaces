package addimports

import (
	"go/format"
	"go/parser"
	"go/token"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
)

func NewImport(name, path string) Import {
	return Import{
		name: name,
		path: path,
	}
}

type Import struct {
	name string
	path string
}

func AddImports(file string, src any, imports []Import, output io.Writer) error {
	// TODO - Fix problem with imports containing the same package identifier choosing the wrong package.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		return errors.Wrap(err, "unable to add import")
	}
	chk := &hasPath{}
	for _, i := range f.Imports {
		name := ``
		if i.Name != nil {
			name = i.Name.String()
		}
		path := i.Path.Value
		chk.add(name, path)
	}
	for _, i := range imports {
		if chk.contains(i.name, i.path) {
			continue
		}
		if i.name != `` {
			astutil.AddNamedImport(fset, f, i.name, i.path)
		} else {
			astutil.AddImport(fset, f, i.path)
		}
	}
	return format.Node(output, fset, f)
}

type hasPath map[string]map[string]any

func (h hasPath) add(name, path string) {
	if _, ok := h[name]; !ok {
		h[name] = make(map[string]any)
	}
	if _, ok := h[name][path]; !ok {
		h[name][path] = struct{}{}
	}
}

func (h hasPath) contains(name, path string) bool {
	if _, ok := h[name]; !ok {
		return ok
	}
	_, ok := h[name][path]
	return ok
}
