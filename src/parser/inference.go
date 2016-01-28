package parser

import (
	"fmt"
	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
	"os"
	"reflect"
)

type TypeVariable struct {
	metaType
	Id int
}

func (v *TypeVariable) Equals(other Type) bool {
	if ot, ok := other.(*TypeVariable); ok {
		return v.Id == ot.Id
	}
	return false
}

func (v *TypeVariable) String() string {
	return NewASTStringer("TypeVariable").AddType(v).Finish()
}

func (v *TypeVariable) TypeName() string {
	return fmt.Sprintf("$%d", v.Id)
}

func (v *TypeVariable) ActualType() Type {
	return v
}

type ConstructorId int

const (
	ConstructorInvalid ConstructorId = iota
	ConstructorStructMember

	// TODO: This guy goes away once we remove tuple indexing and replace with
	// tuple destructuring
	ConstructorTupleIndex
)

type ConstructorType struct {
	metaType
	Id   ConstructorId
	Args []Type

	// Some constructors need additional data
	Data interface{}
}

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

type AnnotatedTyped struct {
	Pos   lexer.Position
	Typed Typed
	Id    int
}

type SideType int

const (
	IdentSide SideType = iota
	TypeSide
)

type Side struct {
	SideType SideType
	Id       int
	Type     Type
}

func SideFromType(t Type) Side {
	if tv, ok := t.(*TypeVariable); ok {
		return Side{SideType: IdentSide, Id: tv.Id}
	}
	return Side{SideType: TypeSide, Type: t}
}

func (v Side) Subs(id int, what Side) Side {
	switch v.SideType {
	case IdentSide:
		if v.Id == id {
			return what
		}
		return v

	case TypeSide:
		var nt Type
		if what.SideType == TypeSide {
			nt = SubsType(v.Type, id, what.Type)
		} else {
			nt = SubsType(v.Type, id, &TypeVariable{Id: what.Id})
		}
		return Side{SideType: TypeSide, Type: nt}

	default:
		panic("Invalid SideType")
	}
}

func SubsType(typ Type, id int, what Type) Type {
	switch t := typ.(type) {
	case *TypeVariable:
		if t.Id == id {
			return what
		}
		return t

	case *ConstructorType:
		nargs := make([]Type, len(t.Args))
		for idx, arg := range t.Args {
			nargs[idx] = SubsType(arg, id, what)
		}

		// Handle special cases
		switch t.Id {
		case ConstructorStructMember:
			// Method check
			if typ, ok := TypeWithoutPointers(nargs[0]).(*NamedType); ok {
				fn := typ.GetMethod(t.Data.(string))
				if fn != nil {
					return fn.Type
				}
			}

			// Struct member
			typ := nargs[0]
			if pt, ok := typ.(PointerType); ok {
				typ = pt.Addressee
			}
			if st, ok := typ.ActualType().(StructType); ok {
				mem := st.GetMember(t.Data.(string))
				return mem.Type
			}

		case ConstructorTupleIndex:
			if tt, ok := nargs[0].ActualType().(TupleType); ok {
				return tt.Members[t.Data.(uint64)]
			}
		}

		return &ConstructorType{Id: t.Id, Args: nargs, Data: t.Data}

	case FunctionType:
		newRet := SubsType(t.Return, id, what)
		np := make([]Type, len(t.Parameters))
		for idx, param := range t.Parameters {
			np[idx] = SubsType(param, id, what)
		}

		return FunctionType{
			attrs:      t.attrs,
			IsVariadic: t.IsVariadic,
			Parameters: np,
			Return:     newRet,
		}

	case TupleType:
		nm := make([]Type, len(t.Members))
		for idx, mem := range t.Members {
			nm[idx] = SubsType(mem, id, what)
		}
		return tupleOf(nm...)

	case PointerType:
		return PointerTo(SubsType(t.Addressee, id, what))

	case ArrayType:
		return ArrayOf(SubsType(t.MemberType, id, what))

	case PrimitiveType, StructType, *NamedType, ConstantReferenceType,
		MutableReferenceType, InterfaceType, EnumType:
		return t

	default:
		panic("Unhandled type in Side.Subs(): " + reflect.TypeOf(t).String())
	}
}

func (v Side) String() string {
	switch v.SideType {
	case IdentSide:
		return fmt.Sprintf("$%d", v.Id)
	case TypeSide:
		return fmt.Sprintf("type `%s`", v.Type.TypeName())
	}
	panic("Invalid side type")
}

type Constraint struct {
	Left, Right Side
}

func ConstraintFromTypes(left Type, right Type) *Constraint {
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

type NewInferer struct {
	Submodule   *Submodule
	Functions   []*Function
	Typeds      map[int]*AnnotatedTyped
	TypedLookup map[Typed]*AnnotatedTyped
	Constraints []*Constraint
	IdCount     int
}

func (v *NewInferer) Function() *Function {
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
		inf := &NewInferer{
			Submodule:   submod,
			Typeds:      make(map[int]*AnnotatedTyped),
			TypedLookup: make(map[Typed]*AnnotatedTyped),
		}
		vis := NewASTVisitor(inf)
		vis.VisitSubmodule(submod)
		inf.Finalize()
	})

}

func (v *NewInferer) AddConstraint(c *Constraint) {
	v.Constraints = append(v.Constraints, c)
}

func (v *NewInferer) AddEqualsConstraint(a, b int) {
	c := &Constraint{
		Left:  Side{Id: a, SideType: IdentSide},
		Right: Side{Id: b, SideType: IdentSide},
	}
	v.AddConstraint(c)
}

func (v *NewInferer) AddIsConstraint(id int, typ Type) {
	c := &Constraint{
		Left:  Side{Id: id, SideType: IdentSide},
		Right: Side{Type: typ, SideType: TypeSide},
	}
	v.AddConstraint(c)
}

func (v *NewInferer) EnterScope() {}

func (v *NewInferer) ExitScope() {}

func (v *NewInferer) PostVisit(node *Node) {
	switch (*node).(type) {
	case *FunctionDecl:
		idx := len(v.Functions) - 1
		v.Functions[idx] = nil
		v.Functions = v.Functions[:idx]
		return
	}
}

func (v *NewInferer) Visit(node *Node) bool {
	switch n := (*node).(type) {
	case *FunctionDecl:
		v.Functions = append(v.Functions, n.Function)
		return true
	}

	switch n := (*node).(type) {
	case *VariableDecl:
		a := v.HandleTyped(n.Pos(), n.Variable)
		if n.Assignment != nil {
			b := v.HandleExpr(n.Assignment)
			v.AddEqualsConstraint(a, b)
		}

	case *AssignStat:
		a := v.HandleExpr(n.Access)
		b := v.HandleExpr(n.Assignment)
		v.AddEqualsConstraint(a, b)

	case *BinopAssignStat:
		a := v.HandleExpr(n.Access)
		b := v.HandleExpr(n.Assignment)
		v.AddEqualsConstraint(a, b)

	case *CallStat:
		v.HandleExpr(n.Call)

	case *DeferStat:
		v.HandleExpr(n.Call)

	case *IfStat:
		for _, expr := range n.Exprs {
			id := v.HandleExpr(expr)
			v.AddIsConstraint(id, PRIMITIVE_bool)
		}

	case *ReturnStat:
		if n.Value != nil {
			id := v.HandleExpr(n.Value)
			v.AddIsConstraint(id, v.Function().Type.Return)
		}

	case *LoopStat:
		if n.Condition != nil {
			id := v.HandleExpr(n.Condition)
			v.AddIsConstraint(id, PRIMITIVE_bool)
		}

	case *MatchStat:
		// TODO: Implement once we actuall do match statement
	}

	return true
}

func (v *NewInferer) HandleExpr(expr Expr) int {
	return v.HandleTyped(expr.Pos(), expr)
}

func (v *NewInferer) HandleTyped(pos lexer.Position, typed Typed) int {
	if ann, ok := v.TypedLookup[typed]; ok {
		return ann.Id
	}

	ann := &AnnotatedTyped{Pos: pos, Id: v.IdCount, Typed: typed}
	v.Typeds[ann.Id] = ann
	v.TypedLookup[typed] = ann
	v.IdCount++

	switch typed := typed.(type) {
	case *BinaryExpr:
		a := v.HandleExpr(typed.Lhand)
		b := v.HandleExpr(typed.Rhand)
		switch typed.Op {
		case BINOP_EQ, BINOP_NOT_EQ, BINOP_GREATER, BINOP_LESS,
			BINOP_GREATER_EQ, BINOP_LESS_EQ:
			v.AddEqualsConstraint(a, b)
			v.AddIsConstraint(ann.Id, PRIMITIVE_bool)

		case BINOP_BIT_AND, BINOP_BIT_OR, BINOP_BIT_XOR:
			v.AddEqualsConstraint(a, b)
			v.AddEqualsConstraint(ann.Id, a)

		case BINOP_ADD, BINOP_SUB, BINOP_MUL, BINOP_DIV, BINOP_MOD:
			// TODO: These assumptions don't hold once we add operator overloading
			v.AddEqualsConstraint(a, b)
			v.AddEqualsConstraint(ann.Id, a)

		case BINOP_BIT_LEFT, BINOP_BIT_RIGHT:
			v.AddEqualsConstraint(ann.Id, a)

		case BINOP_LOG_AND, BINOP_LOG_OR:
			v.AddIsConstraint(a, PRIMITIVE_bool)
			v.AddIsConstraint(b, PRIMITIVE_bool)
			v.AddIsConstraint(ann.Id, PRIMITIVE_bool)

		default:
			panic("Unhandled binary operator in type inference")

		}

	case *UnaryExpr:
		id := v.HandleExpr(typed.Expr)
		switch typed.Op {
		case UNOP_LOG_NOT:
			v.AddIsConstraint(id, PRIMITIVE_bool)

		case UNOP_BIT_NOT:
			v.AddEqualsConstraint(ann.Id, id)

		case UNOP_NEGATIVE:
			v.AddEqualsConstraint(ann.Id, id)

		}

	case *CallExpr:
		if typed.ReceiverAccess != nil {
			v.HandleExpr(typed.ReceiverAccess)
		}

		fnId := v.HandleExpr(typed.Function)
		argIds := make([]int, len(typed.Arguments))
		for idx, arg := range typed.Arguments {
			argIds[idx] = v.HandleExpr(arg)
		}

		fnType := FunctionType{Return: &TypeVariable{Id: ann.Id}}
		for _, argId := range argIds {
			fnType.Parameters = append(fnType.Parameters, &TypeVariable{Id: argId})
		}
		v.AddIsConstraint(fnId, fnType)

	case *CastExpr:
		v.HandleExpr(typed.Expr)
		v.AddIsConstraint(ann.Id, typed.Type)

	case *AddressOfExpr:
		id := v.HandleExpr(typed.Access)
		v.AddIsConstraint(ann.Id, PointerTo(&TypeVariable{Id: id}))

	case *DerefAccessExpr:
		id := v.HandleExpr(typed.Expr)
		v.AddIsConstraint(id, PointerTo(&TypeVariable{Id: ann.Id}))

	case *SizeofExpr:
		if typed.Expr != nil {
			v.HandleExpr(typed.Expr)
		}
		v.AddIsConstraint(ann.Id, PRIMITIVE_uint)

	case *VariableAccessExpr:
		id := v.HandleTyped(typed.Pos(), typed.Variable)
		v.AddEqualsConstraint(ann.Id, id)

	case *StructAccessExpr:
		id := v.HandleExpr(typed.Struct)
		v.AddIsConstraint(ann.Id, &ConstructorType{
			Id:   ConstructorStructMember,
			Args: []Type{&TypeVariable{Id: id}},
			Data: typed.Member,
		})

	case *TupleAccessExpr:
		id := v.HandleExpr(typed.Tuple)
		v.AddIsConstraint(ann.Id, &ConstructorType{
			Id:   ConstructorTupleIndex,
			Args: []Type{&TypeVariable{Id: id}},
			Data: typed.Index,
		})

	case *ArrayAccessExpr:
		id := v.HandleExpr(typed.Array)
		v.HandleExpr(typed.Subscript)
		v.AddIsConstraint(id, ArrayOf(&TypeVariable{Id: ann.Id}))

	case *ArrayLenExpr:
		v.HandleExpr(typed.Expr)
		v.AddIsConstraint(ann.Id, PRIMITIVE_uint)

	case *EnumLiteral:
		if typed.Type == nil {
			panic("INTERNAL ERROR: Encountered enum literal without a type")
		}

		var id int
		if typed.TupleLiteral != nil {
			id = v.HandleExpr(typed.TupleLiteral)
		} else if typed.CompositeLiteral != nil {
			id = v.HandleExpr(typed.CompositeLiteral)
		}
		v.AddIsConstraint(id, typed.Type)
		v.AddIsConstraint(ann.Id, typed.Type)

	case *BoolLiteral:
		v.AddIsConstraint(ann.Id, PRIMITIVE_bool)

	case *StringLiteral:
		if typed.IsCString {
			v.AddIsConstraint(ann.Id, PointerTo(PRIMITIVE_u8))
		} else {
			v.AddIsConstraint(ann.Id, stringType)
		}

	case *RuneLiteral:
		v.AddIsConstraint(ann.Id, PRIMITIVE_rune)

	case *CompositeLiteral:
		ids := make([]int, len(typed.Values))
		for idx, mem := range typed.Values {
			ids[idx] = v.HandleExpr(mem)
		}

		typ := typed.Type.ActualType()
		if at, ok := typ.(ArrayType); ok {
			v.AddIsConstraint(ann.Id, at)
			for _, id := range ids {
				v.AddIsConstraint(id, at.MemberType)
			}
		} else if st, ok := typ.(StructType); ok {
			v.AddIsConstraint(ann.Id, st)
			for idx, id := range ids {
				field := typed.Fields[idx]
				mem := st.GetMember(field)
				v.AddIsConstraint(id, mem.Type)
			}
		}

	case *TupleLiteral:
		var tt TupleType
		var ok bool
		if typed.Type != nil {
			tt, ok = typed.Type.(TupleType)
			v.AddIsConstraint(ann.Id, tt)
		}

		nt := make([]Type, len(typed.Members))
		for idx, mem := range typed.Members {
			id := v.HandleExpr(mem)
			nt[idx] = &TypeVariable{Id: id}
			if ok {
				v.AddIsConstraint(id, tt.Members[idx])
				nt[idx] = tt.Members[idx]
			}
		}

		v.AddIsConstraint(ann.Id, tupleOf(nt...))

	case *Variable:
		if typed.GetType() != nil {
			v.AddIsConstraint(ann.Id, typed.GetType())
		}

	case *FunctionAccessExpr:
		v.AddIsConstraint(ann.Id, typed.Function.Type)

	case *NumericLiteral:
		// noop

	default:
		log.Errorln("inferer", "Unhandled Typed type `%T`", typed)
	}

	return ann.Id
}

func (v *NewInferer) Unify() []*Constraint {
	stack := make([]*Constraint, len(v.Constraints))
	copy(stack, v.Constraints)

	var substitutions []*Constraint
	subsAll := func(id int, what Side) {
		for idx, cons := range stack {
			stack[idx] = cons.Subs(id, what)
		}
		for idx, cons := range substitutions {
			substitutions[idx] = cons.Subs(id, what)
		}
	}

	for len(stack) > 0 {
		idx := len(stack) - 1

		// Given a constraint X = Y
		element := stack[idx]
		stack = stack[:idx]
		x, y := element.Left, element.Right

		// 1. If X and Y are identical identifiers, do nothing.
		if x.SideType == IdentSide && y.SideType == IdentSide && x.Id == y.Id {
			continue
		}

		// 2. If X is an identifier, replace all occurrences of X by Y both on
		// the stack and in the substitution, and add X → Y to the substitution.
		if x.SideType == IdentSide {
			subsAll(x.Id, y)
			substitutions = append(substitutions, &Constraint{
				Left: x, Right: y,
			})
			continue
		}

		// 3. If Y is an identifier, replace all occurrences of Y by X both on
		// the stack and in the substitution, and add Y → X to the substitution.
		if y.SideType == IdentSide {
			subsAll(y.Id, x)
			substitutions = append(substitutions, &Constraint{Left: y, Right: x})
			continue
		}

		// 4. If X is of the form C(X_1, ..., X_n) for some constructor C, and
		// Y is of the form C(Y_1, ..., Y_n) (i.e., it has the same constructor),
		// then push X_i = Y_i for all 1 ≤ i ≤ n onto the stack.

		// 4.0.1. Equal types
		if x.SideType == TypeSide && y.SideType == TypeSide {
			xtyp := x.Type.ActualType()
			ytyp := y.Type.ActualType()
			if xtyp.Equals(ytyp) {
				continue
			}

		}

		// 4.1. {^, &mut, &}x = {^, &mut, &}y
		if x.SideType == TypeSide && y.SideType == TypeSide {
			xAddressee := getAdressee(x.Type)
			yAddressee := getAdressee(y.Type)
			if xAddressee != nil && yAddressee != nil {
				stack = append(stack, ConstraintFromTypes(xAddressee, yAddressee))
				continue
			}
		}

		// 4.2. []x = []y
		if x.SideType == TypeSide && y.SideType == TypeSide {
			atX, okX := x.Type.ActualType().(ArrayType)
			atY, okY := y.Type.ActualType().(ArrayType)
			if okX && okY {
				stack = append(stack, ConstraintFromTypes(atX.MemberType, atY.MemberType))
				continue
			}
		}

		// 4.3 C(x1, ..., xn).d = C(y1, ... yn).d
		// NOTE: This currently handles both struct members and tuple members
		if x.SideType == TypeSide && y.SideType == TypeSide {
			conX, okX := x.Type.(*ConstructorType)
			conY, okY := y.Type.(*ConstructorType)
			if okX && okY && conX.Id == conY.Id && len(conX.Args) == len(conY.Args) &&
				conX.Data == conY.Data {
				for idx, argX := range conX.Args {
					argY := conY.Args[idx]
					stack = append(stack, ConstraintFromTypes(argX, argY))
				}
				continue
			}
		}

		// 4.4. fn(x1, ...) -> xn = fn(y1, ...) -> yn
		if x.SideType == TypeSide && y.SideType == TypeSide {
			xFunc, okX := x.Type.ActualType().(FunctionType)
			yFunc, okY := y.Type.ActualType().(FunctionType)

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
					xRet = PRIMITIVE_void
				}
				if yRet == nil {
					yRet = PRIMITIVE_void
				}

				stack = append(stack, ConstraintFromTypes(xRet, yRet))
				continue
			}
		}

		// 5. Otherwise, X and Y do not unify. Report an error.
		// NOTE: We defer handling error until the semantic type check
		// TODO: Verify if continuing is ok, or if we should return now
	}

	return substitutions
}

func (v *NewInferer) Finalize() {
	substitutions := v.Unify()

	subList := make([]*Constraint, v.IdCount)
	for _, subs := range substitutions {
		if subs.Left.SideType != IdentSide {
			panic("INTERNAL ERROR: Left side of substitution was not ident")
		}
		ann := v.Typeds[subs.Left.Id]
		subList[ann.Id] = subs
	}

	resolved := true
	for _, val := range subList {
		resolved = resolved && (val == nil || val.Right.SideType != TypeSide)
	}

	// If it didn't all resolve the first time, inject default integer types
	// and try once again
	if !resolved {
		v.Constraints = nil
		for idx := 0; idx < v.IdCount; idx++ {
			ann := v.Typeds[idx]
			subs := subList[idx]
			if subs != nil && subs.Right.SideType == TypeSide {
				v.AddConstraint(subs)
				continue
			}

			if lit, ok := ann.Typed.(*NumericLiteral); ok {
				typ := PRIMITIVE_int
				if lit.IsFloat {
					typ = PRIMITIVE_f32
					switch lit.floatSizeHint {
					case 'f':
						typ = PRIMITIVE_f32
					case 'd':
						typ = PRIMITIVE_f64
					case 'q':
						typ = PRIMITIVE_f128
					}

				}
				v.AddIsConstraint(idx, typ)
			} else if subs != nil {
				v.AddConstraint(subs)
			}
		}

		substitutions = v.Unify()
	}

	// Apply all substitutions
	for _, subs := range substitutions {
		if subs.Left.SideType != IdentSide {
			panic("INTERNAL ERROR: Left side of substitution was not ident")
		}
		ann := v.Typeds[subs.Left.Id]

		if subs.Right.SideType != TypeSide {
			log.Errorln("inferer", "Couldn't infer type of expression:")
			log.Errorln("inferer", "%s", v.Submodule.File.MarkPos(ann.Pos))
			os.Exit(util.EXIT_FAILURE_SEMANTIC)
		}

		if typ, ok := subs.Right.Type.(*ConstructorType); ok {
			log.Debugln("inferer", "%v", typ)
			panic("INTERNAL ERROR: ConstructorType escaped inference pass")
		}

		ann.Typed.setTypeHint(subs.Right.Type)
	}

	// Type specific touch ups
	for idx := 0; idx < v.IdCount; idx++ {
		ann := v.Typeds[idx]

		switch n := ann.Typed.(type) {
		case *CallExpr:
			// Resolve methods
			if sae, ok := n.Function.(*StructAccessExpr); ok {
				fn := TypeWithoutPointers(sae.Struct.GetType()).(*NamedType).GetMethod(sae.Member)
				n.Function = &FunctionAccessExpr{Function: fn}
				if n.Function == nil {
					log.Errorln("inferer", "Type `%s` has no method `%s`", TypeWithoutPointers(sae.Struct.GetType()).TypeName(), sae.Member)
					log.Errorln("inferer", "%s", v.Submodule.File.MarkPos(sae.Pos()))
					os.Exit(util.EXIT_FAILURE_SEMANTIC)
				}
			}

			if n.Function != nil {
				if _, ok := n.Function.GetType().(FunctionType); !ok {
					log.Errorln("inferer", "Attempt to call non-function `%s`", n.Function.GetType().TypeName())
					log.Errorln("inferer", "%s", v.Submodule.File.MarkPos(n.Function.Pos()))
					os.Exit(util.EXIT_FAILURE_SEMANTIC)
				}

				// Insert dereference if needed
				if n.Function.GetType().(FunctionType).Receiver != nil {
					recType := n.Function.GetType().(FunctionType).Receiver
					accessType := n.ReceiverAccess.GetType()

					if accessType.LevelsOfIndirection() == recType.LevelsOfIndirection()+1 {
						n.ReceiverAccess = &DerefAccessExpr{Expr: n.ReceiverAccess}
					}
				}
			}

		case *StructAccessExpr:
			// Check if we're dealing with a method and exit early
			baseType := TypeWithoutPointers(n.Struct.GetType())
			if nt, ok := baseType.(*NamedType); ok && nt.GetMethod(n.Member) != nil {
				break
			}

			// Insert dereference if needed
			if n.Struct.GetType().ActualType().LevelsOfIndirection() == 1 {
				n.Struct = &DerefAccessExpr{Expr: n.Struct}
			}

			typ := n.Struct.GetType()
			structType, ok := typ.ActualType().(StructType)
			if !ok {
				log.Errorln("inferer", "%s Cannot access member of type `%s`", util.Red("error:"), typ.TypeName())
				log.Errorln("inferer", "%s", v.Submodule.File.MarkPos(n.Pos()))
				os.Exit(util.EXIT_FAILURE_SEMANTIC)
			}

			mem := structType.GetMember(n.Member)
			if mem == nil {
				log.Errorln("inferer", "Struct `%s` does not contain member or method `%s`", typ.TypeName(), n.Member)
				log.Errorln("inferer", "%s", v.Submodule.File.MarkPos(n.Pos()))
				os.Exit(util.EXIT_FAILURE_SEMANTIC)
			}

		case *BinaryExpr:
			// Some wiggling of default numeric literal types
			nll, ok1 := n.Lhand.(*NumericLiteral)
			nlr, ok2 := n.Rhand.(*NumericLiteral)

			if ok1 && !ok2 {
				nll.setTypeHint(n.Rhand.GetType())
				break
			}

			if ok2 && !ok1 {
				nlr.setTypeHint(n.Lhand.GetType())
				break
			}

			if ok1 && ok2 && nll.IsFloat {
				nlr.setTypeHint(nll.GetType())
				break
			}

			if ok1 && ok2 && nlr.IsFloat {
				nll.setTypeHint(nlr.GetType())
				break
			}

		case *CastExpr:
			expr, ok := n.Expr.(*NumericLiteral)
			if ok && n.Type.LevelsOfIndirection() > 0 {
				expr.setTypeHint(PRIMITIVE_uintptr)
			}
		}
	}
}

// UnaryExpr
func (v *UnaryExpr) setTypeHint(t Type) {
	v.Type = t
}

// BinaryExpr
func (v *BinaryExpr) setTypeHint(t Type) {
	v.Type = t
}

// NumericLiteral
func (v *NumericLiteral) setTypeHint(t Type) {
	var actual Type
	if t != nil {
		actual = t.ActualType()
	}

	if v.IsFloat {
		switch actual {
		case PRIMITIVE_f32, PRIMITIVE_f64, PRIMITIVE_f128:
			v.Type = t

		default:
			v.Type = PRIMITIVE_f64
		}
	} else {
		switch actual {
		case PRIMITIVE_int, PRIMITIVE_uint, PRIMITIVE_uintptr,
			PRIMITIVE_s8, PRIMITIVE_s16, PRIMITIVE_s32, PRIMITIVE_s64, PRIMITIVE_s128,
			PRIMITIVE_u8, PRIMITIVE_u16, PRIMITIVE_u32, PRIMITIVE_u64, PRIMITIVE_u128,
			PRIMITIVE_f32, PRIMITIVE_f64, PRIMITIVE_f128:
			v.Type = t

		default:
			v.Type = PRIMITIVE_int
		}
	}
}

// ArrayLiteral
func (v *CompositeLiteral) setTypeHint(t Type) {
	if t == nil {
		return
	}

	switch t.ActualType().(type) {
	case StructType, ArrayType:
		v.Type = t
	}
}

// StringLiteral
func (v *StringLiteral) setTypeHint(t Type) {
	v.Type = t
}

// TupleLiteral
func (v *TupleLiteral) setTypeHint(t Type) {
	if t == nil {
		return
	}

	_, ok := t.ActualType().(TupleType)
	if ok {
		v.Type = t
	}
}

// Variable
func (v *Variable) setTypeHint(t Type) {
	if v.Type == nil {
		v.Type = t
	}
}

// Noops
func (_ AddressOfExpr) setTypeHint(t Type)      {}
func (_ ArrayAccessExpr) setTypeHint(t Type)    {}
func (_ ArrayLenExpr) setTypeHint(t Type)       {}
func (_ BoolLiteral) setTypeHint(t Type)        {}
func (_ CastExpr) setTypeHint(t Type)           {}
func (_ CallExpr) setTypeHint(t Type)           {}
func (_ DefaultMatchBranch) setTypeHint(t Type) {}
func (_ DerefAccessExpr) setTypeHint(t Type)    {}
func (_ EnumLiteral) setTypeHint(t Type)        {}
func (_ FunctionAccessExpr) setTypeHint(t Type) {}
func (_ LambdaExpr) setTypeHint(t Type)         {}
func (_ RuneLiteral) setTypeHint(t Type)        {}
func (_ VariableAccessExpr) setTypeHint(t Type) {}
func (_ SizeofExpr) setTypeHint(t Type)         {}
func (_ StructAccessExpr) setTypeHint(t Type)   {}
func (_ TupleAccessExpr) setTypeHint(t Type)    {}
