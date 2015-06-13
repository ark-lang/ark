package codegen

import (
	"github.com/ark-lang/ark/parser"
)

type Codegen interface {
	Generate(input []*parser.Module, verbose bool)
}
