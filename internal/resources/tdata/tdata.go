// tdata Template data
package tdata

import (
	"bufio"
	"bytes"
	"errors"
	"regexp"
	"strings"

	"github.com/mitchellh/go-wordwrap"
)

var (
	ErrorDuplicateInterface = errors.New(`can not add duplicate interface`)
	ErrorDuplicateMethod    = errors.New(`can not add duplicate method`)
	reFirstWord             = regexp.MustCompile(`^\w+`)
)

type TData struct {
	Comment string // Comment comment at the top of the file
	NoFDoc  bool   // NoFDoc omit copying function documentation
	NoTDoc  bool   // NoTDoc omit copying type documentation
	Pkg     string // Pkg package name
	Post    string // Post postfix to interface name
	Pre     string // Pre prefix to interface name
	Ifaces  []*Interface
	unique  map[string]*Interface
}

// Add add an interface
func (t *TData) Add(iface *Interface) error {
	if t.unique == nil {
		t.unique = map[string]*Interface{}
	}
	if _, ok := t.unique[iface.Type.name]; ok {
		return ErrorDuplicateInterface
	}
	t.unique[iface.Type.name] = iface
	t.Ifaces = append(t.Ifaces, iface)
	return nil
}

func (t TData) Get(iface string) *Interface {
	if i, ok := t.unique[iface]; ok {
		return i
	}
	return nil
}

type Interface struct {
	Type    *Type     // TypeDecl type declaration
	Methods []*Method // Methods list of methods
	unique  map[string]*Method
}

// Add add method to the interface
func (i *Interface) Add(method *Method) error {
	if i.unique == nil {
		i.unique = map[string]*Method{}
	}
	if _, ok := i.unique[method.name]; ok {
		return ErrorDuplicateMethod
	}
	i.unique[method.name] = method
	i.Methods = append(i.Methods, method)
	return nil
}

func NewType(name, doc string, noTypeDoc bool) *Type {
	return &Type{
		name:      name,
		doc:       doc,
		noTypeDoc: noTypeDoc,
	}
}

type Type struct {
	noTypeDoc bool
	name      string
	doc       string
}

func (t Type) Doc() string {
	if t.noTypeDoc {
		return ``
	}
	doc := reFirstWord.ReplaceAllString(t.doc, t.Name())
	return wrapDoc(doc, false)
}

func (t Type) Name() string {
	return t.name
}

func NewMethod(name, signature, doc string, nofdoc bool) *Method {
	return &Method{
		name:      name,
		doc:       doc,
		signature: signature,
		noFuncDoc: nofdoc,
	}
}

type Method struct {
	noFuncDoc bool
	name      string
	doc       string
	signature string
}

func (r Method) Doc() string {
	if r.noFuncDoc {
		return ``
	}
	doc := wrapDoc(r.doc, true)
	if len(doc) > 0 {
		return doc + "\t"
	}
	return ``
}

func (r Method) Signature() string {
	return r.signature
}

func wrapDoc(doc string, indent bool) string {
	doc = strings.Replace(doc, "\n", " ", -1)
	doc = wordwrap.WrapString(doc, 76)
	scanner := bufio.NewScanner(strings.NewReader(doc))
	first := true
	buf := bytes.Buffer{}
	for scanner.Scan() {
		if first {
			buf.WriteString("// " + scanner.Text() + "\n")
			first = false
		} else {
			if indent {
				buf.WriteString("\t")
			}
			buf.WriteString("// " + scanner.Text() + "\n")
		}
	}
	return buf.String()
}
