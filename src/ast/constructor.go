package ast

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/parser"
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

	curTree   *parser.ParseTree
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

func (v *Constructor) constructSubmodule(tree *parser.ParseTree) {
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

func (v *Constructor) constructNode(node parser.ParseNode) Node {
	switch node := node.(type) {
	case *parser.TypeDeclNode:
		return v.constructTypeDeclNode(node)
	case *parser.LinkDirectiveNode:
		return v.constructLinkDirectiveNode(node)
	case *parser.UseDirectiveNode:
		return v.constructUseDirectiveNode(node)
	case *parser.FunctionDeclNode:
		return v.constructFunctionDeclNode(node)
	case *parser.VarDeclNode:
		return v.constructVarDeclNode(node)
	case *parser.DestructVarDeclNode:
		return v.constructDestructVarDeclNode(node)
	case *parser.DeferStatNode:
		return v.constructDeferStatNode(node)
	case *parser.IfStatNode:
		return v.constructIfStatNode(node)
	case *parser.MatchStatNode:
		return v.constructMatchStatNode(node)
	case *parser.LoopStatNode:
		return v.constructLoopStatNode(node)
	case *parser.ReturnStatNode:
		return v.constructReturnStatNode(node)
	case *parser.BreakStatNode:
		return v.constructBreakStatNode(node)
	case *parser.NextStatNode:
		return v.constructNextStatNode(node)
	case *parser.BlockStatNode:
		return v.constructBlockStatNode(node)
	case *parser.BlockNode:
		return v.constructBlockNode(node)
	case *parser.CallStatNode:
		return v.constructCallStatNode(node)
	case *parser.AssignStatNode:
		return v.constructAssignStatNode(node)
	case *parser.BinopAssignStatNode:
		return v.constructBinopAssignStatNode(node)

	default:
		log.Infoln("constructor", "Type of node: %s", reflect.TypeOf(node))
		panic("Encountered un-constructable node")
	}
}

func (v *Constructor) constructType(node parser.ParseNode) Type {
	switch node := node.(type) {
	case *parser.ReferenceTypeNode:
		return v.constructReferenceTypeNode(node)
	case *parser.PointerTypeNode:
		return v.constructPointerTypeNode(node)
	case *parser.TupleTypeNode:
		return v.constructTupleTypeNode(node)
	case *parser.FunctionTypeNode:
		return v.constructFunctionTypeNode(node)
	case *parser.ArrayTypeNode:
		return v.constructArrayTypeNode(node)
	case *parser.NamedTypeNode:
		return v.constructNamedTypeNode(node)
	case *parser.InterfaceTypeNode:
		return v.constructInterfaceTypeNode(node)
	case *parser.StructTypeNode:
		return v.constructStructTypeNode(node)
	case *parser.EnumTypeNode:
		return v.constructEnumTypeNode(node)

	default:
		log.Infoln("constructor", "Type of node: %s", reflect.TypeOf(node))
		panic("Encountered un-constructable node")
	}
}

func (v *Constructor) constructExpr(node parser.ParseNode) Expr {
	switch node := node.(type) {
	case *parser.BinaryExprNode:
		return v.constructBinaryExprNode(node)
	case *parser.ArrayLenExprNode:
		return v.constructArrayLenExprNode(node)
	case *parser.SizeofExprNode:
		return v.constructSizeofExprNode(node)
	case *parser.AddrofExprNode:
		return v.constructAddrofExprNode(node)
	case *parser.CastExprNode:
		return v.constructCastExprNode(node)
	case *parser.UnaryExprNode:
		return v.constructUnaryExprNode(node)
	case *parser.CallExprNode:
		return v.constructCallExprNode(node)
	case *parser.VariableAccessNode:
		return v.constructVariableAccessNode(node)
	case *parser.StructAccessNode:
		return v.constructStructAccessNode(node)
	case *parser.ArrayAccessNode:
		return v.constructArrayAccessNode(node)
	case *parser.DiscardAccessNode:
		return v.constructDiscardAccessNode(node)
	case *parser.EnumPatternNode:
		return v.constructEnumPatternNode(node)
	case *parser.TupleLiteralNode:
		return v.constructTupleLiteralNode(node)
	case *parser.CompositeLiteralNode:
		return v.constructCompositeLiteralNode(node)
	case *parser.BoolLitNode:
		return v.constructBoolLitNode(node)
	case *parser.NumberLitNode:
		return v.constructNumberLitNode(node)
	case *parser.StringLitNode:
		return v.constructStringLitNode(node)
	case *parser.RuneLitNode:
		return v.constructRuneLitNode(node)
	case *parser.LambdaExprNode:
		return v.constructLambdaExprNode(node)

	default:
		log.Infoln("constructor", "Type of node: %s", reflect.TypeOf(node))
		panic("Encountered un-constructable node")
	}
}

func (v *Constructor) constructNodes(nodes []parser.ParseNode) []Node {
	var res []Node
	for _, node := range nodes {
		res = append(res, v.constructNode(node))
	}
	return res
}

func (v *Constructor) constructTypes(nodes []parser.ParseNode) []Type {
	var res []Type
	for _, node := range nodes {
		res = append(res, v.constructType(node))
	}
	return res
}

func (v *Constructor) constructExprs(nodes []parser.ParseNode) []Expr {
	var res []Expr
	for _, node := range nodes {
		res = append(res, v.constructExpr(node))
	}
	return res
}

func (c *Constructor) constructTypeReferenceNode(v *parser.TypeReferenceNode) *TypeReference {
	args := c.constructTypeReferences(v.GenericArguments)
	res := &TypeReference{BaseType: c.constructType(v.Type), GenericArguments: args}
	return res
}

func (v *Constructor) constructTypeReferences(refs []*parser.TypeReferenceNode) []*TypeReference {
	ret := make([]*TypeReference, 0, len(refs))
	for _, ref := range refs {
		ret = append(ret, v.constructTypeReferenceNode(ref))
	}
	return ret
}

func (c *Constructor) constructReferenceTypeNode(v *parser.ReferenceTypeNode) ReferenceType {
	targetType := c.constructTypeReferenceNode(v.TargetType)
	return ReferenceTo(targetType, v.Mutable)
}

func (c *Constructor) constructPointerTypeNode(v *parser.PointerTypeNode) PointerType {
	targetType := c.constructTypeReferenceNode(v.TargetType)
	return PointerTo(targetType, v.Mutable)
}

func (c *Constructor) constructTupleTypeNode(v *parser.TupleTypeNode) TupleType {
	res := TupleType{}
	res.Members = c.constructTypeReferences(v.MemberTypes)
	return res
}

func (c *Constructor) constructFunctionTypeNode(v *parser.FunctionTypeNode) FunctionType {
	res := FunctionType{
		IsVariadic: v.IsVariadic,
		Parameters: c.constructTypeReferences(v.ParameterTypes),
		attrs:      v.Attrs(),
	}

	if v.ReturnType != nil {
		res.Return = c.constructTypeReferenceNode(v.ReturnType)
	} else {
		res.Return = &TypeReference{BaseType: PRIMITIVE_void}
	}
	return res
}

func (c *Constructor) constructArrayTypeNode(v *parser.ArrayTypeNode) ArrayType {
	memberType := c.constructTypeReferenceNode(v.MemberType)
	return ArrayOf(memberType, v.IsFixedLength, v.Length)
}

func (c *Constructor) constructNamedTypeNode(v *parser.NamedTypeNode) UnresolvedType {
	return UnresolvedType{Name: toUnresolvedName(v.Name)}
}

func (c *Constructor) constructInterfaceTypeNode(v *parser.InterfaceTypeNode) InterfaceType {
	interfaceType := InterfaceType{
		attrs:        v.Attrs(),
		GenericSigil: c.constructGenericSigilNode(v.GenericSigil),
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
		funcData := c.constructFunctionNode(&parser.FunctionNode{Header: function})
		interfaceType = interfaceType.addFunction(funcData)
	}

	return interfaceType
}

func (c *Constructor) constructStructTypeNode(v *parser.StructTypeNode) StructType {
	structType := StructType{
		attrs:             v.Attrs(),
		GenericParameters: c.constructGenericSigilNode(v.GenericSigil),
	}

	for _, member := range v.Members {
		structType = structType.addMember(member.Name.Value, c.constructTypeReferenceNode(member.Type))
	}

	return structType
}

func (c *Constructor) constructEnumTypeNode(v *parser.EnumTypeNode) EnumType {
	enumType := EnumType{
		Simple:            true,
		Members:           make([]EnumTypeMember, len(v.Members)),
		GenericParameters: c.constructGenericSigilNode(v.GenericSigil),
	}

	lastValue := 0
	for idx, mem := range v.Members {
		enumType.Members[idx].Name = mem.Name.Value

		if mem.TupleBody != nil {
			enumType.Members[idx].Type = c.constructTupleTypeNode(mem.TupleBody)
			enumType.Simple = false
		} else if mem.StructBody != nil {
			enumType.Members[idx].Type = c.constructStructTypeNode(mem.StructBody)
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

func (c *Constructor) constructTypeDeclNode(v *parser.TypeDeclNode) *TypeDecl {
	var paramNodes []parser.ParseNode

	if v.GenericSigil != nil {
		paramNodes = make([]parser.ParseNode, len(v.GenericSigil.GenericParameters))
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
	res.SetPos(v.Where().Start())

	return res
}

func (c *Constructor) constructLinkDirectiveNode(v *parser.LinkDirectiveNode) Node {
	c.module.LinkedLibraries = append(c.module.LinkedLibraries, v.Library.Value)
	return nil
}

func (c *Constructor) constructUseDirectiveNode(v *parser.UseDirectiveNode) *UseDirective {
	res := &UseDirective{}
	res.ModuleName = toUnresolvedName(v.Module)
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructFunctionDeclNode(v *parser.FunctionDeclNode) *FunctionDecl {
	function := c.constructFunctionNode(v.Function)
	function.Type.attrs = v.Attrs()

	res := &FunctionDecl{
		docs:      v.DocComments(),
		Function:  function,
		Prototype: v.Function.Body == nil,
	}

	res.SetPublic(v.IsPublic())
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructVarDeclNode(v *parser.VarDeclNode) *VariableDecl {
	if parser.IsReservedKeyword(v.Name.Value) {
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
		variable.Type = c.constructTypeReferenceNode(v.Type)
	}

	res := &VariableDecl{
		docs:     v.DocComments(),
		Variable: variable,
	}

	if v.Value != nil {
		res.Assignment = c.constructExpr(v.Value)
	}

	res.SetPublic(v.IsPublic())
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructDestructVarDeclNode(v *parser.DestructVarDeclNode) *DestructVarDecl {
	res := &DestructVarDecl{
		docs:          v.DocComments(),
		Variables:     make([]*Variable, len(v.Names)),
		ShouldDiscard: make([]bool, len(v.Names)),
		Assignment:    c.constructExpr(v.Value),
	}
	res.SetPos(v.Where().Start())

	for idx, name := range v.Names {
		mutable := v.Mutable[idx]

		if name.Value == parser.KEYWORD_DISCARD {
			res.ShouldDiscard[idx] = true
		} else {
			res.Variables[idx] = &Variable{
				Name:         name.Value,
				Attrs:        make(parser.AttrGroup),
				Mutable:      mutable,
				ParentModule: c.module,
			}
		}
	}

	return res
}

func (c *Constructor) constructDeferStatNode(v *parser.DeferStatNode) *DeferStat {
	res := &DeferStat{}
	res.Call = c.constructCallExprNode(v.Call)
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructIfStatNode(v *parser.IfStatNode) *IfStat {
	res := &IfStat{}
	for _, part := range v.Parts {
		res.Exprs = append(res.Exprs, c.constructExpr(part.Condition))
		res.Bodies = append(res.Bodies, c.constructBlockNode(part.Body))
	}
	if v.ElseBody != nil {
		res.Else = c.constructBlockNode(v.ElseBody)
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructMatchStatNode(v *parser.MatchStatNode) *MatchStat {
	res := &MatchStat{}
	res.Target = c.constructExpr(v.Value)
	res.Branches = make(map[Expr]Node)
	for _, branch := range v.Cases {
		pattern := c.constructExpr(branch.Pattern)
		body := c.constructNode(branch.Body)
		res.Branches[pattern] = body
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructLoopStatNode(v *parser.LoopStatNode) *LoopStat {
	res := &LoopStat{}
	if v.Condition != nil {
		res.LoopType = LOOP_TYPE_CONDITIONAL
		res.Condition = c.constructExpr(v.Condition)
	} else {
		res.LoopType = LOOP_TYPE_INFINITE
	}
	res.Body = c.constructBlockNode(v.Body)
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructReturnStatNode(v *parser.ReturnStatNode) *ReturnStat {
	res := &ReturnStat{}
	if v.Value != nil {
		res.Value = c.constructExpr(v.Value)
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructBreakStatNode(v *parser.BreakStatNode) *BreakStat {
	res := &BreakStat{}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructNextStatNode(v *parser.NextStatNode) *NextStat {
	res := &NextStat{}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructBlockStatNode(v *parser.BlockStatNode) *BlockStat {
	res := &BlockStat{}
	res.Block = c.constructBlockNode(v.Body)
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructBlockNode(v *parser.BlockNode) *Block {
	res := &Block{}
	res.NonScoping = v.NonScoping
	res.Nodes = c.constructNodes(v.Nodes)
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructCallStatNode(v *parser.CallStatNode) *CallStat {
	res := &CallStat{}
	res.Call = c.constructExpr(v.Call).(*CallExpr)
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructAssignStatNode(v *parser.AssignStatNode) Node {
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

	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructBinopAssignStatNode(v *parser.BinopAssignStatNode) Node {
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

	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructBinaryExprNode(v *parser.BinaryExprNode) *BinaryExpr {
	res := &BinaryExpr{
		Lhand: c.constructExpr(v.Lhand),
		Rhand: c.constructExpr(v.Rhand),
		Op:    v.Operator,
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructArrayLenExprNode(v *parser.ArrayLenExprNode) *ArrayLenExpr {
	res := &ArrayLenExpr{}
	if v.ArrayExpr != nil {
		res.Expr = c.constructExpr(v.ArrayExpr)
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructSizeofExprNode(v *parser.SizeofExprNode) *SizeofExpr {
	res := &SizeofExpr{}
	if v.Value != nil {
		res.Expr = c.constructExpr(v.Value)
	} else if v.Type != nil {
		res.Type = c.constructTypeReferenceNode(v.Type)
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructAddrofExprNode(v *parser.AddrofExprNode) Expr {
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
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructCastExprNode(v *parser.CastExprNode) *CastExpr {
	res := &CastExpr{
		Type: c.constructTypeReferenceNode(v.Type),
		Expr: c.constructExpr(v.Value),
	}
	if ttyp, ok := res.Type.BaseType.(TupleType); ok && len(ttyp.Members) == 1 {
		res.Type = ttyp.Members[0]
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructUnaryExprNode(v *parser.UnaryExprNode) Expr {
	var res Expr
	subExpr := c.constructExpr(v.Value)
	if v.Operator == parser.UNOP_DEREF {
		res = &DerefAccessExpr{
			Expr: subExpr,
		}
	} else if numlit, ok := subExpr.(*NumericLiteral); ok && v.Operator == parser.UNOP_NEGATIVE {
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

	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructCallExprNode(v *parser.CallExprNode) *CallExpr {
	// TODO: when we allow function types, allow all access forms (eg. `thing[0]()``)
	res := &CallExpr{
		Arguments: c.constructExprs(v.Arguments),
		Function:  c.constructExpr(v.Function),
	}

	if sae, ok := v.Function.(*parser.StructAccessNode); ok {
		res.ReceiverAccess = c.constructStructAccessNode(sae).Struct
	}

	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructVariableAccessNode(v *parser.VariableAccessNode) *VariableAccessExpr {
	res := &VariableAccessExpr{
		Name:             toUnresolvedName(v.Name),
		GenericArguments: c.constructTypeReferences(v.GenericParameters),
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructStructAccessNode(v *parser.StructAccessNode) *StructAccessExpr {
	res := &StructAccessExpr{
		Member: v.Member.Value,
	}
	res.Struct = c.constructExpr(v.Struct).(AccessExpr) // TODO: Error message
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructArrayAccessNode(v *parser.ArrayAccessNode) *ArrayAccessExpr {
	res := &ArrayAccessExpr{
		Subscript: c.constructExpr(v.Index),
	}
	res.Array = c.constructExpr(v.Array).(AccessExpr) // TODO: Error message
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructDiscardAccessNode(v *parser.DiscardAccessNode) *DiscardAccessExpr {
	res := &DiscardAccessExpr{}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructEnumPatternNode(v *parser.EnumPatternNode) *EnumPatternExpr {
	res := &EnumPatternExpr{
		MemberName: toUnresolvedName(v.MemberName),
		Variables:  make([]*Variable, len(v.Names)),
	}
	for idx, name := range v.Names {
		if name.Value != parser.KEYWORD_DISCARD {
			res.Variables[idx] = &Variable{
				Name:         name.Value,
				ParentModule: c.module,
			}
		}
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructTupleLiteralNode(v *parser.TupleLiteralNode) Expr {
	res := &TupleLiteral{
		Members: c.constructExprs(v.Values),
	}
	if len(res.Members) == 1 {
		return res.Members[0]
	}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructCompositeLiteralNode(v *parser.CompositeLiteralNode) *CompositeLiteral {
	res := &CompositeLiteral{}
	if v.Type != nil {
		res.Type = c.constructTypeReferenceNode(v.Type)
	}

	for i, val := range v.Values {
		res.Fields = append(res.Fields, v.Fields[i].Value)
		res.Values = append(res.Values, c.constructExpr(val))
	}

	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructBoolLitNode(v *parser.BoolLitNode) *BoolLiteral {
	res := &BoolLiteral{Value: v.Value}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructNumberLitNode(v *parser.NumberLitNode) *NumericLiteral {
	res := &NumericLiteral{
		IsFloat:    v.IsFloat,
		IntValue:   v.IntValue,
		FloatValue: v.FloatValue,
	}

	res.floatSizeHint = v.FloatSize
	res.SetPos(v.Where().Start())
	return res

}

func (c *Constructor) constructStringLitNode(v *parser.StringLitNode) *StringLiteral {
	res := &StringLiteral{Value: v.Value, IsCString: v.IsCString}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructRuneLitNode(v *parser.RuneLitNode) *RuneLiteral {
	res := &RuneLiteral{Value: v.Value}
	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructLambdaExprNode(v *parser.LambdaExprNode) *LambdaExpr {
	function := c.constructFunctionNode(v.Function)
	function.Type.attrs = v.Attrs()

	res := &LambdaExpr{
		Function: function,
	}

	res.SetPos(v.Where().Start())
	return res
}

func (c *Constructor) constructFunctionNode(v *parser.FunctionNode) *Function {
	function := &Function{
		Name:         v.Header.Name.Value,
		ParentModule: c.module,
		Type: FunctionType{
			IsVariadic: v.Header.Variadic,
			//attrs:      v.Attrs(),
		},
	}

	if len(v.Attrs()) != 0 {
		panic("functionnode shouldn't have attributes")
	}

	if v.Header.Receiver != nil {
		function.Receiver = c.constructVarDeclNode(v.Header.Receiver)
		if !function.Receiver.Variable.IsReceiver {
			panic("INTERNAL ERROR: Reciever variable was not marked a reciever")
		}

		// Propagate generic parameters from reciever to method
		typ := TypeReferenceWithoutPointers(function.Receiver.Variable.Type)
		if len(typ.GenericArguments) > 0 {
			for _, param := range typ.GenericArguments {
				if unres, ok := param.BaseType.(UnresolvedType); ok {
					name := unres.Name.Name
					subt := &SubstitutionType{Name: name}
					function.Type.GenericParameters = append(function.Type.GenericParameters, subt)
				} else {
					panic("INTERNAL ERROR: Any type parameters used in reciever must not be actual types")
				}

			}
		}

		function.Type.Receiver = function.Receiver.Variable.Type
	} else if v.Header.StaticReceiverType != nil {
		function.StaticReceiverType = c.constructType(v.Header.StaticReceiverType)
	}

	var arguments []parser.ParseNode
	for _, arg := range v.Header.Arguments { // TODO rename v.Header.Arguments to v.Header.Parameters
		arguments = append(arguments, arg)
		decl := c.constructVarDeclNode(arg)
		decl.Variable.IsParameter = true
		function.Parameters = append(function.Parameters, decl)
		function.Type.Parameters = append(function.Type.Parameters, decl.Variable.Type)
	}

	if v.Header.ReturnType != nil {
		function.Type.Return = c.constructTypeReferenceNode(v.Header.ReturnType)
	} else {
		// set it to void since we haven't specified a type
		function.Type.Return = &TypeReference{BaseType: PRIMITIVE_void}
	}

	if v.Header.GenericSigil != nil {
		function.Type.GenericParameters = c.constructGenericSigilNode(v.Header.GenericSigil)
	}

	if v.Expr != nil {
		v.Stat = &parser.ReturnStatNode{Value: v.Expr}
	}
	if v.Stat != nil {
		v.Body = &parser.BlockNode{Nodes: []parser.ParseNode{v.Stat}}
	}
	if v.Body != nil {
		function.Body = c.constructBlockNode(v.Body)
	} else if v.Header.Anonymous {
		c.err(v.Where(), "Lambda cannot be prototype")
	}

	return function
}

func (c *Constructor) constructGenericSigilNode(v *parser.GenericSigilNode) GenericSigil {
	if v == nil {
		return nil
	}

	ret := make([]*SubstitutionType, 0, len(v.GenericParameters))
	for _, p := range v.GenericParameters {
		ret = append(ret, NewSubstitutionType(p.Name.Value, c.constructTypes(p.Constraints)))
	}

	return ret
}

func toUnresolvedName(node *parser.NameNode) UnresolvedName {
	res := UnresolvedName{Name: node.Name.Value}
	for _, module := range node.Modules {
		res.ModuleNames = append(res.ModuleNames, module.Value)
	}
	return res
}

func toParentName(node *parser.NameNode) UnresolvedName {
	res := UnresolvedName{}
	for _, moduleName := range node.Modules[:len(node.Modules)-1] {
		res.ModuleNames = append(res.ModuleNames, moduleName.Value)
	}
	res.Name = node.Modules[len(node.Modules)-1].Value

	return res
}
