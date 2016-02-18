package codegen

import (
	"fmt"

	"github.com/ark-lang/ark/src/ast"
)

type Codegen interface {
	Generate(input []*ast.Module)
}

type OutputType int

const (
	OutputUnknown OutputType = iota
	OutputExectuably
	OutputObject
	OutputAssembly
	OutputLLVMIR
)

var typeMapping = map[string]OutputType{
	"executable": OutputExectuably,
	"object":     OutputObject,
	"assembly":   OutputAssembly,
	"llvm-ir":    OutputLLVMIR,
}

func ParseOutputType(input string) (OutputType, error) {
	typ, ok := typeMapping[input]
	if !ok {
		return OutputUnknown, fmt.Errorf("ark-lang/codegen: Unknown output type `%s`", input)
	}
	return typ, nil
}
