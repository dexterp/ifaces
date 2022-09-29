// Code generated by ifaces DO NOT EDIT.

package parser

// MethodIface receiver or interface method
type MethodIface interface {
	// SetPkg set the prefix when exporting to a new package. E.G. MyType will be
	// converted to pkg.MyType.
	SetPkg(pkg string)
	// Line return line number in source code
	Line() int
	// TypeName return the type name or receiver name for this method
	TypeName() string
	// Name method name
	Name() string
	// Doc method documentation
	Doc() string
	// Signature return the function signature
	Signature() string
	// UsesTypeParams returns true if function declaration contains type parameters
	UsesTypeParams() bool
}
