package lexer

// Note that most of the panic() calls should be removed once the lexer is bug-free.

import (
	"strings"
	"log"
	"fmt"
	"unicode"
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
	if ahead < 0 {
		panic(fmt.Sprintf("Tried to peek a negative number: %d", ahead))
	}
	
	if v.endPos + ahead >= len(v.input) {
		return 0;
	}
	return v.input[v.endPos + ahead]
}

func (v *lexer) consume(num int) {
	if num <= 0 {
		panic(fmt.Sprintf("Tried to consume %d runes", num))
	}
	v.endPos += num;
}

func (v *lexer) expect(r rune) {
	if v.peek(0) == r {
		v.consume(1)
	} else {
		v.err(fmt.Sprintf("Expected `%c`, found `%c`", r, v.peek(0)))
	}
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
	
	if true {
		fmt.Printf("[%4d:%4d:%-17s] `%s`\n", v.startPos, v.endPos, tok.Type, tok.Contents)
	}
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
		
		if isDecimalDigit(v.peek(0)) {
			v.recognizeNumberToken()
		} else if isLetter(v.peek(0)) || v.peek(0) == '_' {
			v.recognizeIdentifierToken()
		} else if v.peek(0) == '"' {
			v.recognizeStringToken()
		} else if v.peek(0) == '\'' {
			v.recognizeCharacterToken()
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
		v.consume(1)
		if v.peek(0) == '/' {
			v.consume(1)
		}
		
		for {
			if isEOL(v.peek(0)) || isEOF(v.peek(0)) {
				v.consume(1)
				goto start
			}
			v.consume(1)
		}
	}
	
	//v.printBuffer()
	v.discardBuffer()
}

func (v *lexer) recognizeNumberToken() {
	v.consume(1)
	
	if v.peek(0) == 'x' || v.peek(0) == 'X' {
		// Hexadecimal
		v.consume(1)
		for isHexDigit(v.peek(0)) || v.peek(0) == '_' {
			v.consume(1)
		}
		v.pushToken(TOKEN_NUMBER)
	} else if v.peek(0) == 'b' {
		// Binary
		v.consume(1)
		for isBinaryDigit(v.peek(0)) || v.peek(0) == '_' {
			v.consume(1)
		}
		v.pushToken(TOKEN_NUMBER)
	} else if v.peek(0) == 'o' {
		// Octal
		v.consume(1)
		for isOctalDigit(v.peek(0)) || v.peek(0) == '_' {
			v.consume(1)
		}
		v.pushToken(TOKEN_NUMBER)
	} else {
		// Decimal
		for {
			if isDecimalDigit(v.peek(0)) || v.peek(0) == '_' {
				v.consume(1)
				continue
			} else if v.peek(0) == 'f' || v.peek(0) == 'd' {
				v.consume(1)
			}
			v.pushToken(TOKEN_NUMBER)
			return;
		}
	}
}

func (v *lexer) recognizeIdentifierToken() {
	v.consume(1)
	
	for isLetter(v.peek(0)) || isDecimalDigit(v.peek(0)) || v.peek(0) == '_' {
		v.consume(1)
	}
	
	v.pushToken(TOKEN_IDENTIFIER)
}

func (v *lexer) recognizeStringToken() {
	v.expect('"')
	
	for {
		if v.peek(0) == '\\' && v.peek(1) == '"' {
			v.consume(2)
		} else if v.peek(0) == '"' {
			v.consume(1)
			v.pushToken(TOKEN_STRING)
			return
		} else if isEOF(v.peek(0))	{
			v.err("Unterminated string constant")
		} else {
			v.consume(1)
		}
	}
}

func (v *lexer) recognizeCharacterToken() {
	v.expect('\'')
	
	if v.peek(0) == '\'' {
		v.err("Empty character constant")
	}
	
	for {
		if v.peek(0) == '\\' && v.peek(1) == '\'' {
			v.consume(2)
		} else if v.peek(0) == '\'' {
			v.consume(1)
			v.pushToken(TOKEN_CHARACTER)
			return
		} else if isEOF(v.peek(0)) {
			v.err("Unterminated character constant")
		} else {
			v.consume(1)
		}
	}
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

func isLetter(r rune) bool {
	return unicode.IsLetter(r)
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
	return (r <= ' ' || unicode.IsSpace(r)) && !isEOF(r)
}

