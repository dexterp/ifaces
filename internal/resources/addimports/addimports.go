package addimports

import (
	"go/format"
	"go/parser"
	"go/token"
	"io"

	"golang.org/x/tools/go/ast/astutil"
)

type Import interface {
	Name() string
	Path() string
}

func AddImports(file string, src any, imports []Import, output io.Writer) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, parser.ParseComments)
	if err != nil {
		return err
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
		if chk.contains(i.Name(), i.Path()) {
			continue
		}
		if i.Name() != `` {
			astutil.AddNamedImport(fset, f, i.Name(), i.Path())
		} else {
			astutil.AddImport(fset, f, i.Path())
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
