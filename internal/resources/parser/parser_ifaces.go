// DO NOT EDIT. DO NOT USE. generated for ifaces testing

package parser

// ParserIface parse method
type ParserIface interface {
	// GetRecvs GetRecvs returns all the function interfaces for
	GetRecvs(typ string) (recvs []*Recv)
	// GetType GetType returns the type at or after line. returns nil if the end of
	// file is reached or it encounters a iface generator comment.
	GetType(line int) *Type
	// Imports list of imports
	Imports() (imports []*Import)
	// Package package file
	Package() string
	// TypeByPattern use pattern to match types and return a list of type
	// decleartions
	TypeByPattern(pattern string) []*Type
	// TypeByType return type
	TypeByType(ifacetype int) []*Type
}

// ImportIface
type ImportIface interface {
	Name() string
	Path() string
}

// TypeIface type declaration
type TypeIface interface {
	// Line line number
	Line() int
	// Doc type documentation
	Doc() string
	// Name type name
	Name() string
	// Type get type of
	Type() int
}

// RecvIface function receiver
type RecvIface interface {
	// Line line number
	Line() int
	// TypeName typename
	TypeName() string
	// Doc returns function documentation
	Doc() string
	// Signature return the function signature
	Signature() string
}
