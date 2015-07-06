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

	"github.com/ark-lang/ark/lexer"
	"github.com/ark-lang/ark/util"
)

type parser struct {
	module       *Module
	input        []*lexer.Token
	currentToken int
	verbose      bool

	modules           map[string]*Module
	scope             *Scope
	binOpPrecedences  map[BinOpType]int
	attrs             AttrGroup
	curNodeTokenStart int
	docCommentsBuf    []*DocComment
}

func (v *parser) err(err string, stuff ...interface{}) {
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Parser error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		v.peek(0).Filename, v.peek(0).LineNumber, v.peek(0).CharNumber, fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_PARSE)
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

func (v *parser) consumeTokens(num int) {
	for i := 0; i < num; i++ {
		v.consumeToken()
	}
}

func (v *parser) pushNode(node Node) {
	v.module.Nodes = append(v.module.Nodes, node)
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

func (v *parser) fetchAttrs() AttrGroup {
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

func Parse(input *lexer.Sourcefile, modules map[string]*Module, verbose bool) *Module {
	p := &parser{
		module: &Module{
			Nodes: make([]Node, 0),
			Path:  input.Path,
			Name:  input.Name,
		},
		input:            input.Tokens,
		verbose:          verbose,
		scope:            NewGlobalScope(),
		binOpPrecedences: newBinOpPrecedenceMap(),
	}
	p.module.GlobalScope = p.scope
	p.modules = modules
	modules[input.Name] = p.module

	// use the C module by default.
	// TODO: should we allow this?
	// it means the errors are a bit
	// more clear in some cases
	p.useModule("C")

	if verbose {
		fmt.Println(util.TEXT_BOLD+util.TEXT_GREEN+"Started parsing"+util.TEXT_RESET, input.Name)
	}

	t := time.Now()
	p.parse()

	sem := &semanticAnalyzer{module: p.module}
	sem.analyze(modules)

	dur := time.Since(t)

	if verbose {
		for _, n := range p.module.Nodes {
			fmt.Println(n.String())
		}

		fmt.Printf(util.TEXT_BOLD+util.TEXT_GREEN+"Finished parsing"+util.TEXT_RESET+" %s (%.2fms)\n",
			input.Name, float32(dur.Nanoseconds())/1000000)
	}

	return p.module
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

// reads an identifier, with module access, without consuming it
// returns the name and the number of tokens it takes up
func (v *parser) peekName() (unresolvedName, int) {
	var name unresolvedName

	numTok := 0

	for {
		if v.tokenMatches(numTok, lexer.TOKEN_IDENTIFIER, "") {
			ident := v.peek(numTok).Contents
			numTok++
			if v.tokenMatches(numTok, lexer.TOKEN_OPERATOR, "::") {
				numTok++
				name.moduleNames = append(name.moduleNames, ident)
			} else {
				name.name = ident
				return name, numTok
			}
		} else {
			v.err("Expected identifier, found `%s`", v.peek(0).Contents)
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
	for v.tokenMatches(0, lexer.TOKEN_DOCCOMMENT, "") {
		v.docCommentsBuf = append(v.docCommentsBuf, v.parseDocComment())
	}

	// this is a little dirty, but allows for attribute block without reflection (part 1 / 2)
	if v.attrs == nil {
		v.attrs = v.parseAttrs()
	} else {
		v.attrs.Extend(v.parseAttrs())
	}

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
	locationToken := v.peek(0)
	filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber

	if ifStat := v.parseIfStat(); ifStat != nil {
		ret = ifStat
	} else if matchStat := v.parseMatchStat(); matchStat != nil {
		ret = matchStat
	} else if loopStat := v.parseLoopStat(); loopStat != nil {
		ret = loopStat
	} else if returnStat := v.parseReturnStat(); returnStat != nil {
		ret = returnStat
	} else if blockStat := v.parseBlockStat(); blockStat != nil {
		ret = blockStat
	} else if callStat := v.parseCallStat(); callStat != nil {
		ret = callStat
	} else if assignStat := v.parseAssignStat(); assignStat != nil {
		ret = assignStat
	} else {
		return nil
	}

	ret.setPos(filename, line, char)
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

	locationToken := v.peek(0)
	filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber

	if deref := v.parseDerefExpr(); deref != nil {
		deref.setPos(filename, line, char)
		assign.Deref = deref
	} else if access := v.parseAccessExpr(); access != nil {
		access.setPos(filename, line, char)
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

func (v *parser) parseBlockStat() *BlockStat {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_DO) && !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		return nil
	}

	// messy but fuck it
	hasDo := false
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_DO) {
		v.consumeToken()
		hasDo = true
	}

	var blockStat *BlockStat
	if block := v.parseBlock(true); block != nil {
		blockStat = &BlockStat{Block: block}
	} else if hasDo {
		v.err("Expected block after `%d` keyword, found `%s`", KEYWORD_DO, v.peek(0).Contents)
	} else {
		v.err("Expected block, found `%s`", v.peek(0).Contents)
	}

	return blockStat
}

func (v *parser) parseCallStat() *CallStat {
	locationToken := v.peek(0)
	filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber
	if call := v.parseCallExpr(); call != nil {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
			v.consumeToken()
			callStat := &CallStat{Call: call}
			call.setPos(filename, line, char)
			return callStat
		}
		v.err("Expected semicolon after function call statement, found `%s`", v.peek(0).Contents)
	}
	return nil
}

func (v *parser) moduleInUse(name string) bool {
	if _, ok := v.modules[name]; ok {
		if v.scope.Outer != nil {
			if _, ok := v.scope.Outer.UsedModules[name]; ok {
				return true
			}
		} else {
			if _, ok := v.scope.UsedModules[name]; ok {
				return true
			}
		}
	}
	return false
}

func (v *parser) useModule(name string) {
	// check if the module exists in the modules that are
	// parsed to avoid any weird errors
	if moduleToUse, ok := v.modules[name]; ok {
		if v.scope.Outer != nil {
			v.scope.Outer.UsedModules[name] = moduleToUse
		} else {
			v.scope.UsedModules[name] = moduleToUse
		}
	}
}

func (v *parser) parseUseDecl() Decl {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_USE) {
		return nil
	}

	// consume use, since we know it's
	// already there due to the previous check
	v.consumeToken()

	var useDecl *UseDecl

	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
		moduleName := v.consumeToken().Contents
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
			v.consumeToken()
			useDecl = &UseDecl{ModuleName: moduleName, Scope: v.scope}
		} else {
			v.err("Expected semicolon after use declaration, found `%s`", v.peek(0).Contents)
		}
	}

	v.useModule(useDecl.ModuleName)
	return useDecl

	/*v.err("attempting to use undefined module `%s`", useDecl.ModuleName)
	return nil*/ // TODO?
}

func (v *parser) parseDecl() Decl {
	var ret Decl

	locationToken := v.peek(0)
	filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber

	if structureDecl := v.parseStructDecl(); structureDecl != nil {
		ret = structureDecl
	} else if useDecl := v.parseUseDecl(); useDecl != nil {
		ret = useDecl
	} else if traitDecl := v.parseTraitDecl(); traitDecl != nil {
		ret = traitDecl
	} else if implDecl := v.parseImplDecl(); implDecl != nil {
		ret = implDecl
	} else if moduleDecl := v.parseModuleDecl(); moduleDecl != nil {
		ret = moduleDecl
	} else if functionDecl := v.parseFunctionDecl(); functionDecl != nil {
		ret = functionDecl
	} else if enumDecl := v.parseEnumDecl(); enumDecl != nil {
		ret = enumDecl
	} else if variableDecl := v.parseVariableDecl(true); variableDecl != nil {
		ret = variableDecl
	} else {
		return nil
	}

	ret.setPos(filename, line, char)

	return ret
}

func (v *parser) parseType() Type {
	if !(v.peek(0).Type == lexer.TOKEN_IDENTIFIER || v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") || v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(")) {
		panic("expected type")
	}

	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "^") {
		v.consumeToken()
		if innerType := v.parseType(); innerType != nil {
			return pointerTo(innerType)
		} else {
			v.err("Standalone `^` is not a valid type")
		}
	}

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.consumeToken()

		innerType := v.parseType()
		if innerType == nil {
			v.err("Expected at least one type in tuple")
		}

		var members []Type
		for innerType != nil {
			members = append(members, innerType)

			if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
				break
			}

			v.consumeToken()
			innerType = v.parseType()
			if innerType == nil {
				// TODO: handle this more gracefully
				panic("got a type that's not in scope?")
			}
		}

		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
			v.err("Expected closing parens")
		}
		v.consumeToken()

		return tupleOf(members...)
	}

	return &UnresolvedType{Name: v.consumeToken().Contents}
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

	scopeToInsertTo := v.scope
	for _, attr := range function.Attrs {
		switch attr.Key {
		case "c":
			if mod, ok := v.modules["C"]; ok {
				scopeToInsertTo = mod.GlobalScope
			} else {
				v.err("Could not find C module to insert C binding into")
			}
		}
	}

	if vname := scopeToInsertTo.InsertFunction(function); vname != nil {
		v.err("Illegal redeclaration of function `%s`", function.Name)
	}

	funcDecl := &FunctionDecl{
		Function: function,
		docs:     v.fetchDocComments(),
	}

	// Arguments
	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.err("Expected `(` after function identifier, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	v.pushScope()

	function.Parameters = make([]*VariableDecl, 0)
	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.consumeToken()
	} else {
		for {

			// either I'm just really sleep deprived,
			// or this is the best way to do this?
			//
			// this parses the ellipse for variable
			// function arguments, it will check for
			// a . and then consume the other 2, it's
			// kind of a weird way to do this and it
			// should probably be lexed as an entire token
			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ".") {
				weird_counter := 0
				for i := 0; i < 2; i++ {
					if v.tokenMatches(i, lexer.TOKEN_SEPARATOR, ".") {
						v.consumeToken()
						weird_counter = weird_counter + i + 1
					}
				}
				if weird_counter == 3 {
					v.consumeToken() // last .
					function.IsVariadic = true
				}
			} else if decl := v.parseVariableDecl(false); decl != nil {
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

		if function.ReturnType = v.parseType(); function.ReturnType == nil {
			v.err("Invalid function return type after colon for function `%s`", function.Name)
		}
	}

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.consumeToken()
		funcDecl.Prototype = true
	}

	// block
	if block := v.parseBlock(false); block != nil {
		if funcDecl.Prototype {
			v.err("Function prototype cannot have a block")
		}
		funcDecl.Function.Body = block
	} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "->") && !funcDecl.Prototype {
		v.consumeToken()

		v.pushScope()
		funcDecl.Function.Body = &Block{scope: v.scope}
		if stat := v.parseStat(); stat != nil {
			funcDecl.Function.Body.appendNode(stat)
		} else {
			// messy...
			// parses the expression appends the node to
			// a fake "body", then checks for a semi colon
			if expr := v.parseExpr(); expr != nil {
				funcDecl.Function.Body.appendNode(&ReturnStat{Value: expr})
				if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
					v.consumeToken()
				} else {
					v.err("Expected semi-colon at the end of single line function `%s`", funcDecl.Function.Name)
				}
			} else {
				v.err("Single line function `%s` expects a statement or an expression, found `%s`", funcDecl.Function.Name, v.peek(0).Contents)
			}

			v.popScope()
		}
	} else if !funcDecl.Prototype {
		v.err("Expecting block or semi-colon (prototype) after function signature")
	}

	v.popScope()

	return funcDecl
}

func (v *parser) parseBlock(pushNewScope bool) *Block {
	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		return nil
	}

	v.consumeToken()

	if pushNewScope {
		v.pushScope()
	}

	block := &Block{scope: v.scope}
	attrs := v.fetchAttrs()
	for _, attr := range attrs {
		attr.FromBlock = true
	}

	for {
		for v.tokenMatches(0, lexer.TOKEN_DOCCOMMENT, "") {
			v.consumeToken() // TODO error here?
		}
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			v.consumeToken()
			break
		}

		// this is a little dirty, but allows for attribute block without reflection (part 2 / 2)
		v.attrs = attrs
		if s := v.parseNode(); s != nil {
			block.appendNode(s)
		} else {
			v.err("Expected statement, found something else")
		}
	}

	if pushNewScope {
		v.popScope()
	}
	return block
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

	// same here, you wouldn't have to do this
	// if every statement parsed a ";" at the end??
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

		body := v.parseBlock(true)
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
				body := v.parseBlock(true)
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

func (v *parser) parseMatchStat() *MatchStat {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MATCH) {
		return nil
	}
	v.consumeToken()

	match := newMatch()

	if target := v.parseExpr(); target != nil {
		match.Target = target
	} else {
		v.err("Expected match target, found `%s`", v.peek(0).Contents)
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.err("Expected body after match target, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	var parsedDefaultMatch bool
	for {
		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
			v.consumeToken()
			break
		}

		// a "pattern" is currently either the default branch "_" or an expression
		var pattern Expr
		if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "_") {
			v.consumeToken()
			if parsedDefaultMatch {
				v.err("Duplicate \"default\" branch in match statement")
			}
			pattern = &DefaultMatchBranch{}
			parsedDefaultMatch = true
		} else if pattern = v.parseExpr(); pattern == nil {
			v.err("Expected pattern in match body, found `%s`", v.peek(0).Contents)
		}

		if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "->") {
			v.err("Expected \"->\" between pattern and statement "+
				"in match body, found `%s`", v.peek(0).Contents)
		}
		v.consumeToken() // consume "->"

		v.pushScope()
		var stmt Stat
		if stmt = v.parseStat(); stmt == nil {
			v.err("Expected statement or block in match body, "+
				"found `%s`", v.peek(0).Contents)
		}
		v.popScope()

		match.Branches[pattern] = stmt
	}

	return match
}

func (v *parser) parseLoopStat() *LoopStat {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FOR) {
		return nil
	}
	v.consumeToken()

	loop := &LoopStat{}

	// matches for {} which is
	// an infinite loop
	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") { // infinite loop
		loop.LoopType = LOOP_TYPE_INFINITE

		loop.Body = v.parseBlock(true)
		if loop.Body == nil {
			v.err("Malformed infinite loop body")
		}
		return loop
	}

	// matches for condition {}
	// basically a white loop
	// or a "conditional for loop"
	if cond := v.parseExpr(); cond != nil {
		loop.LoopType = LOOP_TYPE_CONDITIONAL
		loop.Condition = cond

		loop.Body = v.parseBlock(true)
		if loop.Body == nil {
			v.err("Malformed infinite loop body")
		}
		return loop
	}

	v.err("Malformed `%s` loop", KEYWORD_FOR)
	return nil
}

func (v *parser) parseModuleDecl() *ModuleDecl {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MODULE) {
		return nil
	}
	module := &Module{}

	v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
		v.err("Expected identifier after `module` keyword, found `%s`", v.peek(0).Contents)
	}

	module.Name = v.consumeToken().Contents
	if isReservedKeyword(module.Name) {
		v.err("Cannot name module reserved keyword `%s`", module.Name)
	}

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.consumeToken()

		for {
			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
				v.consumeToken()
				break
			}

			// maybe decl for this instead?
			// also refactor how it's stored in the
			// module and just store Decl?
			// idk might be cleaner
			if function := v.parseFunctionDecl(); function != nil {
				module.Functions = append(module.Functions, function)
			} else if variable := v.parseVariableDecl(true); variable != nil {
				module.Variables = append(module.Variables, variable)
			} else {
				v.err("invalid item in module `%s`", module.Name)
			}
		}
	}

	return &ModuleDecl{Module: module}
}

func (v *parser) parseStructDecl() *StructDecl {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_STRUCT) {
		return nil
	}
	structure := &StructType{}

	v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
		v.err("Expected identifier after `struct` keyword, found `%s`", v.peek(0).Contents)
	}
	structure.Name = v.consumeToken().Contents

	if isReservedKeyword(structure.Name) {
		v.err("Cannot name struct reserved keyword `%s`", structure.Name)
	}

	if sname := v.scope.InsertType(structure); sname != nil {
		v.err("Illegal redeclaration of type `%s`", structure.Name)
	}

	structure.attrs = v.fetchAttrs()

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ";") {
		v.consumeToken()
	} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.consumeToken()

		v.pushScope()

		var itemCount = 0
		for {
			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
				v.consumeToken()
				break
			}

			locationToken := v.peek(0)
			filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber

			if variable := v.parseVariableDecl(false); variable != nil {
				if variable.Variable.Mutable {
					v.err("Cannot specify `mut` keyword on struct member: `%s`", variable.Variable.Name)
				}
				structure.addVariableDecl(variable)
				variable.setPos(filename, line, char)
				itemCount++
			} else {
				v.err("Invalid structure item in structure `%s`", structure.Name)
			}

			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
				v.consumeToken()
			}
		}

		v.popScope()
	} else {
		v.err("Expected block or semi-colon after structure declaration, found `%s`", v.peek(0).Contents)
	}

	return &StructDecl{Struct: structure}
}

func (v *parser) parseTraitDecl() *TraitDecl {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_TRAIT) {
		return nil
	}
	trait := &TraitType{}

	v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
		v.err("Expected identifier after `trait` keyword, found `%s`", v.peek(0).Contents)
	}
	trait.Name = v.consumeToken().Contents

	if isReservedKeyword(trait.Name) {
		v.err("Cannot name trait reserved keyword `%s`", trait.Name)
	}
	if tname := v.scope.InsertType(trait); tname != nil {
		v.err("Illegal redeclaration of type `%s`", trait.Name)
	}

	trait.attrs = v.fetchAttrs()

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.consumeToken()

		v.pushScope()

		for {
			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
				v.consumeToken()
				break
			}

			locationToken := v.peek(0)
			filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber

			if fn := v.parseFunctionDecl(); fn != nil {
				trait.addFunctionDecl(fn)
				fn.setPos(filename, line, char)
			} else {
				v.err("Invalid function declaration in trait `%s`", trait.Name)
			}
		}

		v.popScope()
	} else {
		v.err("Expected body after trait identifier, found `%s`", v.peek(0).Contents)
	}

	return &TraitDecl{Trait: trait}
}

func (v *parser) parseEnumDecl() *EnumDecl {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_ENUM) {
		return nil
	}

	v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
		v.err("Expected identifier after `enum` keyword, found `%s`", v.peek(0).Contents)
	}

	enum := &EnumDecl{Name: v.consumeToken().Contents, Body: make([]*EnumVal, 0)}

	if isReservedKeyword(enum.Name) {
		v.err("Cannot define `enum` for reserved keyword `%s`", enum.Name)
	}

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.consumeToken()
		for {
			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
				v.consumeToken()
				break
			}
			if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
				name := v.consumeToken().Contents
				var value Expr

				if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
					v.consumeToken()

					if expr := v.parseExpr(); expr != nil {
						value = expr
					} else {
						v.err("expecting expression in enum item")
					}
				}

				if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
					v.consumeToken()
				} else if !v.tokenMatches(1, lexer.TOKEN_SEPARATOR, "}") {
					v.err("Missing comma in `enum` %s", enum.Name)
				}

				enum.Body = append(enum.Body, &EnumVal{Name: name, Value: value})
			}
		}

	}
	return enum
}

func (v *parser) parseImplDecl() *ImplDecl {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_IMPL) {
		return nil
	}

	v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
		v.err("Expected identifier after `impl` keyword, found `%s`", v.peek(0).Contents)
	}

	impl := &ImplDecl{}
	impl.StructName = v.consumeToken().Contents

	if isReservedKeyword(impl.StructName) {
		v.err("Cannot define `impl` for reserved keyword `%s`", impl.StructName)
	}

	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FOR) {
		v.consumeToken()

		if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
			v.err("Expected identifier after `for` keyword, found `%s`", v.peek(0).Contents)
		}
		impl.TraitName = v.consumeToken().Contents
		if isReservedKeyword(impl.TraitName) {
			v.err("Cannot define `impl` for reserved keyword `%s`", impl.TraitName)
		}

	}

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "{") {
		v.consumeToken()

		v.pushScope()

		for {
			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "}") {
				v.consumeToken()
				break
			}

			locationToken := v.peek(0)
			filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber

			if fn := v.parseFunctionDecl(); fn != nil {
				impl.Functions = append(impl.Functions, fn)
				fn.setPos(filename, line, char)
			} else {
				v.err("Invalid function in `impl`")
			}
		}

		v.popScope()
	} else {
		v.err("Expected body after `impl` identifier, found `%s`", v.peek(0).Contents)
	}

	return impl
}

func (v *parser) parseVariableDecl(needSemicolon bool) *VariableDecl {
	variable := &Variable{}
	varDecl := &VariableDecl{Variable: variable}

	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_MUT) {
		variable.Mutable = true
		v.consumeToken()
	}

	if !v.tokensMatch(lexer.TOKEN_IDENTIFIER, "", lexer.TOKEN_OPERATOR, ":") {
		return nil
	}
	variable.Name = v.consumeToken().Contents // consume name
	variable.Attrs = v.fetchAttrs()

	if isReservedKeyword(variable.Name) {
		v.err("Cannot name variable reserved keyword `%s`", variable.Name)
	}

	v.consumeToken() // consume :

	if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
		// type is inferred
		variable.Type = nil
	} else {
		variable.Type = v.parseType()
		if variable.Type == nil {
			v.err("Could not decipher type for variable `%s`", variable.Name)
		}
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

	// might be better to just have every statement need a ";"
	// so you would do this after parsing the statement instead?
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

	locationToken := v.peek(0)
	filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber

	if pri == nil {
		return nil
	}
	pri.setPos(filename, line, char)

	if bin := v.parseBinaryOperator(0, pri); bin != nil {
		bin.setPos(filename, line, char)
		return bin
	}
	return pri
}

func (v *parser) parseBinaryOperator(upperPrecedence int, lhand Expr) Expr {
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
			panic("yep")
		}

		v.consumeToken()

		locationToken := v.peek(0)
		filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber
		rhand := v.parsePrimaryExpr()
		rhand.setPos(filename, line, char)

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
	if sizeofExpr := v.parseSizeofExpr(); sizeofExpr != nil {
		return sizeofExpr
	} else if addressOfExpr := v.parseAddressOfExpr(); addressOfExpr != nil {
		return addressOfExpr
	} else if litExpr := v.parseLiteral(); litExpr != nil {
		return litExpr
	} else if castExpr := v.parseCastExpr(); castExpr != nil {
		return castExpr
	} else if derefExpr := v.parseDerefExpr(); derefExpr != nil {
		return derefExpr
	} else if unaryExpr := v.parseUnaryExpr(); unaryExpr != nil {
		return unaryExpr
	} else if callExpr := v.parseCallExpr(); callExpr != nil {
		return callExpr
	} else if accessExpr := v.parseAccessExpr(); accessExpr != nil {
		return accessExpr
	}

	return nil
}

func (v *parser) parseSizeofExpr() *SizeofExpr {
	if !v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_SIZEOF) {
		return nil
	}
	v.consumeToken()

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		v.err("Expected `()` in sizeof expression, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	sizeof := &SizeofExpr{}

	// TODO types
	/*if castExpr.typeName = v.parseTypeToString(); castExpr.typeName != "" {
		castExpr.Type = getTypeFromString(v.scope, castExpr.typeName)
		if castExpr.Type == nil {
			v.err("nope")
		}

	}*/

	if expr := v.parseExpr(); expr != nil {
		sizeof.Expr = expr
	} else {
		v.err("Expected expression in sizeof expression, found `%s`", v.peek(0).Contents)
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Expected `)` in sizeof expression, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	return sizeof
}

func (v *parser) parseTupleLiteral() Expr {
	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}
	v.consumeToken()

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Empty tuple literal")
	}

	tupleLit := &TupleLiteral{}

	for {
		if expr := v.parseExpr(); expr != nil {
			tupleLit.Members = append(tupleLit.Members, expr)

			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
				v.consumeToken()
				break
			} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
				v.consumeToken()
			} else {
				v.err("Expected `,` after tuple literal member, found `%s`", v.peek(0).Contents)
			}
		} else {
			v.err("Expected expression in tuple literal, found `%s`", v.peek(0).Contents)
		}
	}

	if len(tupleLit.Members) == 1 {
		return tupleLit.Members[0]
	} else {
		return tupleLit
	}
}

func (v *parser) parseAccessExpr() *AccessExpr {
	access := &AccessExpr{}

	if _, numNameToks := v.peekName(); numNameToks > 0 {

	} else {
		return nil
	}

	for {
		if name, numNameToks := v.peekName(); numNameToks > 0 {
			v.consumeTokens(numNameToks)

			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ".") {
				// struct access
				v.consumeToken()
				access.Accesses = append(access.Accesses, &Access{AccessType: ACCESS_STRUCT, variableName: name})
			} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
				// array access
				v.consumeToken()

				subscript := v.parseExpr()
				if subscript == nil {
					v.err("Expected expression for array subscript, found `%s`", v.peek(0).Contents)
				}

				if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
					v.err("Expected `]` after array subscript, found `%s`", v.peek(0).Contents)
				}
				v.consumeToken()

				access.Accesses = append(access.Accesses, &Access{AccessType: ACCESS_ARRAY, variableName: name, Subscript: subscript})

				if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ".") {
					return access
				}
			} else if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "|") {
				// tuple access
				v.consumeToken()

				index := v.parseNumericLiteral()
				if index == nil {
					v.err("Expected integer for tuple index, found `%s`", v.peek(0).Contents)
				}

				indexLit, ok := index.(*IntegerLiteral)
				if !ok {
					v.err("Expected integer for tuple index, found `%s`", index)
				}

				if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "|") {
					v.err("Expected `|` after tuple index, found `%s`", v.peek(0).Contents)
				}
				v.consumeToken()

				access.Accesses = append(access.Accesses, &Access{AccessType: ACCESS_TUPLE, variableName: name, Index: indexLit.Value})

				if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ".") {
					return access
				}

			} else {
				access.Accesses = append(access.Accesses, &Access{AccessType: ACCESS_VARIABLE, variableName: name})
				return access
			}
		} else {
			panic("shit")
		}
	}
}

func (v *parser) parseCallExpr() *CallExpr {
	callExpr := &CallExpr{}

	numNameToks := 0
	if callExpr.functionName, numNameToks = v.peekName(); v.tokenMatches(numNameToks, lexer.TOKEN_SEPARATOR, "(") && numNameToks > 0 {
		v.consumeTokens(numNameToks)
	} else {
		return nil
	}

	v.consumeToken() // consume (

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.consumeToken()
	} else {
		for {
			if expr := v.parseExpr(); expr != nil {
				callExpr.Arguments = append(callExpr.Arguments, expr)
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

	return callExpr
}

// TODO: move cast exprs into CallExpr
func (v *parser) parseCastExpr() *CastExpr {
	if !v.tokensMatch(lexer.TOKEN_IDENTIFIER, KEYWORD_CAST, lexer.TOKEN_SEPARATOR, "(") {
		return nil
	}
	v.consumeToken()
	v.consumeToken()

	castExpr := &CastExpr{}

	if castExpr.Type = v.parseType(); castExpr.Type == nil {
		v.err("Invalid type in cast expr")
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
		v.err("Expected `,`, found `%s`", v.peek(0).Contents)
	}
	v.consumeToken()

	castExpr.Expr = v.parseExpr()
	if castExpr.Expr == nil {
		v.err("Expected expression in typecast, found `%s`", v.peek(0))
	}

	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ")") {
		v.err("Exprected `)` at the end of typecast, found `%s`", v.peek(0))
	}
	v.consumeToken()

	return castExpr
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

func (v *parser) parseAddressOfExpr() *AddressOfExpr {
	if !v.tokenMatches(0, lexer.TOKEN_OPERATOR, "&") {
		return nil
	}
	v.consumeToken()

	locationToken := v.peek(0)
	filename, line, char := locationToken.Filename, locationToken.LineNumber, locationToken.CharNumber

	access := v.parseAccessExpr()
	if access == nil {
		v.err("Expected variable access after `&` operator, found `%s`", v.peek(0).Contents)
	}
	access.setPos(filename, line, char)

	return &AddressOfExpr{
		Access: access,
	}
}

func (v *parser) parseLiteral() Expr {
	if arrayLit := v.parseArrayLiteral(); arrayLit != nil {
		return arrayLit
	} else if tupleLit := v.parseTupleLiteral(); tupleLit != nil {
		return tupleLit
	} else if boolLit := v.parseBoolLiteral(); boolLit != nil {
		return boolLit
	} else if numLit := v.parseNumericLiteral(); numLit != nil {
		return numLit
	} else if stringLit := v.parseStringLiteral(); stringLit != nil {
		return stringLit
	} else if runeLit := v.parseRuneLiteral(); runeLit != nil {
		return runeLit
	}
	return nil
}

func (v *parser) parseArrayLiteral() *ArrayLiteral {
	if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		return nil
	}
	v.consumeToken()

	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
		v.err("Empty array literal")
	}

	arrayLit := &ArrayLiteral{}

	for {
		if expr := v.parseExpr(); expr != nil {
			arrayLit.Members = append(arrayLit.Members, expr)

			if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
				v.consumeToken()
				return arrayLit
			} else if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
				v.consumeToken()
			} else {
				v.err("Expected `,` after array literal member, found `%s`", v.peek(0).Contents)
			}
		} else {
			v.err("Expected expression in array literal, found `%s`", v.peek(0).Contents)
		}
	}
}

func (v *parser) parseBoolLiteral() *BoolLiteral {
	if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_TRUE) {
		v.consumeToken()
		return &BoolLiteral{Value: true}
	} else if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FALSE) {
		v.consumeToken()
		return &BoolLiteral{Value: false}
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
	} else if lastRune := unicode.ToLower([]rune(num)[len([]rune(num))-1]); strings.ContainsRune(num, '.') || lastRune == 'f' || lastRune == 'd' || lastRune == 'q' {
		if strings.Count(num, ".") > 1 {
			v.err("Floating-point cannot have multiple periods: `%s`", num)
			return nil
		}

		f := &FloatingLiteral{}

		fnum := num
		hasSuffix := true

		switch lastRune {
		case 'f':
			f.Type = PRIMITIVE_f32
		case 'd':
			f.Type = PRIMITIVE_f64
		case 'q':
			f.Type = PRIMITIVE_f128
		default:
			hasSuffix = false
		}

		if hasSuffix {
			fnum = fnum[:len(fnum)-1]
		}

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
