package parser

// Note that you should include a lot of calls to panic() where something's happening that shouldn't be.
// This will help to find bugs. Once the compiler is in a better state, a lot of these calls can be removed.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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
}

func Parse(input *lexer.Sourcefile) *ParseTree {
	p := &parser{
		input:            input,
		binOpPrecedences: newBinOpPrecedenceMap(),
		tree:             &ParseTree{Source: input},
	}

	log.Verboseln("parser", util.TEXT_BOLD+util.TEXT_GREEN+"Started parsing "+util.TEXT_RESET+input.Name)
	t := time.Now()

	p.parse()

	dur := time.Since(t)
	log.Verbose("parser", util.TEXT_BOLD+util.TEXT_GREEN+"Finished parsing"+util.TEXT_RESET+" %s (%.2fms)\n",
		input.Name, float32(dur.Nanoseconds())/1000000)

	return p.tree
}

func (v *parser) err(err string, stuff ...interface{}) {
	v.errPos(err, stuff...)
}

func (v *parser) errToken(err string, stuff ...interface{}) {
	v.errTokenSpecific(v.peek(0), err, stuff...)
}

func (v *parser) errPos(err string, stuff ...interface{}) {
	v.errPosSpecific(v.peek(0).Where.Start(), err, stuff...)
}

func (v *parser) errTokenSpecific(tok *lexer.Token, err string, stuff ...interface{}) {
	v.dumpRules()
	log.Errorln("parser",
		util.TEXT_RED+util.TEXT_BOLD+"Parser error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
		tok.Where.Filename, tok.Where.StartLine, tok.Where.StartChar,
		fmt.Sprintf(err, stuff...))

	log.Error("parser", v.input.MarkSpan(tok.Where))

	os.Exit(util.EXIT_FAILURE_PARSE)
}

func (v *parser) errPosSpecific(pos lexer.Position, err string, stuff ...interface{}) {
	v.dumpRules()
	log.Errorln("parser",
		util.TEXT_RED+util.TEXT_BOLD+"Parser error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
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

func (v *parser) nextIs(typ lexer.TokenType) bool {
	return v.peek(0).Type == typ
}

func (v *parser) expect(typ lexer.TokenType, val string) *lexer.Token {
	if !v.tokenMatches(0, typ, val) {
		tok := v.peek(0)
		if val != "" {
			v.errToken("Expected `%s` (%s), got `%s` (%s)", val, typ, tok.Contents, tok.Type)
		} else {
			v.errToken("Expected %s, got %s (`%s`)", typ, tok.Type, tok.Contents)
		}
	}
	return v.consumeToken()
}

func (v *parser) parse() {
	for v.peek(0) != nil {
		if n := v.parseNode(); n != nil {
			v.tree.AddNode(n)
		} else {
			panic("what's this over here?")
		}
	}
}

func (v *parser) parseNode() ParseNode {
	defer un(trace(v, "node"))

	var ret ParseNode

	if decl := v.parseDecl(); decl != nil {
		ret = decl
	} else if stat := v.parseStat(); stat != nil {
		ret = stat
	}

	return ret
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

func (v *parser) parseDecl() ParseNode {
	defer un(trace(v, "decl"))

	var res ParseNode
	docComments := v.parseDocComments()
	attrs := v.parseAttributes()

	if structDecl := v.parseStructDecl(); structDecl != nil {
		res = structDecl
	} else if useDecl := v.parseUseDecl(); useDecl != nil {
		res = useDecl
	} else if traitDecl := v.parseTraitDecl(); traitDecl != nil {
		res = traitDecl
	} else if implDecl := v.parseImplDecl(); implDecl != nil {
		res = implDecl
	} else if funcDecl := v.parseFuncDecl(); funcDecl != nil {
		res = funcDecl
	} else if enumDecl := v.parseEnumDecl(); enumDecl != nil {
		res = enumDecl
	} else if varDecl := v.parseVarDecl(); varDecl != nil {
		res = varDecl
	}

	if len(docComments) != 0 && res != nil {
		res.SetDocComments(docComments)
	}

	if attrs != nil && res != nil {
		res.SetAttrs(attrs)
	}

	return res
}

func (v *parser) parseStructDecl() *StructDeclNode {
	defer un(trace(v, "structdecl"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_STRUCT) {
		return nil
	}
	startToken := v.consumeToken()

	name := v.expect(lexer.TOKEN_IDENTIFIER, "")
	if isReservedKeyword(name.Contents) {
		v.err("Cannot use reserved keyword `%s` as name for struct", name.Contents)
	}

	body := v.parseStructBody()
	if body == nil {
		v.err("Expected starting `{` after struct name")
	}

	res := &StructDeclNode{Name: NewLocatedString(name), Body: body}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), body.Where().End()))
	return res
}

func (v *parser) parseStructBody() *StructBodyNode {
	defer un(trace(v, "structbody"))

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		return nil
	}
	startToken := v.consumeToken()

	var members []*VarDeclNode
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			break
		}

		member := v.parseVarDeclBody()
		if member == nil {
			v.err("Expected valid variable declaration in struct")
		}
		members = append(members, member)

		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			v.consumeToken()
		}
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "}")

	res := &StructBodyNode{Members: members}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseUseDecl() *UseDeclNode {
	defer un(trace(v, "usedecl"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_USE) {
		return nil
	}
	startToken := v.consumeToken()

	module := v.parseName()
	if module == nil {
		v.err("Expected valid module name after `use` keyword")
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ";")

	res := &UseDeclNode{Module: module}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseTraitDecl() *TraitDeclNode {
	defer un(trace(v, "traitdecl"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_TRAIT) {
		return nil
	}
	startToken := v.consumeToken()

	name := v.expect(lexer.TOKEN_IDENTIFIER, "")
	if isReservedKeyword(name.Contents) {
		v.err("Cannot use reserved keyword `%s` as name for trait", name.Contents)
	}

	v.expect(lexer.TOKEN_IDENTIFIER, "{")

	var members []*FunctionDeclNode
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			break
		}

		member, ok := v.parseDecl().(*FunctionDeclNode)
		if member == nil || !ok {
			v.err("Expected valid function declaration in trait declaration")
		}
		members = append(members, member)
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "}")

	res := &TraitDeclNode{Name: NewLocatedString(name), Members: members}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseImplDecl() *ImplDeclNode {
	defer un(trace(v, "impldecl"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_IMPL) {
		return nil
	}
	startToken := v.consumeToken()

	structName := v.expect(lexer.TOKEN_IDENTIFIER, "")
	if isReservedKeyword(structName.Contents) {
		v.err("Cannot use reserved keyword `%s` as struct name", structName.Contents)
	}

	var traitName *lexer.Token
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FOR) {
		v.consumeToken()

		traitName = v.expect(lexer.TOKEN_IDENTIFIER, "")
		if isReservedKeyword(traitName.Contents) {
			v.err("Cannot use reserved keyword `%s` as trait name", traitName.Contents)
		}
	}

	v.expect(lexer.TOKEN_SEPARATOR, "{")

	var members []*FunctionDeclNode
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			break
		}

		member, ok := v.parseDecl().(*FunctionDeclNode)
		if member == nil || !ok {
			v.err("Expected valid function declaration in impl declaration")
		}
		members = append(members, member)
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "}")

	res := &ImplDeclNode{StructName: NewLocatedString(structName), Members: members}
	if traitName != nil {
		res.TraitName = NewLocatedString(traitName)
	}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseFuncDecl() *FunctionDeclNode {
	defer un(trace(v, "funcdecl"))

	funcHeader := v.parseFuncHeader()
	if funcHeader == nil {
		return nil
	}

	var body *BlockNode
	var stat, expr ParseNode
	var endPosition lexer.Position
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "->") {
		v.consumeToken()

		if stat = v.parseStat(); stat != nil {
			endPosition = stat.Where().End()
		} else if expr = v.parseExpr(); expr != nil {
			tok := v.expect(lexer.TOKEN_SEPARATOR, ";")
			endPosition = tok.Where.End()
		} else {
			v.err("Expected valid statement or expression after `->` in function declaration")
		}
	} else {
		body = v.parseBlock()
		if body != nil {
			endPosition = body.Where().End()
		}
	}

	if body == nil && stat == nil && expr == nil {
		tok := v.expect(lexer.TOKEN_SEPARATOR, ";")
		endPosition = tok.Where.End()
	}

	res := &FunctionDeclNode{Header: funcHeader, Body: body, Stat: stat, Expr: expr}
	res.SetWhere(lexer.NewSpan(funcHeader.Where().Start(), endPosition))
	return res
}

func (v *parser) parseFuncHeader() *FunctionHeaderNode {
	defer un(trace(v, "funcheader"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
		return nil
	}
	startToken := v.consumeToken()

	name := v.expect(lexer.TOKEN_IDENTIFIER, "")
	v.expect(lexer.TOKEN_SEPARATOR, "(")

	var args []*VarDeclNode
	variadic := false
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
			break
		}

		if v.tokensMatch(lexer.TOKEN_SEPARATOR, ".", lexer.TOKEN_SEPARATOR, ".", lexer.TOKEN_SEPARATOR, ".") {
			v.consumeTokens(3)
			if !variadic {
				variadic = true
			} else {
				v.err("Duplicate `...` in function arguments")
			}
		} else {
			arg := v.parseVarDeclBody()
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

	var returnType ParseNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, ":") {
		v.consumeToken()

		returnType = v.parseType(true)
		if returnType == nil {
			v.err("Expected valid type after `:` in function header")
		}
	}

	res := &FunctionHeaderNode{Name: NewLocatedString(name), Arguments: args, Variadic: variadic}
	if returnType != nil {
		res.ReturnType = returnType
		res.SetWhere(lexer.NewSpan(startToken.Where.Start(), returnType.Where().End()))
	} else {
		res.SetWhere(lexer.NewSpanFromTokens(startToken, maybeEndToken))
	}
	return res
}

func (v *parser) parseEnumDecl() *EnumDeclNode {
	defer un(trace(v, "enumdecl"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_ENUM) {
		return nil
	}
	startToken := v.consumeToken()

	name := v.expect(lexer.TOKEN_IDENTIFIER, "")
	if isReservedKeyword(name.Contents) {
		v.err("Cannot use reserved keyword `%s` as name for enum", name.Contents)
	}

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

	res := &EnumDeclNode{Name: NewLocatedString(name), Members: members}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
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
	var structBody *StructBodyNode
	var tupleBody *TupleTypeNode
	var lastPos lexer.Position
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.consumeToken()

		value = v.parseNumberLit()
		if value == nil || value.IsFloat {
			v.err("Expected valid integer after `=` in enum entry")
		}
		lastPos = value.Where().End()
	} else if tupleBody = v.parseTupleType(); tupleBody != nil {
		lastPos = tupleBody.Where().End()
	} else if structBody = v.parseStructBody(); structBody != nil {
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

func (v *parser) parseVarDecl() *VarDeclNode {
	defer un(trace(v, "vardecl"))

	body := v.parseVarDeclBody()
	if body == nil {
		return nil
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ";")

	res := body
	res.SetWhere(lexer.NewSpan(body.Where().Start(), endToken.Where.End()))
	return res
}

func (v *parser) parseVarDeclBody() *VarDeclNode {
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

	varType := v.parseType(true)
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

	res := &VarDeclNode{Name: NewLocatedString(name), Type: varType}
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

func (v *parser) parseStat() ParseNode {
	defer un(trace(v, "stat"))

	var res ParseNode

	if deferStat := v.parseDeferStat(); deferStat != nil {
		res = deferStat
	} else if ifStat := v.parseIfStat(); ifStat != nil {
		res = ifStat
	} else if matchStat := v.parseMatchStat(); matchStat != nil {
		res = matchStat
	} else if loopStat := v.parseLoopStat(); loopStat != nil {
		res = loopStat
	} else if returnStat := v.parseReturnStat(); returnStat != nil {
		res = returnStat
	} else if blockStat := v.parseBlockStat(); blockStat != nil {
		res = blockStat
	} else if callStat := v.parseCallStat(); callStat != nil {
		res = callStat
	} else if assignStat := v.parseAssignStat(); assignStat != nil {
		res = assignStat
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

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ";")

	res := &DeferStatNode{Call: call}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
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

		v.expect(lexer.TOKEN_OPERATOR, "->")

		body := v.parseStat()
		if body == nil {
			v.err("Expected valid statement as body in match statement")
		}

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
		v.err("Expected valid block as body of loop statement")
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

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ";")

	res := &ReturnStatNode{Value: value}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
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
		node := v.parseNode()
		if node == nil {
			break
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

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ";")

	res := &CallStatNode{Call: callExpr}
	res.SetWhere(lexer.NewSpan(callExpr.Where().Start(), endToken.Where.End()))
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

	if value == nil {
		v.err("Expected valid expression in assignment statement")
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ";")

	res := &AssignStatNode{Target: accessExpr, Value: value}
	res.SetWhere(lexer.NewSpan(accessExpr.Where().Start(), endToken.Where.End()))
	return res
}

func (v *parser) parseType(doRefs bool) ParseNode {
	defer un(trace(v, "type"))

	var res ParseNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		res = v.parsePointerType()
	} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		res = v.parseTupleType()
	} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		res = v.parseArrayType()
	} else if doRefs && v.nextIs(lexer.TOKEN_IDENTIFIER) {
		res = v.parseTypeReference()
	}

	return res
}

func (v *parser) parsePointerType() *PointerTypeNode {
	defer un(trace(v, "pointertype"))

	if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		return nil
	}
	startToken := v.consumeToken()

	target := v.parseType(true)
	if target == nil {
		v.err("Expected valid type after `^` in pointer type")
	}

	res := &PointerTypeNode{TargetType: target}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), target.Where().End()))

	return res
}

func (v *parser) parseTupleType() *TupleTypeNode {
	defer un(trace(v, "tupletype"))

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}
	startToken := v.consumeToken()

	var members []ParseNode
	for {
		memberType := v.parseType(true)
		if memberType == nil {
			v.err("Expected valid type in tuple type")
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

	memberType := v.parseType(true)
	if memberType == nil {
		v.err("Expected valid type in array type")
	}

	res := &ArrayTypeNode{MemberType: memberType}
	if length != nil {
		// TODO: Defend against overflow
		res.Length = int(length.IntValue)
	}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), memberType.Where().End()))
	return res
}

func (v *parser) parseTypeReference() *TypeReferenceNode {
	defer un(trace(v, "typereference"))

	name := v.parseName()
	if name == nil {
		return nil
	}

	res := &TypeReferenceNode{Reference: name}
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
		} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "|") {
			// tuple index
			v.consumeToken()
			defer un(trace(v, "tupleindex"))

			index := v.parseNumberLit()
			if index == nil || index.IsFloat {
				v.err("Expected integer for tuple index")
			}

			endToken := v.expect(lexer.TOKEN_OPERATOR, "|")

			res := &TupleAccessNode{Tuple: expr, Index: int(index.IntValue)}
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
	} else if addrofExpr := v.parseAddrofExpr(); addrofExpr != nil {
		res = addrofExpr
	} else if litExpr := v.parseLitExpr(); litExpr != nil {
		res = litExpr
	} else if castExpr := v.parseCastExpr(); castExpr != nil {
		res = castExpr
	} else if unaryExpr := v.parseUnaryExpr(); unaryExpr != nil {
		res = unaryExpr
	} else if name := v.parseName(); name != nil {
		res = &VariableAccessNode{Name: name}
	}

	return res
}

func (v *parser) parseSizeofExpr() *SizeofExprNode {
	defer un(trace(v, "sizeofexpr"))

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_SIZEOF) {
		return nil
	}
	startToken := v.consumeToken()

	v.expect(lexer.TOKEN_SEPARATOR, "(")

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expression in sizeof expression")
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ")")

	res := &SizeofExprNode{Value: value}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseAddrofExpr() *AddrofExprNode {
	defer un(trace(v, "addrofexpr"))

	if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "&") {
		return nil
	}
	startToken := v.consumeToken()

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expression after addrof expression")
	}

	res := &AddrofExprNode{Value: value}
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

	typ := v.parseType(false)
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

	if !v.nextIs(lexer.TOKEN_OPERATOR) {
		return nil
	}

	op := stringToUnOpType(v.peek(0).Contents)
	if op == UNOP_ERR {
		return nil
	}
	startToken := v.consumeToken()

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expression after unary operator")
	}

	res := &UnaryExprNode{Value: value, Operator: op}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), value.Where().End()))
	return res
}

func (v *parser) parseCompositeLiteral() ParseNode {
	defer un(trace(v, "complit"))

	var res ParseNode
	if arrayLit := v.parseArrayLit(); arrayLit != nil {
		res = arrayLit
	} else if structLit := v.parseStructLit(); structLit != nil {
		res = structLit
	}
	return res
}

func (v *parser) parseArrayLit() *ArrayLiteralNode {
	defer un(trace(v, "arraylit"))

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		return nil
	}
	startToken := v.consumeToken()

	var values []ParseNode
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
			break
		}

		value := v.parseExpr()
		if value == nil {
			v.err("Expected valid expression in array literal")
		}
		values = append(values, value)

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "]")

	res := &ArrayLiteralNode{Values: values}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
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
			v.err("Expected valid expression in tuple literal")
		}
		values = append(values, value)

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, ")")

	// Dirty hack
	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ".") {
		v.currentToken = startPos
		return nil
	}

	res := &TupleLiteralNode{Values: values}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseStructLit() *StructLiteralNode {
	defer un(trace(v, "structlit"))

	startPos := v.currentToken
	name := v.parseName()
	if !v.tokensMatch(lexer.TOKEN_SEPARATOR, "{", lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_OPERATOR, ":") {
		v.currentToken = startPos
		return nil
	}
	startToken := v.consumeToken()

	var members []LocatedString
	var values []ParseNode
	for {
		member := v.expect(lexer.TOKEN_IDENTIFIER, "")
		members = append(members, NewLocatedString(member))

		v.expect(lexer.TOKEN_OPERATOR, ":")

		value := v.parseExpr()
		if value == nil {
			v.err("Expected valid expression in struct literal")
		}
		values = append(values, value)

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}

	endToken := v.expect(lexer.TOKEN_SEPARATOR, "}")

	res := &StructLiteralNode{Name: name, Members: members, Values: values}
	if name != nil {
		res.SetWhere(lexer.NewSpan(name.Where().Start(), endToken.Where.End()))
	} else {
		res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	}
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

func (v *parser) parseNumberLit() *NumberLitNode {
	defer un(trace(v, "numberlit"))

	// TODO: Arbitrary base numbers?
	if !v.nextIs(lexer.TOKEN_NUMBER) {
		return nil
	}
	token := v.consumeToken()

	num := token.Contents
	var err error

	res := &NumberLitNode{}

	if strings.HasPrefix(num, "0x") || strings.HasPrefix(num, "0X") {
		// Hexadecimal integer
		for _, r := range num[2:] {
			if r == '_' {
				continue
			}
			res.IntValue *= 16
			if val := uint64(hexRuneToInt(r)); val >= 0 {
				res.IntValue += val
			} else {
				v.err("Malformed hex literal: `%s`", num)
			}
		}
	} else if strings.HasPrefix(num, "0b") {
		// Binary integer
		for _, r := range num[2:] {
			if r == '_' {
				continue
			}
			res.IntValue *= 2
			if val := uint64(binRuneToInt(r)); val >= 0 {
				res.IntValue += val
			} else {
				v.err("Malformed binary literal: `%s`", num)
			}
		}
	} else if strings.HasPrefix(num, "0o") {
		// Octal integer
		for _, r := range num[2:] {
			if r == '_' {
				continue
			}
			res.IntValue *= 8
			if val := uint64(octRuneToInt(r)); val >= 0 {
				res.IntValue += val
			} else {
				v.err("Malformed octal literal: `%s`", num)
			}
		}
	} else if lastRune := unicode.ToLower([]rune(num)[len([]rune(num))-1]); strings.ContainsRune(num, '.') || lastRune == 'f' || lastRune == 'd' || lastRune == 'q' {
		if strings.Count(num, ".") > 1 {
			v.err("Floating-point cannot have multiple periods: `%s`", num)
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
				v.err("Malformed floating-point literal: `%s`", num)
			} else if err.(*strconv.NumError).Err == strconv.ErrRange {
				v.err("Floating-point literal cannot be represented: `%s`", num)
			} else {
				v.err("Unexpected error from floating-point literal: %s", err)
			}
		}
	} else {
		// Decimal integer
		for _, r := range num {
			if r == '_' {
				continue
			}
			res.IntValue *= 10
			res.IntValue += uint64(r - '0')
		}
	}

	res.SetWhere(token.Where)
	return res
}

func (v *parser) parseStringLit() *StringLitNode {
	defer un(trace(v, "stringlit"))

	if !v.nextIs(lexer.TOKEN_STRING) {
		return nil
	}
	token := v.consumeToken()
	strLen := len(token.Contents)
	res := &StringLitNode{Value: UnescapeString(token.Contents[1 : strLen-1])}
	res.SetWhere(token.Where)
	return res
}

func (v *parser) parseRuneLit() *RuneLitNode {
	defer un(trace(v, "runelit"))

	if !v.nextIs(lexer.TOKEN_RUNE) {
		return nil
	}
	token := v.consumeToken()
	c := UnescapeString(token.Contents)

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
