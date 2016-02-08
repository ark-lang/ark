package parser

// Note that you should include a lot of calls to panic() where something's happening that shouldn't be.
// This will help to find bugs. Once the compiler is in a better state, a lot of these calls can be removed.

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/ark-lang/ark/src/util/log"

	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util"
)

type parser struct {
	input        *lexer.Sourcefile
	currentToken int
	tree         *ParseTree

	binOpPrecedences  map[BinOpType]int
	curNodeTokenStart int
	ruleStack         []string
	deps              []*NameNode
}

func Parse(input *lexer.Sourcefile) (*ParseTree, []*NameNode) {
	p := &parser{
		input:            input,
		binOpPrecedences: newBinOpPrecedenceMap(),
		tree:             &ParseTree{Source: input},
	}

	log.Timed("parsing", input.Name, func() {
		p.parse()
	})

	return p.tree, p.deps
}

func (v *parser) err(err string, stuff ...interface{}) {
	v.errPos(err, stuff...)
}

func (v *parser) errToken(err string, stuff ...interface{}) {
	tok := v.peek(0)
	if tok != nil {
		v.errTokenSpecific(tok, err, stuff...)
	} else {
		lastTok := v.input.Tokens[len(v.input.Tokens)-1]
		v.errTokenSpecific(lastTok, err, stuff...)
	}

}

func (v *parser) errPos(err string, stuff ...interface{}) {
	tok := v.peek(0)
	if tok != nil {
		v.errPosSpecific(v.peek(0).Where.Start(), err, stuff...)
	} else {
		lastTok := v.input.Tokens[len(v.input.Tokens)-1]
		v.errPosSpecific(lastTok.Where.Start(), err, stuff...)
	}

}

func (v *parser) errTokenSpecific(tok *lexer.Token, err string, stuff ...interface{}) {
	v.dumpRules()
	log.Errorln("parser",
		util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
		tok.Where.Filename, tok.Where.StartLine, tok.Where.StartChar,
		fmt.Sprintf(err, stuff...))

	log.Error("parser", v.input.MarkSpan(tok.Where))

	os.Exit(util.EXIT_FAILURE_PARSE)
}

func (v *parser) errPosSpecific(pos lexer.Position, err string, stuff ...interface{}) {
	v.dumpRules()
	log.Errorln("parser",
		util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
		pos.Filename, pos.Line, pos.Char,
		fmt.Sprintf(err, stuff...))

	log.Error("parser", v.input.MarkPos(pos))

	os.Exit(util.EXIT_FAILURE_PARSE)
}

func (v *parser) pushRule(name string) {
	v.ruleStack = append(v.ruleStack, name)
}

func (v *parser) popRule() {
	v.ruleStack = v.ruleStack[:len(v.ruleStack)-1]
}

func (v *parser) dumpRules() {
	log.Debugln("parser", strings.Join(v.ruleStack, " / "))
}

func (v *parser) peek(ahead int) *lexer.Token {
	if ahead < 0 {
		panic(fmt.Sprintf("Tried to peek a negative number: %d", ahead))
	}

	if v.currentToken+ahead >= len(v.input.Tokens) {
		return nil
	}

	return v.input.Tokens[v.currentToken+ahead]
}

func (v *parser) consumeToken() *lexer.Token {
	ret := v.peek(0)
	v.currentToken++
	return ret
}

func (v *parser) consumeTokens(num int) {
	for i := 0; i < num; i++ {
		v.consumeToken()
	}
}

func (v *parser) tokenMatches(ahead int, t lexer.TokenType, contents string) bool {
	tok := v.peek(ahead)
	return tok != nil && tok.Type == t && (contents == "" || (tok.Contents == contents))
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

func (v *parser) nextIs(typ lexer.TokenType) bool {
	return v.peek(0).Type == typ
}

func (v *parser) expect(typ lexer.TokenType, val string) *lexer.Token {
	if !v.tokenMatches(0, typ, val) {
		tok := v.peek(0)
		if tok == nil {
			if val != "" {
				v.err("Expected `%s` (%s), got EOF", val, typ)
			} else {
				v.err("Expected %s, got EOF", typ)
			}
		} else {
			if val != "" {
				v.errToken("Expected `%s` (%s), got `%s` (%s)", val, typ, tok.Contents, tok.Type)
			} else {
				v.errToken("Expected %s, got %s (`%s`)", typ, tok.Type, tok.Contents)
			}
		}

	}
	return v.consumeToken()
}

func (v *parser) parse() {
	for v.peek(0) != nil {
		if n := v.parseDecl(true); n != nil {
			v.tree.AddNode(n)
		} else if n := v.parseToplevelDirective(); n != nil {
			v.tree.AddNode(n)
		} else {
			v.err("Unexpected token at toplevel: `%s` (%s)", v.peek(0).Contents, v.peek(0).Type)
		}
	}
}

func (v *parser) parseToplevelDirective() ParseNode {
	defer un(trace(v, "toplevel-directive"))

	if !v.tokensMatch(lexer.TOKEN_OPERATOR, "#", lexer.TOKEN_IDENTIFIER, "") {
		return nil
	}
	start := v.expect(lexer.TOKEN_OPERATOR, "#")

	directive := v.expect(lexer.TOKEN_IDENTIFIER, "")
	switch directive.Contents {
	case "link":
		library := v.expect(lexer.TOKEN_STRING, "")
		res := &LinkDirectiveNode{Library: NewLocatedString(library)}
		res.SetWhere(lexer.NewSpanFromTokens(start, library))
		return res

	case "use":
		module := v.parseName()
		if module == nil {
			v.errPosSpecific(directive.Where.End(), "Expected name after use directive")
		}

		v.deps = append(v.deps, module)

		res := &UseDirectiveNode{Module: module}
		res.SetWhere(lexer.NewSpan(start.Where.Start(), module.Where().End()))
		return res

	default:
		v.errTokenSpecific(directive, "No such directive `%s`", directive.Contents)
		return nil
	}
}

func (v *parser) parseNode() (ParseNode, bool) {
	defer un(trace(v, "node"))

	var ret ParseNode

	is_cond := false

	if decl := v.parseDecl(false); decl != nil {
		ret = decl
	} else if cond := v.parseConditionalStat(); cond != nil {
		ret = cond
		is_cond = true
	} else if blockStat := v.parseBlockStat(); blockStat != nil {
		ret = blockStat
		is_cond = true
	} else if stat := v.parseStat(); stat != nil {
		ret = stat
	}

	return ret, is_cond
}

func (v *parser) parseDocComments() []*DocComment {
	defer un(trace(v, "doccomments"))

	var dcs []*DocComment

	for v.nextIs(lexer.TOKEN_DOCCOMMENT) {
		tok := v.consumeToken()

		var contents string
		if strings.HasPrefix(tok.Contents, "/**") {
			contents = tok.Contents[3 : len(tok.Contents)-2]
		} else if strings.HasPrefix(tok.Contents, "///") {
			contents = tok.Contents[3:]
		} else {
			panic(fmt.Sprintf("How did this doccomment get through the lexer??\n`%s`", tok.Contents))
		}

		dcs = append(dcs, &DocComment{Where: tok.Where, Contents: contents})
	}

	return dcs
}

func (v *parser) parseAttributes() AttrGroup {
	defer un(trace(v, "attributes"))

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		return nil
	}
	attrs := make(AttrGroup)

	for v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		v.consumeToken()
		for {
			attr := &Attr{}

			keyToken := v.expect(lexer.TOKEN_IDENTIFIER, "")
			attr.setPos(keyToken.Where.Start())
			attr.Key = keyToken.Contents

			if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
				v.consumeToken()
				attr.Value = v.expect(lexer.TOKEN_STRING, "").Contents
			}

			if attrs.Set(attr.Key, attr) {
				// TODO: I feel kinda dirty having this here
				v.err("Duplicate attribute `%s`", attr.Key)
			}

			if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
				break
			}
			v.consumeToken()
		}

		v.expect(lexer.TOKEN_SEPARATOR, "]")
	}

	return attrs
}

func (v *parser) parseName() *NameNode {
	defer un(trace(v, "name"))

	if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
		return nil
	}

	var parts []LocatedString
	for {
		part := v.expect(lexer.TOKEN_IDENTIFIER, "")
		parts = append(parts, NewLocatedString(part))

		if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "::") {
			break
		}
		v.consumeToken()
	}

	name, parts := parts[len(parts)-1], parts[:len(parts)-1]
	res := &NameNode{Modules: parts, Name: name}
	if len(parts) > 0 {
		res.SetWhere(lexer.NewSpan(parts[0].Where.Start(), name.Where.End()))
	} else {
		res.SetWhere(name.Where)
	}
	return res
}

func (v *parser) parseDecl(isTopLevel bool) ParseNode {
	defer un(trace(v, "decl"))

	var res ParseNode
	docComments := v.parseDocComments()
	attrs := v.parseAttributes()

	var pub bool
	if isTopLevel {
		if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_PUB) {
			pub = true
			v.consumeToken()
		}
	}

	if typeDecl := v.parseTypeDecl(isTopLevel); typeDecl != nil {
		res = typeDecl
	} else if funcDecl := v.parseFuncDecl(isTopLevel); funcDecl != nil {
		res = funcDecl
	} else if varDecl := v.parseVarDecl(isTopLevel); varDecl != nil {
		res = varDecl
	} else if varTupleDecl := v.parseDestructVarDecl(isTopLevel); varTupleDecl != nil {
		res = varTupleDecl
	} else {
		return nil
	}

	res.(DeclNode).SetPublic(pub)

	if len(docComments) != 0 {
		res.SetDocComments(docComments)
	}

	if attrs != nil {
		res.SetAttrs(attrs)
	}

	return res
}

func (v *parser) parseFuncDecl(isTopLevel bool) *FunctionDeclNode {
	fn := v.parseFunc(false, isTopLevel)
	if fn == nil {
		return nil
	}

	res := &FunctionDeclNode{Function: fn}
	res.SetWhere(fn.Where())
	return res
}

func (v *parser) parseLambdaExpr() *LambdaExprNode {
	fn := v.parseFunc(true, false)
	if fn == nil {
		return nil
	}

	res := &LambdaExprNode{Function: fn}
	res.SetWhere(fn.Where())
	return res
}

// If lambda is true, we're parsing an expression.
// If lambda is false, we're parsing a proper function declaration.
func (v *parser) parseFunc(lambda bool, topLevelNode bool) *FunctionNode {
	defer un(trace(v, "func"))

	funcHeader := v.parseFuncHeader(lambda)
	if funcHeader == nil {
		return nil
	}

	var body *BlockNode
	var stat, expr ParseNode
	var end lexer.Position

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		terminator := v.consumeToken()
		end = terminator.Where.End()
	} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=>") {
		v.consumeToken()

		isCond := false
		if stat = v.parseStat(); stat != nil {
			end = stat.Where().End()
		} else if stat = v.parseConditionalStat(); stat != nil {
			end = stat.Where().End()
			isCond = true
		} else if expr = v.parseExpr(); expr != nil {
			end = expr.Where().End()
		} else {
			v.err("Expected valid statement or expression after => operator in function declaration")
		}

		if topLevelNode && !isCond {
			v.expect(lexer.TOKEN_SEPARATOR, ";")
		}
	} else {
		body = v.parseBlock()
		if body == nil {
			v.err("Expected block after function declaration, or terminating semi-colon")
		}
		end = body.Where().End()
	}

	res := &FunctionNode{Header: funcHeader, Body: body, Stat: stat, Expr: expr}
	res.SetWhere(lexer.NewSpan(funcHeader.Where().Start(), end))
	return res
}

// If lambda is true, don't parse name and set Anonymous to true.
func (v *parser) parseFuncHeader(lambda bool) *FunctionHeaderNode {
	defer un(trace(v, "funcheader"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
		return nil
	}
	startToken := v.consumeToken()

	res := &FunctionHeaderNode{}

	if !lambda {
		// parses the function receiver if there is one.
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
			// we have a method receiver
			v.consumeToken()

			if v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_SEPARATOR, ")") {
				res.StaticReceiverType = v.parseNamedType()
				if res.StaticReceiverType == nil {
					v.errToken("Expected type name in method receiver, found `%s`", v.peek(0).Contents)
				}
			} else {
				res.Receiver = v.parseVarDeclBody(true)
				if res.Receiver == nil {
					v.errToken("Expected variable declaration in method receiver, found `%s`", v.peek(0).Contents)
				}
			}

			v.expect(lexer.TOKEN_SEPARATOR, ")")
		}

		// parses the function identifier/name
		name := v.expect(lexer.TOKEN_IDENTIFIER, "")
		res.Name = NewLocatedString(name)
	}

	genericSigil := v.parseGenericSigil()
	v.expect(lexer.TOKEN_SEPARATOR, "(")

	var args []*VarDeclNode
	variadic := false
	// parse the function arguments
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
			break
		}

		// parse our variadic sigil (three magical dots)
		if v.tokensMatch(lexer.TOKEN_SEPARATOR, ".", lexer.TOKEN_SEPARATOR, ".", lexer.TOKEN_SEPARATOR, ".") {
			v.consumeTokens(3)
			if !variadic {
				variadic = true
			} else {
				v.err("Duplicate `...` in function arguments")
			}
		} else {
			arg := v.parseVarDeclBody(false)
			if arg == nil {
				v.err("Expected valid variable declaration in function args")
			}
			args = append(args, arg)
		}

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}

	maybeEndToken := v.expect(lexer.TOKEN_SEPARATOR, ")")

	var returnType *TypeReferenceNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "->") {
		v.consumeToken()

		returnType = v.parseTypeReference(true, false, true)
		if returnType == nil {
			v.err("Expected valid type after `->` in function header")
		}
	}

	res.Arguments = args
	res.Variadic = variadic
	res.GenericSigil = genericSigil
	res.Anonymous = lambda

	if returnType != nil {
		res.ReturnType = returnType
		res.SetWhere(lexer.NewSpan(startToken.Where.Start(), returnType.Where().End()))
	} else {
		res.SetWhere(lexer.NewSpanFromTokens(startToken, maybeEndToken))
	}

	return res
}

func (v *parser) parseTypeDecl(isTopLevel bool) *TypeDeclNode {
	defer un(trace(v, "typdecl"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "type") {
		return nil
	}

	startToken := v.consumeToken()

	name := v.expect(lexer.TOKEN_IDENTIFIER, "")
	if isReservedKeyword(name.Contents) {
		v.err("Cannot use reserved keyword `%s` as type name", name.Contents)
	}

	typ := v.parseType(true, false, true)

	if isTopLevel {
		v.expect(lexer.TOKEN_SEPARATOR, ";")
	}

	res := &TypeDeclNode{
		Name: NewLocatedString(name),
		Type: typ,
	}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), typ.Where().End()))

	return res
}

func (v *parser) parseGenericSigil() *GenericSigilNode {
	defer un(trace(v, "genericsigil"))

	if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "<") {
		return nil
	}
	startToken := v.consumeToken()

	var parameters []*TypeParameterNode
	for {
		parameter := v.parseTypeParameter()
		if parameter == nil {
			v.err("Expected valid type parameter in generic sigil")
		}
		parameters = append(parameters, parameter)

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}
	endToken := v.expect(lexer.TOKEN_OPERATOR, ">")

	res := &GenericSigilNode{GenericParameters: parameters}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseTypeParameter() *TypeParameterNode {
	name := v.expect(lexer.TOKEN_IDENTIFIER, "")

	var constraints []ParseNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, ":") {
		v.consumeToken()
		for {
			constraint := v.parseType(true, false, false)
			if constraint == nil {
				v.err("Expected valid name in type restriction")
			}
			constraints = append(constraints, constraint)

			if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "&") {
				break
			}
			v.consumeToken()
		}
	}

	res := &TypeParameterNode{Name: NewLocatedString(name), Constraints: constraints}
	if idx := len(constraints) - 1; idx >= 0 {
		res.SetWhere(lexer.NewSpan(name.Where.Start(), constraints[idx].Where().End()))
	} else {
		res.SetWhere(lexer.NewSpanFromTokens(name, name))
	}
	return res
}

func (v *parser) parseEnumEntry() *EnumEntryNode {
	defer un(trace(v, "enumentry"))

	if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
		return nil
	}
	name := v.consumeToken()

	if isReservedKeyword(name.Contents) {
		v.err("Cannot use reserved keyword `%s` as name for enum entry", name.Contents)
	}

	var value *NumberLitNode
	var structBody *StructTypeNode
	var tupleBody *TupleTypeNode
	var lastPos lexer.Position
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.consumeToken()

		value = v.parseNumberLit()
		if value == nil || value.IsFloat {
			v.err("Expected valid integer after `=` in enum entry")
		}
		lastPos = value.Where().End()
	} else if tupleBody = v.parseTupleType(true); tupleBody != nil {
		lastPos = tupleBody.Where().End()
	} else if structBody = v.parseStructType(false); structBody != nil {
		lastPos = structBody.Where().End()
	}

	res := &EnumEntryNode{Name: NewLocatedString(name), Value: value, TupleBody: tupleBody, StructBody: structBody}
	if value != nil || structBody != nil || tupleBody != nil {
		res.SetWhere(lexer.NewSpan(name.Where.Start(), lastPos))
	} else {
		res.SetWhere(name.Where)
	}
	return res
}

func (v *parser) parseVarDecl(isTopLevel bool) *VarDeclNode {
	defer un(trace(v, "vardecl"))

	body := v.parseVarDeclBody(false)
	if body == nil {
		return nil
	}
	if isTopLevel {
		v.expect(lexer.TOKEN_SEPARATOR, ";")
	}

	return body
}

func (v *parser) parseVarDeclBody(isReceiver bool) *VarDeclNode {
	defer un(trace(v, "vardeclbody"))

	startPos := v.currentToken

	var mutable *lexer.Token
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MUT) {
		mutable = v.consumeToken()
	}

	if !v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_OPERATOR, ":") {
		v.currentToken = startPos
		return nil
	}

	name := v.consumeToken()

	// consume ':'
	v.consumeToken()

	var varType *TypeReferenceNode
	var sigil *GenericSigilNode
	if isReceiver {
		typ := v.parseType(true, false, true)
		if typ != nil {
			varType = &TypeReferenceNode{Type: typ}
		}

		sigil = v.parseGenericSigil()
	} else {
		varType = v.parseTypeReference(true, false, true)
	}
	if varType == nil && !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.err("Expected valid type in variable declaration")
	}

	var value ParseNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.consumeToken()

		value = v.parseCompositeLiteral()
		if value == nil {
			value = v.parseExpr()
		}

		if value == nil {
			v.err("Expected valid expression after `=` in variable declaration")
		}
	}

	res := &VarDeclNode{
		Name:                 NewLocatedString(name),
		Type:                 varType,
		IsReceiver:           isReceiver,
		ReceiverGenericSigil: sigil,
	}
	start := name.Where.Start()
	if mutable != nil {
		res.Mutable = NewLocatedString(mutable)
		start = mutable.Where.Start()
	}

	var end lexer.Position
	if value != nil {
		res.Value = value
		end = value.Where().End()
	} else {
		end = varType.Where().End()
	}

	res.SetWhere(lexer.NewSpan(start, end))
	return res
}

func (v *parser) parseDestructVarDecl(isTopLevel bool) *DestructVarDeclNode {
	defer un(trace(v, "vartupledecl"))

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}
	start := v.consumeToken()

	var names []LocatedString
	var mutable []bool
	for {
		isMutable := false
		if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MUT) {
			isMutable = true
			v.consumeToken()
		}

		if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
			v.errPos("Expected identifier in tuple destructuring variable declaration, got %s", v.peek(0).Type)
		}
		name := v.consumeToken()

		names = append(names, NewLocatedString(name))
		mutable = append(mutable, isMutable)

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}

	v.expect(lexer.TOKEN_SEPARATOR, ")")
	v.expect(lexer.TOKEN_OPERATOR, ":")
	v.expect(lexer.TOKEN_OPERATOR, "=")

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expression after tuple destructuring variable declaration")
	}

	res := &DestructVarDeclNode{
		Names:   names,
		Mutable: mutable,
		Value:   value,
	}
	res.SetWhere(lexer.NewSpan(start.Where.Start(), value.Where().End()))
	return res
}

func (v *parser) parseConditionalStat() ParseNode {
	defer un(trace(v, "conditionalstat"))

	var res ParseNode

	// conditional structures
	if ifStat := v.parseIfStat(); ifStat != nil {
		res = ifStat
	} else if matchStat := v.parseMatchStat(); matchStat != nil {
		res = matchStat
	} else if loopStat := v.parseLoopStat(); loopStat != nil {
		res = loopStat
	}

	return res
}

func (v *parser) parseStat() ParseNode {
	defer un(trace(v, "stat"))

	var res ParseNode

	if breakStat := v.parseBreakStat(); breakStat != nil {
		res = breakStat
	} else if nextStat := v.parseNextStat(); nextStat != nil {
		res = nextStat
	} else if deferStat := v.parseDeferStat(); deferStat != nil {
		res = deferStat
	} else if returnStat := v.parseReturnStat(); returnStat != nil {
		res = returnStat
	} else if callStat := v.parseCallStat(); callStat != nil {
		res = callStat
	} else if assignStat := v.parseAssignStat(); assignStat != nil {
		res = assignStat
	} else if binopAssignStat := v.parseBinopAssignStat(); binopAssignStat != nil {
		res = binopAssignStat
	}

	return res
}

func (v *parser) parseDeferStat() *DeferStatNode {
	defer un(trace(v, "deferstat"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_DEFER) {
		return nil
	}
	startToken := v.consumeToken()

	call, ok := v.parseExpr().(*CallExprNode)
	if !ok {
		v.err("Expected valid call expression in defer statement")
	}

	res := &DeferStatNode{Call: call}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), call.Where().End()))
	return res
}

func (v *parser) parseIfStat() *IfStatNode {
	defer un(trace(v, "ifstat"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_IF) {
		return nil
	}
	startToken := v.consumeToken()

	var parts []*ConditionBodyNode
	var lastPart *ConditionBodyNode
	for {
		condition := v.parseExpr()
		if condition == nil {
			v.err("Expected valid expression as condition in if statement")
		}

		body := v.parseBlock()
		if body == nil {
			v.err("Expected valid block after condition in if statement")
		}

		lastPart = &ConditionBodyNode{Condition: condition, Body: body}
		lastPart.SetWhere(lexer.NewSpan(condition.Where().Start(), body.Where().End()))
		parts = append(parts, lastPart)

		if !v.tokensMatch(lexer.TOKEN_IDENTIFIER, KEYWORD_ELSE, lexer.TOKEN_IDENTIFIER, KEYWORD_IF) {
			break
		}
		v.consumeTokens(2)
	}

	var elseBody *BlockNode
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_ELSE) {
		v.consumeToken()

		elseBody = v.parseBlock()
		if elseBody == nil {
			v.err("Expected valid block after `else` keyword in if statement")
		}
	}

	res := &IfStatNode{Parts: parts, ElseBody: elseBody}
	if elseBody != nil {
		res.SetWhere(lexer.NewSpan(startToken.Where.Start(), elseBody.Where().End()))
	} else {
		res.SetWhere(lexer.NewSpan(startToken.Where.Start(), lastPart.Where().End()))
	}
	return res
}

func (v *parser) parseMatchStat() *MatchStatNode {
	defer un(trace(v, "matchstat"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MATCH) {
		return nil
	}
	startToken := v.consumeToken()

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expresson as value in match statement")
	}

	v.expect(lexer.TOKEN_SEPARATOR, "{")

	var cases []*MatchCaseNode
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			break
		}

		var pattern ParseNode
		if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "_") {
			patTok := v.consumeToken()

			pattern = &DefaultPatternNode{}
			pattern.SetWhere(patTok.Where)
		} else {
			pattern = v.parseExpr()
		}

		if pattern == nil {
			v.err("Expected valid expression as pattern in match statement")
		}

		v.expect(lexer.TOKEN_OPERATOR, "=>")

		var body ParseNode
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
			body = v.parseBlock()
		} else {
			body = v.parseStat()
		}
		if body == nil {
			v.err("Expected valid arm statement in match clause")
		}

		v.expect(lexer.TOKEN_SEPARATOR, ",")

		caseNode := &MatchCaseNode{Pattern: pattern, Body: body}
		caseNode.SetWhere(lexer.NewSpan(pattern.Where().Start(), body.Where().End()))
		cases = append(cases, caseNode)
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "}")

	res := &MatchStatNode{Value: value, Cases: cases}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseLoopStat() *LoopStatNode {
	defer un(trace(v, "loopstat"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FOR) {
		return nil
	}
	startToken := v.consumeToken()

	condition := v.parseExpr()

	body := v.parseBlock()
	if body == nil {
		v.err("Expected valid block as body of loop statement ", v.peek(0))
	}

	res := &LoopStatNode{Condition: condition, Body: body}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), body.Where().End()))
	return res
}

func (v *parser) parseReturnStat() *ReturnStatNode {
	defer un(trace(v, "returnstat"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_RETURN) {
		return nil
	}
	startToken := v.consumeToken()

	value := v.parseExpr()

	var end lexer.Position
	if value != nil {
		end = value.Where().End()
	} else {
		end = startToken.Where.End()
	}

	res := &ReturnStatNode{Value: value}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), end))
	return res
}

func (v *parser) parseBreakStat() *BreakStatNode {
	defer un(trace(v, "breakstat"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_BREAK) {
		return nil
	}
	startToken := v.consumeToken()

	res := &BreakStatNode{}
	res.SetWhere(startToken.Where)
	return res
}

func (v *parser) parseNextStat() *NextStatNode {
	defer un(trace(v, "nextstat"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_NEXT) {
		return nil
	}
	startToken := v.consumeToken()

	res := &NextStatNode{}
	res.SetWhere(startToken.Where)
	return res
}

func (v *parser) parseBlockStat() *BlockStatNode {
	defer un(trace(v, "blockstat"))

	startPos := v.currentToken
	var doToken *lexer.Token
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_DO) {
		doToken = v.consumeToken()
	}

	body := v.parseBlock()
	if body == nil {
		v.currentToken = startPos
		return nil
	}

	res := &BlockStatNode{Body: body}
	if doToken != nil {
		body.NonScoping = true
		res.SetWhere(lexer.NewSpan(doToken.Where.Start(), body.Where().End()))
	} else {
		res.SetWhere(body.Where())
	}
	return res
}

func (v *parser) parseBlock() *BlockNode {
	defer un(trace(v, "block"))

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		return nil
	}
	startToken := v.consumeToken()

	var nodes []ParseNode
	for {
		node, is_cond := v.parseNode()
		if node == nil {
			break
		}
		if !is_cond {
			v.expect(lexer.TOKEN_SEPARATOR, ";")
		}
		nodes = append(nodes, node)
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "}")

	res := &BlockNode{Nodes: nodes}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseCallStat() *CallStatNode {
	defer un(trace(v, "callstat"))

	startPos := v.currentToken

	callExpr, ok := v.parseExpr().(*CallExprNode)
	if !ok {
		v.currentToken = startPos
		return nil
	}

	res := &CallStatNode{Call: callExpr}
	res.SetWhere(lexer.NewSpan(callExpr.Where().Start(), callExpr.Where().End()))
	return res
}

func (v *parser) parseAssignStat() ParseNode {
	defer un(trace(v, "assignstat"))

	startPos := v.currentToken

	accessExpr := v.parseExpr()
	if accessExpr == nil || !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.currentToken = startPos
		return nil
	}

	// consume '='
	v.consumeToken()

	var value ParseNode
	value = v.parseCompositeLiteral()
	if value == nil {
		value = v.parseExpr()
	}

	// not a composite or expr = error
	if value == nil {
		v.err("Expected valid expression in assignment statement")
	}

	res := &AssignStatNode{Target: accessExpr, Value: value}
	res.SetWhere(lexer.NewSpan(accessExpr.Where().Start(), value.Where().End()))
	return res
}

func (v *parser) parseBinopAssignStat() ParseNode {
	defer un(trace(v, "binopassignstat"))

	startPos := v.currentToken

	accessExpr := v.parseExpr()
	if accessExpr == nil || !v.tokensMatch(lexer.TOKEN_OPERATOR, "", lexer.TOKEN_OPERATOR, "=") {
		v.currentToken = startPos
		return nil
	}

	typ := stringToBinOpType(v.peek(0).Contents)
	if typ == BINOP_ERR || typ.Category() == OP_COMPARISON {
		v.err("Invalid binary operator `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	// consume '='
	v.consumeToken()

	var value ParseNode
	value = v.parseCompositeLiteral()
	if value == nil {
		value = v.parseExpr()
	}

	// no composite and no expr = err
	if value == nil {
		v.err("Expected valid expression in assignment statement")
	}

	res := &BinopAssignStatNode{Target: accessExpr, Operator: typ, Value: value}
	res.SetWhere(lexer.NewSpan(accessExpr.Where().Start(), value.Where().End()))
	return res
}

func (v *parser) parseTypeReference(doNamed bool, onlyComposites bool, mustParse bool) *TypeReferenceNode {
	typ := v.parseType(doNamed, onlyComposites, mustParse)
	if typ == nil {
		return nil
	}

	var gargs []*TypeReferenceNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "<") {
		v.consumeToken()

		for {
			typ := v.parseTypeReference(true, false, true)
			if typ == nil {
				v.err("Expected valid type as type parameter")
			}
			gargs = append(gargs, typ)

			if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
				break
			}
			v.consumeToken()
		}

		v.expect(lexer.TOKEN_OPERATOR, ">")
	}

	res := &TypeReferenceNode{
		Type:             typ,
		GenericArguments: gargs,
	}

	res.SetWhere(lexer.NewSpan(typ.Where().Start(), typ.Where().End()))
	if len(gargs) > 0 {
		last := gargs[len(gargs)-1]
		res.SetWhere(lexer.NewSpan(typ.Where().Start(), last.Where().End()))
	}

	return res
}

// NOTE onlyComposites does not affect doRefs.
func (v *parser) parseType(doNamed bool, onlyComposites bool, mustParse bool) ParseNode {
	defer un(trace(v, "type"))

	var res ParseNode
	var attrs AttrGroup

	defer func() {
		if res != nil && attrs != nil {
			res.SetAttrs(attrs)
			// TODO: Update start position of result
		}
	}()

	// If the next token is a [ and identifier it must be a group of attributes
	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") && v.tokenMatches(1, lexer.TOKEN_IDENTIFIER, "") {
		attrs = v.parseAttributes()
	}

	if !onlyComposites {
		if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
			res = v.parseFunctionType()
		} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
			res = v.parsePointerType()
		} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "&") {
			res = v.parseReferenceType()
		} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
			res = v.parseTupleType(mustParse)
		} else if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_INTERFACE) {
			res = v.parseInterfaceType()
		}
	}

	if res != nil {
		return res
	}

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		res = v.parseArrayType()
	} else if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_STRUCT) {
		res = v.parseStructType(true)
	} else if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_ENUM) {
		res = v.parseEnumType()
	} else if doNamed && v.nextIs(lexer.TOKEN_IDENTIFIER) {
		res = v.parseNamedType()
	}

	return res
}

func (v *parser) parseEnumType() *EnumTypeNode {
	defer un(trace(v, "enumtype"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_ENUM) {
		return nil
	}
	startToken := v.consumeToken()

	genericsigil := v.parseGenericSigil()

	v.expect(lexer.TOKEN_SEPARATOR, "{")

	var members []*EnumEntryNode
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			break
		}

		member := v.parseEnumEntry()
		if member == nil {
			v.err("Expected valid enum entry in enum")
		}
		members = append(members, member)

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "}")

	res := &EnumTypeNode{
		Members:      members,
		GenericSigil: genericsigil,
	}

	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseInterfaceType() *InterfaceTypeNode {
	defer un(trace(v, "interfacetype"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_INTERFACE) {
		return nil
	}

	startToken := v.consumeToken()
	v.expect(lexer.TOKEN_SEPARATOR, "{")

	// when we hit a };
	// this means our interface is done...
	var functions []*FunctionHeaderNode
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") && v.tokenMatches(1, lexer.TOKEN_SEPARATOR, ";") {
			break
		}

		function := v.parseFuncHeader(false)
		if function != nil {
			// TODO trailing comma
			v.expect(lexer.TOKEN_SEPARATOR, ",")
			functions = append(functions, function)
		} else {
			v.err("Failed to parse function in interface")
		}
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "}")

	res := &InterfaceTypeNode{
		Functions: functions,
	}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseStructType(requireKeyword bool) *StructTypeNode {
	defer un(trace(v, "structtype"))

	var startToken *lexer.Token

	var sigil *GenericSigilNode

	if requireKeyword {
		if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_STRUCT) {
			return nil
		}
		startToken = v.consumeToken()

		sigil = v.parseGenericSigil()
		v.expect(lexer.TOKEN_SEPARATOR, "{")
	} else {
		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
			return nil
		}
		startToken = v.consumeToken()
	}

	var members []*StructMemberNode
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			break
		}

		member := v.parseStructMember()
		if member == nil {
			v.err("Expected valid member declaration in struct")
		}
		members = append(members, member)

		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			v.consumeToken()
		}
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "}")

	res := &StructTypeNode{Members: members, GenericSigil: sigil}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseStructMember() *StructMemberNode {
	if !v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_OPERATOR, ":") {
		return nil
	}

	name := v.consumeToken()

	// consume ':'
	v.consumeToken()

	memType := v.parseTypeReference(true, false, true)
	if memType == nil {
		v.err("Expected valid type in struct member")
	}

	res := &StructMemberNode{Name: NewLocatedString(name), Type: memType}
	res.SetWhere(lexer.NewSpan(name.Where.Start(), memType.Where().End()))
	return res
}

func (v *parser) parseFunctionType() *FunctionTypeNode {
	defer un(trace(v, "functiontype"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
		return nil
	}
	startToken := v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.err("Expected `(` after `func` keyword")
	}
	lastParens := v.consumeToken()

	var pars []*TypeReferenceNode
	variadic := false

	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
			lastParens = v.consumeToken()
			break
		}

		if variadic {
			v.err("Variadic signifier must be the last argument in a variadic function")
		}

		if v.tokensMatch(lexer.TOKEN_SEPARATOR, ".", lexer.TOKEN_SEPARATOR, ".", lexer.TOKEN_SEPARATOR, ".") {
			v.consumeTokens(3)
			if !variadic {
				variadic = true
			} else {
				v.err("Duplicate variadic signifier `...` in function header")
			}
		} else {
			par := v.parseTypeReference(true, false, true)
			if par == nil {
				v.err("Expected type in function argument, found `%s`", v.peek(0).Contents)
			}

			pars = append(pars, par)
		}

		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			v.consumeToken()
			continue
		} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
			lastParens = v.consumeToken()
			break
		} else {
			v.err("Unexpected `%s`", v.peek(0).Contents)
		}
	}

	var returnType *TypeReferenceNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "->") {
		v.consumeToken()
		returnType = v.parseTypeReference(true, false, true)
		if returnType == nil {
			v.err("Expected return type in function header, found `%s`", v.peek(0).Contents)
		}
	}

	var end lexer.Position
	if returnType != nil {
		end = returnType.Where().End()
	} else {
		end = lastParens.Where.End()
	}

	res := &FunctionTypeNode{
		ParameterTypes: pars,
		ReturnType:     returnType,
	}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), end))

	return res
}

func (v *parser) parsePointerType() *PointerTypeNode {
	defer un(trace(v, "pointertype"))

	mutable, target, where := v.parsePointerlikeType("^")
	if target == nil {
		return nil
	}

	res := &PointerTypeNode{Mutable: mutable, TargetType: target}
	res.SetWhere(where)
	return res
}

func (v *parser) parseReferenceType() *ReferenceTypeNode {
	defer un(trace(v, "referencetype"))

	mutable, target, where := v.parsePointerlikeType("&")
	if target == nil {
		return nil
	}

	res := &ReferenceTypeNode{Mutable: mutable, TargetType: target}
	res.SetWhere(where)
	return res
}

func (v *parser) parsePointerlikeType(symbol string) (mutable bool, target *TypeReferenceNode, where lexer.Span) {
	defer un(trace(v, "pointerliketype"))

	if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, symbol) {
		return false, nil, lexer.Span{}
	}
	startToken := v.consumeToken()

	mutable = false
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MUT) {
		v.consumeToken()
		mutable = true
	}

	target = v.parseTypeReference(true, false, true)
	if target == nil {
		v.err("Expected valid type after '%s' in pointer/reference type", symbol)
	}

	where = lexer.NewSpan(startToken.Where.Start(), target.Where().End())
	return
}

func (v *parser) parseTupleType(mustParse bool) *TupleTypeNode {
	defer un(trace(v, "tupletype"))

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}
	startToken := v.consumeToken()

	var members []*TypeReferenceNode
	for {
		memberType := v.parseTypeReference(true, false, mustParse)
		if memberType == nil {
			if mustParse {
				v.err("Expected valid type in tuple type")
			} else {
				return nil
			}

		}
		members = append(members, memberType)

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ")")

	res := &TupleTypeNode{MemberTypes: members}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseArrayType() *ArrayTypeNode {
	defer un(trace(v, "arraytype"))

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		return nil
	}
	startToken := v.consumeToken()

	length := v.parseNumberLit()
	if length != nil && length.IsFloat {
		v.err("Expected integer length for array type")
	}

	v.expect(lexer.TOKEN_SEPARATOR, "]")

	memberType := v.parseTypeReference(true, false, true)
	if memberType == nil {
		v.err("Expected valid type in array type")
	}

	res := &ArrayTypeNode{MemberType: memberType}
	if length != nil {
		// TODO: Defend against overflow
		res.Length = int(length.IntValue.Int64())
		res.IsFixedLength = true
	}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), memberType.Where().End()))
	return res
}

func (v *parser) parseNamedType() *NamedTypeNode {
	defer un(trace(v, "typereference"))

	name := v.parseName()
	if name == nil {
		return nil
	}

	res := &NamedTypeNode{Name: name}
	res.SetWhere(name.Where())
	return res
}

func (v *parser) parseExpr() ParseNode {
	defer un(trace(v, "expr"))

	pri := v.parsePostfixExpr()
	if pri == nil {
		return nil
	}

	if bin := v.parseBinaryOperator(0, pri); bin != nil {
		return bin
	}

	return pri
}

func (v *parser) parseBinaryOperator(upperPrecedence int, lhand ParseNode) ParseNode {
	defer un(trace(v, "binop"))

	// TODO: I have a suspicion this might break with some combinations of operators
	startPos := v.currentToken

	tok := v.peek(0)
	if tok.Type != lexer.TOKEN_OPERATOR || v.peek(1).Contents == ";" {
		return nil
	}

	for {
		tokPrecedence := v.getPrecedence(stringToBinOpType(v.peek(0).Contents))
		if tokPrecedence < upperPrecedence {
			return lhand
		}

		typ := stringToBinOpType(v.peek(0).Contents)
		if typ == BINOP_ERR {
			v.err("Invalid binary operator `%s`", v.peek(0).Contents)
		}
		v.consumeToken()

		rhand := v.parsePostfixExpr()
		if rhand == nil {
			v.currentToken = startPos
			return nil
		}

		nextPrecedence := v.getPrecedence(stringToBinOpType(v.peek(0).Contents))
		if tokPrecedence < nextPrecedence {
			rhand = v.parseBinaryOperator(tokPrecedence+1, rhand)
			if rhand == nil {
				v.currentToken = startPos
				return nil
			}
		}

		temp := &BinaryExprNode{
			Lhand:    lhand,
			Rhand:    rhand,
			Operator: typ,
		}
		temp.SetWhere(lexer.NewSpan(lhand.Where().Start(), rhand.Where().Start()))
		lhand = temp
	}
}

func (v *parser) parsePostfixExpr() ParseNode {
	defer un(trace(v, "postfixexpr"))

	expr := v.parsePrimaryExpr()
	if expr == nil {
		return nil
	}

	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ".") {
			// struct access
			v.consumeToken()
			defer un(trace(v, "structaccess"))

			member := v.expect(lexer.TOKEN_IDENTIFIER, "")

			res := &StructAccessNode{Struct: expr, Member: NewLocatedString(member)}
			res.SetWhere(lexer.NewSpan(expr.Where().Start(), member.Where.End()))
			expr = res
		} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
			// array index
			v.consumeToken()
			defer un(trace(v, "arrayindex"))

			index := v.parseExpr()
			if index == nil {
				v.err("Expected valid expression as array index")
			}

			endToken := v.expect(lexer.TOKEN_SEPARATOR, "]")

			res := &ArrayAccessNode{Array: expr, Index: index}
			res.SetWhere(lexer.NewSpan(expr.Where().Start(), endToken.Where.End()))
			expr = res
		} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
			// call expr
			v.consumeToken()
			defer un(trace(v, "callexpr"))

			var args []ParseNode
			for {
				if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
					break
				}

				arg := v.parseExpr()
				if arg == nil {
					v.err("Expected valid expression as call argument")
				}
				args = append(args, arg)

				if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
					break
				}
				v.consumeToken()
			}

			endToken := v.expect(lexer.TOKEN_SEPARATOR, ")")

			res := &CallExprNode{Function: expr, Arguments: args}
			res.SetWhere(lexer.NewSpan(expr.Where().Start(), endToken.Where.End()))
			expr = res
		} else {
			break
		}
	}

	return expr
}

func (v *parser) parsePrimaryExpr() ParseNode {
	defer un(trace(v, "primaryexpr"))

	var res ParseNode

	if sizeofExpr := v.parseSizeofExpr(); sizeofExpr != nil {
		res = sizeofExpr
	} else if arrayLenExpr := v.parseArrayLenExpr(); arrayLenExpr != nil {
		res = arrayLenExpr
	} else if addrofExpr := v.parseAddrofExpr(); addrofExpr != nil {
		res = addrofExpr
	} else if litExpr := v.parseLitExpr(); litExpr != nil {
		res = litExpr
	} else if lambdaExpr := v.parseLambdaExpr(); lambdaExpr != nil {
		res = lambdaExpr
	} else if unaryExpr := v.parseUnaryExpr(); unaryExpr != nil {
		res = unaryExpr
	} else if castExpr := v.parseCastExpr(); castExpr != nil {
		res = castExpr
	} else if name := v.parseName(); name != nil {
		startPos := v.currentToken

		var parameters []*TypeReferenceNode
		if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "<") {
			v.consumeToken()

			for {
				typ := v.parseTypeReference(true, false, false)
				if typ == nil {
					break
				}
				parameters = append(parameters, typ)

				if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
					break
				}
				v.consumeToken()
			}

			if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, ">") {
				v.currentToken = startPos
				parameters = nil
			} else {
				endToken := v.consumeToken()
				_ = endToken // TODO: Do somethign with end token?
			}
		}

		res = &VariableAccessNode{Name: name, GenericParameters: parameters}
		res.SetWhere(lexer.NewSpan(name.Where().Start(), name.Where().End()))
	}

	return res
}

func (v *parser) parseArrayLenExpr() *ArrayLenExprNode {
	defer un(trace(v, "arraylenexpr"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_LEN) {
		return nil
	}
	startToken := v.consumeToken()

	v.expect(lexer.TOKEN_SEPARATOR, "(")

	var array ParseNode
	array = v.parseCompositeLiteral()
	if array == nil {
		array = v.parseExpr()
	}
	if array == nil {
		v.err("Expected valid expression in array length expression")
	}

	v.expect(lexer.TOKEN_SEPARATOR, ")")

	end := v.peek(0)
	res := &ArrayLenExprNode{ArrayExpr: array}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), end.Where.Start()))
	return res
}

func (v *parser) parseSizeofExpr() *SizeofExprNode {
	defer un(trace(v, "sizeofexpr"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_SIZEOF) {
		return nil
	}
	startToken := v.consumeToken()

	v.expect(lexer.TOKEN_SEPARATOR, "(")

	var typ *TypeReferenceNode
	value := v.parseExpr()
	if value == nil {
		typ = v.parseTypeReference(true, false, true)
		if typ == nil {
			v.err("Expected valid expression or type in sizeof expression")
		}
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ")")

	res := &SizeofExprNode{Value: value, Type: typ}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseAddrofExpr() *AddrofExprNode {
	defer un(trace(v, "addrofexpr"))
	startPos := v.currentToken

	isReference := false
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "&") {
		isReference = true
	} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		isReference = false
	} else {
		return nil
	}
	startToken := v.consumeToken()

	mutable := false
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MUT) {
		v.consumeToken()
		mutable = true
	}

	value := v.parseExpr()
	if value == nil {
		// TODO: Restore this error once we go through with #655
		//v.err("Expected valid expression after addrof expression")
		v.currentToken = startPos
		return nil
	}

	res := &AddrofExprNode{Mutable: mutable, Value: value, IsReference: isReference}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), value.Where().End()))
	return res
}

func (v *parser) parseLitExpr() ParseNode {
	defer un(trace(v, "litexpr"))

	var res ParseNode

	if tupleLit := v.parseTupleLit(); tupleLit != nil {
		res = tupleLit
	} else if boolLit := v.parseBoolLit(); boolLit != nil {
		res = boolLit
	} else if numberLit := v.parseNumberLit(); numberLit != nil {
		res = numberLit
	} else if stringLit := v.parseStringLit(); stringLit != nil {
		res = stringLit
	} else if runeLit := v.parseRuneLit(); runeLit != nil {
		res = runeLit
	}

	return res
}

func (v *parser) parseCastExpr() *CastExprNode {
	defer un(trace(v, "castexpr"))

	startPos := v.currentToken

	typ := v.parseTypeReference(false, false, false)
	if typ == nil || !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.currentToken = startPos
		return nil
	}
	v.consumeToken()

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expression in cast expression")
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ")")

	res := &CastExprNode{Type: typ, Value: value}
	res.SetWhere(lexer.NewSpan(typ.Where().Start(), endToken.Where.End()))
	return res
}

func (v *parser) parseUnaryExpr() *UnaryExprNode {
	defer un(trace(v, "unaryexpr"))

	startPos := v.currentToken

	if !v.nextIs(lexer.TOKEN_OPERATOR) {
		return nil
	}

	op := stringToUnOpType(v.peek(0).Contents)
	if op == UNOP_ERR {
		return nil
	}
	startToken := v.consumeToken()

	value := v.parsePostfixExpr()
	if value == nil {
		// TODO: Restore this error once we go through with #655
		//v.err("Expected valid expression after unary operator")
		v.currentToken = startPos
		return nil
	}

	res := &UnaryExprNode{Value: value, Operator: op}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), value.Where().End()))
	return res
}

func (v *parser) parseCompositeLiteral() ParseNode {
	defer un(trace(v, "complit"))

	startPos := v.currentToken
	typ := v.parseTypeReference(true, true, true)

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.currentToken = startPos
		return nil
	}
	start := v.consumeToken() // eat opening bracket

	res := &CompositeLiteralNode{
		Type: typ,
	}

	var lastToken *lexer.Token

	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			lastToken = v.consumeToken()
			break
		}

		var field LocatedString

		if v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_OPERATOR, ":") {
			field = NewLocatedString(v.consumeToken())
			v.consumeToken()
		}

		val := v.parseCompositeLiteral()
		if val == nil {
			val = v.parseExpr()
		}
		if val == nil {
			v.err("Expected value in composite literal, found `%s`", v.peek(0).Contents)
		}

		res.Fields = append(res.Fields, field)
		res.Values = append(res.Values, val)

		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			v.consumeToken()
			continue
		} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			lastToken = v.consumeToken()
			break
		} else {
			v.err("Unexpected `%s`", v.peek(0).Contents)
		}
	}

	if typ != nil {
		res.SetWhere(lexer.NewSpan(typ.Where().Start(), lastToken.Where.End()))
	} else {
		res.SetWhere(lexer.NewSpanFromTokens(start, lastToken))
	}

	return res
}

func (v *parser) parseTupleLit() *TupleLiteralNode {
	defer un(trace(v, "tuplelit"))

	startPos := v.currentToken
	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}
	startToken := v.consumeToken()

	var values []ParseNode
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
			break
		}

		value := v.parseExpr()
		if value == nil {
			// TODO: Restore this error once we go through with #655
			//v.err("Expected valid expression in tuple literal")
			v.currentToken = startPos
			return nil
		}
		values = append(values, value)

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}

	endToken := v.peek(0)
	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		// TODO: Restore this error once we go through wiht #655
		// endToken := v.expect(lexer.TOKEN_SEPARATOR, ")")
		v.currentToken = startPos
		return nil
	}
	v.currentToken++

	// Dirty hack
	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ".") {
		v.currentToken = startPos
		return nil
	}

	res := &TupleLiteralNode{Values: values}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseBoolLit() *BoolLitNode {
	defer un(trace(v, "boollit"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "true") && !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "false") {
		return nil
	}
	token := v.consumeToken()

	var value bool
	if token.Contents == "true" {
		value = true
	} else {
		value = false
	}

	res := &BoolLitNode{Value: value}
	res.SetWhere(token.Where)
	return res
}

func parseInt(num string, base int) (*big.Int, bool) {
	num = strings.ToLower(strings.Replace(num, "_", "", -1))

	var splitNum []string
	if base == 10 {
		splitNum = strings.Split(num, "e")
	} else {
		splitNum = []string{num}
	}

	if !(len(splitNum) == 1 || len(splitNum) == 2) {
		return nil, false
	}

	numVal := splitNum[0]

	ret := big.NewInt(0)

	_, ok := ret.SetString(numVal, base)
	if !ok {
		return nil, false
	}

	// handle standard form
	if len(splitNum) == 2 {
		expVal := splitNum[1]

		exp := big.NewInt(0)
		_, ok = exp.SetString(expVal, base)
		if !ok {
			return nil, false
		}

		if exp.BitLen() > 64 {
			panic("TODO handle this better")
		}
		expInt := exp.Int64()

		ten := big.NewInt(10)

		if expInt < 0 {
			for ; expInt < 0; expInt++ {
				ret.Div(ret, ten)
			}
		} else if expInt > 0 {
			for ; expInt > 0; expInt-- {
				ret.Mul(ret, ten)
			}
		}
	}

	return ret, true
}

func (v *parser) parseNumberLit() *NumberLitNode {
	defer un(trace(v, "numberlit"))

	if !v.nextIs(lexer.TOKEN_NUMBER) {
		return nil
	}
	token := v.consumeToken()

	num := token.Contents
	var err error

	res := &NumberLitNode{}

	if strings.HasPrefix(num, "0x") || strings.HasPrefix(num, "0X") {
		ok := false
		res.IntValue, ok = parseInt(num[2:], 16)
		if !ok {
			v.errTokenSpecific(token, "Malformed hex literal: `%s`", num)
		}
	} else if strings.HasPrefix(num, "0b") {
		ok := false
		res.IntValue, ok = parseInt(num[2:], 2)
		if !ok {
			v.errTokenSpecific(token, "Malformed binary literal: `%s`", num)
		}
	} else if strings.HasPrefix(num, "0o") {
		ok := false
		res.IntValue, ok = parseInt(num[2:], 8)
		if !ok {
			v.errTokenSpecific(token, "Malformed octal literal: `%s`", num)
		}
	} else if lastRune := unicode.ToLower([]rune(num)[len([]rune(num))-1]); strings.ContainsRune(num, '.') || lastRune == 'f' || lastRune == 'd' || lastRune == 'q' {
		if strings.Count(num, ".") > 1 {
			v.errTokenSpecific(token, "Floating-point cannot have multiple periods: `%s`", num)
			return nil
		}
		res.IsFloat = true

		switch lastRune {
		case 'f', 'd', 'q':
			res.FloatSize = lastRune
		}

		if res.FloatSize != 0 {
			res.FloatValue, err = strconv.ParseFloat(num[:len(num)-1], 64)
		} else {
			res.FloatValue, err = strconv.ParseFloat(num, 64)
		}

		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrSyntax {
				v.errTokenSpecific(token, "Malformed floating-point literal: `%s`", num)
			} else if err.(*strconv.NumError).Err == strconv.ErrRange {
				v.errTokenSpecific(token, "Floating-point literal cannot be represented: `%s`", num)
			} else {
				v.errTokenSpecific(token, "Unexpected error from floating-point literal: %s", err)
			}
		}
	} else {
		ok := false
		res.IntValue, ok = parseInt(num, 10)
		if !ok {
			v.errTokenSpecific(token, "Malformed hex literal: `%s`", num)
		}
	}

	res.SetWhere(token.Where)
	return res
}

func (v *parser) parseStringLit() *StringLitNode {
	defer un(trace(v, "stringlit"))

	var cstring bool
	var firstToken, stringToken *lexer.Token

	if v.tokenMatches(0, lexer.TOKEN_STRING, "") {
		cstring = false
		firstToken = v.consumeToken()
		stringToken = firstToken
	} else if v.tokensMatch(lexer.TOKEN_IDENTIFIER, "c", lexer.TOKEN_STRING, "") {
		cstring = true
		firstToken = v.consumeToken()
		stringToken = v.consumeToken()
	} else {
		return nil
	}

	unescaped, err := UnescapeString(stringToken.Contents)
	if err != nil {
		v.errTokenSpecific(stringToken, "Invalid string literal: %s", err)
	}

	res := &StringLitNode{Value: unescaped, IsCString: cstring}
	res.SetWhere(lexer.NewSpan(firstToken.Where.Start(), stringToken.Where.End()))
	return res
}

func (v *parser) parseRuneLit() *RuneLitNode {
	defer un(trace(v, "runelit"))

	if !v.nextIs(lexer.TOKEN_RUNE) {
		return nil
	}
	token := v.consumeToken()
	c, err := UnescapeString(token.Contents)
	if err != nil {
		v.errTokenSpecific(token, "Invalid character literal: %s", err)
	}

	res := &RuneLitNode{Value: []rune(c)[1]}
	res.SetWhere(token.Where)
	return res
}

func trace(v *parser, name string) *parser {
	v.pushRule(name)
	return v
}

func un(v *parser) {
	v.popRule()
}
