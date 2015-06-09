package parser

// Note that you should include a lot of calls to panic() where something's happening that shouldn't be.
// This will help to find bugs. Once the compiler is in a better state, a lot of these calls can be removed.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ark-lang/ark/common"
	"github.com/ark-lang/ark/lexer"
	"github.com/ark-lang/ark/util"
)

type parser struct {
	file         *File
	input        []*lexer.Token
	currentToken int
	verbose      bool

	scope             *Scope
	binOpPrecedences  map[BinOpType]int
	attrs             []*Attr
	curNodeTokenStart int
	docCommentsBuf    []*DocComment
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
	v.file.Nodes = append(v.file.Nodes, node)
}

func (v *parser) pushScope() {
	v.scope = newScope(v.scope)
}

func (v *parser) popScope() {
	v.scope = v.scope.Outer
	if v.scope == nil {
		panic("popped too many scopes")
	}
}

func (v *parser) fetchAttrs() []*Attr {
	ret := v.attrs
	v.attrs = nil
	return ret
}

func (v *parser) fetchDocComments() []*DocComment {
	ret := v.docCommentsBuf
	v.docCommentsBuf = nil
	return ret
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

func Parse(input *common.Sourcefile, verbose bool) *File {
	p := &parser{
		file: &File{
			Nodes: make([]Node, 0),
			Name:  input.Filename,
		},
		input:            input.Tokens,
		verbose:          verbose,
		scope:            newGlobalScope(),
		binOpPrecedences: newBinOpPrecedenceMap(),
	}

	if verbose {
		fmt.Println(util.TEXT_BOLD+util.TEXT_GREEN+"Started parsing"+util.TEXT_RESET, input.Filename)
	}
	t := time.Now()
	p.parse()
	sem := &semanticAnalyzer{file: p.file}
	sem.analyze()
	dur := time.Since(t)
	if verbose {
		for _, n := range p.file.Nodes {
			fmt.Println(n.String())
		}
		fmt.Printf(util.TEXT_BOLD+util.TEXT_GREEN+"Finished parsing"+util.TEXT_RESET+" %s (%.2fms)\n",
			input.Filename, float32(dur.Nanoseconds())/1000000)
	}

	return p.file
}

func (v *parser) parse() {
	for v.peek(0) != nil {
		if n := v.parseNode(); n != nil {
			v.pushNode(n)
		} else {
			panic("what's this over here?")
		}
	}
}

func (v *parser) parseDocComment() *DocComment {
	if !v.tokenMatches(0, lexer.TOKEN_DOCCOMMENT, "") {
		return nil
	}
	tok := v.consumeToken()
	doc := &DocComment{
		StartLine: tok.LineNumber,
		EndLine:   tok.EndLineNumber,
	}

	if strings.HasPrefix(tok.Contents, "/**") {
		doc.Contents = tok.Contents[3 : len(tok.Contents)-2]
	} else if strings.HasPrefix(tok.Contents, "///") {
		doc.Contents = tok.Contents[3:]
	} else {
		panic(fmt.Sprintf("How did this doccomment get through the lexer??\n`%s`", tok.Contents))
	}

	return doc
}

func (v *parser) parseNode() Node {
	v.docCommentsBuf = make([]*DocComment, 0)
	for v.tokenMatches(0, lexer.TOKEN_COMMENT, "") || v.tokenMatches(0, lexer.TOKEN_DOCCOMMENT, "") {
		if v.tokenMatches(0, lexer.TOKEN_DOCCOMMENT, "") {
			v.docCommentsBuf = append(v.docCommentsBuf, v.parseDocComment())
		} else {
			v.consumeToken()
		}
	}

	v.attrs = v.parseAttrs()
	var ret Node
	if decl := v.parseDecl(); decl != nil {
		ret = decl
	} else if stat := v.parseStat(); stat != nil {
		ret = stat
	}

	if ret != nil {
		if len(v.attrs) > 0 {
			v.err("%s does not accept attributes", util.CapitalizeFirst(ret.NodeName()))
		}
		if len(v.docCommentsBuf) > 0 {
			v.err("%s does not accept documentation comments", util.CapitalizeFirst(ret.NodeName())) // TODO fix for func decls
		}
		return ret
	}

	return nil
}

func (v *parser) parseStat() Stat {
	var ret Stat
	line, char := v.peek(0).LineNumber, v.peek(0).CharNumber

	if ifStat := v.parseIfStat(); ifStat != nil {
		ret = ifStat
	} else if loopStat := v.parseLoopStat(); loopStat != nil {
		ret = loopStat
	} else if returnStat := v.parseReturnStat(); returnStat != nil {
		ret = returnStat
	} else if callStat := v.parseCallStat(); callStat != nil {
		ret = callStat
	} else if assignStat := v.parseAssignStat(); assignStat != nil {
		ret = assignStat
	} else {
		return nil
	}

	ret.setPos(line, char)
	return ret
}

func (v *parser) parseAssignStat() *AssignStat {
	for i := 0; true; i++ {
		if v.peek(i) == nil || v.tokenMatches(i, lexer.TOKEN_SEPARATOR, ";") {
			return nil
		} else if v.tokenMatches(i, lexer.TOKEN_OPERATOR, "=") {
			break
		}
	}

	assign := &AssignStat{}
	line, char := v.peek(0).LineNumber, v.peek(0).CharNumber

	if deref := v.parseDerefExpr(); deref != nil {
		deref.setPos(line, char)
		assign.Deref = deref
	} else if access := v.parseAccessExpr(); access != nil {
		access.setPos(line, char)
		assign.Access = access
	} else {
		v.err("Malformed assignment statement")
	}

	if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.err("Expected `=`, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	if expr := v.parseExpr(); expr != nil {
		assign.Assignment = expr
	} else {
		v.err("Expected expression in variable assignment, found `%s`", v.peek(0).Contents)
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.err("Expected semicolon after assignment statement, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	return assign

}

func (v *parser) parseCallStat() *CallStat {
	if call := v.parseCallExpr(); call != nil {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
			v.consumeToken()
			return &CallStat{Call: call}
		}
		v.err("Expected semicolon after function call statement, found `%s`", v.peek(0))
	}
	return nil
}

func (v *parser) parseDecl() Decl {
	var ret Decl
	line, char := v.peek(0).LineNumber, v.peek(0).CharNumber
	if structureDecl := v.parseStructDecl(); structureDecl != nil {
		ret = structureDecl
	} else if functionDecl := v.parseFunctionDecl(); functionDecl != nil {
		ret = functionDecl
	} else if variableDecl := v.parseVariableDecl(true); variableDecl != nil {
		ret = variableDecl
	} else {
		return nil
	}
	ret.setPos(line, char)
	return ret
}

func (v *parser) parseType() Type {
	if !(v.peek(0).Type == lexer.TOKEN_IDENTIFIER || v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^")) {
		return nil
	}

	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		v.consumeToken()
		if innerType := v.parseType(); innerType != nil {
			return pointerTo(innerType)
		} else {
			v.err("Expected type name")
		}
	}

	typeName := v.consumeToken().Contents // consume type

	typ := v.scope.GetType(typeName)
	if typ == nil {
		v.err("Unrecognized type `%s`", typeName)
	}
	return typ
}

func (v *parser) parseFunctionDecl() *FunctionDecl {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
		return nil
	}

	function := &Function{Attrs: v.fetchAttrs()}

	v.consumeToken()

	// name
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
		function.Name = v.consumeToken().Contents
	} else {
		v.err("Function expected an identifier")
	}

	if isReservedKeyword(function.Name) {
		v.err("Cannot name function reserved keyword `%s`", function.Name)
	}

	if vname := v.scope.InsertFunction(function); vname != nil {
		v.err("Illegal redeclaration of function `%s`", function.Name)
	}

	funcDecl := &FunctionDecl{
		Function: function,
		docs:     v.fetchDocComments(),
	}

	v.pushScope()

	// Arguments
	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.err("Expected `(` after function identifier, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	function.Parameters = make([]*VariableDecl, 0)
	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.consumeToken()
	} else {
		for {
			if decl := v.parseVariableDecl(false); decl != nil {
				if decl.Assignment != nil {
					v.err("Assignment in function parameter `%s`", decl.Variable.Name)
				}

				function.Parameters = append(function.Parameters, decl)
			} else {
				v.err("Expected function parameter, found `%s`", v.peek(0).Contents)
			}

			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
				v.consumeToken()
				break
			} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
				v.consumeToken()
			} else {
				v.err("Expected `)` or `,` after function parameter, found `%s`", v.peek(0).Contents)
			}
		}
	}

	// return type
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, ":") {
		v.consumeToken()

		// mutable return type
		if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MUT) {
			v.consumeToken()
			function.Mutable = true
		}

		// actual return type
		if typ := v.parseType(); typ != nil {
			function.ReturnType = typ
		} else {
			v.err("Expected function return type after colon for function `%s`", function.Name)
		}
	}

	// block
	if block := v.parseBlock(); block != nil {
		funcDecl.Function.Body = block
	} else {
		v.err("Expecting block after function decl even though some point prototypes should be support lol whatever")
	}

	v.popScope()

	return funcDecl
}

func (v *parser) parseBlock() *Block {
	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		return nil
	}

	v.consumeToken()

	block := newBlock()

	for {
		for v.tokenMatches(0, lexer.TOKEN_COMMENT, "") || v.tokenMatches(0, lexer.TOKEN_DOCCOMMENT, "") {
			v.consumeToken()
		}
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			v.consumeToken()
			return block
		}

		if s := v.parseNode(); s != nil {
			block.appendNode(s)
		} else {
			v.err("Expected statment, found something else")
		}
	}
}

func (v *parser) parseReturnStat() *ReturnStat {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_RETURN) {
		return nil
	}

	v.consumeToken()

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.consumeToken()
		return &ReturnStat{}
	}

	expr := v.parseExpr()
	if expr == nil {
		v.err("Expected expression in return statement, found `%s`", v.peek(0).Contents)
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.err("Expected semicolon after return statement, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()
	return &ReturnStat{Value: expr}
}

func (v *parser) parseIfStat() *IfStat {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_IF) {
		return nil
	}
	v.consumeToken()

	ifStat := &IfStat{
		Exprs:  make([]Expr, 0),
		Bodies: make([]*Block, 0),
	}

	for {
		expr := v.parseExpr()
		if expr == nil {
			v.err("Expected expression for if condition, found `%s`", v.peek(0).Contents)
		}
		ifStat.Exprs = append(ifStat.Exprs, expr)

		v.pushScope()
		body := v.parseBlock()
		v.popScope()
		if body == nil {
			v.err("Expected body after if condition, found `%s`", v.peek(0).Contents)
		}
		ifStat.Bodies = append(ifStat.Bodies, body)

		if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_ELSE) {
			v.consumeToken()
			if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_IF) {
				v.consumeToken()
				continue
			} else {
				v.pushScope()
				body := v.parseBlock()
				v.popScope()
				if body == nil {
					v.err("Expected else body, found `%s`", v.peek(0).Contents)
				}
				ifStat.Else = body
				return ifStat
			}
		} else {
			return ifStat
		}
	}
}

func (v *parser) parseLoopStat() *LoopStat {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FOR) {
		return nil
	}
	v.consumeToken()

	loop := &LoopStat{}

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") { // infinite loop
		loop.LoopType = LOOP_TYPE_INFINITE

		loop.Body = v.parseBlock()
		if loop.Body == nil {
			v.err("Malformed infinite loop body")
		}
		return loop
	}

	if cond := v.parseExpr(); cond != nil {
		loop.LoopType = LOOP_TYPE_CONDITIONAL
		loop.Condition = cond

		loop.Body = v.parseBlock()
		if loop.Body == nil {
			v.err("Malformed infinite loop body")
		}
		return loop
	}

	v.err("Malformed `%s` loop", KEYWORD_FOR)
	return nil
}

func (v *parser) parseStructDecl() *StructDecl {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_STRUCT) {
		return nil
	}
	struc := &StructType{}

	v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
		v.err("Expected identifier after `struct` keyword, found `%s`", v.peek(0).Contents)
	}
	struc.Name = v.consumeToken().Contents

	if isReservedKeyword(struc.Name) {
		v.err("Cannot name struct reserved keyword `%s`", struc.Name)
	}

	if sname := v.scope.InsertType(struc); sname != nil {
		v.err("Illegal redeclaration of type `%s`", struc.Name)
	}

	struc.attrs = v.fetchAttrs()

	// TODO semi colons i.e. struct with no body?
	var itemCount = 0
	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.consumeToken()

		v.pushScope()

		for {
			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
				v.consumeToken()
				break
			}

			line, char := v.peek(0).LineNumber, v.peek(0).CharNumber
			if variable := v.parseVariableDecl(false); variable != nil {
				if variable.Variable.Mutable {
					v.err("Cannot specify `mut` keyword on struct member: `%s`", variable.Variable.Name)
				}
				struc.addVariableDecl(variable)
				variable.setPos(line, char)
				itemCount++
			} else {
				v.err("Invalid structure item in structure `%s`", struc.Name)
			}

			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
				v.consumeToken()
			}
		}

		v.popScope()
	} else {
		v.err("Expected body after struct identifier, found `%s`", v.peek(0).Contents)
	}

	return &StructDecl{Struct: struc}
}

func (v *parser) parseVariableDecl(needSemicolon bool) *VariableDecl {
	variable := &Variable{Attrs: v.fetchAttrs()}
	varDecl := &VariableDecl{Variable: variable}

	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MUT) {
		variable.Mutable = true
		v.consumeToken()
	}

	if !v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_OPERATOR, ":") {
		return nil
	}
	variable.Name = v.consumeToken().Contents // consume name

	if isReservedKeyword(variable.Name) {
		v.err("Cannot name variable reserved keyword `%s`", variable.Name)
	}

	v.consumeToken() // consume :

	if typ := v.parseType(); typ != nil {
		variable.Type = typ
	} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		variable.Type = nil
	}

	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.consumeToken() // consume =
		varDecl.Assignment = v.parseExpr()
		if varDecl.Assignment == nil {
			v.err("Expected expression in assignment to variable `%s`", variable.Name)
		}
	}

	if sname := v.scope.InsertVariable(variable); sname != nil {
		v.err("Illegal redeclaration of variable `%s`", variable.Name)
	}

	if needSemicolon {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
			v.consumeToken()
		} else {
			v.err("Expected semicolon at end of variable declaration, found `%s`", v.peek(0).Contents)
		}
	}

	varDecl.docs = v.fetchDocComments()

	return varDecl
}

func (v *parser) parseExpr() Expr {
	pri := v.parsePrimaryExpr()
	line, char := v.peek(0).LineNumber, v.peek(0).CharNumber
	if pri == nil {
		return nil
	}
	pri.setPos(line, char)

	if bin := v.parseBinaryOperator(0, pri); bin != nil {
		bin.setPos(line, char)
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
	if bracketExpr := v.parseBracketExpr(); bracketExpr != nil {
		return bracketExpr
	} else if litExpr := v.parseLiteral(); litExpr != nil {
		return litExpr
	} else if derefExpr := v.parseDerefExpr(); derefExpr != nil {
		return derefExpr
	} else if unaryExpr := v.parseUnaryExpr(); unaryExpr != nil {
		return unaryExpr
	} else if castExpr := v.parseCastExpr(); castExpr != nil {
		return castExpr
	} else if callExpr := v.parseCallExpr(); callExpr != nil {
		return callExpr
	} else if accessExpr := v.parseAccessExpr(); accessExpr != nil {
		return accessExpr
	}

	return nil
}

func (v *parser) parseBracketExpr() *BracketExpr {
	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}
	v.consumeToken()

	expr := v.parseExpr()
	if expr == nil {
		v.err("Expected expression after `)`, found `%s`", v.peek(0).Contents)
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Expected matching `)`, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	return &BracketExpr{Expr: expr}
}

func (v *parser) parseAccessExpr() *AccessExpr {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
		return nil
	}

	access := &AccessExpr{}

	ident := v.consumeToken().Contents
	access.Variable = v.scope.GetVariable(ident)
	if access.Variable == nil {
		v.err("Unresolved variable `%s`", ident)
	}

	for {
		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ".") {
			return access
		}
		v.consumeToken()

		structType, ok := access.Variable.Type.(*StructType)
		if !ok {
			v.err("Cannot access member of `%s`, type `%s`", access.Variable.Name, access.Variable.Type.TypeName())
		}

		memberName := v.consumeToken().Contents
		decl := structType.getVariableDecl(memberName)
		if decl == nil {
			v.err("Struct `%s` does not contain member `%s`", structType.TypeName(), memberName)
		}

		access.StructVariables = append(access.StructVariables, access.Variable)
		access.Variable = decl.Variable
	}
}

func (v *parser) parseCallExpr() *CallExpr {
	if !v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}

	function := v.scope.GetFunction(v.peek(0).Contents)
	if function == nil {
		v.err("Call to undefined function `%s`", v.peek(0).Contents)
	}
	v.consumeToken()
	v.consumeToken()

	args := make([]Expr, 0)
	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.consumeToken()
	} else {
		for {
			if expr := v.parseExpr(); expr != nil {
				args = append(args, expr)
			} else {
				v.err("Expected function argument, found `%s`", v.peek(0).Contents)
			}

			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
				v.consumeToken()
				break
			} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
				v.consumeToken()
			} else {
				v.err("Expected `)` or `,` after function argument, found `%s`", v.peek(0).Contents)
			}
		}
	}

	return &CallExpr{
		Arguments: args,
		Function:  function,
	}
}

func (v *parser) parseCastExpr() *CastExpr {
	if !v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}

	typ := v.scope.GetType(v.peek(0).Contents)
	if typ == nil {
		return nil
	}
	v.consumeToken()
	v.consumeToken()

	expr := v.parseExpr()
	if expr == nil {
		v.err("Expected expression in typecast, found `%s`", v.peek(0))
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Exprected `)` at the end of typecase, found `%s`", v.peek(0))
	}
	v.consumeToken()

	return &CastExpr{
		Type: typ,
		Expr: expr,
	}
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

func (v *parser) parseDerefExpr() *DerefExpr {
	if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		return nil
	}
	v.consumeToken()

	e := v.parseExpr()
	if e == nil {
		v.err("Expected expression after dereference operator")
	}

	return &DerefExpr{Expr: e}
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
	strLen := len(c)
	return &StringLiteral{Value: UnescapeString(c[1 : strLen-1]), StrLen: strLen}
}

func (v *parser) parseRuneLiteral() *RuneLiteral {
	if !v.tokenMatches(0, lexer.TOKEN_RUNE, "") {
		return nil
	}

	raw := v.consumeToken().Contents
	c := UnescapeString(raw)

	if l := len([]rune(c)); l == 3 {
		return &RuneLiteral{Value: []rune(c)[1]}
	} else if l < 3 {
		panic("lexer problem")
	} else {
		v.err("Rune literal contains more than one rune: `%s`", raw)
		return nil
	}
}
