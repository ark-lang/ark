package parser

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
)

type UnresolvedName struct {
	ModuleNames []string
	Name        string
}

func (v UnresolvedName) String() string {
	ret := ""
	for _, mod := range v.ModuleNames {
		ret += mod + "::"
	}
	return ret + v.Name
}

func (v UnresolvedName) Split() (UnresolvedName, string) {
	if len(v.ModuleNames) > 0 {
		res := UnresolvedName{}
		res.ModuleNames = v.ModuleNames[:len(v.ModuleNames)-1]
		res.Name = v.ModuleNames[len(v.ModuleNames)-1]
		return res, v.Name
	} else {
		return UnresolvedName{}, ""
	}
}

type Resolver struct {
	modules       *ModuleLookup
	module        *Module
	cModule       *Module
	curSubmod     *Submodule
	functionStack []*Function
	curScope      *Scope
}

func (v *Resolver) pushFunction(fn *Function) {
	v.functionStack = append(v.functionStack, fn)
}

func (v *Resolver) popFunction() {
	v.functionStack = v.functionStack[:len(v.functionStack)-1]
}

func (v Resolver) currentFunction() *Function {
	if len(v.functionStack) == 0 {
		return nil
	}
	return v.functionStack[len(v.functionStack)-1]
}

func Resolve(mod *Module, mods *ModuleLookup) {
	if mod.resolved {
		return
	}
	mod.resolved = true

	res := &Resolver{
		modules: mods,
		module:  mod,
		cModule: &Module{
			Name:    &ModuleName{Parts: []string{"C"}},
			Parts:   make(map[string]*Submodule),
			Dirpath: "", // not really a path for this module
		},
	}

	res.cModule.ModScope = NewCScope(res.cModule)

	res.curScope = NewGlobalScope(mod)
	mod.ModScope = res.curScope

	// add a C module here which will contain all of the c bindings and what
	// not to keep everything separate
	mod.ModScope.UsedModules["C"] = res.cModule

	res.ResolveUsedModules()
	log.Timed("resolving module", mod.Name.String(), func() {
		res.ResolveTopLevelDecls()
		res.ResolveDescent()
	})
	res.module.ModScope.Dump(0)
}

func (v *Resolver) ResolveUsedModules() {
	for _, submod := range v.module.Parts {
		// TODO: Verify whether we need the outer scope
		submod.UseScope = newScope(nil, v.module, nil)

		for _, node := range submod.Nodes {
			switch node := node.(type) {
			case *UseDirective:
				// TODO: Propagate this down into the parser/constructor
				modName := ModuleNameFromUnresolvedName(node.ModuleName)
				usedMod, err := v.modules.Get(modName)
				if err == nil {
					Resolve(usedMod.Module, v.modules)
				} else {
					panic("INTERNAL ERROR: Used module not loaded")
				}
				submod.UseScope.UseModule(usedMod.Module)

			default:
				continue
			}
		}
	}
}

func (v *Resolver) ResolveTopLevelDecls() {
	modScope := v.module.ModScope

	for _, submod := range v.module.Parts {
		for _, node := range submod.Nodes {
			switch node := node.(type) {
			// TODO: We might need to do more that just insert this into the
			// scope at the current point.
			case *TypeDecl:
				if modScope.InsertType(node.NamedType, node.IsPublic()) != nil {
					v.err(node, "Illegal redeclaration of type `%s`", node.NamedType.Name)
				}

			case *FunctionDecl:
				if node.Function.Receiver == nil {
					scope := v.curScope
					if node.Function.Type.Attrs().Contains("c") {
						scope = v.cModule.ModScope
						node.SetPublic(true)
					}

					if scope.InsertFunction(node.Function, node.IsPublic()) != nil {
						v.err(node, "Illegal redeclaration of function `%s`", node.Function.Name)
					}
				}

			case *VariableDecl:
				if modScope.InsertVariable(node.Variable, node.IsPublic()) != nil {
					v.err(node, "Illegal redeclaration of variable `%s`", node.Variable.Name)
				}

			default:
				continue
			}
		}
	}
}

func (v *Resolver) ResolveDescent() {
	vis := NewASTVisitor(v)
	for _, submod := range v.module.Parts {
		// TODO: Remove if not needed
		v.curSubmod = submod
		vis.VisitSubmodule(submod)
	}
}

func (v *Resolver) err(thing Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Error("resolve", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Error("resolve", v.curSubmod.File.MarkPos(pos))

	os.Exit(util.EXIT_FAILURE_SEMANTIC)
}

func (v *Resolver) getIdent(loc Locatable, name UnresolvedName) *Ident {
	// TODO: Decide whether we should actually allow shadowing a module
	ident := v.curScope.GetIdent(name)
	if ident == nil {
		ident = v.curSubmod.UseScope.GetIdent(name)
	}

	if ident == nil {
		v.err(loc, "Cannot resolve `%s`", name.String())
		return nil
	}

	if !ident.Public && ident.Scope.Module != v.module {
		v.err(loc, "Cannot access private identifier `%s`", name)
	}

	// make sure lambda can't access variables of enclosing function
	if ident.Scope.Function != nil && v.currentFunction() != ident.Scope.Function {
		v.err(loc, "Cannot access local identifier `%s` from lambda", name)
	}

	return ident
}

func (v *Resolver) Visit(n *Node) bool {
	v.ResolveNode(n)
	return true
}

func (v *Resolver) PostVisit(node *Node) {
	switch n := (*node).(type) {
	case *FunctionDecl:
		// Store the method in the type of the reciever
		if n.Function.Type.Receiver != nil {
			if named, ok := TypeWithoutPointers(n.Function.Receiver.Variable.Type.BaseType).(*NamedType); ok {
				named.addMethod(n.Function)
			}
		}

		v.ExitScope()
		v.popFunction()

	case *LambdaExpr:
		v.popFunction()

	case *DerefAccessExpr:
		if ce, ok := n.Expr.(*CastExpr); ok {
			res := &CastExpr{Type: &TypeReference{BaseType: PointerTo(ce.Type)}, Expr: ce.Expr}
			res.setPos(n.Pos())
			*node = res
		}
	}
}

func (v *Resolver) EnterScope() {
	v.curScope = newScope(v.curScope, v.module, v.currentFunction())
}

func (v *Resolver) ExitScope() {
	if v.curScope.Outer == nil {
		panic("INTERNAL ERROR: Trying to exit highest scope")
	}
	v.curScope = v.curScope.Outer
}

// returns true if no error
func checkReceiverType(res *Resolver, loc Locatable, t Type, purpose string) bool {
	if named, ok := TypeWithoutPointers(t).(*NamedType); ok {
		if named.ParentModule != res.module {
			res.err(loc, "Cannot use type `%s` declared in module `%s` as %s",
				t.TypeName(), named.ParentModule.Name, purpose)
			return false
		}
	} else {
		res.err(loc, "Expected named type for %s, found `%s`", purpose, t.TypeName())
		return false
	}
	return true
}

func (v *Resolver) ResolveNode(node *Node) {
	// TODO: I'm pretty sure the way we do pointers to everything
	// mean that we don't actually need a Node pointer.

	switch n := (*node).(type) {
	case *TypeDecl:
		// Only resolve non-generic type, generic types will currently be
		// resolved when they are used, as the type parameters can only be
		// resolved when we know what they are.
		n.NamedType.Type = v.ResolveType(n, n.NamedType.Type)

	case *FunctionDecl:
		v.EnterScope()
		v.pushFunction(n.Function)
		for _, par := range n.Function.Type.GenericParameters {
			if v.curScope.InsertType(par, false) != nil {
				v.err(n, "Illegal redeclaration of generic type parameter `%s`", par.TypeName())
			}
		}

		n.Function.Type = v.ResolveType(n, n.Function.Type).(FunctionType)

		if n.Function.StaticReceiverType != nil {
			n.Function.StaticReceiverType = v.ResolveType(n, n.Function.StaticReceiverType)
			if checkReceiverType(v, n, n.Function.StaticReceiverType, "static receiver") {
				n.Function.StaticReceiverType.(*NamedType).addMethod(n.Function)
			}
		}

	case *VariableDecl:
		if n.Variable.Type != nil {
			n.Variable.Type = v.ResolveTypeReference(n, n.Variable.Type)
		}
		if v.curScope.InsertVariable(n.Variable, n.IsPublic()) != nil {
			v.err(n, "Illegal redeclaration of variable `%s`", n.Variable.Name)
		}

	// Expr

	case *LambdaExpr:
		v.pushFunction(n.Function)

		n.Function.Type = v.ResolveType(n, n.Function.Type).(FunctionType)

	case *CastExpr:
		n.Type = v.ResolveTypeReference(n, n.Type)

	case *ArrayLenExpr:
		if n.Type != nil {
			n.Type = v.ResolveType(n, n.Type)
		}

	case *EnumLiteral:
		n.Type = v.ResolveTypeReference(n, n.Type)

	case *VariableAccessExpr:
		// TODO: Check if we can clean this up
		// NOTE: Here we check whether this is actually a variable access or an enum member.
		if len(n.Name.ModuleNames) > 0 {
			enumName, memberName := n.Name.Split()
			ident := v.getIdent(n, enumName)
			if ident != nil && ident.Type == IDENT_TYPE {
				itype := ident.Value.(Type)
				if etype, ok := itype.ActualType().(EnumType); ok {
					if _, ok := etype.GetMember(memberName); !ok {
						v.err(n, "No such member in enum `%s`: `%s`", itype.TypeName(), memberName)
						break
					}

					enum := &EnumLiteral{}
					enum.Member = memberName
					enum.Type = &TypeReference{
						BaseType: UnresolvedType{
							Name: enumName,
						},
						GenericArguments: v.ResolveTypeReferences(n, n.GenericArguments),
					}
					enum.Type = v.ResolveTypeReference(n, enum.Type)
					enum.setPos(n.Pos())

					*node = enum
					break
				}
			}
		}

		ident := v.getIdent(n, n.Name)
		if ident == nil {
			// do nothing
		} else if ident.Type == IDENT_FUNCTION {
			fan := &FunctionAccessExpr{
				Function:         ident.Value.(*Function),
				GenericArguments: v.ResolveTypeReferences(n, n.GenericArguments),
			}
			fan.Function.Accesses = append(fan.Function.Accesses, fan)
			*node = fan
			(*node).setPos(n.Pos())
			break
		} else if ident.Type == IDENT_VARIABLE {
			n.Variable = ident.Value.(*Variable)
		} else {
			v.err(n, "Expected variable identifier, found %s `%s`", ident.Type, n.Name)
		}

		if n.Variable.Type != nil {
			n.Variable.Type = v.ResolveTypeReference(n, n.Variable.Type)
		}

	case *SizeofExpr:
		// TODO: Check if we can clean this up
		if n.Expr != nil {
			if typ, ok := v.exprToType(n.Expr); ok {
				n.Expr = nil
				n.Type = &TypeReference{BaseType: typ}
			}
		}

		if n.Type != nil {
			n.Type = v.ResolveTypeReference(n, n.Type)
		}

	case *CompositeLiteral:
		// TODO: why is this here?
		if n.InEnum {
			break
		}

		// NOTE: Here we check if we are referencing an actual struct,
		// or the struct part of an enum type
		if name, ok := n.Type.BaseType.(UnresolvedType); ok {
			enumName, memberName := name.Name.Split()
			if memberName != "" {
				ident := v.getIdent(n, enumName)
				if ident.Type == IDENT_TYPE {
					itype := ident.Value.(Type)
					if _, ok := itype.ActualType().(EnumType); ok {
						et := v.ResolveTypeReference(n, &TypeReference{
							BaseType: UnresolvedType{
								Name: enumName,
							},
							GenericArguments: v.ResolveTypeReferences(n, n.Type.GenericArguments),
						})

						member, ok := et.BaseType.ActualType().(EnumType).GetMember(memberName)
						if !ok {
							v.err(n, "Enum `%s` has no member `%s`", enumName.String(), memberName)
						}

						// TODO sort out all the type duplication
						enum := &EnumLiteral{}
						enum.Member = memberName
						enum.Type = &TypeReference{BaseType: itype, GenericArguments: et.GenericArguments} // TODO should this be `et`?
						enum.CompositeLiteral = n
						enum.CompositeLiteral.Type = &TypeReference{BaseType: member.Type, GenericArguments: et.GenericArguments}
						enum.CompositeLiteral.InEnum = true
						enum.setPos(n.Pos())

						*node = enum
						break
					}
				}
			}
		}

		if n.Type != nil {
			n.Type = v.ResolveTypeReference(n, n.Type)
		}

	case *CallExpr:
		// NOTE: Here we check whether this is a call or an enum tuple lit.
		// way too much duplication with all this enum literal creating stuff
		if vae, ok := n.Function.(*VariableAccessExpr); ok {
			if len(vae.Name.ModuleNames) > 0 {
				enumName, memberName := vae.Name.Split()
				ident := v.getIdent(n, enumName)
				if ident != nil && ident.Type == IDENT_TYPE {
					itype := ident.Value.(Type)
					if _, ok := itype.ActualType().(EnumType); ok {
						et := v.ResolveTypeReference(n, &TypeReference{
							BaseType: UnresolvedType{
								Name: enumName,
							},
							GenericArguments: v.ResolveTypeReferences(vae, vae.GenericArguments),
						})

						member, ok := et.BaseType.ActualType().(EnumType).GetMember(memberName)
						if !ok {
							v.err(n, "Enum `%s` has no member `%s`", enumName.String(), memberName)
						}

						enum := &EnumLiteral{}
						enum.Member = memberName
						enum.Type = et
						enum.TupleLiteral = &TupleLiteral{
							Members:           n.Arguments,
							Type:              &TypeReference{BaseType: member.Type, GenericArguments: et.GenericArguments},
							ParentEnumLiteral: enum,
						}
						enum.TupleLiteral.setPos(n.Pos())
						enum.setPos(n.Pos())

						*node = enum
						break
					}
				}
			}
		}

		// NOTE: Here we check whether this is a call or a cast
		// Unwrap any deref access expressions as these might signify pointer types

		if typ, ok := v.exprToType(n.Function); ok {
			if len(n.Arguments) != 1 {
				v.err(n, "Casts must recieve exactly one argument")
			}

			cast := &CastExpr{}
			cast.Type = &TypeReference{BaseType: typ}
			cast.Expr = n.Arguments[0]
			cast.setPos(n.Pos())
			*node = cast
		}

	// No-Ops
	case *Block, *DefaultMatchBranch, *UseDirective, *AssignStat, *BinopAssignStat,
		*BlockStat, *BreakStat, *CallStat, *DeferStat, *IfStat, *MatchStat,
		*LoopStat, *NextStat, *ReturnStat, *AddressOfExpr, *ArrayAccessExpr,
		*BinaryExpr, *DerefAccessExpr, *UnaryExpr, *StructAccessExpr,
		*BoolLiteral, *NumericLiteral, *RuneLiteral, *StringLiteral,
		*TupleLiteral:
		break

	default:
		panic("INTERNAL ERROR: Unhandled node in resolve pass `" + reflect.TypeOf(n).String() + "`")
	}
}

func (v *Resolver) exprToType(expr Expr) (Type, bool) {
	derefs := 0
	for {
		if dae, ok := expr.(*DerefAccessExpr); ok {
			derefs++
			expr = dae.Expr
		} else {
			break
		}
	}

	if vae, ok := expr.(*VariableAccessExpr); ok {
		ident := v.getIdent(vae, vae.Name)
		if ident != nil && ident.Type == IDENT_TYPE {
			res := ident.Value.(Type)
			for i := 0; i < derefs; i++ {
				res = PointerTo(&TypeReference{BaseType: res})
			}
			return res, true
		}
	}

	return nil, false
}

func (v *Resolver) ResolveTypes(src Locatable, ts []Type) []Type {
	res := make([]Type, 0, len(ts))
	for _, t := range ts {
		res = append(res, v.ResolveType(src, t))
	}
	return res
}

func (v *Resolver) ResolveTypeReferences(src Locatable, ts []*TypeReference) []*TypeReference {
	res := make([]*TypeReference, 0, len(ts))
	for _, t := range ts {
		res = append(res, v.ResolveTypeReference(src, t))
	}
	return res
}

func (v *Resolver) ResolveTypeReference(src Locatable, t *TypeReference) *TypeReference {
	return &TypeReference{
		BaseType:         v.ResolveType(src, t.BaseType),
		GenericArguments: v.ResolveTypeReferences(src, t.GenericArguments),
	}
}

func (v *Resolver) ResolveType(src Locatable, t Type) Type {
	switch t := t.(type) {
	case PrimitiveType, *NamedType, InterfaceType:
		return t

	case ArrayType:
		return ArrayOf(v.ResolveTypeReference(src, t.MemberType))

	case ReferenceType:
		return ReferenceTo(v.ResolveTypeReference(src, t.Referrer), t.IsMutable)

	case PointerType:
		return PointerTo(v.ResolveTypeReference(src, t.Addressee))

	case *SubstitutionType:
		return t

	case StructType:
		v.EnterScope()

		for _, gpar := range t.GenericParameters {
			v.curScope.InsertType(gpar, false)
		}

		nt := StructType{
			Members:           make([]*StructMember, len(t.Members)),
			attrs:             t.attrs,
			GenericParameters: t.GenericParameters,
		}

		v.EnterScope()
		for idx, mem := range t.Members {
			nt.Members[idx] = &StructMember{
				Name: mem.Name,
				Type: v.ResolveTypeReference(src, mem.Type),
			}
		}
		v.ExitScope()

		v.ExitScope()

		return nt

	case TupleType:
		nt := TupleType{Members: make([]*TypeReference, len(t.Members))}

		for idx, mem := range t.Members {
			nt.Members[idx] = v.ResolveTypeReference(src, mem)
		}

		return nt

	case EnumType:
		v.EnterScope()

		for _, gpar := range t.GenericParameters {
			v.curScope.InsertType(gpar, false)
		}

		nv := EnumType{
			Simple:            t.Simple,
			Members:           make([]EnumTypeMember, len(t.Members)),
			attrs:             t.attrs,
			GenericParameters: t.GenericParameters,
		}

		for idx, mem := range t.Members {
			nv.Members[idx].Name = mem.Name
			nv.Members[idx].Tag = mem.Tag
			nv.Members[idx].Type = v.ResolveType(src, mem.Type)
		}

		v.ExitScope()

		return nv

	case FunctionType:
		nv := FunctionType{
			attrs:             t.attrs,
			IsVariadic:        t.IsVariadic,
			Parameters:        v.ResolveTypeReferences(src, t.Parameters),
			GenericParameters: t.GenericParameters,
		}

		if t.Receiver != nil {
			nv.Receiver = v.ResolveType(src, t.Receiver)
			checkReceiverType(v, src, nv.Receiver, "receiver")
		}
		if t.Return != nil { // TODO can this ever be nil
			nv.Return = v.ResolveTypeReference(src, t.Return)
		}

		return nv

	case UnresolvedType:
		ident := v.getIdent(src, t.Name)
		if ident == nil {
			// do nothing
		} else if ident.Type != IDENT_TYPE {
			v.err(src, "Expected type identifier, found %s `%s`", ident.Type, t.Name)
		} else {
			typ := ident.Value.(Type)

			// TODO what is this stuff?
			/*if namedType, ok := typ.(*NamedType); ok && len(t.GenericParameters) > 0 {
				v.EnterScope()
				name := namedType.Name + "<"
				for idx, param := range namedType.GenericParameters {
					paramType := *SubstitutionType{
						Name: param,
						Type: v.ResolveType(src, t.GenericParameters[idx]),
					}
					v.curScope.InsertType(paramType, ident.Public)

					name += paramType.Type.TypeName()
					if idx < len(namedType.GenericParameters)-1 {
						name += ", "
					}
				}
				name += ">"

				typ = &NamedType{
					Name:         name,
					Type:         v.ResolveType(src, namedType.Type),
					ParentModule: namedType.ParentModule,
					Methods:      namedType.Methods,
				}
				v.ExitScope()
			} else {*/
			typ = v.ResolveType(src, typ)
			//}

			return typ
		}

		panic("unreachable")

	default:
		typeName := reflect.TypeOf(t).String()
		panic("INTERNAL ERROR: Unhandled type in resolve pass: " + typeName)
	}
}
