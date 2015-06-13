package parser

import (
	"llvm.org/llvm/bindings/go/llvm"
)

type File struct {
	Nodes       []Node
	Name        string
	GlobalScope *Scope
	Module      llvm.Module
}
