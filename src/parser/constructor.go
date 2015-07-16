package parser

import (
	"fmt"
	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
	"os"
	"reflect"
	"strings"
	"time"
)

type ConstructableNode interface {
	construct(*Constructor) Node
}

type ConstructableType interface {
	construct(*Constructor) Type
}

type ConstructableExpr interface {
	construct(*Constructor) Expr
}

type Constructor struct {
	tree    *ParseTree
	module  *Module
	modules map[string]*Module
	scope   *Scope
}

func (v *Constructor) err(pos lexer.Span, err string, stuff ...interface{}) {
	v.errPos(pos.Start(), err, stuff...)
}

func (v *Constructor) errPos(pos lexer.Position, err string, stuff ...interface{}) {
	log.Errorln("constructor", util.TEXT_RED+util.TEXT_BOLD+"Constructor error:"+util.TEXT_RESET)

	line := v.tree.Source.GetLine(pos.Line)
	pad := pos.Char - 1

	log.Errorln("constructor", strings.Replace(line, "%", "%%", -1))
	log.Error("constructor", strings.Repeat(" ", pad))
	log.Errorln("constructor", "^")

	log.Errorln("constructor", "[%s:%d:%d] %s",
		pos.Filename, pos.Line, pos.Char,
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_CONSTRUCTOR)
}

func (v *Constructor) errSpan(pos lexer.Span, err string, stuff ...interface{}) {
	log.Errorln("constructor", util.TEXT_RED+util.TEXT_BOLD+"Constructor error:"+util.TEXT_RESET)

	for line := pos.Start().Line; line <= pos.End().Line; line++ {
		lineString := v.tree.Source.GetLine(line)

		var pad int
		if line == pos.Start().Line {
			pad = pos.Start().Char - 1
		} else {
			pad = 0
		}

		var length int
		if line == pos.End().Line {
			length = pos.End().Char
		} else {
			length = len([]rune(lineString))
		}

		log.Errorln("constructor", strings.Replace(lineString, "%", "%%", -1))
		log.Error("constructor", strings.Repeat(" ", pad))
		log.Errorln("constructor", strings.Repeat("^", length))
	}

	log.Errorln("constructor", "[%s:%d:%d] %s",
		pos.Filename, pos.StartLine, pos.StartChar,
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_CONSTRUCTOR)
}

func (v *Constructor) pushScope() {
	v.scope = newScope(v.scope)
}

func (v *Constructor) popScope() {
	v.scope = v.scope.Outer
	if v.scope == nil {
		panic("popped too many scopes")
	}
}

func (v *Constructor) useModule(name string) {
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

func Construct(tree *ParseTree, modules map[string]*Module) *Module {
	c := &Constructor{
		tree: tree,
		module: &Module{
			Nodes: make([]Node, 0),
			Path:  tree.Source.Path,
			Name:  tree.Source.Name,
		},
		scope: NewGlobalScope(),
	}
	c.module.GlobalScope = c.scope
	c.modules = modules
	modules[tree.Source.Name] = c.module

	// add a C module here which will contain
	// all of the c bindings and what not to
	// keep everything separate
	cModule := &Module{
		Nodes:       make([]Node, 0),
		Path:        "", // not really a path for this module
		Name:        "C",
		GlobalScope: NewGlobalScope(),
	}
	c.module.GlobalScope.UsedModules["C"] = cModule

	log.Verboseln("constructor", util.TEXT_BOLD+util.TEXT_GREEN+"Started constructing "+util.TEXT_RESET+tree.Source.Name)
	t := time.Now()

	c.construct()

	dur := time.Since(t)
	log.Verbose("constructor", util.TEXT_BOLD+util.TEXT_GREEN+"Finished parsing"+util.TEXT_RESET+" %s (%.2fms)\n",
		tree.Source.Name, float32(dur.Nanoseconds())/1000000)

	return c.module
}

func (v *Constructor) construct() {
	for _, node := range v.tree.Nodes {
		v.module.Nodes = append(v.module.Nodes, v.constructNode(node))
	}
}

func (v *Constructor) constructNode(node ParseNode) Node {
	switch node.(type) {
	case ConstructableNode:
		return node.(ConstructableNode).construct(v)

	default:
		println(reflect.TypeOf(node).String())
		panic("Encountered un-constructable node")
	}
}

func (v *Constructor) constructType(node ParseNode) Type {
	switch node.(type) {
	case ConstructableType:
		return node.(ConstructableType).construct(v)

	default:
		println(reflect.TypeOf(node).String())
		panic("Encountered un-constructable node")
	}
}

func (v *Constructor) constructExpr(node ParseNode) Expr {
	switch node.(type) {
	case ConstructableExpr:
		return node.(ConstructableExpr).construct(v)

	default:
		println(reflect.TypeOf(node).String())
		panic("Encountered un-constructable node")
	}
}

func (v *PointerTypeNode) construct(c *Constructor) Type {
	targetType := c.constructType(v.TargetType)
	return pointerTo(targetType)
}

func (v *TupleTypeNode) construct(c *Constructor) Type {
	res := &TupleType{}
	for _, member := range v.MemberTypes {
		res.Members = append(res.Members, c.constructType(member))
	}
	return res
}

func (v *ArrayTypeNode) construct(c *Constructor) Type {
	memberType := c.constructType(v.MemberType)
	return arrayOf(memberType)
}

func (v *TypeReferenceNode) construct(c *Constructor) Type {
	res := &UnresolvedType{}
	res.Name.name = v.Reference.Name.Value
	for _, module := range v.Reference.Modules {
		res.Name.moduleNames = append(res.Name.moduleNames, module.Value)
	}
	return res
}

func (v *StructDeclNode) construct(c *Constructor) Node {
	structType := &StructType{}
	structType.attrs = v.Attrs()
	structType.Name = v.Name.Value
	c.pushScope()
	for _, member := range v.Members {
		structType.addVariableDecl(c.constructNode(member).(*VariableDecl)) // TODO: Error message
	}
	c.popScope()

	if c.scope.InsertType(structType) != nil {
		c.err(v.Where(), "Illegal redeclaration of structure `%s`", structType.Name)
	}

	res := &StructDecl{}
	res.Struct = structType
	res.setPos(v.Where().Start())
	return res
}

func (v *UseDeclNode) construct(c *Constructor) Node {
	res := &UseDecl{}
	res.ModuleName = v.Module.Name.Value
	res.Scope = c.scope
	c.useModule(res.ModuleName)
	res.setPos(v.Where().Start())
	return res
}

func (v *TraitDeclNode) construct(c *Constructor) Node {
	trait := &TraitType{}
	trait.attrs = v.Attrs()
	trait.Name = v.Name.Value
	c.pushScope()
	for _, member := range v.Members {
		trait.addFunctionDecl(c.constructNode(member).(*FunctionDecl)) // TODO: Error message
	}
	c.popScope()

	if c.scope.InsertType(trait) != nil {
		c.err(v.Where(), "Illegal redeclaration of trait `%s`", trait.Name)
	}

	res := &TraitDecl{}
	res.Trait = trait
	res.setPos(v.Where().Start())
	return res
}

func (v *ImplDeclNode) construct(c *Constructor) Node {
	res := &ImplDecl{}
	res.StructName = v.StructName.Value
	res.TraitName = v.TraitName.Value
	c.pushScope()
	for _, member := range v.Members {
		res.Functions = append(res.Functions, c.constructNode(member).(*FunctionDecl)) // TODO: Error message
	}
	c.popScope()
	res.setPos(v.Where().Start())
	return res
}

func (v *ModuleDeclNode) construct(c *Constructor) Node {
	mod := &Module{}
	mod.Name = v.Name.Value

	for _, member := range v.Members {
		thing := c.constructNode(member)

		// Comment from ages past:
		// maybe decl for this instead?
		// also refactor how it's stored in the
		// module and just store Decl?
		// idk might be cleaner
		switch thing.(type) {
		case *FunctionDecl:
			mod.Functions = append(mod.Functions, thing.(*FunctionDecl))

		case *VariableDecl:
			mod.Variables = append(mod.Variables, thing.(*VariableDecl))

		default:
			panic("invalid item in module `" + mod.Name + "`")
		}

	}

	res := &ModuleDecl{Module: mod}
	res.docs = v.DocComments()
	res.setPos(v.Where().Start())
	return res
}

func (v *FunctionDeclNode) construct(c *Constructor) Node {
	function := &Function{}
	function.Name = v.Header.Name.Value
	function.Attrs = v.Attrs()
	function.IsVariadic = v.Header.Variadic
	c.pushScope()
	for _, arg := range v.Header.Arguments {
		function.Parameters = append(function.Parameters, c.constructNode(arg).(*VariableDecl)) // TODO: Error message
	}

	if v.Header.ReturnType != nil {
		function.ReturnType = c.constructType(v.Header.ReturnType)
	}

	res := &FunctionDecl{}
	res.docs = v.DocComments()
	res.Function = function
	if v.Body != nil {
		c.pushScope()
		function.Body = c.constructNode(v.Body).(*Block) // TODO: Error message
		c.popScope()
	} else {
		res.Prototype = true
	}
	c.popScope()

	scopeToInsertTo := c.scope
	if function.Attrs.Contains("c") {
		if mod, ok := c.module.GlobalScope.UsedModules["C"]; ok {
			scopeToInsertTo = mod.GlobalScope
		} else {
			panic("Could not find C module to insert C binding into")
		}
	}

	if scopeToInsertTo.InsertFunction(function) != nil {
		c.err(v.Where(), "Illegal redeclaration of function `%s`", function.Name)
	}

	res.setPos(v.Where().Start())
	return res
}

func (v *EnumDeclNode) construct(c *Constructor) Node {
	res := &EnumDecl{}
	res.Name = v.Name.Value
	for _, member := range v.Members {
		val := &EnumVal{Name: member.Name.Value}
		if member.Value != nil {
			val.Value = c.constructExpr(member.Value)
		}
		res.Body = append(res.Body, val)
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *VarDeclNode) construct(c *Constructor) Node {
	variable := &Variable{}
	variable.Name = v.Name.Value
	variable.Attrs = v.Attrs()
	variable.Mutable = v.Mutable.Value != ""
	if v.Type != nil {
		variable.Type = c.constructType(v.Type)
	}

	if c.scope.InsertVariable(variable) != nil {
		c.err(v.Where(), "Illegal redeclaration of variable `%s`", variable.Name)
	}

	res := &VariableDecl{}
	res.docs = v.DocComments()
	res.Variable = variable
	if v.Value != nil {
		res.Assignment = c.constructExpr(v.Value)
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *DeferStatNode) construct(c *Constructor) Node {
	res := &DeferStat{}
	res.Call = c.constructExpr(v.Call).(*CallExpr) // TODO: Error message
	res.setPos(v.Where().Start())
	return res
}

func (v *IfStatNode) construct(c *Constructor) Node {
	res := &IfStat{}
	for _, part := range v.Parts {
		res.Exprs = append(res.Exprs, c.constructExpr(part.Condition))       // TODO: Error message
		res.Bodies = append(res.Bodies, c.constructNode(part.Body).(*Block)) // TODO: Error message
	}
	if v.ElseBody != nil {
		res.Else = c.constructNode(v.ElseBody).(*Block) // TODO: Error message
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *MatchStatNode) construct(c *Constructor) Node {
	res := &MatchStat{}
	res.Target = c.constructExpr(v.Value)
	res.Branches = make(map[Expr]Node)
	for _, branch := range v.Cases {
		var pattern Expr
		if dpn, ok := branch.Pattern.(*DefaultPatternNode); ok {
			pattern = &DefaultMatchBranch{}
			pattern.setPos(dpn.Where().Start())
		} else {
			pattern = c.constructExpr(branch.Pattern)
		}

		body := c.constructNode(branch.Body)
		res.Branches[pattern] = body
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *DefaultPatternNode) construct(c *Constructor) Node {
	res := &DefaultMatchBranch{}
	res.setPos(v.Where().Start())
	return res
}

func (v *LoopStatNode) construct(c *Constructor) Node {
	res := &LoopStat{}
	if v.Condition != nil {
		res.LoopType = LOOP_TYPE_CONDITIONAL
		res.Condition = c.constructExpr(v.Condition)
	} else {
		res.LoopType = LOOP_TYPE_INFINITE
	}
	res.Body = c.constructNode(v.Body).(*Block) // TODO: Error message
	res.setPos(v.Where().Start())
	return res
}

func (v *ReturnStatNode) construct(c *Constructor) Node {
	res := &ReturnStat{}
	if v.Value != nil {
		res.Value = c.constructExpr(v.Value)
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *BlockStatNode) construct(c *Constructor) Node {
	res := &BlockStat{}
	res.Block = c.constructNode(v.Body).(*Block) // TODO: Error message
	res.setPos(v.Where().Start())
	return res
}

func (v *BlockNode) construct(c *Constructor) Node {
	res := &Block{}
	res.scope = c.scope
	if !v.NonScoping {
		c.pushScope()
	}
	res.scope = c.scope
	for _, node := range v.Nodes {
		res.Nodes = append(res.Nodes, c.constructNode(node))
	}
	if !v.NonScoping {
		c.popScope()
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *CallStatNode) construct(c *Constructor) Node {
	res := &CallStat{}
	res.Call = c.constructExpr(v.Call).(*CallExpr)
	res.setPos(v.Where().Start())
	return res
}

func (v *AssignStatNode) construct(c *Constructor) Node {
	res := &AssignStat{}
	res.Access = c.constructExpr(v.Target).(AccessExpr) // TODO: Error message
	res.Assignment = c.constructExpr(v.Value)
	res.setPos(v.Where().Start())
	return res
}

func (v *BinaryExprNode) construct(c *Constructor) Expr {
	res := &BinaryExpr{}
	res.Lhand = c.constructExpr(v.Lhand)
	res.Rhand = c.constructExpr(v.Rhand)
	res.Op = v.Operator
	res.setPos(v.Where().Start())
	return res
}

func (v *SizeofExprNode) construct(c *Constructor) Expr {
	res := &SizeofExpr{}
	res.Expr = c.constructExpr(v.Value)
	res.setPos(v.Where().Start())
	return res
}

func (v *AddrofExprNode) construct(c *Constructor) Expr {
	res := &AddressOfExpr{}
	res.Access = c.constructExpr(v.Value)
	res.setPos(v.Where().Start())
	return res
}

func (v *CastExprNode) construct(c *Constructor) Expr {
	res := &CastExpr{}
	res.Type = c.constructType(v.Type)
	res.Expr = c.constructExpr(v.Value)
	res.setPos(v.Where().Start())
	return res
}

func (v *UnaryExprNode) construct(c *Constructor) Expr {
	res := &UnaryExpr{}
	res.Expr = c.constructExpr(v.Value)
	res.Op = v.Operator
	res.setPos(v.Where().Start())
	return res
}

func (v *CallExprNode) construct(c *Constructor) Expr {
	res := &CallExpr{}
	for _, arg := range v.Arguments {
		res.Arguments = append(res.Arguments, c.constructExpr(arg))
	}
	res.functionSource = c.constructExpr(v.Function)
	res.setPos(v.Where().Start())
	return res
}

func (v *VariableAccessNode) construct(c *Constructor) Expr {
	res := &VariableAccessExpr{}
	for _, module := range v.Name.Modules {
		res.Name.moduleNames = append(res.Name.moduleNames, module.Value)
	}
	res.Name.name = v.Name.Name.Value
	res.setPos(v.Where().Start())
	return res
}

func (v *DerefAccessNode) construct(c *Constructor) Expr {
	res := &DerefAccessExpr{}
	res.Expr = c.constructExpr(v.Value)
	res.setPos(v.Where().Start())
	return res
}

func (v *StructAccessNode) construct(c *Constructor) Expr {
	res := &StructAccessExpr{}
	res.Struct = c.constructExpr(v.Struct).(AccessExpr) // TODO: Error message
	res.Member = v.Member.Value
	res.setPos(v.Where().Start())
	return res
}

func (v *ArrayAccessNode) construct(c *Constructor) Expr {
	res := &ArrayAccessExpr{}
	res.Array = c.constructExpr(v.Array).(AccessExpr) // TODO: Error message
	res.Subscript = c.constructExpr(v.Index)
	res.setPos(v.Where().Start())
	return res
}

func (v *TupleAccessNode) construct(c *Constructor) Expr {
	res := &TupleAccessExpr{}
	res.Tuple = c.constructExpr(v.Tuple).(AccessExpr) // TODO: Error message
	res.Index = uint64(v.Index)
	res.setPos(v.Where().Start())
	return res
}

func (v *ArrayLiteralNode) construct(c *Constructor) Expr {
	res := &ArrayLiteral{}
	for _, member := range v.Values {
		res.Members = append(res.Members, c.constructExpr(member))
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *TupleLiteralNode) construct(c *Constructor) Expr {
	res := &TupleLiteral{}
	for _, member := range v.Values {
		res.Members = append(res.Members, c.constructExpr(member))
	}
	if len(res.Members) == 1 {
		return res.Members[0]
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *BoolLitNode) construct(c *Constructor) Expr {
	res := &BoolLiteral{}
	res.Value = v.Value
	res.setPos(v.Where().Start())
	return res
}

func (v *NumberLitNode) construct(c *Constructor) Expr {
	res := &NumericLiteral{}
	res.IsFloat = v.IsFloat
	res.IntValue = v.IntValue
	res.FloatValue = v.FloatValue

	switch v.FloatSize {
	case 'f':
		res.Type = PRIMITIVE_f32
	case 'd':
		res.Type = PRIMITIVE_f64
	case 'q':
		res.Type = PRIMITIVE_f128
	}

	res.setPos(v.Where().Start())
	return res

}

func (v *StringLitNode) construct(c *Constructor) Expr {
	res := &StringLiteral{}
	res.Value = v.Value
	res.setPos(v.Where().Start())
	return res
}

func (v *RuneLitNode) construct(c *Constructor) Expr {
	res := &RuneLiteral{}
	res.Value = v.Value
	res.setPos(v.Where().Start())
	return res
}
