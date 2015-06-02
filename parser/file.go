package parser

import (
    "llvm.org/llvm/bindings/go/llvm"
)

type File struct {
	Nodes []Node
	Name  string
    Module llvm.Module
}
