package parser

import (
	"fmt"
	"os"
	"reflect"

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
	modules *ModuleLookup
	module  *Module

	curTree   *ParseTree
	curSubmod *Submodule
}

func (v *Constructor) err(pos lexer.Span, err string, stuff ...interface{}) {
	v.errPos(pos.Start(), err, stuff...)
}

func (v *Constructor) errPos(pos lexer.Position, err string, stuff ...interface{}) {
	log.Errorln("constructor",
		util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
		pos.Filename, pos.Line, pos.Char,
		fmt.Sprintf(err, stuff...))

	log.Error("constructor", v.curTree.Source.MarkPos(pos))

	os.Exit(util.EXIT_FAILURE_CONSTRUCTOR)
}

func (v *Constructor) errSpan(pos lexer.Span, err string, stuff ...interface{}) {
	log.Errorln("constructor",
		util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" [%s:%d:%d] %s",
		pos.Filename, pos.StartLine, pos.StartChar,
		fmt.Sprintf(err, stuff...))

	log.Error("constructor", v.curTree.Source.MarkSpan(pos))

	os.Exit(util.EXIT_FAILURE_CONSTRUCTOR)
}

func Construct(module *Module, modules *ModuleLookup) {
	module.Parts = make(map[string]*Submodule)
	con := &Constructor{
		modules: modules,
		module:  module,
	}

	log.Timed("constructing module", module.Name.String(), func() {
		for _, tree := range con.module.Trees {
			log.Timed("constructing submodule", tree.Source.Name, func() {
				con.constructSubmodule(tree)
			})
		}
	})
}

func (v *Constructor) constructSubmodule(tree *ParseTree) {
	v.curTree = tree
	v.curSubmod = &Submodule{
		Parent: v.module,
		File:   tree.Source,
	}

	for _, node := range v.curTree.Nodes {
		cnode := v.constructNode(node)
		if cnode != nil {
			v.curSubmod.Nodes = append(v.curSubmod.Nodes, cnode)
		}
	}

	v.module.Parts[v.curTree.Source.Name] = v.curSubmod
	v.curSubmod, v.curTree = nil, nil
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

func (v *ReferenceTypeNode) construct(c *Constructor) Type {
	targetType := v.TargetType.construct(c)
	return ReferenceTo(targetType, v.Mutable)
}

func (v *PointerTypeNode) construct(c *Constructor) Type {
	targetType := v.TargetType.construct(c)
	return PointerTo(targetType, v.Mutable)
}

func (v *TupleTypeNode) construct(c *Constructor) Type {
	res := TupleType{}
	res.Members = c.constructTypeReferences(v.MemberTypes)
	return res
}

func (v *FunctionTypeNode) construct(c *Constructor) Type {
	res := FunctionType{
		IsVariadic: v.IsVariadic,
		Parameters: c.constructTypeReferences(v.ParameterTypes),
		attrs:      v.Attrs(),
	}

	if v.ReturnType != nil {
		res.Return = v.ReturnType.construct(c)
	} else {
		res.Return = &TypeReference{BaseType: PRIMITIVE_void}
	}
	return res
}

func (v *ArrayTypeNode) construct(c *Constructor) Type {
	memberType := v.MemberType.construct(c)
	return ArrayOf(memberType, v.IsFixedLength, v.Length)
}

func (v *NamedTypeNode) construct(c *Constructor) Type {
	return UnresolvedType{Name: toUnresolvedName(v.Name)}
}

func (v *TypeReferenceNode) construct(c *Constructor) *TypeReference {
	args := c.constructTypeReferences(v.GenericArguments)
	res := &TypeReference{BaseType: c.constructType(v.Type), GenericArguments: args}
	return res
}

func (v *Constructor) constructTypeReferences(refs []*TypeReferenceNode) []*TypeReference {
	ret := make([]*TypeReference, 0, len(refs))
	for _, ref := range refs {
		ret = append(ret, ref.construct(v))
	}
	return ret
}

func (v *InterfaceTypeNode) construct(c *Constructor) Type {
	interfaceType := InterfaceType{
		attrs: v.Attrs(),
	}

	for _, function := range v.Functions {
		/*funcData := &Function{
			Name:         function.Name.Value,
			ParentModule: c.module,
			Type: FunctionType{
				IsVariadic: function.Variadic,
				attrs:      v.Attrs(),
			},
		}*/
		funcData := (&FunctionNode{Header: function}).construct(c)
		interfaceType = interfaceType.addFunction(funcData)
	}

	return interfaceType
}

func (v *StructTypeNode) construct(c *Constructor) Type {
	structType := StructType{
		attrs:             v.Attrs(),
		GenericParameters: v.GenericSigil.construct(c),
	}

	for _, member := range v.Members {
		structType = structType.addMember(member.Name.Value, member.Type.construct(c))
	}

	return structType
}

func (v *TypeDeclNode) construct(c *Constructor) Node {
	var paramNodes []ParseNode

	if v.GenericSigil != nil {
		paramNodes = make([]ParseNode, len(v.GenericSigil.GenericParameters))
		for i, p := range v.GenericSigil.GenericParameters {
			paramNodes[i] = p
		}
	}

	namedType := &NamedType{
		Name:         v.Name.Value,
		Type:         c.constructType(v.Type),
		ParentModule: c.module,
	}

	res := &TypeDecl{
		NamedType: namedType,
	}

	res.SetPublic(v.IsPublic())
	res.setPos(v.Where().Start())

	return res
}

func (v *LinkDirectiveNode) construct(c *Constructor) Node {
	c.module.LinkedLibraries = append(c.module.LinkedLibraries, v.Library.Value)
	return nil
}

func (v *UseDirectiveNode) construct(c *Constructor) Node {
	res := &UseDirective{}
	res.ModuleName = toUnresolvedName(v.Module)
	res.setPos(v.Where().Start())
	return res
}

/*func (v *TraitDeclNode) construct(c *Constructor) Node {
	trait := &TraitType{
		attrs: v.Attrs(),
		Name:  v.Name.Value,
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
}*/

func (v *FunctionNode) construct(c *Constructor) *Function {
	function := &Function{
		Name:         v.Header.Name.Value,
		ParentModule: c.module,
		Type: FunctionType{
			IsVariadic: v.Header.Variadic,
			//attrs:      v.Attrs(),
		},
	}

	if len(v.attrs) != 0 {
		panic("functionnode shouldn't have attributes")
	}

	if v.Header.Receiver != nil {
		function.Receiver = c.constructNode(v.Header.Receiver).(*VariableDecl) // TODO: error
		if !function.Receiver.Variable.IsReceiver {
			panic("oh no")
		}
		function.Type.Receiver = function.Receiver.Variable.Type.BaseType
	} else if v.Header.StaticReceiverType != nil {
		function.StaticReceiverType = c.constructType(v.Header.StaticReceiverType)
	}

	var arguments []ParseNode
	for _, arg := range v.Header.Arguments { // TODO rename v.Header.Arguments to v.Header.Parameters
		arguments = append(arguments, arg)
		decl := c.constructNode(arg).(*VariableDecl) // TODO: Error message
		decl.Variable.IsParameter = true
		function.Parameters = append(function.Parameters, decl)
		function.Type.Parameters = append(function.Type.Parameters, decl.Variable.Type)
	}

	if v.Header.ReturnType != nil {
		function.Type.Return = v.Header.ReturnType.construct(c)
	} else {
		// set it to void since we haven't specified a type
		function.Type.Return = &TypeReference{BaseType: PRIMITIVE_void}
	}

	if v.Header.GenericSigil != nil {
		function.Type.GenericParameters = v.Header.GenericSigil.construct(c)
	}

	if v.Expr != nil {
		v.Stat = &ReturnStatNode{Value: v.Expr}
	}
	if v.Stat != nil {
		v.Body = &BlockNode{Nodes: []ParseNode{v.Stat}}
	}
	if v.Body != nil {
		function.Body = c.constructNode(v.Body).(*Block) // TODO: Error message
	} else if v.Header.Anonymous {
		c.err(v.Where(), "Lambda cannot be prototype")
	}

	return function
}

func (v *FunctionDeclNode) construct(c *Constructor) Node {
	function := v.Function.construct(c)
	function.Type.attrs = v.Attrs()

	res := &FunctionDecl{
		docs:      v.DocComments(),
		Function:  function,
		Prototype: v.Function.Body == nil,
	}

	res.SetPublic(v.IsPublic())
	res.setPos(v.Where().Start())
	return res
}

func (v *LambdaExprNode) construct(c *Constructor) Expr {
	function := v.Function.construct(c)
	function.Type.attrs = v.Attrs()

	res := &LambdaExpr{
		Function: function,
	}

	res.setPos(v.Where().Start())
	return res
}

func (v *GenericSigilNode) construct(c *Constructor) GenericSigil {
	if v == nil {
		return nil
	}

	ret := make([]*SubstitutionType, 0, len(v.GenericParameters))
	for _, p := range v.GenericParameters {
		ret = append(ret, NewSubstitutionType(p.Name.Value, c.constructTypes(p.Constraints)))
	}

	return ret
}

func (v *EnumTypeNode) construct(c *Constructor) Type {
	enumType := EnumType{
		Simple:            true,
		Members:           make([]EnumTypeMember, len(v.Members)),
		GenericParameters: v.GenericSigil.construct(c),
	}

	lastValue := 0
	for idx, mem := range v.Members {
		enumType.Members[idx].Name = mem.Name.Value

		if mem.TupleBody != nil {
			enumType.Members[idx].Type = mem.TupleBody.construct(c)
			enumType.Simple = false
		} else if mem.StructBody != nil {
			enumType.Members[idx].Type = mem.StructBody.construct(c)
			enumType.Simple = false
		} else {
			enumType.Members[idx].Type = tupleOf()
		}

		if mem.Value != nil {
			// TODO: Check for overflow
			lastValue = int(mem.Value.IntValue.Int64())
		}
		enumType.Members[idx].Tag = lastValue
		lastValue += 1
	}

	// this should probably be somewhere else
	usedNames := make(map[string]bool)
	usedTags := make(map[int]bool)
	for _, mem := range enumType.Members {
		if usedNames[mem.Name] {
			c.err(v.Where(), "Duplicate member name `%s`", mem.Name)
		}
		usedNames[mem.Name] = true

		if usedTags[mem.Tag] {
			c.err(v.Where(), "Duplciate enum tag `%d` on member `%s`", mem.Tag, mem.Name)
		}
		usedTags[mem.Tag] = true
	}

	return enumType
}

func (v *VarDeclNode) construct(c *Constructor) Node {
	if isReservedKeyword(v.Name.Value) {
		c.err(v.Name.Where, "Variable name was reserved keyword `%s`", v.Name.Value)
	}

	variable := &Variable{
		Name:         v.Name.Value,
		Attrs:        v.Attrs(),
		Mutable:      v.Mutable.Value != "",
		ParentModule: c.module,
		IsReceiver:   v.IsReceiver,
	}

	if v.Type != nil {
		variable.Type = v.Type.construct(c)
	}

	res := &VariableDecl{
		docs:     v.DocComments(),
		Variable: variable,
	}

	if v.Value != nil {
		res.Assignment = c.constructExpr(v.Value)
	}

	res.SetPublic(v.IsPublic())
	res.setPos(v.Where().Start())
	return res
}

func (v *DestructVarDeclNode) construct(c *Constructor) Node {
	res := &DestructVarDecl{
		docs:          v.DocComments(),
		Variables:     make([]*Variable, len(v.Names)),
		ShouldDiscard: make([]bool, len(v.Names)),
		Assignment:    c.constructExpr(v.Value),
	}
	res.setPos(v.Where().Start())

	for idx, name := range v.Names {
		mutable := v.Mutable[idx]

		if name.Value == KEYWORD_DISCARD {
			res.ShouldDiscard[idx] = true
		} else {
			res.Variables[idx] = &Variable{
				Name:         name.Value,
				Attrs:        make(AttrGroup),
				Mutable:      mutable,
				ParentModule: c.module,
			}
		}
	}

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
		pattern := c.constructExpr(branch.Pattern)
		body := c.constructNode(branch.Body)
		res.Branches[pattern] = body
	}
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

func (v *BreakStatNode) construct(c *Constructor) Node {
	res := &BreakStat{}
	res.setPos(v.Where().Start())
	return res
}

func (v *NextStatNode) construct(c *Constructor) Node {
	res := &NextStat{}
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
	res.NonScoping = v.NonScoping
	res.Nodes = c.constructNodes(v.Nodes)
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
	var res Node

	acc := c.constructExpr(v.Target)
	if tl, ok := acc.(*TupleLiteral); ok {
		accesses := make([]AccessExpr, len(tl.Members))
		for idx, mem := range tl.Members {
			if ae, ok := mem.(AccessExpr); ok {
				accesses[idx] = ae
			} else {
				c.errPos(mem.Pos(), "Cannot assign to non-access expression")
			}
		}

		res = &DestructAssignStat{
			Accesses:   accesses,
			Assignment: c.constructExpr(v.Value),
		}

	} else if ae, ok := acc.(AccessExpr); ok {
		res = &AssignStat{
			Access:     ae,
			Assignment: c.constructExpr(v.Value),
		}
	} else {
		c.errSpan(v.Target.Where(), "Cannot assign to non-access expression")
	}

	res.setPos(v.Where().Start())
	return res
}

func (v *BinopAssignStatNode) construct(c *Constructor) Node {
	var res Node

	acc := c.constructExpr(v.Target)
	if tl, ok := acc.(*TupleLiteral); ok {
		accesses := make([]AccessExpr, len(tl.Members))
		for idx, mem := range tl.Members {
			if ae, ok := mem.(AccessExpr); ok {
				accesses[idx] = ae
			} else {
				c.errPos(mem.Pos(), "Cannot assign to non-access expression")
			}
		}

		res = &DestructBinopAssignStat{
			Operator:   v.Operator,
			Accesses:   accesses,
			Assignment: c.constructExpr(v.Value),
		}

	} else if ae, ok := acc.(AccessExpr); ok {
		res = &BinopAssignStat{
			Operator:   v.Operator,
			Access:     ae,
			Assignment: c.constructExpr(v.Value),
		}
	} else {
		c.errSpan(v.Target.Where(), "Cannot assign to non-access expression")
	}

	res.setPos(v.Where().Start())
	return res
}

func (v *BinaryExprNode) construct(c *Constructor) Expr {
	res := &BinaryExpr{
		Lhand: c.constructExpr(v.Lhand),
		Rhand: c.constructExpr(v.Rhand),
		Op:    v.Operator,
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *ArrayLenExprNode) construct(c *Constructor) Expr {
	res := &ArrayLenExpr{}
	if v.ArrayExpr != nil {
		res.Expr = c.constructExpr(v.ArrayExpr)
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *SizeofExprNode) construct(c *Constructor) Expr {
	res := &SizeofExpr{}
	if v.Value != nil {
		res.Expr = c.constructExpr(v.Value)
	} else if v.Type != nil {
		res.Type = v.Type.construct(c)
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *AddrofExprNode) construct(c *Constructor) Expr {
	var res Expr
	if v.IsReference {
		res = &ReferenceToExpr{
			IsMutable: v.Mutable,
			Access:    c.constructExpr(v.Value),
		}
	} else {
		res = &PointerToExpr{
			IsMutable: v.Mutable,
			Access:    c.constructExpr(v.Value),
		}
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *CastExprNode) construct(c *Constructor) Expr {
	res := &CastExpr{
		Type: v.Type.construct(c),
		Expr: c.constructExpr(v.Value),
	}
	if ttyp, ok := res.Type.BaseType.(TupleType); ok && len(ttyp.Members) == 1 {
		res.Type = ttyp.Members[0]
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *UnaryExprNode) construct(c *Constructor) Expr {
	var res Expr
	subExpr := c.constructExpr(v.Value)
	if v.Operator == UNOP_DEREF {
		res = &DerefAccessExpr{
			Expr: subExpr,
		}
	} else if numlit, ok := subExpr.(*NumericLiteral); ok && v.Operator == UNOP_NEGATIVE {
		if numlit.IsFloat {
			numlit.FloatValue *= -1
		} else {
			numlit.IntValue.Neg(numlit.IntValue)
		}
		res = numlit
	} else {
		res = &UnaryExpr{
			Expr: subExpr,
			Op:   v.Operator,
		}
	}

	res.setPos(v.Where().Start())
	return res
}

func (v *CallExprNode) construct(c *Constructor) Expr {
	// TODO: when we allow function types, allow all access forms (eg. `thing[0]()``)
	res := &CallExpr{
		Arguments: c.constructExprs(v.Arguments),
		Function:  c.constructExpr(v.Function),
	}

	if sae, ok := v.Function.(*StructAccessNode); ok {
		res.ReceiverAccess = sae.construct(c).(*StructAccessExpr).Struct
	}

	res.setPos(v.Where().Start())
	return res
}

func (v *VariableAccessNode) construct(c *Constructor) Expr {
	res := &VariableAccessExpr{
		Name:             toUnresolvedName(v.Name),
		GenericArguments: c.constructTypeReferences(v.GenericParameters),
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *StructAccessNode) construct(c *Constructor) Expr {
	res := &StructAccessExpr{
		Member: v.Member.Value,
	}
	res.Struct = c.constructExpr(v.Struct).(AccessExpr) // TODO: Error message
	res.setPos(v.Where().Start())
	return res
}

func (v *ArrayAccessNode) construct(c *Constructor) Expr {
	res := &ArrayAccessExpr{
		Subscript: c.constructExpr(v.Index),
	}
	res.Array = c.constructExpr(v.Array).(AccessExpr) // TODO: Error message
	res.setPos(v.Where().Start())
	return res
}

func (v *DiscardAccessNode) construct(c *Constructor) Expr {
	res := &DiscardAccessExpr{}
	res.setPos(v.Where().Start())
	return res
}

func (v *EnumPatternNode) construct(c *Constructor) Expr {
	res := &EnumPatternExpr{
		MemberName: toUnresolvedName(v.MemberName),
		Variables:  make([]*Variable, len(v.Names)),
	}
	for idx, name := range v.Names {
		if name.Value != KEYWORD_DISCARD {
			res.Variables[idx] = &Variable{
				Name:         name.Value,
				ParentModule: c.module,
			}
		}
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *TupleLiteralNode) construct(c *Constructor) Expr {
	res := &TupleLiteral{
		Members: c.constructExprs(v.Values),
	}
	if len(res.Members) == 1 {
		return res.Members[0]
	}
	res.setPos(v.Where().Start())
	return res
}

func (v *CompositeLiteralNode) construct(c *Constructor) Expr {
	res := &CompositeLiteral{}
	if v.Type != nil {
		res.Type = v.Type.construct(c)
	}

	for i, val := range v.Values {
		res.Fields = append(res.Fields, v.Fields[i].Value)
		res.Values = append(res.Values, c.constructExpr(val))
	}

	res.setPos(v.Where().Start())
	return res
}

func (v *BoolLitNode) construct(c *Constructor) Expr {
	res := &BoolLiteral{Value: v.Value}
	res.setPos(v.Where().Start())
	return res
}

func (v *NumberLitNode) construct(c *Constructor) Expr {
	res := &NumericLiteral{
		IsFloat:    v.IsFloat,
		IntValue:   v.IntValue,
		FloatValue: v.FloatValue,
	}

	res.floatSizeHint = v.FloatSize
	res.setPos(v.Where().Start())
	return res

}

func (v *StringLitNode) construct(c *Constructor) Expr {
	res := &StringLiteral{Value: v.Value, IsCString: v.IsCString}
	res.setPos(v.Where().Start())
	return res
}

func (v *RuneLitNode) construct(c *Constructor) Expr {
	res := &RuneLiteral{Value: v.Value}
	res.setPos(v.Where().Start())
	return res
}

func toUnresolvedName(node *NameNode) UnresolvedName {
	res := UnresolvedName{Name: node.Name.Value}
	for _, module := range node.Modules {
		res.ModuleNames = append(res.ModuleNames, module.Value)
	}
	return res
}

func toParentName(node *NameNode) UnresolvedName {
	res := UnresolvedName{}
	for _, moduleName := range node.Modules[:len(node.Modules)-1] {
		res.ModuleNames = append(res.ModuleNames, moduleName.Value)
	}
	res.Name = node.Modules[len(node.Modules)-1].Value

	return res
}
