package parser

import (
	"fmt"
	"os"
	
	"github.com/alloy-lang/alloy-go/lexer"
	"github.com/alloy-lang/alloy-go/util"
)

type File struct {
	nodes []Node
}

type parser struct {
	file *File
	input []*lexer.Token
	currentToken int
	verbose bool
}

func (v *parser) err(err string) {
	fmt.Printf(util.TEXT_RED + util.TEXT_BOLD + "Lexer error:" + util.TEXT_RESET + " [%s:%d:%d] %s\n",
			v.peek(0).Filename, v.peek(0).LineNumber, v.peek(0).CharNumber, err)
	os.Exit(2)
}

func (v *parser) peek(ahead int) *lexer.Token {
	if ahead < 0 {
		panic(fmt.Sprintf("Tried to peek a negative number: %d", ahead))
	}
	
	return v.input[v.currentToken + ahead]
}

func (v *parser) consume() {
	v.currentToken++
}

func Parse(tokens []*lexer.Token, verbose bool) *File {
	p := &parser {
		file: &File {
			nodes: make([]Node, 0),
		},
		input: tokens,
		verbose: verbose,
	}
	
	if verbose {
		fmt.Println(util.TEXT_BOLD + util.TEXT_GREEN + "Started parsing" + util.TEXT_RESET, tokens[0].Filename)
	}
	p.parse()
	if verbose {
		fmt.Println(util.TEXT_BOLD + util.TEXT_GREEN + "Finished parsing" + util.TEXT_RESET, tokens[0].Filename)
	}
	
	return p.file
}

func (v *parser) parse() {
	
}
