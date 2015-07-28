package parser

import (
	"fmt"
	"os"

	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
)

type unresolvedName struct {
	moduleNames []string
	name        string
	modules     map[string]*Module
}

func (v unresolvedName) String() string {
	ret := ""
	for _, mod := range v.moduleNames {
		ret += mod + "::"
	}
	return ret + v.name
}

type Resolver struct {
	Module   *Module
	modules  map[string]*Module
	resolved map[interface{}]bool
}

func (v *Resolver) err(thing Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Error("resolve", util.TEXT_RED+util.TEXT_BOLD+"Resolve error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Error("resolve", v.Module.File.MarkPos(pos))

	os.Exit(util.EXIT_FAILURE_SEMANTIC)
}

func (v *Resolver) errCannotResolve(thing Locatable, name unresolvedName) {
	v.err(thing, "Cannot resolve `%s`", name.String())
}

func (v *Resolver) Resolve(modules map[string]*Module) {
	v.resolved = make(map[interface{}]bool)
	v.modules = modules

	for _, node := range v.Module.Nodes {
		node.resolve(v, v.Module.GlobalScope)
	}
}

func (v *Block) resolve(res *Resolver, s *Scope) {
	if res.resolved[v] {
		return
	}
	res.resolved[v] = true

	for _, n := range v.Nodes {
		n.resolve(res, v.scope)
	}
}

/**
 * Declarations
 */

func (v *VariableDecl) resolve(res *Resolver, s *Scope) {
	if v.Assignment != nil {
		v.Assignment.resolve(res, s)
	}

	if v.Variable.Type != nil {
		v.Variable.Type = v.Variable.Type.resolveType(v, res, s)
	}
}

func (v *StructDecl) resolve(res *Resolver, s *Scope) {
	v.Struct = v.Struct.resolveType(v, res, s).(*StructType)
}

func (v *EnumDecl) resolve(res *Resolver, s *Scope) {
	v.Enum = v.Enum.resolveType(v, res, s).(*EnumType)
}

func (v *TraitDecl) resolve(res *Resolver, s *Scope) {
	v.Trait = v.Trait.resolveType(v, res, s).(*TraitType)
}

func (v *ImplDecl) resolve(res *Resolver, s *Scope) {
	for _, fun := range v.Functions {
		fun.resolve(res, s)
	}
}

func (v *FunctionDecl) resolve(res *Resolver, s *Scope) {
	for _, param := range v.Function.Parameters {
		param.resolve(res, s)
	}

	if v.Function.ReturnType != nil {
		v.Function.ReturnType = v.Function.ReturnType.resolveType(v, res, s)
	}

	if !v.Prototype {
		v.Function.Body.resolve(res, s)
	}
}

func (v *UseDecl) resolve(res *Resolver, s *Scope) {
	// later...
}

/*
 * Statements
 */

func (v *ReturnStat) resolve(res *Resolver, s *Scope) {
	if v.Value != nil {
		v.Value.resolve(res, s)
	}
}

func (v *IfStat) resolve(res *Resolver, s *Scope) {
	for _, expr := range v.Exprs {
		expr.resolve(res, s)
	}

	for _, body := range v.Bodies {
		body.resolve(res, s)
	}

	if v.Else != nil {
		v.Else.resolve(res, s)
	}

}

func (v *BlockStat) resolve(res *Resolver, s *Scope) {
	v.Block.resolve(res, s)
}

func (v *CallStat) resolve(res *Resolver, s *Scope) {
	v.Call.resolve(res, s)
}

func (v *DeferStat) resolve(res *Resolver, s *Scope) {
	v.Call.resolve(res, s)
}

func (v *AssignStat) resolve(res *Resolver, s *Scope) {
	v.Assignment.resolve(res, s)
	v.Access.resolve(res, s)
}

func (v *BinopAssignStat) resolve(res *Resolver, s *Scope) {
	v.Assignment.resolve(res, s)
	v.Access.resolve(res, s)
}

func (v *LoopStat) resolve(res *Resolver, s *Scope) {
	v.Body.resolve(res, s)

	switch v.LoopType {
	case LOOP_TYPE_INFINITE:
	case LOOP_TYPE_CONDITIONAL:
		v.Condition.resolve(res, s)
	default:
		panic("invalid loop type")
	}
}

func (v *MatchStat) resolve(res *Resolver, s *Scope) {
	v.Target.resolve(res, s)

	for pattern, stmt := range v.Branches {
		pattern.resolve(res, s)
		stmt.resolve(res, s)
	}
}

func (v *DefaultStat) resolve(res *Resolver, s *Scope) {
	v.Target.resolve(res, s)
}

/*
 * Expressions
 */

func (v *NumericLiteral) resolve(res *Resolver, s *Scope) {}
func (v *StringLiteral) resolve(res *Resolver, s *Scope)  {}
func (v *RuneLiteral) resolve(res *Resolver, s *Scope)    {}
func (v *BoolLiteral) resolve(res *Resolver, s *Scope)    {}

func (v *UnaryExpr) resolve(res *Resolver, s *Scope) {
	v.Expr.resolve(res, s)
}

func (v *BinaryExpr) resolve(res *Resolver, s *Scope) {
	v.Lhand.resolve(res, s)
	v.Rhand.resolve(res, s)
}

func (v *ArrayLiteral) resolve(res *Resolver, s *Scope) {
	for _, mem := range v.Members {
		mem.resolve(res, s)
	}
}

func (v *CastExpr) resolve(res *Resolver, s *Scope) {
	v.Type = v.Type.resolveType(v, res, s)
	v.Expr.resolve(res, s)
}

func (v *CallExpr) resolve(res *Resolver, s *Scope) {
	// TODO: This will be cleaner once we get around to implementing function types
	var name unresolvedName
	switch v.functionSource.(type) {
	case *VariableAccessExpr:
		vae := v.functionSource.(*VariableAccessExpr)
		name = vae.Name

	case *StructAccessExpr:
		sae := v.functionSource.(*StructAccessExpr)
		sae.Struct.resolve(res, s)
		name = unresolvedName{name: sae.Struct.GetType().TypeName() + "." + sae.Member}

	default:
		panic("Invalid function source (for now)")
	}

	ident := s.GetIdent(name)
	if ident == nil {
		res.errCannotResolve(v, name)
	} else if ident.Type != IDENT_FUNCTION {
		res.err(v, "Expected function identifier, found %s `%s`", ident.Type, name)
	} else {
		v.Function = ident.Value.(*Function)
	}

	for _, arg := range v.Arguments {
		arg.resolve(res, s)
	}
}

func (v *VariableAccessExpr) resolve(res *Resolver, s *Scope) {
	ident := s.GetIdent(v.Name)
	if ident == nil {
		res.errCannotResolve(v, v.Name)
	} else if ident.Type != IDENT_VARIABLE {
		res.err(v, "Expected variable identifier, found %s `%s`", ident.Type, v.Name)
	} else {
		v.Variable = ident.Value.(*Variable)
	}

	if v.Variable == nil {
		res.errCannotResolve(v, v.Name)
	} else if v.Variable.Type != nil {
		v.Variable.Type.resolveType(v, res, s)
	}
}

func (v *StructAccessExpr) resolve(res *Resolver, s *Scope) {
	v.Struct.resolve(res, s)

	pointerType, ok := v.Struct.GetType().(PointerType)
	if _, isStruct := pointerType.Addressee.(*StructType); ok && isStruct {
		v.Struct = &DerefAccessExpr{Expr: v.Struct, Type: pointerType.Addressee}
		v.resolve(res, s)
	}

	structType, ok := v.Struct.GetType().(*StructType)
	if !ok {
		if v.Struct.GetType() == nil {
			res.err(v, "Type of access expression was nil")
		} else {
			res.err(v, "Cannot access member of type `%s`", v.Struct.GetType().TypeName())
		}
	}

	// TODO check no mod access
	decl := structType.GetVariableDecl(v.Member)
	if decl == nil {
		res.err(v, "Struct `%s` does not contain member `%s`", structType.TypeName(), v.Member)
	}

	v.Variable = decl.Variable
}

func (v *ArrayAccessExpr) resolve(res *Resolver, s *Scope) {
	v.Array.resolve(res, s)
	v.Subscript.resolve(res, s)
}

func (v *TupleAccessExpr) resolve(res *Resolver, s *Scope) {
	v.Tuple.resolve(res, s)
}

func (v *DerefAccessExpr) resolve(res *Resolver, s *Scope) {
	v.Expr.resolve(res, s)
	if ptr, ok := v.Expr.GetType().(PointerType); ok {
		v.Type = ptr.Addressee
	}
}

func (v *AddressOfExpr) resolve(res *Resolver, s *Scope) {
	v.Access.resolve(res, s)
}

func (v *SizeofExpr) resolve(res *Resolver, s *Scope) {
	if v.Expr != nil {
		v.Expr.resolve(res, s)
	} else if v.Type != nil {
		v.Type = v.Type.resolveType(v, res, s)
	} else {
		panic("invalid state")
	}
}

func (v *TupleLiteral) resolve(res *Resolver, s *Scope) {
	for _, mem := range v.Members {
		mem.resolve(res, s)
	}
}

func (v *StructLiteral) resolve(res *Resolver, s *Scope) {
	for _, mem := range v.Values {
		mem.resolve(res, s)
	}
}

func (v *EnumLiteral) resolve(res *Resolver, s *Scope) {
	v.Type = v.Type.resolveType(v, res, s)
	if v.TupleLiteral != nil {
		v.TupleLiteral.resolve(res, s)
	} else if v.StructLiteral != nil {
		v.StructLiteral.resolve(res, s)
	}
}

func (v *DefaultMatchBranch) resolve(res *Resolver, s *Scope) {}

func (v *DefaultExpr) resolve(res *Resolver, s *Scope) {
	v.Type = v.Type.resolveType(v, res, s)
}

/*
 * Types
 */

func (v PrimitiveType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return v
}

func (v *StructType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	if res.resolved[v] {
		return v
	}
	res.resolved[v] = true

	for _, vari := range v.Variables {
		vari.resolve(res, s)
	}
	return v
}

func (v ArrayType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return arrayOf(v.MemberType.resolveType(src, res, s))
}

func (v *TraitType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	if res.resolved[v] {
		return v
	}
	res.resolved[v] = true

	for _, fun := range v.Functions {
		fun.resolve(res, s)
	}
	return v
}

func (v PointerType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return pointerTo(v.Addressee.resolveType(src, res, s))
}

func (v *TupleType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	for idx, mem := range v.Members {
		v.Members[idx] = mem.resolveType(src, res, s)
	}
	return v
}

func (v *EnumType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	if res.resolved[v] {
		return v
	}
	res.resolved[v] = true

	for _, mem := range v.Members {
		mem.Type = mem.Type.resolveType(src, res, s)
	}
	return v
}

func (v *NamedType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	v.Type = v.Type.resolveType(src, res, s)
	return v
}

func (v *UnresolvedType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	ident := s.GetIdent(v.Name)
	if ident == nil {
		res.err(src, "Cannot resolve `%s`", v.Name)
	} else if ident.Type != IDENT_TYPE {
		res.err(src, "Expected type identifier, found %s `%s`", ident.Type, v.Name)
	} else {
		return ident.Value.(Type)
	}

	panic("unreachable")
}
