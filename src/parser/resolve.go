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
			if named, ok := TypeWithoutPointers(n.Function.Receiver.Variable.Type).(*NamedType); ok {
				named.addMethod(n.Function)
			}
		}

		v.popFunction()

	case *LambdaExpr:
		v.popFunction()

	case *DerefAccessExpr:
		if ce, ok := n.Expr.(*CastExpr); ok {
			*node = &CastExpr{Type: PointerTo(ce.Type), Expr: ce.Expr}
		} else if ptr, ok := n.Expr.GetType().(PointerType); ok {
			n.Type = ptr.Addressee
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
		if n.NamedType.Parameters == nil {
			n.NamedType.Type = v.ResolveType(n, n.NamedType.Type)
		}

	case *FunctionDecl:
		v.pushFunction(n.Function)

		n.Function.Type = v.ResolveType(n, n.Function.Type).(FunctionType)

		if n.Function.StaticReceiverType != nil {
			n.Function.StaticReceiverType = v.ResolveType(n, n.Function.StaticReceiverType)
			if checkReceiverType(v, n, n.Function.StaticReceiverType, "static receiver") {
				n.Function.StaticReceiverType.(*NamedType).addMethod(n.Function)
			}
		}

	case *VariableDecl:
		if n.Variable.Type != nil {
			n.Variable.Type = v.ResolveType(n, n.Variable.Type)
		}
		if v.curScope.InsertVariable(n.Variable, n.IsPublic()) != nil {
			v.err(n, "Illegal redeclaration of variable `%s`", n.Variable.Name)
		}

	// Expr

	case *LambdaExpr:
		v.pushFunction(n.Function)

		n.Function.Type = v.ResolveType(n, n.Function.Type).(FunctionType)

	case *CastExpr:
		n.Type = v.ResolveType(n, n.Type)

	case *ArrayLenExpr:
		if n.Type != nil {
			n.Type = v.ResolveType(n, n.Type)
		}

	case *EnumLiteral:
		n.Type = v.ResolveType(n, n.Type)

	case *DefaultExpr:
		n.Type = v.ResolveType(n, n.Type)

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
					enum.Type = UnresolvedType{
						Name:       enumName,
						Parameters: n.parameters,
					}
					enum.Type = v.ResolveType(n, enum.Type)
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
			*node = &FunctionAccessExpr{
				Function:   ident.Value.(*Function),
				parameters: n.parameters,
			}
			break
		} else if ident.Type == IDENT_VARIABLE {
			n.Variable = ident.Value.(*Variable)
		} else {
			v.err(n, "Expected variable identifier, found %s `%s`", ident.Type, n.Name)
		}

		if n.Variable.Type != nil {
			n.Variable.Type = v.ResolveType(n, n.Variable.Type)
		}

	case *SizeofExpr:
		// TODO: Check if we can clean this up
		if n.Expr != nil {
			// NOTE: Here we recurse down any deref ops, to check whether we are dealing
			// with a variable getting dereferenced, or a pointer type.
			var inner Expr = n.Expr
			depth := 0

			for {
				if unaryExpr, ok := inner.(*UnaryExpr); ok && unaryExpr.Op == UNOP_DEREF {
					inner = unaryExpr.Expr
					depth++
					continue
				} else if vaExpr, ok := inner.(*VariableAccessExpr); ok {
					ident := v.getIdent(vaExpr, vaExpr.Name)
					if ident.Type == IDENT_TYPE {
						// NOTE: If it turened out to be a pointer type we
						// reconstruct the type based on the stored pointer depth
						var newType Type = ident.Value.(Type)
						for i := 0; i < depth; i++ {
							newType = PointerTo(newType)
						}
						n.Type = newType
						n.Expr = nil
					}
				}
				break
			}
		}

		if n.Type != nil {
			n.Type = v.ResolveType(n, n.Type)
		}

	case *CompositeLiteral:
		// TODO: why is this here?
		if n.InEnum {
			break
		}

		// NOTE: Here we check if we are referencing an actual struct,
		// or the struct part of an enum type
		if name, ok := n.Type.(UnresolvedType); ok {
			enumName, memberName := name.Name.Split()
			if memberName != "" {
				ident := v.getIdent(n, enumName)
				if ident.Type == IDENT_TYPE {
					itype := ident.Value.(Type)
					if _, ok := itype.ActualType().(EnumType); ok {
						enum := &EnumLiteral{}
						enum.Member = memberName
						enum.Type = itype
						enum.CompositeLiteral = n
						enum.CompositeLiteral.InEnum = true
						enum.setPos(n.Pos())

						*node = enum
						break
					}
				}
			}
		}

		if n.Type != nil {
			n.Type = v.ResolveType(n, n.Type)
		}

	case *CallExpr:
		// NOTE: Here we check whether this is a call or an enum tuple lit.
		if vae, ok := n.Function.(*VariableAccessExpr); ok {
			if len(vae.Name.ModuleNames) > 0 {
				enumName, memberName := vae.Name.Split()
				ident := v.getIdent(n, enumName)
				if ident != nil && ident.Type == IDENT_TYPE {
					itype := ident.Value.(Type)
					if _, ok := itype.ActualType().(EnumType); ok {

						enum := &EnumLiteral{}
						enum.Member = memberName
						enum.Type = v.ResolveType(n, UnresolvedType{
							Name:       enumName,
							Parameters: n.parameters,
						})
						enum.TupleLiteral = &TupleLiteral{Members: n.Arguments}
						enum.setPos(n.Pos())

						*node = enum
						break
					}
				}
			}
		}

		// NOTE: Here we check whether this is a call or a cast
		if vae, ok := n.Function.(*VariableAccessExpr); ok {
			ident := v.getIdent(n, vae.Name)
			if ident != nil && ident.Type == IDENT_TYPE {
				if len(n.Arguments) != 1 {
					v.err(n, "Casts must recieve exactly one argument")
				}

				cast := &CastExpr{}
				cast.Type = v.ResolveType(n, UnresolvedType{Name: vae.Name})
				cast.Expr = n.Arguments[0]
				cast.setPos(n.Pos())

				*node = cast
				break
			}
		}

	// No-Ops
	case *Block, *DefaultMatchBranch, *UseDirective, *AssignStat, *BinopAssignStat,
		*BlockStat, *BreakStat, *CallStat, *DefaultStat, *DeferStat, *IfStat,
		*MatchStat, *LoopStat, *NextStat, *ReturnStat, *AddressOfExpr,
		*ArrayAccessExpr, *BinaryExpr, *DerefAccessExpr, *UnaryExpr,
		*StructAccessExpr, *TupleAccessExpr, *BoolLiteral,
		*NumericLiteral, *RuneLiteral, *StringLiteral, *TupleLiteral:
		break

	default:
		panic("INTERNAL ERROR: Unhandled node in resolve pass `" + reflect.TypeOf(n).String() + "`")
	}
}

func (v *Resolver) ResolveType(src Locatable, t Type) Type {
	switch t := t.(type) {
	case PrimitiveType, *NamedType, InterfaceType:
		return t

	case ArrayType:
		return ArrayOf(v.ResolveType(src, t.MemberType))

	case MutableReferenceType:
		return mutableReferenceTo(v.ResolveType(src, t.Referrer))

	case ConstantReferenceType:
		return constantReferenceTo(v.ResolveType(src, t.Referrer))

	case PointerType:
		return PointerTo(v.ResolveType(src, t.Addressee))

	case ParameterType:
		panic("INTERNAL ERROR: Tried to resolve type parameter early")

	case SubstitutionType:
		return t.Type

	case StructType:
		nt := StructType{
			Variables: make([]*VariableDecl, len(t.Variables)),
			attrs:     t.attrs,
		}

		v.EnterScope()
		for idx, vari := range t.Variables {
			nt.Variables[idx] = &VariableDecl{
				Variable: &Variable{
					Type:         vari.Variable.Type,
					Name:         vari.Variable.Name,
					Mutable:      vari.Variable.Mutable,
					Attrs:        vari.Variable.Attrs,
					FromStruct:   vari.Variable.FromStruct,
					ParentStruct: vari.Variable.ParentStruct,
					ParentModule: vari.Variable.ParentModule,
					IsParameter:  vari.Variable.IsParameter,
				},
				Assignment: vari.Assignment,
				docs:       vari.docs,
			}

			visitor := &ASTVisitor{Visitor: v}
			nt.Variables[idx] = visitor.Visit(Node(nt.Variables[idx])).(*VariableDecl)
		}
		v.ExitScope()

		return nt

	case TupleType:
		nt := TupleType{Members: make([]Type, len(t.Members))}

		for idx, mem := range t.Members {
			nt.Members[idx] = v.ResolveType(src, mem)
		}

		return nt

	case EnumType:
		nv := EnumType{
			Simple:  t.Simple,
			Members: make([]EnumTypeMember, len(t.Members)),
			attrs:   t.attrs,
		}

		for idx, mem := range t.Members {
			nv.Members[idx].Name = mem.Name
			nv.Members[idx].Tag = mem.Tag
			nv.Members[idx].Type = v.ResolveType(src, mem.Type)
		}

		return nv

	case FunctionType:
		nv := FunctionType{
			attrs:      t.attrs,
			IsVariadic: t.IsVariadic,
		}

		for _, par := range t.Parameters {
			nv.Parameters = append(nv.Parameters, v.ResolveType(src, par))
		}
		if t.Receiver != nil {
			nv.Receiver = v.ResolveType(src, t.Receiver)
			checkReceiverType(v, src, nv.Receiver, "receiver")
		}
		if t.Return != nil { // TODO can this ever be nil
			nv.Return = v.ResolveType(src, t.Return)
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
			if namedType, ok := typ.(*NamedType); ok && len(t.Parameters) > 0 {
				v.EnterScope()
				name := namedType.Name + "<"
				for idx, param := range namedType.Parameters {
					paramType := SubstitutionType{
						Name: param.Name,
						Type: v.ResolveType(src, t.Parameters[idx]),
					}
					v.curScope.InsertType(paramType, ident.Public)

					name += t.Parameters[idx].TypeName()
					if idx < len(namedType.Parameters)-1 {
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
			} else {
				typ = v.ResolveType(src, typ)
			}

			return typ
		}

		panic("unreachable")

	default:
		typeName := reflect.TypeOf(t).String()
		panic("INTERNAL ERROR: Unhandled type in resolve pass: " + typeName)
	}
}

//
// The following is preliminary work used for generics and the future redo of
// the type inference system.
//
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
	switch typ := typ.(type) {
	case StructType:
		for _, decl := range typ.Variables {
			dest = append(dest, decl.Variable.Type)
		}

	case *NamedType:
		dest = append(dest, typ.Type)

	case ArrayType:
		dest = append(dest, typ.MemberType)

	case PointerType:
		dest = append(dest, typ.Addressee)

	case TupleType:
		dest = append(dest, typ.Members...)

	case EnumType:
		for _, mem := range typ.Members {
			dest = append(dest, mem.Type)
		}

	case FunctionType:
		if typ.Receiver != nil {
			dest = append(dest, typ.Receiver)
		}
		dest = append(dest, typ.Parameters...)
		if typ.Return != nil { // TODO: can it ever be nil?
			dest = append(dest, typ.Return)
		}

	}
	return dest
}
