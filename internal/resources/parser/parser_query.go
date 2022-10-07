package parser

import (
	"github.com/dexterp/ifaces/internal/resources/match"
)

type Query struct {
	Parser *Parser
}

func NewQuery(p *Parser) *Query {
	return &Query{Parser: p}
}

// GetIfaceMethods returns all the methods of an interface
func (q *Query) GetIfaceMethods(iface string) (methods []*Method) {
	for _, m := range q.Parser.InterfaceMethods {
		if m.TypeName == iface {
			methods = append(methods, m)
		}
	}
	return
}

// GetRecvByLine
func (q *Query) GetRecvByLine(file string, line int) (recv *Method) {
	end := q.NextComment(file, line)
	for _, m := range q.Parser.ReceiverMethods {
		if m.File == file && end == 0 && m.Line >= line {
			return m
		} else if m.File == file && m.Line >= line && m.Line < end {
			return m
		}
	}
	return
}

// GetRecvsByName returns all receivers by a pattern
func (q *Query) GetRecvsByName(name string) (recvs []*Method) {
	for _, recv := range q.Parser.ReceiverMethods {
		if recv.Name == name && match.Capitalized(recv.Name) {
			recvs = append(recvs, recv)
		}
	}
	return
}

// GetRecvsByType returns all the function interfaces for
func (q *Query) GetRecvsByType(typ string) (recvs []*Method) {
	for _, recv := range q.Parser.ReceiverMethods {
		sig := recv.Signature()
		if recv.TypeName == typ && match.Capitalized(sig) {
			recvs = append(recvs, recv)
		}
	}
	return recvs
}

// GetTypeByLine GetTypeByLine returns the type at or after line. returns nil if the end of
// file is reached or it encounters a iface generator comment.
func (q Query) GetTypeByLine(file string, line int) *Type {
	end := q.NextComment(file, line)
	for _, t := range q.Parser.Types {
		if file == t.File && end == 0 && t.Line >= line {
			return &t
		} else if file == t.File && t.Line >= line && t.Line < end {
			return &t
		}
	}
	return nil
}

// GetTypeByName fetch a type by its name
func (q Query) GetTypeByName(name string) *Type {
	for _, t := range q.Parser.Types {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// GetTypeByPattern use pattern to match types and return a list of type declarations
func (q Query) GetTypeByPattern(pattern string) (ts []Type) {
	for _, t := range q.Parser.Types {
		if match.Match(t.Name, pattern) && match.Capitalized(t.Name) {
			ts = append(ts, t)
		}
	}
	return ts
}

// GetTypesByType return type
func (q *Query) GetTypesByType(typ int) (ts []Type) {
	for _, t := range q.Parser.Types {
		if t.Type == typ {
			ts = append(ts, t)
		}
	}
	return ts
}

// NextComment finds the next go:generate comment after line and returns the
// line number. Returns 0 if not found.
func (q Query) NextComment(file string, line int) (end int) {
	for _, c := range q.Parser.Comments {
		if c.File == file && c.Line > line {
			end = c.Line
		}
	}
	return
}
