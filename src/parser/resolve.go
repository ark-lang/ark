package parser

import (
	"fmt"
	"os"
	"reflect"

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
	Module *Module

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
	if v.NamedType.Parameters == nil {
		v.NamedType.Type = v.NamedType.Type.resolveType(v, res, s)
	}
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

func (v StructType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	nv := StructType{
		Variables: make([]*VariableDecl, len(v.Variables)),
		attrs:     v.attrs,
	}

	for idx, vari := range v.Variables {
		nv.Variables[idx] = &VariableDecl{
			Variable: &Variable{
				Type:         vari.Variable.Type,
				Name:         vari.Variable.Name,
				Mutable:      vari.Variable.Mutable,
				Attrs:        vari.Variable.Attrs,
				scope:        vari.Variable.scope,
				FromStruct:   vari.Variable.FromStruct,
				ParentStruct: vari.Variable.ParentStruct,
				ParentModule: vari.Variable.ParentModule,
				IsParameter:  vari.Variable.IsParameter,
			},
			Assignment: vari.Assignment,
			docs:       vari.docs,
		}
		nv.Variables[idx].resolve(res, s)

		if nv.Variables[idx].Assignment != nil {
			visitor := &ASTVisitor{Visitor: res}
			visitor.Visit(nv.Variables[idx].Assignment)
		}
	}

	return nv
}

func (v ArrayType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return arrayOf(v.MemberType.resolveType(src, res, s))
}

func (v PointerType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return pointerTo(v.Addressee.resolveType(src, res, s))
}

func (v TupleType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	nv := TupleType{Members: make([]Type, len(v.Members))}

	for idx, mem := range v.Members {
		nv.Members[idx] = mem.resolveType(src, res, s)
	}

	return nv
}

func (v EnumType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	nv := EnumType{
		Simple:  v.Simple,
		Members: make([]EnumTypeMember, len(v.Members)),
		attrs:   v.attrs,
	}

	for idx, mem := range v.Members {
		nv.Members[idx].Name = mem.Name
		nv.Members[idx].Tag = mem.Tag
		nv.Members[idx].Type = mem.Type.resolveType(src, res, s)
	}

	return nv
}

func (v *NamedType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return v
}

func (v InterfaceType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return v
}

func (v FunctionType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return v
}

func (v ParameterType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	panic("We shouldn't reach this, right?")
}

func (v SubstitutionType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return v.Type
}

func (v UnresolvedType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	ident := s.GetIdent(v.Name)
	if ident == nil {
		res.errCannotResolve(src, v.Name)
	} else if ident.Type != IDENT_TYPE {
		res.err(src, "Expected type identifier, found %s `%s`", ident.Type, v.Name)
	} else {
		typ := ident.Value.(Type)

		if namedType, ok := typ.(*NamedType); ok && len(v.Parameters) > 0 {
			scope := newScope(s)

			name := namedType.Name + "<"
			for idx, param := range namedType.Parameters {
				paramType := SubstitutionType{Name: param.Name, Type: v.Parameters[idx].resolveType(src, res, s)}
				scope.InsertType(paramType)

				name += v.Parameters[idx].TypeName()
				if idx < len(namedType.Parameters)-1 {
					name += ", "
				}
			}
			name += ">"

			typ = &NamedType{
				Name:         name,
				Type:         namedType.Type.resolveType(src, res, scope),
				ParentModule: namedType.ParentModule,
				Methods:      namedType.Methods,
			}
		} else {
			typ = typ.resolveType(src, res, s)
		}

		return typ
	}

	panic("unreachable")
}

func ExtractTypeVariable(pattern Type, value Type) map[string]Type {
	/*
		Pointer($T), Pointer(int) -> {$T: int}
		Arbitrary depth type => Stack containing breadth first traversal
	*/
	res := make(map[string]Type)

	var (
		ps []Type
		vs []Type
	)
	ps = append(ps, pattern)
	vs = append(vs, value)

	for i := 0; i < len(ps); i++ {
		ppart := ps[i]
		vpart := vs[i]
		log.Debugln("resovle", "\nP = `%s`, V = `%s`", ppart.TypeName(), vpart.TypeName())

		ps = AddChildren(ppart, ps)
		vs = AddChildren(vpart, vs)

		if vari, ok := ppart.(ParameterType); ok {
			log.Debugln("resolve", "P was variable (Name: %s)", vari.Name)
			res[vari.Name] = vpart
			continue
		}

		switch ppart.(type) {
		case PrimitiveType, *NamedType:
			if !ppart.Equals(vpart) {
				log.Errorln("resolve", "%s != %s", ppart.TypeName(), vpart.TypeName())
				panic("Part of type did not match pattern")
			}

		default:
			if reflect.TypeOf(ppart) != reflect.TypeOf(vpart) {
				log.Errorln("resolve", "%T != %T", ppart, vpart)
				panic("Part of type did not match pattern")
			}
		}
	}

	return res
}

func AddChildren(typ Type, dest []Type) []Type {
	switch typ.(type) {
	case StructType:
		st := typ.(StructType)
		for _, decl := range st.Variables {
			dest = append(dest, decl.Variable.Type)
		}

	case *NamedType:
		nt := typ.(*NamedType)
		dest = append(dest, nt.Type)

	case ArrayType:
		at := typ.(ArrayType)
		dest = append(dest, at.MemberType)

	case PointerType:
		pt := typ.(PointerType)
		dest = append(dest, pt.Addressee)

	case TupleType:
		tt := typ.(TupleType)
		for _, mem := range tt.Members {
			dest = append(dest, mem)
		}

	case EnumType:
		et := typ.(EnumType)
		for _, mem := range et.Members {
			dest = append(dest, mem.Type)
		}
	}
	return dest
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
