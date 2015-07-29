package parser

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
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
	tree      *ParseTree
	treeFiles map[string]*ParseTree
	module    *Module
	modules   map[string]*Module
	scope     *Scope
	nameMap   *NameMap
}

func (v *Constructor) err(pos lexer.Span, err string, stuff ...interface{}) {
	v.errPos(pos.Start(), err, stuff...)
}

func (v *Constructor) errPos(pos lexer.Position, err string, stuff ...interface{}) {
	log.Errorln("constructor",
		util.TEXT_RED+util.TEXT_BOLD+"Constructor error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
		pos.Filename, pos.Line, pos.Char,
		fmt.Sprintf(err, stuff...))

	log.Error("constructor", v.tree.Source.MarkPos(pos))

	os.Exit(util.EXIT_FAILURE_CONSTRUCTOR)
}

func (v *Constructor) errSpan(pos lexer.Span, err string, stuff ...interface{}) {
	log.Errorln("constructor",
		util.TEXT_RED+util.TEXT_BOLD+"Constructor error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
		pos.Filename, pos.StartLine, pos.StartChar,
		fmt.Sprintf(err, stuff...))

	log.Error("constructor", v.tree.Source.MarkSpan(pos))

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

func Construct(tree *ParseTree, treeFiles map[string]*ParseTree, modules map[string]*Module) *Module {
	c := &Constructor{
		tree:      tree,
		treeFiles: treeFiles,
		module: &Module{
			Nodes: make([]Node, 0),
			File:  tree.Source,
			Path:  tree.Source.Path,
			Name:  tree.Source.Name,
		},
		scope:   NewGlobalScope(),
		nameMap: MapNames(tree.Nodes, tree, treeFiles, nil),
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
		log.Infoln("constructor", "Type of node: %s", reflect.TypeOf(node))
		panic("Encountered un-constructable node")
	}
}

func (v *Constructor) constructType(node ParseNode) Type {
	switch node.(type) {
	case ConstructableType:
		return node.(ConstructableType).construct(v)

	default:
		log.Infoln("constructor", "Type of node: %s", reflect.TypeOf(node))
		panic("Encountered un-constructable node")
	}
}

func (v *Constructor) constructExpr(node ParseNode) Expr {
	switch node.(type) {
	case ConstructableExpr:
		return node.(ConstructableExpr).construct(v)

	default:
		log.Infoln("constructor", "Type of node: %s", reflect.TypeOf(node))
		panic("Encountered un-constructable node")
	}
}

func (v *Constructor) constructNodes(nodes []ParseNode) []Node {
	var res []Node
	for _, node := range nodes {
		res = append(res, v.constructNode(node))
	}
	return res
}

func (v *Constructor) constructTypes(nodes []ParseNode) []Type {
	var res []Type
	for _, node := range nodes {
		res = append(res, v.constructType(node))
	}
	return res
}

func (v *Constructor) constructExprs(nodes []ParseNode) []Expr {
	var res []Expr
	for _, node := range nodes {
		res = append(res, v.constructExpr(node))
	}
	return res
}

func (v *PointerTypeNode) construct(c *Constructor) Type {
	targetType := c.constructType(v.TargetType)
	return pointerTo(targetType)
}

func (v *TupleTypeNode) construct(c *Constructor) Type {
	res := &TupleType{}
	res.Members = c.constructTypes(v.MemberTypes)
	return res
}

func (v *ArrayTypeNode) construct(c *Constructor) Type {
	memberType := c.constructType(v.MemberType)
	return arrayOf(memberType)
}

func (v *TypeReferenceNode) construct(c *Constructor) Type {
	typ := c.nameMap.TypeOfNameNode(v.Reference)
	if !typ.IsType() {
		c.errSpan(v.Reference.Name.Where, "Name `%s` is not a type", v.Reference.Name.Value)
	}

	res := &UnresolvedType{Name: toUnresolvedName(v.Reference)}
	return res
}

func (v *StructDeclNode) construct(c *Constructor) Node {
	structType := &StructType{
		attrs:        v.Attrs(),
		Name:         v.Name.Value,
		ParentModule: c.module,
	}

	c.pushScope()
	for _, member := range v.Body.Members {
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

func (v *TypeDeclNode) construct(c *Constructor) Node {
	namedType := &NamedType{
		Name: v.Name.Value,
		Type: c.constructType(v.Type),
	}

	if c.scope.InsertType(namedType) != nil {
		c.err(v.Where(), "Illegal redeclaration of type `%s`", namedType.Name)
	}

	res := &TypeDecl{
		NamedType: namedType,
	}

	res.setPos(v.Where().Start())

	return res
}

func (v *UseDeclNode) construct(c *Constructor) Node {
	typ := c.nameMap.TypeOfNameNode(v.Module)
	if typ != NODE_MODULE {
		c.errSpan(v.Module.Name.Where, "Name `%s` is not a module", v.Module.Name)
	}

	res := &UseDecl{}
	res.ModuleName = v.Module.Name.Value
	res.Scope = c.scope
	c.useModule(res.ModuleName)
	res.setPos(v.Where().Start())
	return res
}

func (v *TraitDeclNode) construct(c *Constructor) Node {
	trait := &TraitType{
		attrs:        v.Attrs(),
		Name:         v.Name.Value,
		ParentModule: c.module,
	}

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
		fn := c.constructNode(member).(*FunctionDecl) // TODO: Error message

		res.Functions = append(res.Functions, fn)
	}
	c.popScope()
	res.setPos(v.Where().Start())
	return res
}

func (v *FunctionDeclNode) construct(c *Constructor) Node {
	function := &Function{
		Name:         v.Header.Name.Value,
		Attrs:        v.Attrs(),
		IsVariadic:   v.Header.Variadic,
		ParentModule: c.module,
	}

	res := &FunctionDecl{
		docs:     v.DocComments(),
		Function: function,
	}

	c.pushScope()
	var arguments []ParseNode
	for _, arg := range v.Header.Arguments {
		arguments = append(arguments, arg)
		decl := c.constructNode(arg).(*VariableDecl) // TODO: Error message
		decl.Variable.ParentFunction = res
		function.Parameters = append(function.Parameters, decl)
	}
	c.nameMap = MapNames(arguments, c.tree, c.treeFiles, c.nameMap)

	if v.Header.ReturnType != nil {
		function.ReturnType = c.constructType(v.Header.ReturnType)
	}

	if v.Expr != nil {
		v.Stat = &ReturnStatNode{Value: v.Expr}
	}
	if v.Stat != nil {
		v.Body = &BlockNode{Nodes: []ParseNode{v.Stat}}
	}
	if v.Body != nil {
		c.pushScope()
		function.Body = c.constructNode(v.Body).(*Block) // TODO: Error message
		c.popScope()
	} else {
		res.Prototype = true
	}
	c.nameMap = c.nameMap.parent
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
	enumType := &EnumType{
		Name:         v.Name.Value,
		Simple:       true,
		Members:      make([]EnumTypeMember, len(v.Members)),
		ParentModule: c.module,
	}

	lastValue := 0
	for idx, mem := range v.Members {
		enumType.Members[idx].Name = mem.Name.Value

		if mem.TupleBody != nil {
			enumType.Members[idx].Type = c.constructType(mem.TupleBody)
			enumType.Simple = false
		} else if mem.StructBody != nil {
			structType := &StructType{
				Name:       mem.Name.Value,
				ParentEnum: enumType,
			}

			c.pushScope()
			for _, member := range mem.StructBody.Members {
				structType.addVariableDecl(c.constructNode(member).(*VariableDecl)) // TODO: Error message
			}
			c.popScope()

			enumType.Members[idx].Type = structType
			enumType.Simple = false
		} else {
			enumType.Members[idx].Type = PRIMITIVE_void
		}

		if mem.Value != nil {
			// TODO: Check for overflow
			lastValue = int(mem.Value.IntValue.Int64())
		}
		enumType.Members[idx].Tag = lastValue
		lastValue += 1
	}

	if c.scope.InsertType(enumType) != nil {
		c.err(v.Where(), "Illegal redeclaration of enum `%s`", enumType.Name)
	}

	res := &EnumDecl{}
	res.Enum = enumType
	res.setPos(v.Where().Start())
	return res
}

func (v *VarDeclNode) construct(c *Constructor) Node {
	variable := &Variable{
		Name:         v.Name.Value,
		Attrs:        v.Attrs(),
		Mutable:      v.Mutable.Value != "",
		ParentModule: c.module,
	}

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

func (v *DefaultStatNode) construct(c *Constructor) Node {
	res := &DefaultStat{}
	res.Target = c.constructExpr(v.Target).(AccessExpr)
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
	c.nameMap = MapNames(v.Nodes, c.tree, c.treeFiles, c.nameMap)

	res := &Block{}
	res.scope = c.scope
	res.NonScoping = v.NonScoping
	if !v.NonScoping {
		c.pushScope()
	}
	res.scope = c.scope
	res.Nodes = c.constructNodes(v.Nodes)
	if !v.NonScoping {
		c.popScope()
	}
	res.setPos(v.Where().Start())

	c.nameMap = c.nameMap.parent
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

func (v *BinopAssignStatNode) construct(c *Constructor) Node {
	res := &BinopAssignStat{}
	res.Access = c.constructExpr(v.Target).(AccessExpr) // TODO: Error message
	res.Operator = v.Operator
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
	depth := 0
	var inner ParseNode
	inner = v.Value
	for {
		if derefNode, ok := inner.(*UnaryExprNode); ok && derefNode.Operator == UNOP_DEREF {
			inner = derefNode.Value
			depth++
			continue
		} else if varAccNode, ok := inner.(*VariableAccessNode); ok {
			typ := c.nameMap.TypeOfNameNode(varAccNode.Name)
			if typ.IsType() {
				var newType ParseNode
				newType = &TypeReferenceNode{Reference: varAccNode.Name}
				for i := 0; i < depth; i++ {
					newType = &PointerTypeNode{TargetType: newType}
				}
				v.Type = newType
				v.Value = nil
			}
		}
		break
	}

	res := &SizeofExpr{}
	if v.Value != nil {
		res.Expr = c.constructExpr(v.Value)
	} else if v.Type != nil {
		res.Type = c.constructType(v.Type)
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *DefaultExprNode) construct(c *Constructor) Expr {
	res := &DefaultExpr{}
	res.Type = c.constructType(v.Target)
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
	var res Expr
	if v.Operator == UNOP_DEREF {
		expr := c.constructExpr(v.Value)
		if castExpr, ok := expr.(*CastExpr); ok {
			res = &CastExpr{Type: pointerTo(castExpr.Type), Expr: castExpr.Expr}
		} else {
			res = &DerefAccessExpr{
				Expr: expr,
			}
		}

	} else {
		res = &UnaryExpr{
			Expr: c.constructExpr(v.Value),
			Op:   v.Operator,
		}
	}

	res.setPos(v.Where().Start())
	return res
}

func (v *CallExprNode) construct(c *Constructor) Expr {
	van := v.Function.(*VariableAccessNode) // TODO: better error
	typ := c.nameMap.TypeOfNameNode(van.Name)
	if typ == NODE_FUNCTION {
		res := &CallExpr{}
		res.Arguments = c.constructExprs(v.Arguments)
		res.functionSource = c.constructExpr(v.Function)
		res.setPos(v.Where().Start())
		return res
	} else if typ == NODE_ENUM_MEMBER {
		res := &EnumLiteral{}
		res.Member = van.Name.Name.Value
		res.Type = &UnresolvedType{Name: toParentName(van.Name)}
		res.TupleLiteral = &TupleLiteral{Members: c.constructExprs(v.Arguments)}
		res.setPos(v.Where().Start())
		return res
	} else if typ.IsType() {
		if len(v.Arguments) > 1 {
			c.errSpan(v.Where(), "Cast cannot recieve more that one argument")
		}

		res := &CastExpr{}
		res.Type = c.constructType(&TypeReferenceNode{Reference: van.Name})
		res.Expr = c.constructExpr(v.Arguments[0])
		res.setPos(v.Where().Start())
		return res
	} else {
		log.Debugln("constructor", "`%s` was a `%s`", van.Name.Name.Value, typ)
		c.errSpan(van.Name.Name.Where, "Name `%s` is not a function or a enum member", van.Name.Name.Value)
		return nil
	}
}

func (v *VariableAccessNode) construct(c *Constructor) Expr {
	if c.nameMap.TypeOfNameNode(v.Name) == NODE_ENUM_MEMBER {
		res := &EnumLiteral{}
		res.Member = v.Name.Name.Value
		res.Type = &UnresolvedType{Name: toParentName(v.Name)}
		res.setPos(v.Where().Start())
		return res
	} else {
		res := &VariableAccessExpr{}
		res.Name = toUnresolvedName(v.Name)
		res.setPos(v.Where().Start())
		return res
	}
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
	res.Members = c.constructExprs(v.Values)
	res.setPos(v.Where().Start())
	return res
}

func (v *TupleLiteralNode) construct(c *Constructor) Expr {
	res := &TupleLiteral{}
	res.Members = c.constructExprs(v.Values)
	if len(res.Members) == 1 {
		return res.Members[0]
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *StructLiteralNode) construct(c *Constructor) Expr {
	res := &StructLiteral{}
	if v.Name != nil {
		res.Type = &UnresolvedType{Name: toUnresolvedName(v.Name)}
	}
	res.Values = make(map[string]Expr)
	for idx, member := range v.Members {
		res.Values[member.Value] = c.constructExpr(v.Values[idx])
	}

	if v.Name == nil || c.nameMap.TypeOfNameNode(v.Name) == NODE_STRUCT {
		return res
	} else if typ := c.nameMap.TypeOfNameNode(v.Name); typ == NODE_ENUM_MEMBER {
		enum := &EnumLiteral{}
		enum.Member = v.Name.Name.Value
		enum.Type = &UnresolvedType{Name: toParentName(v.Name)}
		enum.StructLiteral = res
		enum.setPos(v.Where().Start())
		return enum
	} else {
		log.Debugln("constructor", "`%s` was a `%s`", v.Name.Name.Value, typ)
		c.errSpan(v.Name.Name.Where, "Name `%s` is not a struct or a enum member", v.Name.Name.Value)
		return nil
	}
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

func toUnresolvedName(node *NameNode) unresolvedName {
	res := unresolvedName{}
	res.name = node.Name.Value
	for _, module := range node.Modules {
		res.moduleNames = append(res.moduleNames, module.Value)
	}
	return res
}

func toParentName(node *NameNode) unresolvedName {
	res := unresolvedName{}
	for _, moduleName := range node.Modules[:len(node.Modules)-1] {
		res.moduleNames = append(res.moduleNames, moduleName.Value)
	}
	res.name = node.Modules[len(node.Modules)-1].Value

	return res
}
