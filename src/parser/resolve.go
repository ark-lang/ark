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

func (v unresolvedName) Split() (unresolvedName, string) {
	if len(v.moduleNames) > 0 {
		res := unresolvedName{}
		res.moduleNames = v.moduleNames[:len(v.moduleNames)-1]
		res.name = v.moduleNames[len(v.moduleNames)-1]
		return res, v.name
	} else {
		return unresolvedName{}, ""
	}
}

type Resolver struct {
	Module *Module

	scope []*Scope
}

type Resolvable interface {
	resolve(*Resolver, *Scope) Node
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

func (v *Resolver) Visit(n *Node) bool {
	if resolveable, ok := (*n).(Resolvable); ok {
		*n = resolveable.resolve(v, v.Scope())
	}
	return true
}

func (v *Resolver) PostVisit(n *Node) {
	switch (*n).(type) {
	case *DerefAccessExpr:
		dae := (*n).(*DerefAccessExpr)
		if ce, ok := dae.Expr.(*CastExpr); ok {
			*n = &CastExpr{Type: pointerTo(ce.Type), Expr: ce.Expr}
		} else if ptr, ok := dae.Expr.GetType().(PointerType); ok {
			dae.Type = ptr.Addressee
		}

	case *FunctionDecl:
		fd := (*n).(*FunctionDecl)
		if fd.Function.Type.Receiver != nil {
			TypeWithoutPointers(fd.Function.Receiver.Variable.Type).(*NamedType).addMethod(fd.Function)
		}
	}
}

///
// LATA
///

func (v *FunctionDecl) resolve(res *Resolver, s *Scope) Node {
	if v.Function.Type.Return != nil {
		v.Function.Type.Return = v.Function.Type.Return.resolveType(v, res, s)
	}

	if v.Function.Type.Receiver != nil {
		if v.Function.StaticReceiverType != nil { //v.Function.IsStatic {
			fmt.Println("good!")
			v.Function.StaticReceiverType = v.Function.StaticReceiverType.resolveType(v, res, s)
			v.Function.StaticReceiverType.(*NamedType).addMethod(v.Function)
		}
	}
	return v
}

func (v *VariableAccessExpr) resolve(res *Resolver, s *Scope) Node {
	// NOTE: Here we check whether this is actually a variable access or an enum member.
	if len(v.Name.moduleNames) > 0 {
		enumName, memberName := v.Name.Split()
		ident := s.GetIdent(enumName)
		if ident.Type == IDENT_TYPE {
			itype := ident.Value.(Type)
			if _, ok := itype.ActualType().(EnumType); ok {

				enum := &EnumLiteral{}
				enum.Member = memberName
				enum.Type = UnresolvedType{
					Name:       enumName,
					Parameters: v.parameters,
				}
				enum.Type = enum.Type.resolveType(v, res, s)
				enum.setPos(v.Pos())
				return enum
			}
		}
	}

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
	return v
}

func (v *VariableDecl) resolve(res *Resolver, s *Scope) Node {
	if v.Variable.Type != nil {
		v.Variable.Type = v.Variable.Type.resolveType(v, res, s)
	}
	return v
}

func (v *TypeDecl) resolve(res *Resolver, s *Scope) Node {
	if v.NamedType.Parameters == nil {
		v.NamedType.Type = v.NamedType.Type.resolveType(v, res, s)
	}
	return v
}

func (v *CastExpr) resolve(res *Resolver, s *Scope) Node {
	v.Type = v.Type.resolveType(v, res, s)
	return v
}

func (v *ArrayLenExpr) resolve(res *Resolver, s *Scope) Node {
	if v.Type != nil {
		v.Type = v.Type.resolveType(v, res, s)
	}
	return v
}

func (v *SizeofExpr) resolve(res *Resolver, s *Scope) Node {
	if v.Expr != nil {
		// NOTE: Here we recurse down any deref ops, to check whether we are dealing
		// with a variable getting dereferenced, or a pointer type.
		var inner Expr = v.Expr
		depth := 0

		for {
			if unaryExpr, ok := inner.(*UnaryExpr); ok && unaryExpr.Op == UNOP_DEREF {
				inner = unaryExpr.Expr
				depth++
				continue
			} else if vaExpr, ok := inner.(*VariableAccessExpr); ok {
				ident := s.GetIdent(vaExpr.Name)
				if ident.Type == IDENT_TYPE {
					// NOTE: If it turened out to be a pointer type we
					// reconstruct the type based on the stored pointer depth
					var newType Type = ident.Value.(Type)
					for i := 0; i < depth; i++ {
						newType = pointerTo(newType)
					}
					v.Type = newType
					v.Expr = nil
				}
			}
			break
		}
	}

	if v.Type != nil {
		v.Type = v.Type.resolveType(v, res, s)
	}
	return v
}

func (v *EnumLiteral) resolve(res *Resolver, s *Scope) Node {
	v.Type = v.Type.resolveType(v, res, s)
	return v
}

func (v *DefaultExpr) resolve(res *Resolver, s *Scope) Node {
	v.Type = v.Type.resolveType(v, res, s)
	return v
}

func (v *UseDecl) resolve(res *Resolver, s *Scope) Node {
	ident := s.GetIdent(v.ModuleName)
	if ident == nil {
		// TODO: Verify whether this case can ever happen
		res.errCannotResolve(v, v.ModuleName)
	} else if ident.Type != IDENT_MODULE {
		res.err(v, "Expected module name, found %s `%s`", ident.Type, v.ModuleName)
	}
	return v
}

func (v *StructLiteral) resolve(res *Resolver, s *Scope) Node {
	if v.InEnum {
		return v
	}

	// NOTE: Here we check if we are referencing an actual struct,
	// or the struct part of an enum type
	if name, ok := v.Type.(UnresolvedType); ok {
		enumName, memberName := name.Name.Split()
		if memberName != "" {
			ident := s.GetIdent(enumName)
			if ident.Type == IDENT_TYPE {
				itype := ident.Value.(Type)
				if _, ok := itype.ActualType().(EnumType); ok {
					enum := &EnumLiteral{}
					enum.Member = memberName
					enum.Type = itype
					enum.StructLiteral = v
					enum.StructLiteral.InEnum = true
					enum.setPos(v.Pos())
					return enum
				}
			}
		}
	}

	if v.Type != nil {
		v.Type = v.Type.resolveType(v, res, s)
	}
	return v
}

func (v *CallExpr) resolve(res *Resolver, s *Scope) Node {
	// NOTE: Here we check whether this is a call or an enum tuple lit.
	if vae, ok := v.functionSource.(*VariableAccessExpr); ok {
		if len(vae.Name.moduleNames) > 0 {
			enumName, memberName := vae.Name.Split()
			ident := s.GetIdent(enumName)
			if ident != nil && ident.Type == IDENT_TYPE {
				itype := ident.Value.(Type)
				if _, ok := itype.ActualType().(EnumType); ok {

					enum := &EnumLiteral{}
					enum.Member = memberName
					enum.Type = UnresolvedType{
						Name:       enumName,
						Parameters: v.parameters,
					}
					enum.Type = enum.Type.resolveType(v, res, s)
					enum.TupleLiteral = &TupleLiteral{Members: v.Arguments}
					enum.setPos(v.Pos())
					return enum
				}
			}
		}
	}

	// NOTE: Here we check whether this is a call or a cast
	if vae, ok := v.functionSource.(*VariableAccessExpr); ok {
		ident := s.GetIdent(vae.Name)
		if ident != nil && ident.Type == IDENT_TYPE {
			if len(v.Arguments) != 1 {
				res.err(v, "Casts must recieve exactly one argument")
			}

			cast := &CastExpr{}
			cast.Type = UnresolvedType{Name: vae.Name}
			cast.Type = cast.Type.resolveType(v, res, s)
			cast.Expr = v.Arguments[0]
			cast.setPos(v.Pos())
			return cast
		}
	}

	return v
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

func (v MutableReferenceType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return mutableReferenceTo(v.Referrer.resolveType(src, res, s))
}

func (v ConstantReferenceType) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	return constantReferenceTo(v.Referrer.resolveType(src, res, s))
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
