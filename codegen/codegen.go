package codegen

import (
	"github.com/ark-lang/ark/parser"
)

type Codegen interface {
	Generate(input []*parser.Module, modules map[string]*parser.Module)
}
