package parser

// QueryParser parse interface
type QueryParser interface {
	Comments() []*Comment
	Imports() []*Import
	InterfaceMethods() []*Method
	ReceiverMethods() []*Method
	Package() string
	Types() []*Type
}
