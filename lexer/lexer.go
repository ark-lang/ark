package lexer

import (
	"strings"
	"log"
	"fmt"
)

type lexer struct {
	input []rune
	startPos, endPos int
	output []*Token
}

func (v *lexer) err(err string) {
	log.Fatal("Lexer error: ", err);
}

func (v *lexer) peek(ahead int) rune {
	if v.endPos + ahead >= len(v.input) {
		return 0;
	}
	return v.input[v.endPos + ahead]
}

func (v *lexer) consume(num int) {
	v.endPos += num;
}

func (v *lexer) discardBuffer() {
	v.startPos = v.endPos
}

// debugging func
func (v *lexer) printBuffer() {
	fmt.Printf("[%d:%d] `%s`\n", v.startPos, v.endPos, string(v.input[v.startPos:v.endPos]))
}

func (v *lexer) pushToken(t TokenType) {
	tok := &Token { Type: t };
	tok.Contents = string(v.input[v.startPos:v.endPos])
	v.startPos = v.endPos
	v.output = append(v.output, tok)
}

func Lex(input []rune) []*Token {
	v := &lexer { input: input, startPos: 0, endPos: 0}
	v.lex()
	return v.output
}

func (v *lexer) lex() {
	for {
		v.skipLayoutAndComments()
		if isEOF(v.peek(0)) {
			return
		}
		
		v.consume(1)
		v.discardBuffer()
	}
}

func (v *lexer) skipLayoutAndComments() {
	start:
	for isLayout(v.peek(0)) {
		v.consume(1)
	}
	
	// Block comments
	if v.peek(0) == '/' && v.peek(1) == '*' {
		v.consume(2)
		
		for {
			if isEOF(v.peek(0)) {
				v.err("Unterminated block comment")
			}
			if v.peek(0) == '*' && v.peek(1) == '/' {
				v.consume(2)
				goto start
			}
			v.consume(1)
		}
	}
	
	// Single-line comments
	if v.peek(0) == '#' || (v.peek(0) == '/' && v.peek(1) == '/') {
		v.consume(2)
		
		for {
			if isEOL(v.peek(0)) || isEOF(v.peek(0)) {
				v.consume(1)
				goto start
			}
			v.consume(1)
		}
	}
	
	v.printBuffer()
	v.discardBuffer()
}

func isDecimalDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHexDigit(r rune) bool {
	return isDecimalDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func isBinaryDigit(r rune) bool {
	return r == '0' || r == '1'
}

func isOctalDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

func isOperator(r rune) bool {
	return strings.ContainsRune("+-*/=><!~?:|&%^\"'", r);
}

func isExpressionOperator(r rune) bool {
	return strings.ContainsRune("+-*/=><!~?:|&%^\"'()", r);
}

func isSeparator(r rune) bool {
	return strings.ContainsRune(" ;,.`@(){}[]", r);
}

func isEOL(r rune) bool {
	return r == '\n'
}

func isEOF(r rune) bool {
	return r == 0
}

func isLayout(r rune) bool {
	return r <= ' ' && !isEOF(r)
}

