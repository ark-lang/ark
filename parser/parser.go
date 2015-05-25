package parser

import (
	"fmt"
	"os"
	
	"github.com/ark-lang/ark-go/lexer"
	"github.com/ark-lang/ark-go/util"
)

type parser struct {
	file *File
	input []*lexer.Token
	currentToken int
	verbose bool
	
	scope *Scope
}

func (v *parser) err(err string, stuff... interface {}) {
	fmt.Printf(util.TEXT_RED + util.TEXT_BOLD + "Parser error:" + util.TEXT_RESET + " [%s:%d:%d] %s\n",
			v.peek(0).Filename, v.peek(0).LineNumber, v.peek(0).CharNumber, fmt.Sprintf(err, stuff...))
	os.Exit(2)
}

func (v *parser) peek(ahead int) *lexer.Token {
	if ahead < 0 {
		panic(fmt.Sprintf("Tried to peek a negative number: %d", ahead))
	}
	
	if v.currentToken + ahead >= len(v.input) {
		return nil
	}
	
	return v.input[v.currentToken + ahead]
}

func (v *parser) consumeToken() *lexer.Token {
	ret := v.peek(0)
	v.currentToken++
	return ret
}

func (v *parser) pushNode(node Node) {
	v.file.nodes = append(v.file.nodes, node)
}

func (v *parser) tokenMatches(ahead int, t lexer.TokenType, contents string) bool {
	tok := v.peek(ahead)
	return tok.Type == t && (contents == "" || (tok.Contents == contents))
}

func (v *parser) tokensMatch(args... interface{}) bool {
	if len(args) % 2 != 0 {
		panic("passed uneven args to tokensMatch")
	}
	
	for i := 0; i < len(args) / 2; i++ {
		if !(v.tokenMatches(i, args[i * 2].(lexer.TokenType), args[i * 2 + 1].(string))) {
			return false
		}
	}
	return true
}

func Parse(tokens []*lexer.Token, verbose bool) *File {
	p := &parser {
		file: &File {
			nodes: make([]Node, 0),
		},
		input: tokens,
		verbose: verbose,
		scope: newGlobalScope(),
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
	for v.peek(0) != nil {
		if n := v.parseStatement(); n != nil {
			v.pushNode(n)
		} else {
			v.consumeToken() // TODO
		}		
	}
}

func (v *parser) parseStatement() Node {
	if decl := v.parseDecl(); decl != nil {
		return decl
	}
	return nil
}

func (v *parser) parseDecl() Decl {
	if variableDecl := v.parseVariableDecl(); variableDecl != nil {
		return variableDecl
	}
	return nil
}

func (v *parser) parseVariableDecl() *VariableDecl {
	variable := &Variable {}
	varDecl := &VariableDecl {
		Variable: variable,
	}
	
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MUT) {
		variable.Mutable = true
		v.consumeToken()
	}
	
	if v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_OPERATOR, ":") {
		fmt.Println("yes")
		variable.Name = v.consumeToken().Contents // consume name
		
		v.consumeToken() // consume :
		
		if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
			typeName := v.consumeToken().Contents // consume type
		
			variable.Type = v.scope.GetType(typeName)
			if variable.Type == nil {
				v.err("Unrecognized type `%s`", typeName)
			}
		} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
			panic("type inference unimplemented")
		}
		
		if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
			v.consumeToken() // consume =
			varDecl.Assignment = v.parseExpr()
			if varDecl.Assignment == nil {
				v.err("Expected expression in assignment to variable `%s`", variable.Name)
			}
			
		} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
			v.consumeToken()
		} else {
			v.err("Missing semicolon at end of variable declaration")
		}
	} else {
		return nil
	}
	
	if sname := v.scope.InsertVariable(variable); sname != nil {
		v.err("Illegal redeclaration of variable `%s`", variable.Name)
	}
	return varDecl
}

func (v *parser) parseExpr() Expr {
	return nil
}

