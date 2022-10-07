package generate

import (
	"github.com/dexterp/ifaces/internal/resources/parser"
	"github.com/dexterp/ifaces/internal/resources/tdata"
)

// target represents the target file
type target struct {
	file     string                 // Output file
	exported bool                   // True if the source file is exported.
	imports  map[*parser.Import]any // Imports
	tdata    *tdata.TData           // Template data
}
