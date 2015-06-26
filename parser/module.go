package parser

import (
	"llvm.org/llvm/bindings/go/llvm"
)

type Module struct {
	Nodes       []Node
	Name        string         // this stores the path too, e.g src/main
    ModuleName  string         // this stores the name, so just main
	GlobalScope *Scope
	Module      llvm.Module
	Functions   []*FunctionDecl
    UsedModules map[string] *Module
}
