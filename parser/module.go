package parser

import (
	"llvm.org/llvm/bindings/go/llvm"
)

type Module struct {
	Nodes       []Node
	Name        string
	GlobalScope *Scope
	Module      llvm.Module
    Functions []*FunctionDecl
}
