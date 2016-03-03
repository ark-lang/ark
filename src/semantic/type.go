package semantic

import (
	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/parser"
)

func (_ TypeCheck) Name() string { return "type" }

// Takes a pointer to the expr, so we can replace it with a cast if necessary.
// TODO: do we need an ImplicitCastExpr node? Or just an InterfaceWrapExpr node.
func expectType(s *SemanticAnalyzer, loc ast.Locatable, expect *ast.TypeReference, expr *ast.Expr) {
	exprType := (*expr).GetType()
	if expect.ActualTypesEqual(exprType) {
		return
	}

	if expectPtr, ok := expect.BaseType.(ast.PointerType); ok {
		if exprPtr, ok := exprType.BaseType.(ast.PointerType); ok {
			if expectPtr.Addressee.ActualTypesEqual(exprPtr.Addressee) && exprPtr.IsMutable && !expectPtr.IsMutable {
				return
			}
		}
	}

	if expectPtr, ok := expect.BaseType.(ast.ReferenceType); ok {
		if exprPtr, ok := exprType.BaseType.(ast.ReferenceType); ok {
			if expectPtr.Referrer.ActualTypesEqual(exprPtr.Referrer) && exprPtr.IsMutable && !expectPtr.IsMutable {
				return
			}
		}
	}

	s.Err(loc, "Mismatched types: %s and %s", expect.String(), exprType.String())
}

type TypeCheck struct {
	functions []*ast.Function
}

func (v *TypeCheck) pushFunction(fn *ast.Function) {
	v.functions = append(v.functions, fn)
}

func (v *TypeCheck) popFunction() {
	v.functions = v.functions[:len(v.functions)-1]
}

func (v *TypeCheck) Function() *ast.Function {
	return v.functions[len(v.functions)-1]
}

func (v *TypeCheck) Init(s *SemanticAnalyzer) {
	v.functions = nil
}

func (v *TypeCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *TypeCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *TypeCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {
	switch n := n.(type) {
	case *ast.FunctionDecl, *ast.LambdaExpr:
		v.popFunction()

	case *ast.AssignStat:
		v.CheckAssignStat(s, n)

	case *ast.BinopAssignStat:
		v.CheckBinopAssignStat(s, n)

	case *ast.DestructAssignStat:
		v.CheckDestructAssignStat(s, n)

	case *ast.DestructBinopAssignStat:
		v.CheckDestructBinopAssignStat(s, n)
	}
}

func (v *TypeCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	switch n := n.(type) {
	case *ast.FunctionDecl:
		v.pushFunction(n.Function)

	case *ast.LambdaExpr:
		v.pushFunction(n.Function)

	case *ast.VariableDecl:
		v.CheckVariableDecl(s, n)

	case *ast.DestructVarDecl:
		v.CheckDestructVarDecl(s, n)

	case *ast.ReturnStat:
		v.CheckReturnStat(s, n)

	case *ast.IfStat:
		v.CheckIfStat(s, n)

	case *ast.MatchStat:
		v.CheckMatchStat(s, n)

	case *ast.ArrayLenExpr:
		v.CheckArrayLenExpr(s, n)

	case *ast.UnaryExpr:
		v.CheckUnaryExpr(s, n)

	case *ast.BinaryExpr:
		v.CheckBinaryExpr(s, n)

	case *ast.CastExpr:
		v.CheckCastExpr(s, n)

	case *ast.CallExpr:
		v.CheckCallExpr(s, n)

	case *ast.ArrayAccessExpr:
		v.CheckArrayAccessExpr(s, n)

	case *ast.DerefAccessExpr:
		v.CheckDerefAccessExpr(s, n)

	case *ast.NumericLiteral:
		v.CheckNumericLiteral(s, n)

	case *ast.CompositeLiteral:
		v.CheckCompositeLiteral(s, n)

	case *ast.TupleLiteral:
		v.CheckTupleLiteral(s, n)

	case *ast.EnumLiteral:
		v.CheckEnumLiteral(s, n)

	case *ast.StructAccessExpr:
		v.CheckStructAccessExpr(s, n)
	}
}

func (v *TypeCheck) Finalize(s *SemanticAnalyzer) {

}

func typeRefTo(typ ast.Type) *ast.TypeReference {
	return ast.NewTypeReference(typ, nil)
}

func (v *TypeCheck) CheckStructAccessExpr(s *SemanticAnalyzer, access *ast.StructAccessExpr) {
	structType := access.Struct.GetType().BaseType.ActualType().(ast.StructType)
	member := structType.GetMember(access.Member)
	if !member.Public && structType.Module != s.Submodule.Parent {
		s.Err(access, "Cannot access private struct member `%s`", access.Member)
	}
}

func (v *TypeCheck) CheckVariableDecl(s *SemanticAnalyzer, decl *ast.VariableDecl) {
	if decl.Variable.Type.BaseType.ActualType() == ast.PRIMITIVE_void {
		s.Err(decl, "Variable cannot be of type `void`")
	}

	if decl.Assignment != nil {
		expectType(s, decl, decl.Variable.Type, &decl.Assignment)
	}
}

func (v *TypeCheck) CheckDestructVarDecl(s *SemanticAnalyzer, decl *ast.DestructVarDecl) {
	tt, ok := decl.Assignment.GetType().BaseType.ActualType().(ast.TupleType)
	if !ok {
		s.Err(decl, "Assignment to destructing variable declaration must be tuple, was `%s`", decl.Assignment.GetType())
	}

	if len(tt.Members) != len(decl.Variables) {
		s.Err(decl.Assignment, "Destructured tuple must have %d values, had %d", len(decl.Variables), len(tt.Members))
	}
}

func (v *TypeCheck) CheckReturnStat(s *SemanticAnalyzer, stat *ast.ReturnStat) {
	if stat.Value == nil {
		if v.Function().Type.Return.BaseType.ActualType() != ast.PRIMITIVE_void {
			s.Err(stat, "Cannot return void from function `%s` of type `%s`",
				v.Function().Name, v.Function().Type.Return.String())
		}
	} else {
		if v.Function().Type.Return.BaseType == ast.PRIMITIVE_void {
			s.Err(stat.Value, "Cannot return expression from void function")
		} else {
			expectType(s, stat.Value, v.Function().Type.Return, &stat.Value)
		}
	}
}

func (v *TypeCheck) CheckIfStat(s *SemanticAnalyzer, stat *ast.IfStat) {
	for _, expr := range stat.Exprs {
		if expr.GetType().BaseType != ast.PRIMITIVE_bool {
			s.Err(expr, "If condition must have a boolean condition")
		}
	}
}

func (v *TypeCheck) CheckMatchStat(s *SemanticAnalyzer, stat *ast.MatchStat) {
	// TODO: Handle string and integer matches
	et, isEnum := stat.Target.GetType().BaseType.ActualType().(ast.EnumType)
	for pattern, _ := range stat.Branches {
		if _, isDiscard := pattern.(*ast.DiscardAccessExpr); isDiscard {
			continue
		}

		if isEnum {
			patt, ok := pattern.(*ast.EnumPatternExpr)
			if !ok {
				s.Err(pattern, "Expected enum pattern in match on enum type `%s`", stat.Target.GetType().String())
				continue
			}

			mem, ok := et.GetMember(patt.MemberName.Name)
			if !ok {
				s.Err(patt, "Enum type `%s` has no such member `%s`", stat.Target.GetType().String(), patt.MemberName.Name)
				continue
			}

			_, isStruct := mem.Type.(ast.StructType)
			_, isTuple := mem.Type.(ast.TupleType)
			if !isStruct && !isTuple && len(patt.Variables) > 0 {
				s.Err(patt, "Tried destructuring simple enum member `%s`", patt.MemberName.Name)
			}
		}
	}

}

func (v *TypeCheck) CheckAssignStat(s *SemanticAnalyzer, stat *ast.AssignStat) {
	if stat.Access.GetType() != nil {
		expectType(s, stat, stat.Access.GetType(), &stat.Assignment)
	}
}

func (v *TypeCheck) CheckBinopAssignStat(s *SemanticAnalyzer, stat *ast.BinopAssignStat) {
	if stat.Access.GetType() != nil {
		expectType(s, stat, stat.Access.GetType(), &stat.Assignment)
	}
}

func (v *TypeCheck) CheckDestructAssignStat(s *SemanticAnalyzer, stat *ast.DestructAssignStat) {
	tt, ok := stat.Assignment.GetType().BaseType.ActualType().(ast.TupleType)
	if !ok {
		s.Err(stat, "Value in destruturing assignment must be tuple, was `%s`", stat.Assignment.GetType())
	}

	if len(tt.Members) != len(stat.Accesses) {
		s.Err(stat.Assignment, "Destructured tuple must have %d values, had %d", len(stat.Accesses), len(tt.Members))
	}

	for idx, acc := range stat.Accesses {
		if acc.GetType() != nil && !acc.GetType().ActualTypesEqual(tt.Members[idx]) {
			s.Err(acc, "Mismatched types: `%s` and `%s`", acc.GetType().String(), tt.Members[idx].String())
		}
	}
}

func (v *TypeCheck) CheckDestructBinopAssignStat(s *SemanticAnalyzer, stat *ast.DestructBinopAssignStat) {
	tt, ok := stat.Assignment.GetType().BaseType.ActualType().(ast.TupleType)
	if !ok {
		s.Err(stat, "Value in destruturing assignment must be tuple, was `%s`", stat.Assignment.GetType())
	}

	if len(tt.Members) != len(stat.Accesses) {
		s.Err(stat.Assignment, "Destructured tuple must have %d values, had %d", len(stat.Accesses), len(tt.Members))
	}

	for idx, acc := range stat.Accesses {
		if acc.GetType() != nil && !acc.GetType().ActualTypesEqual(tt.Members[idx]) {
			s.Err(acc, "Mismatched types: `%s` and `%s`", acc.GetType().String(), tt.Members[idx].String())
		}
	}
}

func (v *TypeCheck) CheckArrayLenExpr(s *SemanticAnalyzer, expr *ast.ArrayLenExpr) {

}

func (v *TypeCheck) CheckUnaryExpr(s *SemanticAnalyzer, expr *ast.UnaryExpr) {
	switch expr.Op {
	case parser.UNOP_LOG_NOT:
		if !expr.Expr.GetType().ActualTypesEqual(typeRefTo(ast.PRIMITIVE_bool)) {
			s.Err(expr, "Used logical not on non-boolean expression")
		}
	case parser.UNOP_BIT_NOT:
		if !(expr.Expr.GetType().BaseType.IsIntegerType() || expr.Expr.GetType().BaseType.IsFloatingType()) {
			s.Err(expr, "Used bitwise not on non-numeric type")
		}
	case parser.UNOP_NEGATIVE:
		if !(expr.Expr.GetType().BaseType.IsIntegerType() || expr.Expr.GetType().BaseType.IsFloatingType()) {
			s.Err(expr, "Used negative on non-numeric type")
		}
	default:
		panic("unknown unary op")
	}
}

func (v *TypeCheck) CheckBinaryExpr(s *SemanticAnalyzer, expr *ast.BinaryExpr) {
	switch expr.Op {
	case parser.BINOP_EQ, parser.BINOP_NOT_EQ:
		if !expr.Lhand.GetType().ActualTypesEqual(expr.Rhand.GetType()) {
			s.Err(expr, "Operands for binary operator `%s` must have the same type, have `%s` and `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().String(), expr.Rhand.GetType().String())
		} else if lht := expr.Lhand.GetType(); !(lht.ActualTypesEqual(typeRefTo(ast.PRIMITIVE_bool)) || lht.BaseType.IsIntegerType() || lht.BaseType.IsFloatingType() || lht.BaseType.LevelsOfIndirection() > 0) {
			s.Err(expr, "Operands for binary operator `%s` must be numeric, or pointers or booleans, have `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().String())
		}

	case parser.BINOP_ADD, parser.BINOP_SUB, parser.BINOP_MUL, parser.BINOP_DIV, parser.BINOP_MOD,
		parser.BINOP_GREATER, parser.BINOP_LESS, parser.BINOP_GREATER_EQ, parser.BINOP_LESS_EQ,
		parser.BINOP_BIT_AND, parser.BINOP_BIT_OR, parser.BINOP_BIT_XOR:
		if !expr.Lhand.GetType().ActualTypesEqual(expr.Rhand.GetType()) {
			s.Err(expr, "Operands for binary operator `%s` must have the same type, have `%s` and `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().String(), expr.Rhand.GetType().String())
		} else if lht := expr.Lhand.GetType(); !(lht.BaseType.IsIntegerType() || lht.BaseType.IsFloatingType() || lht.BaseType.LevelsOfIndirection() > 0) {
			s.Err(expr, "Operands for binary operator `%s` must be numeric or pointers, have `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().String())
		}

	case parser.BINOP_BIT_LEFT, parser.BINOP_BIT_RIGHT:
		if lht := expr.Lhand.GetType(); !(lht.BaseType.IsFloatingType() || lht.BaseType.IsIntegerType() || lht.BaseType.LevelsOfIndirection() > 0) {
			s.Err(expr.Lhand, "Left-hand operand for bitshift operator `%s` must be numeric or a pointer, have `%s`",
				expr.Op.OpString(), lht.String())
		} else if !expr.Rhand.GetType().BaseType.IsIntegerType() {
			s.Err(expr.Rhand, "Right-hand operatnd for bitshift operator `%s` must be an integer, have `%s`",
				expr.Op.OpString(), expr.Rhand.GetType().String())
		}

	case parser.BINOP_LOG_AND, parser.BINOP_LOG_OR:
		if !expr.Lhand.GetType().ActualTypesEqual(typeRefTo(ast.PRIMITIVE_bool)) || !expr.Lhand.GetType().ActualTypesEqual(expr.Rhand.GetType()) {
			s.Err(expr, "Operands for logical operator `%s` must have same boolean type, have `%s` and `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().String(), expr.Rhand.GetType().String())
		}

	default:
		panic("unimplemented bin operation")
	}
}

func (v *TypeCheck) CheckCastExpr(s *SemanticAnalyzer, expr *ast.CastExpr) {
	if expr.Type.Equals(expr.Expr.GetType()) {
		s.Warn(expr, "Casting expression of type `%s` to the same type",
			expr.Type.String())
	} else if !expr.Expr.GetType().CanCastTo(expr.Type) {
		s.Err(expr, "Cannot cast expression of type `%s` to type `%s`",
			expr.Expr.GetType().String(), expr.Type.String())
	}
}

func (v *TypeCheck) CheckCallExpr(s *SemanticAnalyzer, expr *ast.CallExpr) {
	fnType := expr.Function.GetType().BaseType.ActualType().(ast.FunctionType)

	argLen := len(expr.Arguments)
	paramLen := len(fnType.Parameters)

	// attributes defaults
	isVariadic := fnType.IsVariadic
	c := false // if we're calling a C function

	// find them attributes yo
	if fnType.Attrs() != nil {
		c = fnType.Attrs().Contains("c")
	}

	var fnName string
	fae, ok := expr.Function.(*ast.FunctionAccessExpr)
	if ok {
		fnName = fae.Function.Name
	} else {
		fnName = "some func"
	}

	if argLen < paramLen {
		s.Err(expr, "Call to `%s` has too few arguments, expects %d, have %d",
			fnName, paramLen, argLen)
		return
	} else if !isVariadic && argLen > paramLen {
		// we only care if it's not variadic
		s.Err(expr, "Call to `%s` has too many arguments, expects %d, have %d",
			fnName, paramLen, argLen)
		return
	}

	if fnType.Receiver != nil {
		if !expr.ReceiverAccess.GetType().ActualTypesEqual(fnType.Receiver) {
			expectType(s, expr, fnType.Receiver, &expr.ReceiverAccess)
		}
	}

	for i, arg := range expr.Arguments {
		if i >= len(fnType.Parameters) { // we have a variadic arg
			if !isVariadic {
				panic("woah")
			}

			if !c {
				panic("Variadic functions are only legal for C interoperability")
			}

			// varargs take type promotions. If we don't do these, the whole thing fucks up.
			switch arg.GetType().BaseType.ActualType() {
			case ast.PRIMITIVE_f32:
				expr.Arguments[i] = &ast.CastExpr{
					Expr: arg,
					Type: typeRefTo(ast.PRIMITIVE_f64),
				}
			case ast.PRIMITIVE_s8, ast.PRIMITIVE_s16:
				expr.Arguments[i] = &ast.CastExpr{
					Expr: arg,
					Type: typeRefTo(ast.PRIMITIVE_int),
				}
			case ast.PRIMITIVE_u8, ast.PRIMITIVE_u16:
				expr.Arguments[i] = &ast.CastExpr{
					Expr: arg,
					Type: typeRefTo(ast.PRIMITIVE_uint),
				}
			}
		} else {
			par := fnType.Parameters[i]
			if arg.GetType() != nil { // TODO should arg type ever be nil?
				expectType(s, arg, par, &arg)
			}
		}
	}
}

func (v *TypeCheck) CheckArrayAccessExpr(s *SemanticAnalyzer, expr *ast.ArrayAccessExpr) {
	_, isArray := expr.Array.GetType().BaseType.ActualType().(ast.ArrayType)
	_, isPointer := expr.Array.GetType().BaseType.ActualType().(ast.PointerType)
	if !isPointer && !isArray {
		s.Err(expr, "Cannot index type `%s` as an array", expr.Array.GetType().String())
	}

	if !expr.Subscript.GetType().BaseType.IsIntegerType() {
		s.Err(expr, "Array subscript must be an integer type, have `%s`", expr.Subscript.GetType().String())
	}
}

func (v *TypeCheck) CheckDerefAccessExpr(s *SemanticAnalyzer, expr *ast.DerefAccessExpr) {
	if !ast.IsPointerOrReferenceType(expr.Expr.GetType().BaseType) {
		s.Err(expr, "Cannot dereference expression of type `%s`", expr.Expr.GetType().String())
	}
}

func (v *TypeCheck) CheckNumericLiteral(s *SemanticAnalyzer, lit *ast.NumericLiteral) {
	if !(lit.GetType().BaseType.IsIntegerType() || lit.GetType().BaseType.IsFloatingType()) {
		s.Err(lit, "Numeric literal was non-integer, non-float type: %s", lit.GetType().String())
	}

	if lit.IsFloat && lit.GetType().BaseType.IsIntegerType() {
		s.Err(lit, "Floating numeric literal has integer type: %s", lit.GetType().String())
	}

	if lit.GetType().BaseType.IsFloatingType() {
		// TODO
	} else {
		// Guaranteed to be integer type and integer literal
		var bits int

		switch lit.GetType().BaseType.ActualType() {
		case ast.PRIMITIVE_int, ast.PRIMITIVE_uint, ast.PRIMITIVE_uintptr:
			bits = 9000 // FIXME work out proper size
		case ast.PRIMITIVE_u8, ast.PRIMITIVE_s8:
			bits = 8
		case ast.PRIMITIVE_u16, ast.PRIMITIVE_s16:
			bits = 16
		case ast.PRIMITIVE_u32, ast.PRIMITIVE_s32:
			bits = 32
		case ast.PRIMITIVE_u64, ast.PRIMITIVE_s64:
			bits = 64
		case ast.PRIMITIVE_u128, ast.PRIMITIVE_s128:
			bits = 128
		default:
			panic("wrong type here: " + lit.GetType().String())
		}

		/*if lit.Type.IsSigned() {
			bits -= 1
			// FIXME this will give us a warning if a number is the lowest negative it can be
			// because the `-` is a separate expression. eg:
			// x: s8 = -128; // this gives a warning even though it's correct
		}*/

		if bits < lit.IntValue.BitLen() {
			s.Warn(lit, "Integer overflows %s", lit.GetType().String())
		}
	}
}

func exprsToTypeReferences(exprs []ast.Expr) []*ast.TypeReference {
	res := make([]*ast.TypeReference, 0, len(exprs))
	for _, expr := range exprs {
		res = append(res, expr.GetType())
	}
	return res
}

// parentEnum is nil if not in enum
func (v *TypeCheck) CheckTupleLiteral(s *SemanticAnalyzer, lit *ast.TupleLiteral) {
	tupleType, ok := lit.GetType().BaseType.ActualType().(ast.TupleType)
	if !ok {
		panic("Type of tuple literal was not `TupleType`")
	}
	memberTypes := tupleType.Members

	if len(lit.Members) != len(memberTypes) {
		s.Err(lit, "Invalid amount of entries in tuple")
	}

	var gcon *ast.GenericContext
	if lit.ParentEnumLiteral != nil {
		gcon = ast.NewGenericContext(lit.ParentEnumLiteral.GetType().BaseType.ActualType().(ast.EnumType).GenericParameters, lit.ParentEnumLiteral.Type.GenericArguments)
	} else {
		gcon = ast.NewGenericContext(nil, nil)
	}

	for idx, mem := range lit.Members {
		expectType(s, mem, gcon.Get(memberTypes[idx]), &mem)
	}
}

func (v *TypeCheck) CheckCompositeLiteral(s *SemanticAnalyzer, lit *ast.CompositeLiteral) {
	gcon := ast.NewGenericContext([]*ast.SubstitutionType{}, []*ast.TypeReference{})
	if len(lit.Type.GenericArguments) > 0 {
		gcon = ast.NewGenericContextFromTypeReference(lit.Type)
	}

	switch typ := lit.Type.BaseType.ActualType().(type) {
	case ast.ArrayType:
		memType := typ.MemberType
		for i, mem := range lit.Values {
			expectType(s, mem, memType, &mem)

			if lit.Fields[i] != "" {
				s.Err(mem, "Unexpected field in array literal: `%s`", lit.Fields[i])
			}
		}

	case ast.StructType:
		for i, mem := range lit.Values {
			name := lit.Fields[i]

			if name == "" {
				s.Err(mem, "Missing field in struct literal")
				continue
			}

			sMem := typ.GetMember(name)
			if sMem == nil {
				s.Err(lit, "No member named `%s` on struct of type `%s`", name, typ.String())
			}

			sMemType := gcon.Replace(sMem.Type)
			expectType(s, mem, sMemType, &mem)
		}

	default:
		panic("composite literal has neither struct nor array type")
	}
}

func (v *TypeCheck) CheckEnumLiteral(s *SemanticAnalyzer, lit *ast.EnumLiteral) {
	enumType, ok := lit.Type.BaseType.ActualType().(ast.EnumType)
	if !ok {
		panic("Type of enum literal was not `EnumType`")
	}

	memIdx := enumType.MemberIndex(lit.Member)

	if memIdx < 0 || memIdx >= len(enumType.Members) {
		s.Err(lit, "Enum `%s` has no member `%s`", lit.Type.String(), lit.Member)
		return
	}
}
