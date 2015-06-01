package codegen

import (
	"github.com/ark-lang/ark-go/parser"
)

type Codegen interface {
	Generate(binaryName string, input []*parser.File)
}
