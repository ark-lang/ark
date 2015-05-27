package parser

// Note that you should include a lot of calls to panic() where something's happening that shouldn't be.
// This will help to find bugs. Once the compiler is in a better state, a lot of these calls can be removed.

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ark-lang/ark-go/lexer"
	"github.com/ark-lang/ark-go/util"
)

type parser struct {
	file         *File
	input        []*lexer.Token
	currentToken int
	verbose      bool

	scope            *Scope
	binOpPrecedences map[BinOpType]int
}

func (v *parser) err(err string, stuff ...interface{}) {
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Parser error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		v.peek(0).Filename, v.peek(0).LineNumber, v.peek(0).CharNumber, fmt.Sprintf(err, stuff...))
	os.Exit(2)
}

func (v *parser) peek(ahead int) *lexer.Token {
	if ahead < 0 {
		panic(fmt.Sprintf("Tried to peek a negative number: %d", ahead))
	}

	if v.currentToken+ahead >= len(v.input) {
		return nil
	}

	return v.input[v.currentToken+ahead]
}

func (v *parser) consumeToken() *lexer.Token {
	ret := v.peek(0)
	v.currentToken++
	return ret
}

func (v *parser) pushNode(node Node) {
	v.file.nodes = append(v.file.nodes, node)
}

func (v *parser) pushScope() {
	v.scope = newScope(v.scope)
}

func (v *parser) popScope() {
	v.scope = v.scope.Outer
	if v.scope == nil {
		panic("pushed too many scopes")
	}
}

func (v *parser) tokenMatches(ahead int, t lexer.TokenType, contents string) bool {
	tok := v.peek(ahead)
	return tok.Type == t && (contents == "" || (tok.Contents == contents))
}

func (v *parser) tokensMatch(args ...interface{}) bool {
	if len(args)%2 != 0 {
		panic("passed uneven args to tokensMatch")
	}

	for i := 0; i < len(args)/2; i++ {
		if !(v.tokenMatches(i, args[i*2].(lexer.TokenType), args[i*2+1].(string))) {
			return false
		}
	}
	return true
}

func (v *parser) getPrecedence(op BinOpType) int {
	if p := v.binOpPrecedences[op]; p > 0 {
		return p
	}
	return -1
}

func Parse(tokens []*lexer.Token, verbose bool) *File {
	p := &parser{
		file: &File{
			nodes: make([]Node, 0),
		},
		input:            tokens,
		verbose:          verbose,
		scope:            newGlobalScope(),
		binOpPrecedences: newBinOpPrecedenceMap(),
	}

	if verbose {
		fmt.Println(util.TEXT_BOLD+util.TEXT_GREEN+"Started parsing"+util.TEXT_RESET, tokens[0].Filename)
	}
	p.parse()
	if verbose {
		fmt.Println(util.TEXT_BOLD+util.TEXT_GREEN+"Finished parsing"+util.TEXT_RESET, tokens[0].Filename)
	}

	return p.file
}

func (v *parser) parse() {
	for v.peek(0) != nil {
		if n := v.parseStatement(); n != nil {
			v.pushNode(n)
			if v.verbose && n != nil {
				fmt.Println(n)
			}
		} else {
			v.consumeToken() // TODO
		}
	}
}

func (v *parser) parseStatement() Node {
	for v.tokenMatches(0, lexer.TOKEN_COMMENT, "") || v.tokenMatches(0, lexer.TOKEN_DOCCOMMENT, "") {
		v.consumeToken()
	}

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

func (v *parser) parseType() Type {
	if !(v.peek(0).Type == lexer.TOKEN_IDENTIFIER || v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^")) {
		return nil
	}

	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		v.consumeToken()
		if innerType := v.parseType(); innerType != nil {
			return &PointerType{Addressee: innerType}
		} else {
			v.err("TODO")
		}
	}

	typeName := v.consumeToken().Contents // consume type

	typ := v.scope.GetType(typeName)
	if typ == nil {
		v.err("Unrecognized type `%s`", typeName)
	}
	return typ
}

func (v *parser) parseVariableDecl() *VariableDecl {
	variable := &Variable{}
	varDecl := &VariableDecl{
		Variable: variable,
	}

	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MUT) {
		variable.Mutable = true
		v.consumeToken()
	}

	if v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_OPERATOR, ":") {
		variable.Name = v.consumeToken().Contents // consume name

		v.consumeToken() // consume :

		if typ := v.parseType(); typ != nil {
			variable.Type = typ
		} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
			panic("type inference unimplemented")
		}

		if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
			v.consumeToken() // consume =
			varDecl.Assignment = v.parseExpr()
			if varDecl.Assignment == nil {
				v.err("Expected expression in assignment to variable `%s`", variable.Name)
			}
		}

		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
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
	pri := v.parsePrimaryExpr()
	if pri == nil {
		return nil
	}

	if bin := v.parseBinaryOperator(0, pri); bin != nil {
		return bin
	}
	return pri
}

func (v *parser) parseBinaryOperator(upperPrecedence int, lhand Expr) Expr {
	tok := v.peek(0)
	if tok.Type != lexer.TOKEN_OPERATOR {
		return nil
	}

	for {
		tokPrecedence := v.getPrecedence(stringToBinOpType(v.peek(0).Contents))
		if tokPrecedence < upperPrecedence {
			return lhand
		}

		typ := stringToBinOpType(v.peek(0).Contents)
		if typ == BINOP_ERR {
			panic("yep")
		}

		v.consumeToken()

		rhand := v.parsePrimaryExpr()
		if rhand == nil {
			return nil
		}
		nextPrecedence := v.getPrecedence(stringToBinOpType(v.peek(0).Contents))
		if tokPrecedence < nextPrecedence {
			rhand = v.parseBinaryOperator(tokPrecedence+1, rhand)
			if rhand == nil {
				return nil
			}
		}

		temp := &BinaryExpr{
			Lhand: lhand,
			Rhand: rhand,
			Op:    typ,
		}
		lhand = temp
	}
}

func (v *parser) parsePrimaryExpr() Expr {
	if litExpr := v.parseLiteral(); litExpr != nil {
		return litExpr
	} else if unaryExpr := v.parseUnaryExpr(); unaryExpr != nil {
		return unaryExpr
	}
	return nil
}

func (v *parser) parseUnaryExpr() *UnaryExpr {
	if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "") {
		return nil
	}

	contents := v.peek(0).Contents
	op := stringToUnOpType(contents)
	if op == UNOP_ERR {
		return nil
	}

	v.consumeToken()

	e := v.parseExpr()
	if e == nil {
		v.err("Expected expression after unary operator `%s`", contents)
	}

	return &UnaryExpr{Expr: e, Op: op}
}

func (v *parser) parseLiteral() Expr {
	if numLit := v.parseNumericLiteral(); numLit != nil {
		return numLit
	} else if stringLit := v.parseStringLiteral(); stringLit != nil {
		return stringLit
	} else if runeLit := v.parseRuneLiteral(); runeLit != nil {
		return runeLit
	}
	return nil
}

func (v *parser) parseNumericLiteral() Expr {
	if !v.tokenMatches(0, lexer.TOKEN_NUMBER, "") {
		return nil
	}

	num := v.consumeToken().Contents
	var err error

	if strings.HasPrefix(num, "0x") || strings.HasPrefix(num, "0X") {
		// Hexadecimal integer
		hex := &IntegerLiteral{}
		for _, r := range num[2:] {
			if r == '_' {
				continue
			}
			hex.Value *= 16
			if val := uint64(hexRuneToInt(r)); val >= 0 {
				hex.Value += val
			} else {
				v.err("Malformed hex literal: `%s`", num)
			}
		}
		return hex
	} else if strings.HasPrefix(num, "0b") {
		// Binary integer
		bin := &IntegerLiteral{}
		for _, r := range num[2:] {
			if r == '_' {
				continue
			}
			bin.Value *= 2
			if val := uint64(binRuneToInt(r)); val >= 0 {
				bin.Value += val
			} else {
				v.err("Malformed binary literal: `%s`", num)
			}
		}
		return bin
	} else if strings.HasPrefix(num, "0o") {
		// Octal integer
		oct := &IntegerLiteral{}
		for _, r := range num[2:] {
			if r == '_' {
				continue
			}
			oct.Value *= 8
			if val := uint64(octRuneToInt(r)); val >= 0 {
				oct.Value += val
			} else {
				v.err("Malformed octal literal: `%s`", num)
			}
		}
		return oct
	} else if strings.ContainsRune(num, '.') || strings.HasSuffix(num, "f") || strings.HasSuffix(num, "d") {
		if strings.Count(num, ".") > 1 {
			v.err("Floating-point cannot have multiple periods: `%s`", num)
			return nil
		}

		fnum := num
		if strings.HasSuffix(num, "f") || strings.HasSuffix(num, "d") {
			fnum = fnum[:len(fnum)-1]
		}

		f := &FloatingLiteral{}
		f.Value, err = strconv.ParseFloat(fnum, 64)

		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrSyntax {
				v.err("Malformed floating-point literal: `%s`", num)
				return nil
			} else if err.(*strconv.NumError).Err == strconv.ErrRange {
				v.err("Floating-point literal cannot be represented: `%s`", num)
				return nil
			} else {
				panic("shouldn't be here, ever")
			}
		}

		return f
	} else {
		// Decimal integer
		i := &IntegerLiteral{}
		for _, r := range num {
			if r == '_' {
				continue
			}
			i.Value *= 10
			i.Value += uint64(r - '0')
		}
		return i
	}
}

func (v *parser) parseStringLiteral() *StringLiteral {
	if !v.tokenMatches(0, lexer.TOKEN_STRING, "") {
		return nil
	}
	c := v.consumeToken().Contents
	return &StringLiteral{unescapeString(c[1:len(c)-1])}
}

func (v *parser) parseRuneLiteral() *RuneLiteral {
	if !v.tokenMatches(0, lexer.TOKEN_RUNE, "") {
		return nil
	}

	raw := v.consumeToken().Contents
	c := unescapeString(raw)

	if l := len([]rune(c)); l == 3 {
		return &RuneLiteral{Value: []rune(c)[1]}
	} else if l < 3 {
		panic("lexer problem")
	} else {
		v.err("Rune literal contains more than one rune: `%s`", raw)
		return nil
	}
}
