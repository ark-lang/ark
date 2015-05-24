package parser

import (
	"github.com/alloy-lang/alloy-go/lexer"
)

type File struct {
	nodes []Node
}

type parser struct {
	file *File
	input []*lexer.Token
	verbose bool
}

func Parse(tokens []*lexer.Token, verbose bool) *File {
	p := &parser {
		file: &File {
			nodes: make([]Node, 0),
		},
		input: tokens,
		verbose: verbose,
	}
	
	p.parse()
	return p.file
}

func (v *parser) parse() {
	
}
