package lexer

// Note that most of the panic() calls should be removed once the lexer is bug-free.

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/ark-lang/ark/src/util/log"

	"github.com/ark-lang/ark/src/util"
)

type lexer struct {
	input            *Sourcefile
	startPos, endPos int
	curPos           Position
	tokStart         Position
}

func (v *lexer) errPos(pos Position, err string, stuff ...interface{}) {
	log.Errorln("lexer", util.TEXT_RED+util.TEXT_BOLD+"Lexer error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Error("lexer", v.input.MarkPos(pos))

	os.Exit(1)
}

func (v *lexer) err(err string, stuff ...interface{}) {
	v.errPos(v.curPos, err, stuff...)
}

func (v *lexer) peek(ahead int) rune {
	if ahead < 0 {
		panic(fmt.Sprintf("Tried to peek a negative number: %d", ahead))
	}

	if v.endPos+ahead >= len(v.input.Contents) {
		return 0
	}
	return v.input.Contents[v.endPos+ahead]
}

func (v *lexer) consume() {
	v.curPos.Char++
	if v.peek(0) == '\n' {
		v.curPos.Char = 1
		v.curPos.Line++
		v.input.NewLines = append(v.input.NewLines, v.endPos)
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

	v.tokStart = v.curPos
}

// debugging func
func (v *lexer) printBuffer() {
	log.Debug("lexer", "[%d:%d] `%s`\n", v.startPos, v.endPos, string(v.input.Contents[v.startPos:v.endPos]))
}

func (v *lexer) pushToken(t TokenType) {
	tok := &Token{
		Type:     t,
		Contents: string(v.input.Contents[v.startPos:v.endPos]),
		Where:    NewSpan(v.tokStart, v.curPos),
	}

	v.input.Tokens = append(v.input.Tokens, tok)

	log.Verbose("lexer", "[%4d:%4d:%-17s] `%s`\n", v.startPos, v.endPos, tok.Type, tok.Contents)

	v.discardBuffer()
}

func Lex(input *Sourcefile) []*Token {
	v := &lexer{
		input:    input,
		startPos: 0,
		endPos:   0,
		curPos:   Position{Filename: input.Name, Line: 1, Char: 1},
		tokStart: Position{Filename: input.Name, Line: 1, Char: 1},
	}

	log.Verboseln("lexer", util.TEXT_BOLD+util.TEXT_GREEN+"Starting lexing "+util.TEXT_RESET+input.Name)
	t := time.Now()
	v.lex()
	dur := time.Since(t)
	log.Verbose("lexer", util.TEXT_BOLD+util.TEXT_GREEN+"Finished lexing"+util.TEXT_RESET+" %s (%.2fms)\n",
		input.Name, float32(dur)/1000000)
	return v.input.Tokens
}

func (v *lexer) lex() {
	for {
		v.skipLayoutAndComments()

		if isEOF(v.peek(0)) {
			v.input.NewLines = append(v.input.NewLines, v.endPos)
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

	pos := v.curPos

	// Block comments
	// TODO refactor this it's kind of messy
	if v.peek(0) == '/' && v.peek(1) == '*' {
		v.consume()
		v.consume()
		isDoc := v.peek(0) == '*'

		for {
			if isEOF(v.peek(0)) {
				v.errPos(pos, "Unterminated block comment")
			}
			if v.peek(0) == '*' && v.peek(1) == '/' {
				v.consume()
				v.consume()
				if isDoc {
					v.pushToken(TOKEN_DOCCOMMENT)
				} else {
					v.discardBuffer()
				}
				goto start
			}
			v.consume()
		}
	}

	// Single-line comments
	if v.peek(0) == '/' && v.peek(1) == '/' {
		v.consume()
		v.consume()
		isDoc := v.peek(0) == '/'

		for {
			if isEOL(v.peek(0)) || isEOF(v.peek(0)) {
				if isDoc {
					v.pushToken(TOKEN_DOCCOMMENT)
				} else {
					v.discardBuffer()
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
			} else if peek := unicode.ToLower(v.peek(0)); peek == 'f' || peek == 'd' || peek == 'q' {
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
	pos := v.curPos

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
			v.errPos(pos, "Unterminated string literal")
		} else {
			v.consume()
		}
	}
}

func (v *lexer) recognizeCharacterToken() {
	pos := v.curPos

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
			v.errPos(pos, "Unterminated character literal")
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
	return strings.ContainsRune("+-*/=><!~?:|&%^", r)
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
