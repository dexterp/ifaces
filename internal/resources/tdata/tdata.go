// tdata Template data
package tdata

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/dexterp/ifaces/internal/resources/parser"
	"github.com/mitchellh/go-wordwrap"
)

type TData struct {
	Comment string // Comment comment at the top of the file
	NoFDoc  bool   // NoFDoc omit copying function documentation
	NoTDoc  bool   // NoTDoc omit copying type documentation
	Pkg     string // Pkg package name
	Post    string // Post postfix to interface name
	Pre     string // Pre prefix to interface name
	Ifaces  []*Interface
}

type Interface struct {
	Type  *Type   // TypeDecl type declaration
	Recvs []*Recv // Recvs list of recevers to add to template
}

type Import struct {
	imp *parser.Import
}

func (i Import) Name() string {
	return i.imp.Name()
}

func (i Import) Path() string {
	return i.imp.Path()
}

type Type struct {
	NoTDoc bool
	TDoc   string
	Post   string
	Iface  string
	Pre    string
	Decl   *parser.Type
}

func (t Type) Doc() string {
	if t.NoTDoc {
		return ``
	} else if t.TDoc != `` {
		return `// ` + t.TDoc
	}
	re := regexp.MustCompile(`^\w+`)
	doc := t.Decl.Doc()

	doc = re.ReplaceAllString(doc, t.Name())
	return wrapDoc(doc, false)
}

func (t Type) Name() string {
	iface := t.Iface
	if iface == `` {
		iface = t.Decl.Name()
	}
	return fmt.Sprint(t.Pre, iface, t.Post)
}

type Recv struct {
	NoFDoc bool
	Recv   *parser.Recv
}

func (r Recv) Doc() string {
	if r.NoFDoc {
		return ``
	}
	doc := wrapDoc(r.Recv.Doc(), true)
	if len(doc) > 0 {
		return doc + "\t"
	}
	return ``
}

func (r Recv) Signature() string {
	return r.Recv.Signature()
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
