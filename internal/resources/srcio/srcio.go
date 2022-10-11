package srcio

import (
	"bytes"
	"io"
)

type Source struct {
	File string
	Line int
	Src  any
}

type Destination struct {
	File    string
	Package string
	Current *bytes.Buffer
	Output  io.Writer
}
