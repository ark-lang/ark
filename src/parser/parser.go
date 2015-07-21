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
	log.Debugln("parser", strings.Join(v.ruleStack, " / "))
	log.Errorln("parser",
		util.TEXT_RED+util.TEXT_BOLD+"Parser error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
		tok.Where.Filename, tok.Where.StartLine, tok.Where.StartChar,
		fmt.Sprintf(err, stuff...))

	log.Error("parser", v.input.MarkSpan(tok.Where))

	os.Exit(util.EXIT_FAILURE_PARSE)
}

func (v *parser) errPosSpecific(pos lexer.Position, err string, stuff ...interface{}) {
	log.Debugln("parser", strings.Join(v.ruleStack, " / "))
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
	v.pushRule("node")
	defer v.popRule()

	var ret ParseNode

	if decl := v.parseDecl(); decl != nil {
		ret = decl
	} else if stat := v.parseStat(); stat != nil {
		ret = stat
	}

	return ret
}

func (v *parser) parseDocComments() []*DocComment {
	v.pushRule("doccomments")
	defer v.popRule()

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
	v.pushRule("attributes")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		return nil
	}
	attrs := make(AttrGroup)

	for v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		v.consumeToken()
		for {
			attr := &Attr{}

			if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
				v.err("Expected attribute name, got `%s`", v.peek(0).Contents)
			}
			keyToken := v.consumeToken()
			attr.setPos(keyToken.Where.Start())
			attr.Key = keyToken.Contents

			if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
				v.consumeToken()

				if !v.nextIs(lexer.TOKEN_STRING) {
					v.err("Expected attribute value, got `%s`", v.peek(0).Contents)
				}
				attr.Value = v.consumeToken().Contents
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

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
			v.err("Expected closing `]` after attribute, got `%s`", v.peek(0).Contents)
		}
		v.consumeToken()
	}

	return attrs
}

func (v *parser) parseName() *NameNode {
	v.pushRule("name")
	defer v.popRule()

	if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
		return nil
	}

	var parts []LocatedString
	for {
		if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
			v.err("Expected identifier after `::` in name, got `%s`", v.peek(0).Contents)
		}

		parts = append(parts, NewLocatedString(v.consumeToken()))

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
	v.pushRule("decl")
	defer v.popRule()

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
	v.pushRule("structdecl")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_STRUCT) {
		return nil
	}
	startToken := v.consumeToken()

	if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
		v.err("Expected name after struct keyword, got `%s`", v.peek(0).Contents)
	}
	name := v.consumeToken()

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
	v.pushRule("structbody")
	defer v.popRule()

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

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
		v.err("Expected closing `}` after struct, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &StructBodyNode{Members: members}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseUseDecl() *UseDeclNode {
	v.pushRule("usedecl")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_USE) {
		return nil
	}
	startToken := v.consumeToken()

	module := v.parseName()
	if module == nil {
		v.err("Expected valid module name after `use` keyword")
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.err("Expected `;` after use-construct, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &UseDeclNode{Module: module}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseTraitDecl() *TraitDeclNode {
	v.pushRule("traitdecl")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_TRAIT) {
		return nil
	}
	startToken := v.consumeToken()

	if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
		v.err("Expected trait name after `trait` keyword, got `%s`", v.peek(0).Contents)
	}
	name := v.consumeToken()

	if isReservedKeyword(name.Contents) {
		v.err("Cannot use reserved keyword `%s` as name for trait", name.Contents)
	}

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "{") {
		v.err("Expected starting `{` after trait name, got `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

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

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
		v.err("Expected closing `}` after trait declaration, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &TraitDeclNode{Name: NewLocatedString(name), Members: members}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseImplDecl() *ImplDeclNode {
	v.pushRule("impldecl")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_IMPL) {
		return nil
	}
	startToken := v.consumeToken()

	if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
		v.err("Expected struct name after `impl` keyword, got `%s`", v.peek(0).Contents)
	}
	structName := v.consumeToken()

	if isReservedKeyword(structName.Contents) {
		v.err("Cannot use reserved keyword `%s` as struct name", structName.Contents)
	}

	var traitName *lexer.Token
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FOR) {
		v.consumeToken()

		if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
			v.err("Expected trait name after `for` in impl declaration, got `%s`", v.peek(0).Contents)
		}
		traitName = v.consumeToken()

		if isReservedKeyword(traitName.Contents) {
			v.err("Cannot use reserved keyword `%s` as trait name", traitName.Contents)
		}
	}

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "{") {
		v.err("Expected starting `{` after impl start, got `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

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

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
		v.err("Expected closing `}` after impl declaration, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &ImplDeclNode{StructName: NewLocatedString(structName), Members: members}
	if traitName != nil {
		res.TraitName = NewLocatedString(traitName)
	}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseFuncDecl() *FunctionDeclNode {
	v.pushRule("funcdecl")
	defer v.popRule()

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
			if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
				v.err("Expected `;` after function declaration, got `%s`", v.peek(0).Contents)
			}
			v.consumeToken()
			endPosition = expr.Where().End()
		} else {
			v.err("Expected valid statement or expression after `->` in function declaration")
		}
	} else {
		body = v.parseBlock()
		if body != nil {
			endPosition = body.Where().End()
		}
	}

	var maybeEndToken *lexer.Token
	if body == nil && stat == nil && expr == nil {
		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
			v.err("Expected `;` after body-less function declaration, got `%s`", v.peek(0).Contents)
		}
		maybeEndToken = v.consumeToken()
	}

	res := &FunctionDeclNode{Header: funcHeader, Body: body, Stat: stat, Expr: expr}
	if body != nil || stat != nil || expr != nil {
		res.SetWhere(lexer.NewSpan(funcHeader.Where().Start(), endPosition))
	} else {
		res.SetWhere(lexer.NewSpan(funcHeader.Where().Start(), maybeEndToken.Where.End()))
	}
	return res
}

func (v *parser) parseFuncHeader() *FunctionHeaderNode {
	v.pushRule("funcheader")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
		return nil
	}
	startToken := v.consumeToken()

	if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
		v.err("Expected function name after `func` keyword, got `%s`", v.peek(0).Contents)
	}
	name := v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.err("Expected starting `(` after function name, got `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

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

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Expected closing `)` after function args, got `%s`", v.peek(0).Contents)
	}
	maybeEndToken := v.consumeToken()

	var returnType ParseNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, ":") {
		v.consumeToken()

		returnType = v.parseType()
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
	v.pushRule("enumdecl")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_ENUM) {
		return nil
	}
	startToken := v.consumeToken()

	if !v.nextIs(lexer.TOKEN_IDENTIFIER) {
		v.err("Expected enum name after `enum` keyword, got `%s`", v.peek(0).Contents)
	}
	name := v.consumeToken()

	if isReservedKeyword(name.Contents) {
		v.err("Cannot use reserved keyword `%s` as name for enum", name.Contents)
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.err("Expected starting `{` after enum name, got `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

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

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
		v.err("Expected closing `}` after enum, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &EnumDeclNode{Name: NewLocatedString(name), Members: members}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseEnumEntry() *EnumEntryNode {
	v.pushRule("enumentry")
	defer v.popRule()

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
	v.pushRule("vardecl")
	defer v.popRule()

	body := v.parseVarDeclBody()
	if body == nil {
		return nil
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.err("Expected `;` after variable declaration, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := body
	res.SetWhere(lexer.NewSpan(body.Where().Start(), endToken.Where.End()))
	return res
}

func (v *parser) parseVarDeclBody() *VarDeclNode {
	v.pushRule("vardeclbody")
	defer v.popRule()

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

	varType := v.parseType()
	if varType == nil && !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.err("Expected valid type in variable declaration")
	}

	var value ParseNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.consumeToken()

		value = v.parseExpr()
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
	v.pushRule("stat")
	defer v.popRule()

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
	v.pushRule("deferstat")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_DEFER) {
		return nil
	}
	startToken := v.consumeToken()

	call := v.parseCallExpr()
	if call == nil {
		v.err("Expected valid call expression in defer statement")
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.err("Expected `;` after defer statement, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &DeferStatNode{Call: call}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseIfStat() *IfStatNode {
	v.pushRule("ifstat")
	defer v.popRule()

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
	v.pushRule("matchstat")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MATCH) {
		return nil
	}
	startToken := v.consumeToken()

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expresson as value in match statement")
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.err("Expected starting `{` after value in match statement, got `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

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

		if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "->") {
			v.err("Expected `->` after match pattern, got `%s`", v.peek(0).Contents)
		}
		v.consumeToken()

		body := v.parseStat()
		if body == nil {
			v.err("Expected valid statement as body in match statement")
		}

		caseNode := &MatchCaseNode{Pattern: pattern, Body: body}
		caseNode.SetWhere(lexer.NewSpan(pattern.Where().Start(), body.Where().End()))
		cases = append(cases, caseNode)
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
		v.err("Expected closing `}` after match statement, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &MatchStatNode{Value: value, Cases: cases}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseLoopStat() *LoopStatNode {
	v.pushRule("loopstat")
	defer v.popRule()

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
	v.pushRule("returnstat")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_RETURN) {
		return nil
	}
	startToken := v.consumeToken()

	value := v.parseExpr()

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.err("Expected `;` after return statement, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &ReturnStatNode{Value: value}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseBlockStat() *BlockStatNode {
	v.pushRule("blockstat")
	defer v.popRule()

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
	v.pushRule("block")
	defer v.popRule()

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

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
		v.err("Expected closing `}` after block, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &BlockNode{Nodes: nodes}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseCallStat() *CallStatNode {
	v.pushRule("callstat")
	defer v.popRule()

	callExpr := v.parseCallExpr()
	if callExpr == nil {
		return nil
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.err("Expected `;` after call statement")
	}
	endToken := v.consumeToken()

	res := &CallStatNode{Call: callExpr}
	res.SetWhere(lexer.NewSpan(callExpr.Where().Start(), endToken.Where.End()))
	return res
}

func (v *parser) parseAssignStat() ParseNode {
	v.pushRule("assignstat")
	defer v.popRule()

	startPos := v.currentToken

	accessExpr := v.parseAccessExpr()
	if accessExpr == nil || !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		v.currentToken = startPos
		return nil
	}

	// consume '='
	v.consumeToken()

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expression in assignment statement")
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.err("Expected `;` after assignment statement, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &AssignStatNode{Target: accessExpr, Value: value}
	res.SetWhere(lexer.NewSpan(accessExpr.Where().Start(), endToken.Where.End()))
	return res
}

func (v *parser) parseType() ParseNode {
	v.pushRule("type")
	defer v.popRule()

	var res ParseNode
	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		res = v.parsePointerType()
	} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		res = v.parseTupleType()
	} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		res = v.parseArrayType()
	} else if v.nextIs(lexer.TOKEN_IDENTIFIER) {
		res = v.parseTypeReference()
	}

	return res
}

func (v *parser) parsePointerType() *PointerTypeNode {
	v.pushRule("pointertype")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		return nil
	}
	startToken := v.consumeToken()

	target := v.parseType()
	if target == nil {
		v.err("Expected valid type after `^` in pointer type")
	}

	res := &PointerTypeNode{TargetType: target}
	res.SetWhere(lexer.NewSpan(startToken.Where.Start(), target.Where().End()))

	return res
}

func (v *parser) parseTupleType() *TupleTypeNode {
	v.pushRule("tupletype")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}
	startToken := v.consumeToken()

	var members []ParseNode
	for {
		memberType := v.parseType()
		if memberType == nil {
			v.err("Expected valid type in tuple type")
		}
		members = append(members, memberType)

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			break
		}
		v.consumeToken()
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Expected closing `)` after tuple type, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &TupleTypeNode{MemberTypes: members}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseArrayType() *ArrayTypeNode {
	v.pushRule("arraytype")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		return nil
	}
	startToken := v.consumeToken()

	length := v.parseNumberLit()
	if length != nil && length.IsFloat {
		v.err("Expected integer length for array type")
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
		v.err("Expected closing `]` in array type, got `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	memberType := v.parseType()
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
	v.pushRule("typereference")
	defer v.popRule()

	name := v.parseName()
	if name == nil {
		return nil
	}

	res := &TypeReferenceNode{Reference: name}
	res.SetWhere(name.Where())
	return res
}

func (v *parser) parseExpr() ParseNode {
	v.pushRule("expr")
	defer v.popRule()

	pri := v.parsePrimaryExpr()
	if pri == nil {
		return nil
	}

	if bin := v.parseBinaryOperator(0, pri); bin != nil {
		return bin
	}

	return pri
}

func (v *parser) parseBinaryOperator(upperPrecedence int, lhand ParseNode) ParseNode {
	v.pushRule("binop")
	defer v.popRule()

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

		rhand := v.parsePrimaryExpr()
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

func (v *parser) parsePrimaryExpr() ParseNode {
	v.pushRule("primaryexpr")
	defer v.popRule()

	var res ParseNode

	if sizeofExpr := v.parseSizeofExpr(); sizeofExpr != nil {
		res = sizeofExpr
	} else if addrofExpr := v.parseAddrofExpr(); addrofExpr != nil {
		res = addrofExpr
	} else if litExpr := v.parseLitExpr(); litExpr != nil {
		res = litExpr
	} else if unaryExpr := v.parseUnaryExpr(); unaryExpr != nil {
		res = unaryExpr
	} else if callExpr := v.parseCallExpr(); callExpr != nil {
		res = callExpr
	} else if castExpr := v.parseCastExpr(); castExpr != nil {
		res = castExpr
	} else if accessExpr := v.parseAccessExpr(); accessExpr != nil {
		res = accessExpr
	}

	return res
}

func (v *parser) parseSizeofExpr() *SizeofExprNode {
	v.pushRule("sizeofexpr")
	defer v.popRule()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_SIZEOF) {
		return nil
	}
	startToken := v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.err("Expected opening `(` in sizeof expression, got `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expression in sizeof expression")
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Expected closing `)` after sizeof expression, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &SizeofExprNode{Value: value}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseAddrofExpr() *AddrofExprNode {
	v.pushRule("addrofexpr")
	defer v.popRule()

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
	v.pushRule("litexpr")
	defer v.popRule()

	var res ParseNode

	if arrayLit := v.parseArrayLit(); arrayLit != nil {
		res = arrayLit
	} else if tupleLit := v.parseTupleLit(); tupleLit != nil {
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
	v.pushRule("castexpr")
	defer v.popRule()

	startPos := v.currentToken

	typ := v.parseType()
	if typ == nil || !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.currentToken = startPos
		return nil
	}
	v.consumeToken()

	value := v.parseExpr()
	if value == nil {
		v.err("Expected valid expression in cast expression")
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Expected closing `)` after cast expression, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &CastExprNode{Type: typ, Value: value}
	res.SetWhere(lexer.NewSpan(typ.Where().Start(), endToken.Where.End()))
	return res
}

func (v *parser) parseUnaryExpr() *UnaryExprNode {
	v.pushRule("unaryexpr")
	defer v.popRule()

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

func (v *parser) parseCallExpr() *CallExprNode {
	v.pushRule("callexpr")
	defer v.popRule()

	startPos := v.currentToken

	function := v.parseAccessExpr()
	if function == nil || !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.currentToken = startPos
		return nil
	}
	v.consumeToken()

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

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Expected closing `)` after call, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &CallExprNode{Function: function, Arguments: args}
	res.SetWhere(lexer.NewSpan(function.Where().Start(), endToken.Where.End()))
	return res
}

func (v *parser) parseAccessExpr() ParseNode {
	v.pushRule("accessexpr")
	defer v.popRule()

	var lhand ParseNode
	if name := v.parseName(); name != nil {
		lhand = &VariableAccessNode{Name: name}
		lhand.SetWhere(name.Where())
	} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.consumeToken()

		lhand = v.parseAccessExpr()
		if lhand == nil {
			v.err("Expected valid access expression after `(` in access expression")
		}

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
			v.err("Expected closing `)` in access expression, got `%s`", v.peek(0).Contents)
		}
		v.consumeToken()
	} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		startToken := v.consumeToken()

		value := v.parseExpr()
		if value == nil {
			v.err("Expected expression after dereference operator")
		}

		lhand = &DerefAccessNode{Value: value}
		lhand.SetWhere(lexer.NewSpan(startToken.Where.Start(), value.Where().End()))
	} else {
		return nil
	}

	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ".") {
			// struct access
			v.consumeToken()

			if v.peek(0).Type != lexer.TOKEN_IDENTIFIER {
				v.err("Expected identifier after struct access, got `%s`", v.peek(0).Contents)
			}

			member := v.consumeToken()

			res := &StructAccessNode{Struct: lhand, Member: NewLocatedString(member)}
			res.SetWhere(lexer.NewSpan(lhand.Where().Start(), member.Where.End()))
			lhand = res
		} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
			// array access
			v.consumeToken()

			index := v.parseExpr()
			if index == nil {
				v.err("Expected valid expression as array index")
			}

			if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
				v.err("Expected `]` after array index, found `%s`", v.peek(0).Contents)
			}
			endToken := v.consumeToken()

			res := &ArrayAccessNode{Array: lhand, Index: index}
			res.SetWhere(lexer.NewSpan(lhand.Where().Start(), endToken.Where.End()))
			lhand = res
		} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "|") {
			// tuple access
			v.consumeToken()

			index := v.parseNumberLit()
			if index == nil || index.IsFloat {
				v.err("Expected integer for tuple index")
			}

			if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "|") {
				v.err("Expected `|` after tuple index, found `%s`", v.peek(0).Contents)
			}
			endToken := v.consumeToken()

			res := &TupleAccessNode{Tuple: lhand, Index: int(index.IntValue)}
			res.SetWhere(lexer.NewSpan(index.Where().Start(), endToken.Where.End()))
			lhand = res
		} else {
			break
		}
	}

	return lhand
}

func (v *parser) parseArrayLit() *ArrayLiteralNode {
	v.pushRule("arraylit")
	defer v.popRule()

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

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
		v.err("Expected closing `]` after array literal, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

	res := &ArrayLiteralNode{Values: values}
	res.SetWhere(lexer.NewSpanFromTokens(startToken, endToken))
	return res
}

func (v *parser) parseTupleLit() *TupleLiteralNode {
	v.pushRule("tuplelit")
	defer v.popRule()

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

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Expected closing `)` after tuple literal, got `%s`", v.peek(0).Contents)
	}
	endToken := v.consumeToken()

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
	v.pushRule("boollit")
	defer v.popRule()

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
	v.pushRule("numberlit")
	defer v.popRule()

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
	v.pushRule("stringlit")
	defer v.popRule()

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
	v.pushRule("runelit")
	defer v.popRule()

	if !v.nextIs(lexer.TOKEN_RUNE) {
		return nil
	}
	token := v.consumeToken()
	c := UnescapeString(token.Contents)

	res := &RuneLitNode{Value: []rune(c)[1]}
	res.SetWhere(token.Where)
	return res
}
