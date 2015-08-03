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
	resolved map[interface{}]bool

	scope []*Scope
}

type Resolvable interface {
	resolve(*Resolver, *Scope)
}

func (v *Resolver) err(thing Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Error("resolve", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Error("resolve", v.Module.File.MarkPos(pos))

	os.Exit(util.EXIT_FAILURE_SEMANTIC)
}

func (v *Resolver) errCannotResolve(thing Locatable, name unresolvedName) {
	v.err(thing, "Cannot resolve `%s`", name.String())
}

func (v *Resolver) Scope() *Scope {
	return v.scope[len(v.scope)-1]
}

func (v *Resolver) EnterScope(s *Scope) {
	if v.resolved == nil {
		v.resolved = make(map[interface{}]bool)
	}

	if s != nil {
		v.scope = append(v.scope, s)
	}
}

func (v *Resolver) ExitScope(s *Scope) {
	if s != nil {
		v.scope = v.scope[:len(v.scope)-1]
	}
}

func (v *Resolver) Visit(n Node) {
	if resolveable, ok := n.(Resolvable); ok {
		resolveable.resolve(v, v.Scope())
	}

}

func (v *Resolver) PostVisit(n Node) {
	switch n.(type) {
	case *DerefAccessExpr:
		dae := n.(*DerefAccessExpr)
		if ptr, ok := dae.Expr.GetType().(PointerType); ok {
			dae.Type = ptr.Addressee
		}

	case *FunctionDecl:
		fd := n.(*FunctionDecl)
		if fd.Function.IsMethod && !fd.Function.IsStatic {
			TypeWithoutPointers(fd.Function.Receiver.Variable.Type).(*NamedType).addMethod(fd.Function)
		}
	}
}

///
// LATA
///

func (v *FunctionDecl) resolve(res *Resolver, s *Scope) {
	if v.Function.ReturnType != nil {
		v.Function.ReturnType = v.Function.ReturnType.resolveType(v, res, s)
	}

	if v.Function.IsMethod {
		if v.Function.IsStatic {
			v.Function.StaticReceiverType = v.Function.StaticReceiverType.resolveType(v, res, s)
			v.Function.StaticReceiverType.(*NamedType).addMethod(v.Function)
		}
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
		v.Variable.Type = v.Variable.Type.resolveType(v, res, s)
	}
}

func (v *VariableDecl) resolve(res *Resolver, s *Scope) {
	if v.Variable.Type != nil {
		v.Variable.Type = v.Variable.Type.resolveType(v, res, s)
	}
}

func (v *TypeDecl) resolve(res *Resolver, s *Scope) {
	v.NamedType = v.NamedType.resolveType(v, res, s).(*NamedType)
}

func (v *CastExpr) resolve(res *Resolver, s *Scope) {
	v.Type = v.Type.resolveType(v, res, s)
}

func (v *SizeofExpr) resolve(res *Resolver, s *Scope) {
	if v.Type != nil {
		v.Type = v.Type.resolveType(v, res, s)
	}
}

func (v *EnumLiteral) resolve(res *Resolver, s *Scope) {
	v.Type = v.Type.resolveType(v, res, s)
}

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

func (v *InterfaceType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
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

/*func (v *TraitDecl) resolve(res *Resolver, s *Scope) {
	v.Trait = v.Trait.resolveType(v, res, s).(*TraitType)
}

func (v *ImplDecl) resolve(res *Resolver, s *Scope) {
	for _, fun := range v.Functions {
		fun.resolve(res, s)
	}
}*/

/*func (v *TraitType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	if res.resolved[v] {
		return v
	}
	res.resolved[v] = true

	for _, fun := range v.Functions {
		fun.resolve(res, s)
	}
	return v
}*/
