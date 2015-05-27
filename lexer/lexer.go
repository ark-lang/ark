package lexer

// Note that most of the panic() calls should be removed once the lexer is bug-free.

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/ark-lang/ark-go/util"
)

type lexer struct {
	input                  []rune
	startPos, endPos       int
	output                 []*Token
	filename               string
	charNumber, lineNumber int
	verbose                bool
}

func (v *lexer) errWithCustomPosition(ln, cn int, err string, stuff ...interface{}) {
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Lexer error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		v.filename, ln, cn, fmt.Sprintf(err, stuff...))
	os.Exit(1)
}

func (v *lexer) err(err string, stuff ...interface{}) {
	v.errWithCustomPosition(v.lineNumber, v.charNumber, err, stuff...)
}

func (v *lexer) peek(ahead int) rune {
	if ahead < 0 {
		panic(fmt.Sprintf("Tried to peek a negative number: %d", ahead))
	}

	if v.endPos+ahead >= len(v.input) {
		return 0
	}
	return v.input[v.endPos+ahead]
}

func (v *lexer) consume() {
	v.charNumber++
	if v.peek(0) == '\n' {
		v.charNumber = 1
		v.lineNumber++
	}

	v.endPos++
}

func (v *lexer) expect(r rune) {
	if v.peek(0) == r {
		v.consume()
	} else {
		v.err("Expected `%c`, found `%c`", r, v.peek(0))
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
	tok := &Token{
		Type:       t,
		Filename:   v.filename,
		CharNumber: v.charNumber,
		LineNumber: v.lineNumber,
		Contents:   string(v.input[v.startPos:v.endPos]),
	}

	v.output = append(v.output, tok)

	if v.verbose {
		fmt.Printf("[%4d:%4d:%-17s] `%s`\n", v.startPos, v.endPos, tok.Type, tok.Contents)
	}

	v.startPos = v.endPos
}

func Lex(input []rune, filename string, verbose bool) []*Token {
	v := &lexer{
		input:      input,
		startPos:   0,
		endPos:     0,
		filename:   filename,
		charNumber: 1,
		lineNumber: 1,
		verbose:    verbose,
	}

	if v.verbose {
		fmt.Println(util.TEXT_BOLD+util.TEXT_GREEN+"Starting lexing"+util.TEXT_RESET, filename)
	}
	v.lex()
	if v.verbose {
		fmt.Println(util.TEXT_BOLD+util.TEXT_GREEN+"Finished lexing"+util.TEXT_RESET, filename)
	}
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
		} else if isOperator(v.peek(0)) {
			v.recognizeOperatorToken()
		} else if isSeparator(v.peek(0)) {
			v.recognizeSeparatorToken()
		} else {
			v.err("Unrecognised token")
		}
	}
}

func (v *lexer) skipLayoutAndComments() {
start:
	for isLayout(v.peek(0)) {
		v.consume()
	}

	v.discardBuffer()

	lineNumber := v.lineNumber
	charNumber := v.charNumber

	// Block comments
	if v.peek(0) == '/' && v.peek(1) == '*' {
		v.consume()
		v.consume()

		for {
			if isEOF(v.peek(0)) {
				v.errWithCustomPosition(lineNumber, charNumber, "Unterminated block comment")
			}
			if v.peek(0) == '*' && v.peek(1) == '/' {
				v.consume()
				v.consume()
				v.pushToken(TOKEN_COMMENT)
				goto start
			}
			v.consume()
		}
	}

	// Single-line comments
	if v.peek(0) == '#' || (v.peek(0) == '/' && v.peek(1) == '/') {
		isDoc := v.peek(0) == '#'
		v.consume()
		if v.peek(0) == '/' {
			v.consume()
		}

		for {
			if isEOL(v.peek(0)) || isEOF(v.peek(0)) {
				if isDoc {
					v.pushToken(TOKEN_DOCCOMMENT)
				} else {
					v.pushToken(TOKEN_COMMENT)
				}
				v.consume()
				goto start
			}
			v.consume()
		}
	}

	//v.printBuffer()
	v.discardBuffer()
}

func (v *lexer) recognizeNumberToken() {
	v.consume()

	if v.peek(0) == 'x' || v.peek(0) == 'X' {
		// Hexadecimal
		v.consume()
		for isHexDigit(v.peek(0)) || v.peek(0) == '_' {
			v.consume()
		}
		v.pushToken(TOKEN_NUMBER)
	} else if v.peek(0) == 'b' {
		// Binary
		v.consume()
		for isBinaryDigit(v.peek(0)) || v.peek(0) == '_' {
			v.consume()
		}
		v.pushToken(TOKEN_NUMBER)
	} else if v.peek(0) == 'o' {
		// Octal
		v.consume()
		for isOctalDigit(v.peek(0)) || v.peek(0) == '_' {
			v.consume()
		}
		v.pushToken(TOKEN_NUMBER)
	} else {
		// Decimal or floating
		for {
			if isDecimalDigit(v.peek(0)) || v.peek(0) == '_' || v.peek(0) == '.' {
				v.consume()
				continue
			} else if v.peek(0) == 'f' || v.peek(0) == 'd' {
				v.consume()
			}
			v.pushToken(TOKEN_NUMBER)
			return
		}
	}
}

func (v *lexer) recognizeIdentifierToken() {
	v.consume()

	for isLetter(v.peek(0)) || isDecimalDigit(v.peek(0)) || v.peek(0) == '_' {
		v.consume()
	}

	v.pushToken(TOKEN_IDENTIFIER)
}

func (v *lexer) recognizeStringToken() {
	lineNumber := v.lineNumber
	charNumber := v.charNumber

	v.expect('"')

	for {
		if v.peek(0) == '\\' && v.peek(1) == '"' {
			v.consume()
			v.consume()
		} else if v.peek(0) == '"' {
			v.consume()
			v.pushToken(TOKEN_STRING)
			return
		} else if isEOF(v.peek(0)) {
			v.errWithCustomPosition(lineNumber, charNumber, "Unterminated string literal")
		} else {
			v.consume()
		}
	}
}

func (v *lexer) recognizeCharacterToken() {
	lineNumber := v.lineNumber
	charNumber := v.charNumber

	v.expect('\'')

	if v.peek(0) == '\'' {
		v.err("Empty character constant")
	}

	for {
		if v.peek(0) == '\\' && (v.peek(1) == '\'' || v.peek(1) == '\\') {
			v.consume()
			v.consume()
		} else if v.peek(0) == '\'' {
			v.consume()
			v.pushToken(TOKEN_RUNE)
			return
		} else if isEOF(v.peek(0)) {
			v.errWithCustomPosition(lineNumber, charNumber, "Unterminated character literal")
		} else {
			v.consume()
		}
	}
}

func (v *lexer) recognizeOperatorToken() {
	// stop := from being treated as an operator
	// treat them as individual operators instead.
	if v.peek(0) == ':' && v.peek(1) == '=' {
		v.consume()
	} else {
		v.consume()
		if isOperator(v.peek(0)) && v.peek(0) != '^' {
			v.consume()
		}
	}

	v.pushToken(TOKEN_OPERATOR)
}

func (v *lexer) recognizeSeparatorToken() {
	v.consume()
	v.pushToken(TOKEN_SEPARATOR)
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
	return strings.ContainsRune("+-*/=><!~?:|&%^\"'", r)
}

func isExpressionOperator(r rune) bool {
	return strings.ContainsRune("+-*/=><!~?:|&%^\"'()", r) // this is unused?
}

func isSeparator(r rune) bool {
	return strings.ContainsRune(" ;,.`@(){}[]", r)
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
