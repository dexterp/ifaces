package parser

import (
	"github.com/dexterp/ifaces/internal/resources/match"
)

// GetIfaceMethods returns all the methods of an interface
func (p *Parser) GetIfaceMethods(iface string) (methods []Method) {
	for _, m := range *p.InterfaceMethods {
		if m.TypeName() == iface {
			methods = append(methods, m)
		}
	}
	return
}

// GetRecvByLine
func (p *Parser) GetRecvByLine(line int) (recv Method) {
	end := p.NextComment(line)
	for _, m := range *p.ReceiverMethods {
		if end == 0 && m.Line() >= line {
			return m
		} else if m.Line() >= line && m.Line() < end {
			return m
		}
	}
	return
}

// GetRecvsByName returns all receivers by a pattern
func (p *Parser) GetRecvsByName(name string) (recvs []Method) {
	for _, recv := range *p.ReceiverMethods {
		if recv.Name() == name && match.Capitalized(recv.Name()) {
			recvs = append(recvs, recv)
		}
	}
	return
}

// GetRecvsByType returns all the function interfaces for
func (p *Parser) GetRecvsByType(typ string) (recvs []Method) {
	for _, recv := range *p.ReceiverMethods {
		sig := recv.Signature()
		if recv.TypeName() == typ && match.Capitalized(sig) {
			recvs = append(recvs, recv)
		}
	}
	return recvs
}

// GetTypeByLine GetTypeByLine returns the type at or after line. returns nil if the end of
// file is reached or it encounters a iface generator comment.
func (p Parser) GetTypeByLine(line int) *Type {
	end := p.NextComment(line)
	for _, t := range *p.Types {
		if end == 0 && t.Line() >= line {
			return t
		} else if t.Line() >= line && t.Line() < end {
			return t
		}
	}
	return nil
}

// GetTypeByName fetch a type by its name
func (p Parser) GetTypeByName(name string) *Type {
	for _, t := range *p.Types {
		if t.Name() == name {
			return t
		}
	}
	return nil
}

// GetTypeByPattern use pattern to match types and return a list of type declarations
func (p Parser) GetTypeByPattern(pattern string) (ts []*Type) {
	for _, t := range *p.Types {
		if match.Match(t.Name(), pattern) && match.Capitalized(t.Name()) {
			ts = append(ts, t)
		}
	}
	return ts
}

// GetTypesByType return type
func (p *Parser) GetTypesByType(typ int) (ts []*Type) {
	for _, t := range *p.Types {
		if t.Type() == typ {
			ts = append(ts, t)
		}
	}
	return ts
}

// NextComment finds the next go:generate comment after line. Returns 0 if not
// found
func (p Parser) NextComment(line int) (end int) {
	for _, c := range p.comments {
		if c.Line > line {
			end = c.Line
		}
	}
	return
}
