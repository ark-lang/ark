package parser

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
)

// TypeVariable is a type that abstracts the notion of a type variable such
// that we can use our existing types as part of constraints.
type TypeVariable struct {
	metaType
	Id int
}

func (v TypeVariable) Equals(other Type) bool {
	if ot, ok := other.(TypeVariable); ok {
		return v.Id == ot.Id
	}
	return false
}

func (v TypeVariable) String() string {
	return NewASTStringer("TypeVariable").AddType(v).Finish()
}

func (v TypeVariable) TypeName() string {
	return fmt.Sprintf("$%d", v.Id)
}

func (v TypeVariable) ActualType() Type {
	return v
}

// ConstructorType is an abstraction that in principle could represent any type
// that is built from other types. As we can use the actual types for most of
// these, this type is only used to represent the type of a struct member or,
// until removal, the type of tuple member by index.
type ConstructorType struct {
	metaType
	Id   ConstructorId
	Args []*TypeReference

	// Some constructors need additional data
	Data interface{}
}

type ConstructorId int

const (
	ConstructorInvalid ConstructorId = iota
	ConstructorStructMember
	ConstructorDeref
	ConstructorArrayIndex
)

func (v *ConstructorType) Equals(other Type) bool {
	if ot, ok := other.(*ConstructorType); ok {
		if v.Id != ot.Id {
			return false
		}

		if v.Data != ot.Data {
			return false
		}

		if len(v.Args) != len(ot.Args) {
			return false
		}

		for idx, arg := range v.Args {
			oarg := ot.Args[idx]
			if !arg.Equals(oarg) {
				return false
			}
		}

		return true
	}
	return false
}

func (v *ConstructorType) String() string {
	return NewASTStringer("ConstructorType").AddType(v).Finish()
}

func (v *ConstructorType) TypeName() string {
	return fmt.Sprintf("C%d(%v).%v", v.Id, v.Args, v.Data)
}

func (v *ConstructorType) ActualType() Type {
	return v
}

// Constraint represents a single constraint to be solved.
// It consists of two "sides", each representing a type or a type variable.
type Constraint struct {
	Left, Right Side
}

func ConstraintFromTypes(left, right *TypeReference) *Constraint {
	return &Constraint{
		Left:  SideFromType(left),
		Right: SideFromType(right),
	}
}

func (v *Constraint) String() string {
	return fmt.Sprintf("%s = %s", v.Left, v.Right)
}

func (v *Constraint) Subs(id int, side Side) *Constraint {
	res := &Constraint{
		Left:  v.Left.Subs(id, side),
		Right: v.Right.Subs(id, side),
	}
	return res
}

type SideType int

const (
	IdentSide SideType = iota
	TypeSide
)

// Side represents a single side of a constraint.
// It represents either a type (TypeSide) or a type variable (IdentSide)
type Side struct {
	SideType SideType
	Id       int
	Type     *TypeReference
}

// SideFromType creates a new Side from the given type.
// If the given type is a TypeVariable an IdentSide will be created, otherwise
// a TypeSide will be created.
func SideFromType(t *TypeReference) Side {
	if tv, ok := t.BaseType.(TypeVariable); ok {
		return Side{SideType: IdentSide, Id: tv.Id}
	}
	return Side{SideType: TypeSide, Type: t}
}

// Subs descends through the given Side, and replaces all occurenes of the
// given id with the contents of the Side `what`.
func (v Side) Subs(id int, what Side) Side {
	switch v.SideType {
	// If this is an IdentSide we check if the id matches and return the
	// replacement side in case of a match.
	case IdentSide:
		if v.Id == id {
			return what
		}
		return v

	// If this is a TypeSide we create a type from the `what` side,
	// and then delegate the substitution to `SubsType`
	case TypeSide:
		var nt *TypeReference
		if what.SideType == TypeSide {
			nt = SubsType(v.Type, id, what.Type)
		} else {
			nt = SubsType(v.Type, id, &TypeReference{BaseType: TypeVariable{Id: what.Id}})
		}
		return Side{SideType: TypeSide, Type: nt}

	default:
		panic("Invalid SideType")
	}
}

// SubsType descends through a type and replaces all occurences of the given
// type variable by `what`
func SubsType(typ *TypeReference, id int, what *TypeReference) *TypeReference {
	switch t := typ.BaseType.(type) {
	case TypeVariable:
		if t.Id == id {
			return what
		}
		return typ

	case *ConstructorType:
		// Descend through all arguments
		nargs := make([]*TypeReference, len(t.Args))
		for idx, arg := range t.Args {
			nargs[idx] = SubsType(arg, id, what)
		}

		// Handle special cases
		switch t.Id {
		// If we have a struct member, we check whether we can resolve the
		// actual type of the member with the information we have at the
		// current point. If we do, we return the actual type.
		case ConstructorStructMember:
			// Method check
			fn := GetMethod(nargs[0].BaseType, t.Data.(string))
			if fn != nil {
				return &TypeReference{
					BaseType:         fn.Type,
					GenericArguments: typ.GenericArguments,
				}
			}

			// Struct member
			typ := nargs[0]
			if pt, ok := typ.BaseType.(PointerType); ok {
				typ = pt.Addressee
			}
			if st, ok := typ.BaseType.ActualType().(StructType); ok {
				mem := st.GetMember(t.Data.(string))
				if mem != nil {
					mtype := mem.Type
					if len(typ.GenericArguments) > 0 {
						gn := NewGenericContextFromTypeReference(typ)
						mtype = gn.Replace(mtype)
					}

					return mtype
				}
			}

		// If we have a deref member we check if we know the pointer type and
		// if we do we pull out the target type
		case ConstructorDeref:
			adressee := getAdressee(nargs[0].BaseType)
			if adressee != nil {
				if len(typ.GenericArguments) > 0 {
					gn := NewGenericContextFromTypeReference(typ)
					adressee = gn.Replace(adressee)
				}
				return adressee
			}

			// If we have a array member we check if we know the array type and if
		// we do we pull out the member type
		case ConstructorArrayIndex:
			array := nargs[0].BaseType
			if at, ok := array.ActualType().(ArrayType); ok {
				mt := at.MemberType
				if len(typ.GenericArguments) > 0 {
					gn := NewGenericContextFromTypeReference(typ)
					mt = gn.Replace(mt)
				}
				return mt
			}
		}

		return &TypeReference{
			BaseType:         &ConstructorType{Id: t.Id, Args: nargs, Data: t.Data},
			GenericArguments: typ.GenericArguments,
		}

	case FunctionType:
		// Descend into return type
		newRet := SubsType(t.Return, id, what)

		// Descend into parameter types
		np := make([]*TypeReference, len(t.Parameters))
		for idx, param := range t.Parameters {
			np[idx] = SubsType(param, id, what)
		}

		return &TypeReference{
			BaseType: FunctionType{
				attrs:      t.attrs,
				IsVariadic: t.IsVariadic,
				Parameters: np,
				Return:     newRet,
			},
			GenericArguments: typ.GenericArguments,
		}

	case TupleType:
		// Descend into member types
		nm := make([]*TypeReference, len(t.Members))
		for idx, mem := range t.Members {
			nm[idx] = SubsType(mem, id, what)
		}

		return &TypeReference{
			BaseType:         tupleOf(nm...),
			GenericArguments: typ.GenericArguments,
		}

	case ArrayType:
		return &TypeReference{
			BaseType:         ArrayOf(SubsType(t.MemberType, id, what), t.IsFixedLength, t.Length),
			GenericArguments: typ.GenericArguments,
		}

	case PointerType:
		return &TypeReference{
			BaseType:         PointerTo(SubsType(t.Addressee, id, what), t.IsMutable),
			GenericArguments: typ.GenericArguments,
		}

	case ReferenceType:
		return &TypeReference{
			BaseType:         ReferenceTo(SubsType(t.Referrer, id, what), t.IsMutable),
			GenericArguments: typ.GenericArguments,
		}

	// The following are noops at the current time. For NamedType and EnumType
	// this is only temporary, until we finalize implementaiton of generics
	// in a solid maintainable way.
	case PrimitiveType, StructType, *NamedType, InterfaceType, EnumType, *SubstitutionType:
		return typ

	default:
		panic("Unhandled type in Side.Subs(): " + reflect.TypeOf(t).String() + " (" + t.TypeName() + ")")
	}
}

func GetMethod(typ Type, name string) *Function {
	typNp := TypeWithoutPointers(typ)
	if it, ok := typNp.ActualType().(InterfaceType); ok {
		typNp = it
	}

	var fn *Function
	switch t := typNp.(type) {
	case *NamedType:
		fn = t.GetMethod(name)

	case InterfaceType:
		fn = t.GetFunction(name)

	case *SubstitutionType:
		var ifn *Function
		for _, con := range t.Constraints {
			ifn = GetMethod(con, name)
			if ifn != nil {
				break
			}
		}
		fn = ifn
	}

	return fn
}

func (v Side) String() string {
	switch v.SideType {
	case IdentSide:
		return fmt.Sprintf("$%d", v.Id)
	case TypeSide:
		return fmt.Sprintf("type `%s`", v.Type.String())
	}
	panic("Invalid side type")
}

type AnnotatedTyped struct {
	Pos   lexer.Position
	Typed Typed
	Id    int
}

type Inferrer struct {
	Submodule         *Submodule
	Functions         []*Function
	Typeds            map[int]*AnnotatedTyped
	TypedLookup       map[Typed]*AnnotatedTyped
	SimpleConstraints []*Constraint
	Constraints       []*Constraint
	IdCount           int
}

func (v *Inferrer) err(msg string, args ...interface{}) {
	log.Errorln("inferrer", "%s %s", util.Red("error:"), fmt.Sprintf(msg, args...))
	os.Exit(util.EXIT_FAILURE_SEMANTIC)
}

func (v *Inferrer) errPos(pos lexer.Position, msg string, args ...interface{}) {
	log.Errorln("inferrer", "%s: [%s:%d:%d] %s", util.Red("error:"),
		pos.Filename, pos.Line, pos.Char,
		fmt.Sprintf(msg, args...))
	log.Errorln("inferrer", "%s", v.Submodule.File.MarkPos(pos))
	os.Exit(util.EXIT_FAILURE_SEMANTIC)
}

func (v *Inferrer) Function() *Function {
	return v.Functions[len(v.Functions)-1]
}

func Infer(submod *Submodule) {
	if submod.inferred {
		return
	}
	submod.inferred = true

	for _, used := range submod.UseScope.UsedModules {
		for _, submod := range used.Parts {
			Infer(submod)
		}
	}

	log.Timed("inferring submodule", submod.File.Name, func() {
		inf := &Inferrer{
			Submodule:   submod,
			Typeds:      make(map[int]*AnnotatedTyped),
			TypedLookup: make(map[Typed]*AnnotatedTyped),
		}
		vis := NewASTVisitor(inf)
		vis.VisitSubmodule(submod)
		inf.Finalize()
	})

}

func (v *Inferrer) AddConstraint(c *Constraint) {
	v.Constraints = append(v.Constraints, c)
}

// AddEqualsConstraint creates a constraint that indicates that the two given
// ids are equal to one-another and add it to the list of constraints.
func (v *Inferrer) AddEqualsConstraint(a, b int) {
	c := &Constraint{
		Left:  Side{Id: a, SideType: IdentSide},
		Right: Side{Id: b, SideType: IdentSide},
	}
	v.AddConstraint(c)
}

// AddIsConstraint creates a constraing that indicates that the given id is of
// the given type and add it to the list of constraints.
func (v *Inferrer) AddIsConstraint(id int, typref *TypeReference) {
	c := &Constraint{
		Left:  Side{Id: id, SideType: IdentSide},
		Right: Side{Type: typref, SideType: TypeSide},
	}
	v.AddConstraint(c)
}

// AddSimpleIsConstraint creates and adds a constraint to the inferrer, where
// the type given is guaranteed not to contain a type variable.
func (v *Inferrer) AddSimpleIsConstraint(id int, typref *TypeReference) {
	c := &Constraint{
		Left:  Side{Id: id, SideType: IdentSide},
		Right: Side{Type: typref, SideType: TypeSide},
	}
	v.SimpleConstraints = append(v.SimpleConstraints, c)
}

func (v *Inferrer) EnterScope() {}

func (v *Inferrer) ExitScope() {}

func (v *Inferrer) PostVisit(node *Node) {
	switch (*node).(type) {
	case *FunctionDecl, *LambdaExpr:
		idx := len(v.Functions) - 1
		v.Functions[idx] = nil
		v.Functions = v.Functions[:idx]
		return
	}
}

func (v *Inferrer) Visit(node *Node) bool {
	switch n := (*node).(type) {
	case *FunctionDecl:
		v.Functions = append(v.Functions, n.Function)
		return true

	case *LambdaExpr:
		v.Functions = append(v.Functions, n.Function)
		return true
	}

	// Switch on the type of a node. If it is a variable declaration, or a
	// statement that contains an expression it should be in here.
	switch n := (*node).(type) {
	case *VariableDecl:
		if n.Assignment != nil {
			if n.Variable.Type != nil {
				n.Assignment.SetType(n.Variable.Type)
			} else {
				n.Variable.SetType(n.Assignment.GetType())
			}
			aid := v.HandleExpr(n.Assignment)
			vid := v.HandleTyped(n.Pos(), n.Variable)
			v.AddEqualsConstraint(vid, aid)
		}

	case *DestructVarDecl:
		id := v.HandleExpr(n.Assignment)
		if n.Assignment.GetType() != nil {
			if tt, ok := n.Assignment.GetType().BaseType.ActualType().(TupleType); ok {
				for idx, vari := range n.Variables {
					if !n.ShouldDiscard[idx] {
						vari.SetType(tt.Members[idx])
					}
				}
				break
			}
		}

		ids := make([]*TypeReference, len(n.Variables))
		for idx, vari := range n.Variables {
			var vid int
			if n.ShouldDiscard[idx] {
				vid = v.GetDiscardingId()
			} else {
				vid = v.HandleTyped(n.Pos(), vari)
			}
			ids[idx] = &TypeReference{BaseType: TypeVariable{Id: vid}}
		}
		v.AddIsConstraint(id, &TypeReference{BaseType: tupleOf(ids...)})

	case *AssignStat:
		a := v.HandleExpr(n.Access)
		b := v.HandleExpr(n.Assignment)
		if n.Access.GetType() != nil {
			v.AddSimpleIsConstraint(b, n.Access.GetType())
		} else {
			v.AddEqualsConstraint(a, b)
		}

	case *BinopAssignStat:
		a := v.HandleExpr(n.Access)
		b := v.HandleExpr(n.Assignment)
		if n.Access.GetType() != nil {
			v.AddSimpleIsConstraint(b, n.Access.GetType())
		} else {
			v.AddEqualsConstraint(a, b)
		}

	case *DestructAssignStat:
		// If we're dealing with a raw tuple literal we have to patch up the
		// types due to the whole default integer type thing
		tl, ok := n.Assignment.(*TupleLiteral)
		if ok {
			for idx, acc := range n.Accesses {
				if acc.GetType() != nil {
					tl.Members[idx].SetType(acc.GetType())
				}
			}
		}

		assId := v.HandleExpr(n.Assignment)
		accIds := make([]*TypeReference, len(n.Accesses))
		for idx, acc := range n.Accesses {
			id := v.HandleExpr(acc)
			if acc.GetType() != nil {
				accIds[idx] = acc.GetType()
			} else {
				accIds[idx] = &TypeReference{BaseType: TypeVariable{Id: id}}
			}
		}
		v.AddIsConstraint(assId, &TypeReference{
			BaseType: tupleOf(accIds...),
		})

	case *DestructBinopAssignStat:
		// If we're dealing with a raw tuple literal we have to patch up the
		// types due to the whole default integer type thing
		tl, ok := n.Assignment.(*TupleLiteral)
		if ok {
			for idx, acc := range n.Accesses {
				if acc.GetType() != nil {
					tl.Members[idx].SetType(acc.GetType())
				}
			}
		}

		assId := v.HandleExpr(n.Assignment)
		accIds := make([]*TypeReference, len(n.Accesses))
		for idx, acc := range n.Accesses {
			id := v.HandleExpr(acc)
			if acc.GetType() != nil {
				accIds[idx] = acc.GetType()
			} else {
				accIds[idx] = &TypeReference{BaseType: TypeVariable{Id: id}}
			}
		}
		v.AddIsConstraint(assId, &TypeReference{
			BaseType: tupleOf(accIds...),
		})

	case *CallStat:
		v.HandleExpr(n.Call)

	case *DeferStat:
		v.HandleExpr(n.Call)

	case *IfStat:
		for _, expr := range n.Exprs {
			id := v.HandleExpr(expr)
			v.AddSimpleIsConstraint(id, &TypeReference{BaseType: PRIMITIVE_bool})
		}

	case *ReturnStat:
		if n.Value != nil {
			id := v.HandleExpr(n.Value)
			v.AddSimpleIsConstraint(id, v.Function().Type.Return)
		}

	case *LoopStat:
		if n.Condition != nil {
			id := v.HandleExpr(n.Condition)
			v.AddSimpleIsConstraint(id, &TypeReference{BaseType: PRIMITIVE_bool})
		}

	case *MatchStat:
		// TODO: Implement once we actuall do match statement

	}

	return true
}

func (v *Inferrer) GetDiscardingId() int {
	id := v.IdCount
	v.IdCount++
	return id
}

func (v *Inferrer) HandleExpr(expr Expr) int {
	return v.HandleTyped(expr.Pos(), expr)
}

func (v *Inferrer) HandleTyped(pos lexer.Position, typed Typed) int {
	// If we have already handled this type, return now.
	if ann, ok := v.TypedLookup[typed]; ok {
		return ann.Id
	}

	// Wrap and store the typed so we can access it later
	ann := &AnnotatedTyped{Pos: pos, Id: v.IdCount, Typed: typed}
	v.Typeds[ann.Id] = ann
	v.TypedLookup[typed] = ann
	v.IdCount++

	// Switch on the type of the typed. If it is a `Variable`, any expression,
	// or a literal of some sort, it should be handled here.
	switch typed := typed.(type) {
	case *BinaryExpr:
		a := v.HandleExpr(typed.Lhand)
		b := v.HandleExpr(typed.Rhand)
		switch typed.Op.Category() {

		// If we're dealing with a comparison operation, we know that both
		// sides must be of the same type, and that the result will be a bool
		case OP_COMPARISON:
			if typed.Lhand.GetType() == nil || typed.Rhand.GetType() == nil {
				v.AddEqualsConstraint(a, b)
			}
			v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: PRIMITIVE_bool})

		// If we're dealing with bitwise operations we know that both sides
		// must be the same type, and that the result will be of that type
		// aswell.
		case OP_BITWISE:
			if typed.Lhand.GetType() != nil && typed.Rhand.GetType() != nil {
				v.AddSimpleIsConstraint(ann.Id, typed.Lhand.GetType())
			} else {
				v.AddEqualsConstraint(a, b)
				v.AddEqualsConstraint(ann.Id, a)
			}

		// If we're dealing with an arithmetic operation we know that both
		// sides must be of the same type, and that the result will be of that
		// type aswell.
		// TODO: These assumptions don't hold once we add operator overloading
		case OP_ARITHMETIC:
			if typed.Lhand.GetType() != nil && typed.Rhand.GetType() != nil {
				v.AddSimpleIsConstraint(ann.Id, typed.Lhand.GetType())
			} else {
				v.AddEqualsConstraint(a, b)
				v.AddEqualsConstraint(ann.Id, a)
			}

		// If we're dealing with a logical operation, we know that both sides
		// must be booleans, and that the result will also be a boolean.
		case OP_LOGICAL:
			v.AddSimpleIsConstraint(a, &TypeReference{BaseType: PRIMITIVE_bool})
			v.AddSimpleIsConstraint(b, &TypeReference{BaseType: PRIMITIVE_bool})
			v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: PRIMITIVE_bool})

		default:
			panic("Unhandled binary operator in type inference")

		}

	case *UnaryExpr:
		id := v.HandleExpr(typed.Expr)
		switch typed.Op {
		// If we're dealing with a logical not the expression being not'ed must
		// be a boolean, and the resul will also be a boolean.
		case UNOP_LOG_NOT:
			v.AddSimpleIsConstraint(id, &TypeReference{BaseType: PRIMITIVE_bool})
			v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: PRIMITIVE_bool})

		// If we're dealing with a bitwise not, the type will be the same type
		// as the expression acted upon.
		case UNOP_BIT_NOT:
			v.AddEqualsConstraint(ann.Id, id)

		// If we're dealing with a arithmetic negation, the type will be the
		// same type as the expression acted upon.
		case UNOP_NEGATIVE:
			v.AddEqualsConstraint(ann.Id, id)

		}

	case *CallExpr:
		fnId := v.HandleExpr(typed.Function)
		if typed.Function.GetType() != nil {
			ft, ok := typed.Function.GetType().BaseType.ActualType().(FunctionType)
			if ok && len(ft.GenericParameters) == 0 {
				for idx, arg := range typed.Arguments {
					id := v.HandleExpr(arg)
					if idx >= len(ft.Parameters) {
						continue
					}
					v.AddSimpleIsConstraint(id, ft.Parameters[idx])
				}
				v.AddSimpleIsConstraint(ann.Id, ft.Return)
				break
			}
		}

		// TODO generic arguments
		if typed.ReceiverAccess != nil {
			v.HandleExpr(typed.ReceiverAccess)
		}

		argIds := make([]int, len(typed.Arguments))
		for idx, arg := range typed.Arguments {
			argIds[idx] = v.HandleExpr(arg)
		}

		// Construct a function type containing the generated type variables.
		// This will be used to infer the types of the arguments.
		fnType := FunctionType{Return: &TypeReference{BaseType: TypeVariable{Id: ann.Id}}}
		for _, argId := range argIds {
			fnType.Parameters = append(fnType.Parameters, &TypeReference{BaseType: TypeVariable{Id: argId}})
		}
		v.AddIsConstraint(fnId, &TypeReference{BaseType: fnType})

	// The type of a cast will always be the type casted to.
	case *CastExpr:
		v.HandleExpr(typed.Expr)
		v.AddSimpleIsConstraint(ann.Id, typed.Type)

	// Given an reference-to expr or a pointer-to expr, we know that the result
	// will be a pointer to the type of the access of which we took the address
	case *ReferenceToExpr:
		id := v.HandleExpr(typed.Access)
		if typed.Access.GetType() != nil {
			v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: ReferenceTo(typed.Access.GetType(), typed.IsMutable)})
		}
		v.AddIsConstraint(ann.Id, &TypeReference{BaseType: ReferenceTo(&TypeReference{BaseType: TypeVariable{Id: id}}, typed.IsMutable)})

	case *PointerToExpr:
		id := v.HandleExpr(typed.Access)
		if typed.Access.GetType() != nil {
			v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: PointerTo(typed.Access.GetType(), typed.IsMutable)})
		}
		v.AddIsConstraint(ann.Id, &TypeReference{BaseType: PointerTo(&TypeReference{BaseType: TypeVariable{Id: id}}, typed.IsMutable)})

	// Given a deref, we generate a constructor type as inferring the the types
	// while maintaining the mutablility stuff is a pain.
	case *DerefAccessExpr:
		id := v.HandleExpr(typed.Expr)
		if typed.Expr.GetType() != nil {
			addressee := getAdressee(typed.Expr.GetType().BaseType.ActualType())
			if addressee != nil {
				v.AddSimpleIsConstraint(ann.Id, addressee)
				break
			}
		}
		v.AddIsConstraint(ann.Id, &TypeReference{
			BaseType: &ConstructorType{
				Id: ConstructorDeref,
				Args: []*TypeReference{
					&TypeReference{BaseType: TypeVariable{Id: id}},
				},
			},
		})

	// A sizeof expr always return a uint
	case *SizeofExpr:
		if typed.Expr != nil {
			v.HandleExpr(typed.Expr)
		}
		v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: PRIMITIVE_uint})

	// Given a variable access, we know that the type of the access must be
	// equal to the type of the variable being accessed.
	case *VariableAccessExpr:
		id := v.HandleTyped(typed.Pos(), typed.Variable)
		if typed.Variable.Type != nil {
			v.AddSimpleIsConstraint(ann.Id, typed.Variable.Type)
		} else {
			v.AddEqualsConstraint(ann.Id, id)
		}

	// Given a struct access we generate a constructor type. This type is used
	// because inferring an order sensitive type is not practically possible,
	// without a bit of jerry-rigging.
	case *StructAccessExpr:
		id := v.HandleExpr(typed.Struct)
		v.AddIsConstraint(ann.Id, &TypeReference{
			BaseType: &ConstructorType{
				Id:   ConstructorStructMember,
				Args: []*TypeReference{&TypeReference{BaseType: TypeVariable{Id: id}}},
				Data: typed.Member,
			},
		})

	// Given an array access, we know that the type of the expression being
	// accessed must be an array of the same type as the resulting element.
	case *ArrayAccessExpr:
		id := v.HandleExpr(typed.Array)
		v.HandleExpr(typed.Subscript)
		if typed.Array.GetType() != nil {
			at, ok := typed.Array.GetType().BaseType.ActualType().(ArrayType)
			if ok {
				v.AddSimpleIsConstraint(id, at.MemberType)
				break
			}
		}
		v.AddIsConstraint(ann.Id, &TypeReference{
			BaseType: &ConstructorType{
				Id: ConstructorArrayIndex,
				Args: []*TypeReference{
					&TypeReference{BaseType: TypeVariable{Id: id}},
				},
			},
		})

	// An array length expression is always of type uint
	case *ArrayLenExpr:
		v.HandleExpr(typed.Expr)
		v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: PRIMITIVE_uint})

	// An enum literal must always come with a type, so we simply bind its type
	// to it's type variable and to the variable from the contained literal
	case *EnumLiteral:
		if typed.Type == nil {
			panic("INTERNAL ERROR: Encountered enum literal without a type")
		}

		id := -1
		if typed.TupleLiteral != nil {
			id = v.HandleExpr(typed.TupleLiteral)
		} else if typed.CompositeLiteral != nil {
			id = v.HandleExpr(typed.CompositeLiteral)
		}
		if id != -1 {
			v.AddIsConstraint(id, typed.Type)
		}
		v.AddIsConstraint(ann.Id, typed.Type)

	// A bool literal will always be of type bool
	case *BoolLiteral:
		v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: PRIMITIVE_bool})

	// A rune literal will always be of type rune
	case *RuneLiteral:
		v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: runeType})

	// A composite literal is a mess to handle as it can be either an array or
	// a struct, but in either case we go through and generate the type
	// variables for the contained expression, and if we know the type of the
	// literal we bind the generated type variables to their respective types.
	case *CompositeLiteral:
		if typed.Type != nil {
			typ := typed.Type.BaseType.ActualType()
			if at, ok := typ.(ArrayType); ok {
				for _, val := range typed.Values {
					id := v.HandleExpr(val)
					v.AddSimpleIsConstraint(id, at.MemberType)
				}
			} else if st, ok := typ.(StructType); ok {
				for idx, val := range typed.Values {
					field := typed.Fields[idx]
					mem := st.GetMember(field)
					id := v.HandleExpr(val)
					v.AddSimpleIsConstraint(id, mem.Type)
				}
			}
			v.AddSimpleIsConstraint(ann.Id, typed.Type)
		}

	// Given a tuple literal we handle each member, and if we know the type of
	// the tuple we bind their types to their type variables.
	case *TupleLiteral:
		var tt TupleType
		var ok bool
		if typed.Type != nil {
			tt, ok = typed.Type.BaseType.(TupleType)
		}

		nt := make([]*TypeReference, len(typed.Members))
		for idx, mem := range typed.Members {
			id := v.HandleExpr(mem)
			nt[idx] = &TypeReference{BaseType: TypeVariable{Id: id}}
			if ok {
				v.AddSimpleIsConstraint(id, tt.Members[idx])
				nt[idx] = tt.Members[idx]
			}
		}

		if typed.GetType() != nil {
			v.AddSimpleIsConstraint(ann.Id, typed.GetType())
		} else {
			v.AddIsConstraint(ann.Id, &TypeReference{BaseType: tupleOf(nt...)})
		}

	// Given a variable, we bind it's type variable to it's type if its type is known
	case *Variable:
		if typed.GetType() != nil {
			v.AddSimpleIsConstraint(ann.Id, typed.GetType())
		}

	// A function access will always be the type of the function it accesses
	case *FunctionAccessExpr:
		fnType := &TypeReference{BaseType: typed.Function.Type}
		if len(typed.Function.Type.GenericParameters) > 0 {
			if len(typed.GenericArguments) > 0 {
				gcon := NewGenericContext(getTypeGenericParameters(fnType.BaseType), typed.GenericArguments)
				fnType = gcon.Replace(fnType)
				v.AddSimpleIsConstraint(ann.Id, fnType)
			}
		} else {
			v.AddSimpleIsConstraint(ann.Id, fnType)
		}

	// A lambda expr will always be the type of the function it is
	case *LambdaExpr:
		v.AddSimpleIsConstraint(ann.Id, &TypeReference{BaseType: typed.Function.Type})

	case *NumericLiteral, *StringLiteral, *DiscardAccessExpr:
		// noop

	default:
		panic("INTERNAL ERROR: Unhandled Typed type: " + reflect.TypeOf(typed).String())
	}

	return ann.Id
}

// Solve solves the constraints using the unification algorithm.
func (v *Inferrer) Solve() []*Constraint {
	// Create a stack, and copy all constraints to this stack
	stack := make([]*Constraint, len(v.Constraints))
	copy(stack, v.Constraints)

	// Create an array to hold all the final substitutions
	var substitutions []*Constraint

	// Run through the simple constraints
	for _, cons := range v.SimpleConstraints {
		stack, substitutions = v.SolveStep(stack, substitutions, false, cons)
	}

	// As long as we have a constraint on the stack
	for len(stack) > 0 {
		// Remove a constraint X = Y from the stack
		element := stack[0]
		stack[0], stack = nil, stack[1:]

		stack, substitutions = v.SolveStep(stack, substitutions, true, element)
	}

	return substitutions
}

func (v *Inferrer) SolveStep(stackIn []*Constraint, subsIn []*Constraint, addSubs bool, element *Constraint) (stack []*Constraint, substitutions []*Constraint) {
	stack = stackIn
	substitutions = subsIn

	// subsAll runs the substitues a given id for a new side, on all
	// constraints, both on the stack and in the final substitutions
	subsAll := func(id int, what Side) {
		for idx, cons := range stack {
			stack[idx] = cons.Subs(id, what)
		}
		for idx, cons := range substitutions {
			substitutions[idx] = cons.Subs(id, what)
		}
	}

	x, y := element.Left, element.Right

	// 1. If X and Y are identical identifiers, do nothing.
	if x.SideType == IdentSide && y.SideType == IdentSide && x.Id == y.Id {
		return
	}

	// 2. If X is an identifier, replace all occurrences of X by Y both on
	// the stack and in the substitution, and add X → Y to the substitution.
	if x.SideType == IdentSide {
		subsAll(x.Id, y)
		if addSubs {
			substitutions = append(substitutions, &Constraint{
				Left: x, Right: y,
			})
		}
		return
	}

	// 3. If Y is an identifier, replace all occurrences of Y by X both on
	// the stack and in the substitution, and add Y → X to the substitution.
	if y.SideType == IdentSide {
		subsAll(y.Id, x)
		if addSubs {
			substitutions = append(substitutions, &Constraint{Left: y, Right: x})
		}
		return
	}

	// 4. If X is of the form C(X_1, ..., X_n) for some constructor C, and
	// Y is of the form C(Y_1, ..., Y_n) (i.e., it has the same constructor),
	// then push X_i = Y_i for all 1 ≤ i ≤ n onto the stack.

	// 4.0.1. Equal types
	if x.SideType == TypeSide && y.SideType == TypeSide {
		if x.Type.ActualTypesEqual(y.Type) {
			return
		}
	}

	// 4.1. {^, &mut, &}x = {^, &mut, &}y
	if x.SideType == TypeSide && y.SideType == TypeSide {
		xAddressee := getAdressee(x.Type.BaseType)
		yAddressee := getAdressee(y.Type.BaseType)
		if xAddressee != nil && yAddressee != nil {
			stack = append(stack, ConstraintFromTypes(xAddressee, yAddressee))
			return
		}
	}

	// 4.2. []x = []y
	if x.SideType == TypeSide && y.SideType == TypeSide {
		atX, okX := x.Type.BaseType.ActualType().(ArrayType)
		atY, okY := y.Type.BaseType.ActualType().(ArrayType)
		if okX && okY {
			stack = append(stack, ConstraintFromTypes(atX.MemberType, atY.MemberType))
			return
		}
	}

	// 4.3 C(x1, ..., xn).d = C(y1, ... yn).d
	// NOTE: This currently handles both struct members and tuple members
	if x.SideType == TypeSide && y.SideType == TypeSide {
		conX, okX := x.Type.BaseType.(*ConstructorType)
		conY, okY := y.Type.BaseType.(*ConstructorType)
		if okX && okY && conX.Id == conY.Id && len(conX.Args) == len(conY.Args) &&
			conX.Data == conY.Data {
			for idx, argX := range conX.Args {
				argY := conY.Args[idx]
				stack = append(stack, ConstraintFromTypes(argX, argY))
			}
			return
		}
	}

	// 4.4. fn(x1, ...) -> xn = fn(y1, ...) -> yn
	if x.SideType == TypeSide && y.SideType == TypeSide {
		xFunc, okX := x.Type.BaseType.ActualType().(FunctionType)
		yFunc, okY := y.Type.BaseType.ActualType().(FunctionType)

		if okX && okY {
			// Determine minimum parameter list length.
			// This is done to avoid problems with variadic arguments.
			ln := len(xFunc.Parameters)
			if len(yFunc.Parameters) < ln {
				ln = len(yFunc.Parameters)
			}

			// Parameters
			for idx := 0; idx < ln; idx++ {
				stack = append(stack,
					ConstraintFromTypes(xFunc.Parameters[idx], yFunc.Parameters[idx]))
			}

			// Return type
			xRet := xFunc.Return
			yRet := yFunc.Return
			if xRet == nil {
				xRet = &TypeReference{BaseType: PRIMITIVE_void}
			}
			if yRet == nil {
				yRet = &TypeReference{BaseType: PRIMITIVE_void}
			}

			stack = append(stack, ConstraintFromTypes(xRet, yRet))
			return
		}
	}

	// 4.5. (x1, ..., xn) = (y1, ..., yn)
	if x.SideType == TypeSide && y.SideType == TypeSide {
		xTup, okX := x.Type.BaseType.ActualType().(TupleType)
		yTup, okY := y.Type.BaseType.ActualType().(TupleType)

		if okX && okY && len(xTup.Members) == len(yTup.Members) {
			for idx, memX := range xTup.Members {
				memY := yTup.Members[idx]
				stack = append(stack, ConstraintFromTypes(memX, memY))
			}
			return
		}
	}

	// 5. Otherwise, X and Y do not unify. Report an error.
	// NOTE: We defer handling error until the semantic type check
	// TODO: Verify if continuing is ok, or if we should return now
	return
}

// Finalize runs the actual unification, sets default types in cases where
// these are needed, and sets the inferred types on the expressions.
func (v *Inferrer) Finalize() {
	substitutions := v.Solve()

	// Map all substitutions to the id they act upon
	subList := make([]*Constraint, v.IdCount)
	for _, subs := range substitutions {
		if subs.Left.SideType != IdentSide {
			panic("INTERNAL ERROR: Left side of substitution was not ident")
		}
		ann := v.Typeds[subs.Left.Id]
		subList[ann.Id] = subs
	}

	// Add all the simple constraints
	for _, subs := range v.SimpleConstraints {
		if subs.Left.SideType != IdentSide {
			panic("INTERNAL ERROR: Left side of substitution was not ident")
		}
		ann := v.Typeds[subs.Left.Id]
		subList[ann.Id] = subs
	}

	// Apply all substitutions
	for _, subs := range subList {
		if subs == nil {
			continue
		}

		if subs.Left.SideType != IdentSide {
			panic("INTERNAL ERROR: Left side of substitution was not ident")
		}

		ann := v.Typeds[subs.Left.Id]
		if subs.Right.SideType != TypeSide {
			if ann.Typed.GetType() != nil {
				continue
			}
			v.errPos(ann.Pos, "Couldn't infer type of expression")
		}

		if ct, ok := subs.Right.Type.BaseType.(*ConstructorType); ok {
			switch ct.Id {
			case ConstructorStructMember:
				typ := ct.Args[0]
				if tv, ok := typ.BaseType.(TypeVariable); ok {
					typ = subList[tv.Id].Right.Type
				}

				v.errPos(ann.Pos, "Unable to infer type of member `%s` on type `%s`",
					ct.Data.(string), typ.BaseType.TypeName())

			default:
				panic("INTERNAL ERROR: Unhandled ConstructorType escaped inference pass")
			}
		}

		// Set the type of the expression
		ann.Typed.SetType(subs.Right.Type)
	}

	// Type specific touch ups. Here go all the hacky things that was handled
	// in the old inferrence pass, and some new additions to deal with default
	// types.
	for idx := 0; idx < v.IdCount; idx++ {
		ann := v.Typeds[idx]

		switch n := ann.Typed.(type) {
		case *CallExpr:
			// If the function source is a struct access, resolve the method
			// this access represents.
			if sae, ok := n.Function.(*StructAccessExpr); ok {
				// TODO: This will need work once we actually get around to
				// implementing interfaces with all the vtable horribleness
				// it requires.
				fn := GetMethod(sae.Struct.GetType().BaseType, sae.Member)
				if fn == nil {
					v.errPos(sae.Pos(), "Type `%s` has no method `%s`", TypeWithoutPointers(sae.Struct.GetType().BaseType).TypeName(), sae.Member)
				}

				fae := &FunctionAccessExpr{
					Function:         fn,
					ReceiverAccess:   n.ReceiverAccess,
					GenericArguments: sae.GenericArguments,
					ParentFunction:   sae.ParentFunction,
				}
				fae.setPos(sae.Pos())

				n.Function = fae
				fn.Accesses = append(fn.Accesses, fae)
			}

			if n.Function != nil {
				if _, ok := n.Function.GetType().BaseType.(FunctionType); !ok {
					v.errPos(n.Function.Pos(), "Attempt to call non-function `%s`", n.Function.GetType().String())
				}

				// Insert a deref in cases where the code tries to call a value reciver
				// with a pointer type.
				if recType := n.Function.GetType().BaseType.(FunctionType).Receiver; recType != nil {
					accessType := n.ReceiverAccess.GetType()

					if accessType.BaseType.LevelsOfIndirection() == recType.LevelsOfIndirection()+1 {
						deref := &DerefAccessExpr{Expr: n.ReceiverAccess}
						deref.setPos(n.ReceiverAccess.Pos())
						n.ReceiverAccess = deref
					}
				}
			}

		case *StructAccessExpr:
			// Check if we're dealing with a method and exit early
			if GetMethod(n.Struct.GetType().BaseType, n.Member) != nil {
				break
			}

			// Insert a deref in cases where the code tries to access a struct
			// member from a pointer type.
			if n.Struct.GetType().BaseType.ActualType().LevelsOfIndirection() == 1 {
				deref := &DerefAccessExpr{Expr: n.Struct}
				deref.setPos(n.Struct.Pos())
				n.Struct = deref
			}

			// Verify that we're actually dealing with a struct.
			typ := n.Struct.GetType()
			structType, ok := typ.BaseType.ActualType().(StructType)
			if !ok {
				v.errPos(n.Pos(), "Cannot access member of type `%s`", typ.String())
			}

			// Verify that the struct actually has the requested member.
			mem := structType.GetMember(n.Member)
			if mem == nil {
				v.errPos(n.Pos(), "Struct `%s` does not contain member or method `%s`", typ.String(), n.Member)
			}

		case *BinaryExpr:
			nll, ok1 := n.Lhand.(*NumericLiteral)
			nlr, ok2 := n.Rhand.(*NumericLiteral)

			// Here we deal with the case where two numeric literals appear in
			// a binary expression, but where one of them is a float literal
			// and the other isn't.
			if ok1 && ok2 && nll.IsFloat {
				nlr.SetType(nll.GetType())
				break
			}

			if ok1 && ok2 && nlr.IsFloat {
				nll.SetType(nlr.GetType())
				break
			}

			// Patch up the cases wher one side is a numeric literal and the
			// other is not.
			if ok1 {
				nll.SetType(n.Rhand.GetType())
				break
			}

			if ok2 {
				nlr.SetType(n.Lhand.GetType())
			}

		case *CastExpr:
			expr, ok := n.Expr.(*NumericLiteral)

			// Here we handle the case where a numeric literal appear in a cast
			// to a pointer type. We need the default type to be uintptr here
			// as normal integers can't be cast to a pointer.
			if ok && n.Type.BaseType.LevelsOfIndirection() > 0 {
				expr.SetType(&TypeReference{BaseType: PRIMITIVE_uintptr})
			}
		}
	}
}

// SetType Methods

// UnaryExpr
func (v *UnaryExpr) SetType(t *TypeReference) {
	v.Type = t
}

// BinaryExpr
func (v *BinaryExpr) SetType(t *TypeReference) {
	v.Type = t
}

// NumericLiteral
func (v *NumericLiteral) SetType(t *TypeReference) {
	var actual Type
	if t != nil {
		actual = t.BaseType.ActualType()
	}

	if v.IsFloat {
		switch actual {
		case PRIMITIVE_f32, PRIMITIVE_f64, PRIMITIVE_f128:
			v.Type = t

		default:
			v.Type = &TypeReference{BaseType: PRIMITIVE_f64}
		}
	} else {
		switch actual {
		case PRIMITIVE_int, PRIMITIVE_uint, PRIMITIVE_uintptr,
			PRIMITIVE_s8, PRIMITIVE_s16, PRIMITIVE_s32, PRIMITIVE_s64, PRIMITIVE_s128,
			PRIMITIVE_u8, PRIMITIVE_u16, PRIMITIVE_u32, PRIMITIVE_u64, PRIMITIVE_u128,
			PRIMITIVE_f32, PRIMITIVE_f64, PRIMITIVE_f128:
			v.Type = t

		default:
			v.Type = &TypeReference{BaseType: PRIMITIVE_int}
		}
	}
}

// ArrayLiteral
func (v *CompositeLiteral) SetType(t *TypeReference) {
	if t == nil {
		return
	}

	if v.Type == nil {
		switch t.BaseType.ActualType().(type) {
		case StructType, ArrayType:
			v.Type = t
		}
	}
}

// StringLiteral
func (v *StringLiteral) SetType(t *TypeReference) {
	if t.BaseType.ActualType() == stringType {
		v.Type = t
	} // TODO arrays
}

// TupleLiteral
func (v *TupleLiteral) SetType(t *TypeReference) {
	if t == nil {
		return
	}

	_, ok := t.BaseType.ActualType().(TupleType)
	if ok {
		v.Type = t
	}
}

// Variable
func (v *Variable) SetType(t *TypeReference) {
	if v.Type == nil {
		v.Type = t
	}
}

func (v *FunctionAccessExpr) SetType(t *TypeReference) {
	if len(v.GenericArguments) != len(v.Function.Type.GenericParameters) {
		types, err := ExtractTypeVariable(v.Function.Type, t.BaseType)
		if err != nil {
			panic(err)
		}

		genArgs := make([]*TypeReference, len(v.Function.Type.GenericParameters))
		for idx, param := range v.Function.Type.GenericParameters {
			genArgs[idx] = &TypeReference{BaseType: types[param.Name]}
		}
		v.GenericArguments = genArgs
	}
}

func (v *EnumLiteral) SetType(t *TypeReference) {
	et, ok := v.Type.BaseType.ActualType().(EnumType)
	if ok && len(et.GenericParameters) > 0 {
		if len(v.Type.GenericArguments) != len(et.GenericParameters) {
			v.Type.GenericArguments = t.GenericArguments
		}
	}
}

// Noops
func (_ ArrayAccessExpr) SetType(t *TypeReference)    {}
func (_ ArrayLenExpr) SetType(t *TypeReference)       {}
func (_ BoolLiteral) SetType(t *TypeReference)        {}
func (_ CastExpr) SetType(t *TypeReference)           {}
func (_ CallExpr) SetType(t *TypeReference)           {}
func (_ DefaultMatchBranch) SetType(t *TypeReference) {}
func (_ DerefAccessExpr) SetType(t *TypeReference)    {}
func (_ DiscardAccessExpr) SetType(t *TypeReference)  {}
func (_ LambdaExpr) SetType(t *TypeReference)         {}
func (_ PointerToExpr) SetType(t *TypeReference)      {}
func (_ ReferenceToExpr) SetType(t *TypeReference)    {}
func (_ RuneLiteral) SetType(t *TypeReference)        {}
func (_ VariableAccessExpr) SetType(t *TypeReference) {}
func (_ SizeofExpr) SetType(t *TypeReference)         {}
func (_ StructAccessExpr) SetType(t *TypeReference)   {}

// ExtractTypeVariable takes a pattern type containing one or more substitution
// types together with a value type, and generates a map from the substitution
// types to the the corresponding parts of the value type.
// An example would be:
// pattern: Pointer($T)
//   value: Pointer(int)
//  return: {T: int}
func ExtractTypeVariable(pattern Type, value Type) (map[string]Type, error) {
	res := make(map[string]Type)

	// Start with the pattern and the value
	ps := []Type{pattern}
	vs := []Type{value}

	for i := 0; i < len(ps); i++ {
		// Fetch the next types to compare
		ppart := ps[i]
		vpart := vs[i]

		if subst, ok := ppart.(*SubstitutionType); ok {
			// If we reached a substitution type, add an entry to the map
			res[subst.Name] = vpart
		} else {
			// Skip stuff that still contains type variables
			_, ok1 := ppart.(TypeVariable)
			_, ok2 := vpart.(TypeVariable)
			if ok1 || ok2 {
				continue
			}

			// If the pattern part is not a substitution type, delve deeper
			ps = AddChildren(ppart, ps)
			vs = AddChildren(vpart, vs)

			// Also verify that things match up type wise
			switch ppart.(type) {
			case PrimitiveType, *NamedType:
				if !ppart.Equals(vpart) {
					return nil, fmt.Errorf("inferrer: type mismatch %s != %s", ppart.TypeName(), vpart.TypeName())
				}

			default:
				if reflect.TypeOf(ppart) != reflect.TypeOf(vpart) {
					return nil, fmt.Errorf("inferrer: type mismatch %s != %s", ppart.TypeName(), vpart.TypeName())
				}
			}
		}
	}

	return res, nil
}

// AddChildren adds the children of a type to the passed list
func AddChildren(typ Type, dest []Type) []Type {
	switch typ := typ.(type) {
	case StructType:
		for _, mem := range typ.Members {
			dest = append(dest, mem.Type.BaseType)
		}

	case *NamedType:
		dest = append(dest, typ.Type)

	case ArrayType:
		dest = append(dest, typ.MemberType.BaseType)

	case PointerType:
		dest = append(dest, typ.Addressee.BaseType)

	case TupleType:
		for _, tref := range typ.Members {
			dest = append(dest, tref.BaseType)
		}

	case EnumType:
		for _, mem := range typ.Members {
			dest = append(dest, mem.Type)
		}

	case FunctionType:
		if typ.Receiver != nil {
			dest = append(dest, typ.Receiver)
		}
		for _, tref := range typ.Parameters {
			dest = append(dest, tref.BaseType)
		}

		if typ.Return != nil { // TODO: can it ever be nil?
			dest = append(dest, typ.Return.BaseType)
		}

	case PrimitiveType, *SubstitutionType:
		// noops

	default:
		panic("Unhandled type: " + reflect.TypeOf(typ).String())

	}
	return dest
}
