package codegen

import (
	"github.com/ark-lang/ark-go/parser"
)

type Codegen interface {
	Generate(input []*parser.File, verbose bool)
}
