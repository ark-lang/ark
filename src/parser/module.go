package parser

import (
	"github.com/ark-lang/ark/src/lexer"
	"llvm.org/llvm/bindings/go/llvm"
)

type Module struct {
	Nodes       []Node
	File        *lexer.Sourcefile
	Path        string // this stores the path too, e.g src/main
	Name        string // this stores the name, so just main
	GlobalScope *Scope
	Module      llvm.Module
	Functions   []*FunctionDecl
	Variables   []*VariableDecl
	UsedModules map[string]*Module
}
