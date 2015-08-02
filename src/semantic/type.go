package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type TypeCheck struct {
}

func (v *TypeCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *TypeCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *TypeCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.VariableDecl:
		v.CheckVariableDecl(s, n.(*parser.VariableDecl))

	case *parser.ReturnStat:
		v.CheckReturnStat(s, n.(*parser.ReturnStat))

	case *parser.IfStat:
		v.CheckIfStat(s, n.(*parser.IfStat))

	case *parser.AssignStat:
		v.CheckAssignStat(s, n.(*parser.AssignStat))

	case *parser.BinopAssignStat:
		v.CheckBinopAssignStat(s, n.(*parser.BinopAssignStat))

	case *parser.UnaryExpr:
		v.CheckUnaryExpr(s, n.(*parser.UnaryExpr))

	case *parser.BinaryExpr:
		v.CheckBinaryExpr(s, n.(*parser.BinaryExpr))

	case *parser.CastExpr:
		v.CheckCastExpr(s, n.(*parser.CastExpr))

	case *parser.CallExpr:
		v.CheckCallExpr(s, n.(*parser.CallExpr))

	case *parser.ArrayAccessExpr:
		v.CheckArrayAccessExpr(s, n.(*parser.ArrayAccessExpr))

	case *parser.TupleAccessExpr:
		v.CheckTupleAccessExpr(s, n.(*parser.TupleAccessExpr))

	case *parser.DerefAccessExpr:
		v.CheckDerefAccessExpr(s, n.(*parser.DerefAccessExpr))

	case *parser.NumericLiteral:
		v.CheckNumericLiteral(s, n.(*parser.NumericLiteral))

	case *parser.ArrayLiteral:
		v.CheckArrayLiteral(s, n.(*parser.ArrayLiteral))

	case *parser.TupleLiteral:
		v.CheckTupleLiteral(s, n.(*parser.TupleLiteral))

	case *parser.StructLiteral:
		v.CheckStructLiteral(s, n.(*parser.StructLiteral))

	case *parser.EnumLiteral:
		v.CheckEnumLiteral(s, n.(*parser.EnumLiteral))
	}
}

func (v *TypeCheck) CheckVariableDecl(s *SemanticAnalyzer, decl *parser.VariableDecl) {
	if decl.Assignment != nil {
		if !decl.Variable.Type.Equals(decl.Assignment.GetType()) {
			s.Err(decl, "Cannot assign expression of type `%s` to variable of type `%s`",
				decl.Assignment.GetType().TypeName(), decl.Variable.Type.TypeName())
		}
	}
}

func (v *TypeCheck) CheckReturnStat(s *SemanticAnalyzer, stat *parser.ReturnStat) {
	if stat.Value == nil {
		if s.Function.ReturnType != nil {
			s.Err(stat.Value, "Cannot return void from function `%s` of type `%s`",
				s.Function.Name, s.Function.ReturnType.TypeName())
		}
	} else {
		if s.Function.ReturnType == nil {
			s.Err(stat.Value, "Cannot return expression from void function")
		} else {
			if !stat.Value.GetType().Equals(s.Function.ReturnType) {
				s.Err(stat.Value, "Cannot return expression of type `%s` from function `%s` of type `%s`",
					stat.Value.GetType().TypeName(), s.Function.Name, s.Function.ReturnType.TypeName())
			}
		}
	}
}

func (v *TypeCheck) CheckIfStat(s *SemanticAnalyzer, stat *parser.IfStat) {
	for _, expr := range stat.Exprs {
		if expr.GetType() != parser.PRIMITIVE_bool {
			s.Err(expr, "If condition must be of type `bool`")
		}
	}

}

func (v *TypeCheck) CheckAssignStat(s *SemanticAnalyzer, stat *parser.AssignStat) {
	if !stat.Access.GetType().Equals(stat.Assignment.GetType()) {
		s.Err(stat, "Mismatched types: `%s` and `%s`", stat.Access.GetType().TypeName(), stat.Assignment.GetType().TypeName())
	}
}

func (v *TypeCheck) CheckBinopAssignStat(s *SemanticAnalyzer, stat *parser.BinopAssignStat) {
	if !stat.Access.GetType().Equals(stat.Assignment.GetType()) {
		s.Err(stat, "Mismatched types: `%s` and `%s`", stat.Access.GetType().TypeName(), stat.Assignment.GetType().TypeName())
	}
}

func (v *TypeCheck) CheckUnaryExpr(s *SemanticAnalyzer, expr *parser.UnaryExpr) {
	switch expr.Op {
	case parser.UNOP_LOG_NOT:
		if expr.Expr.GetType() != parser.PRIMITIVE_bool {
			s.Err(expr, "Used logical not on non-bool")
		}
	case parser.UNOP_BIT_NOT:
		if !(expr.Expr.GetType().IsIntegerType() || expr.Expr.GetType().IsFloatingType()) {
			s.Err(expr, "Used bitwise not on non-numeric type")
		}
	case parser.UNOP_NEGATIVE:
		if !(expr.Expr.GetType().IsIntegerType() || expr.Expr.GetType().IsFloatingType()) {
			s.Err(expr, "Used negative on non-numeric type")
		}
	default:
		panic("unknown unary op")
	}
}

func (v *TypeCheck) CheckBinaryExpr(s *SemanticAnalyzer, expr *parser.BinaryExpr) {
	switch expr.Op {
	case parser.BINOP_EQ, parser.BINOP_NOT_EQ:
		if !expr.Lhand.GetType().Equals(expr.Rhand.GetType()) {
			s.Err(expr, "Operands for binary operator `%s` must have the same type, have `%s` and `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().TypeName(), expr.Rhand.GetType().TypeName())
		} else if lht := expr.Lhand.GetType(); !(lht == parser.PRIMITIVE_bool || lht == parser.PRIMITIVE_rune || lht.IsIntegerType() || lht.IsFloatingType() || lht.LevelsOfIndirection() > 0) {
			s.Err(expr, "Operands for binary operator `%s` must be numeric, or pointers or booleans, have `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().TypeName())
		}

	case parser.BINOP_ADD, parser.BINOP_SUB, parser.BINOP_MUL, parser.BINOP_DIV, parser.BINOP_MOD,
		parser.BINOP_GREATER, parser.BINOP_LESS, parser.BINOP_GREATER_EQ, parser.BINOP_LESS_EQ,
		parser.BINOP_BIT_AND, parser.BINOP_BIT_OR, parser.BINOP_BIT_XOR:
		if !expr.Lhand.GetType().Equals(expr.Rhand.GetType()) {
			s.Err(expr, "Operands for binary operator `%s` must have the same type, have `%s` and `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().TypeName(), expr.Rhand.GetType().TypeName())
		} else if lht := expr.Lhand.GetType(); !(lht == parser.PRIMITIVE_rune || lht.IsIntegerType() || lht.IsFloatingType() || lht.LevelsOfIndirection() > 0) {
			s.Err(expr, "Operands for binary operator `%s` must be numeric or pointers, have `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().TypeName())
		}

	case parser.BINOP_BIT_LEFT, parser.BINOP_BIT_RIGHT:
		if lht := expr.Lhand.GetType(); !(lht.IsFloatingType() || lht.IsIntegerType() || lht.LevelsOfIndirection() > 0) {
			s.Err(expr.Lhand, "Left-hand operand for bitshift operator `%s` must be numeric or a pointer, have `%s`",
				expr.Op.OpString(), lht.TypeName())
		} else if !expr.Rhand.GetType().IsIntegerType() {
			s.Err(expr.Rhand, "Right-hand operatnd for bitshift operator `%s` must be an integer, have `%s`",
				expr.Op.OpString(), expr.Rhand.GetType().TypeName())
		}

	case parser.BINOP_LOG_AND, parser.BINOP_LOG_OR:
		if expr.Lhand.GetType() != parser.PRIMITIVE_bool || expr.Rhand.GetType() != parser.PRIMITIVE_bool {
			s.Err(expr, "Operands for logical operator `%s` must have the same type, have `%s` and `%s`",
				expr.Op.OpString(), expr.Lhand.GetType().TypeName(), expr.Rhand.GetType().TypeName())
		}

	default:
		panic("unimplemented bin operation")
	}
}

func (v *TypeCheck) CheckCastExpr(s *SemanticAnalyzer, expr *parser.CastExpr) {
	if expr.Type.Equals(expr.Expr.GetType()) {
		s.Warn(expr, "Casting expression of type `%s` to the same type",
			expr.Type.TypeName())
	} else if !expr.Expr.GetType().CanCastTo(expr.Type) {
		s.Err(expr, "Cannot cast expression of type `%s` to type `%s`",
			expr.Expr.GetType().TypeName(), expr.Type.TypeName())
	}
}

func (v *TypeCheck) CheckCallExpr(s *SemanticAnalyzer, expr *parser.CallExpr) {
	argLen := len(expr.Arguments)
	paramLen := len(expr.Function.Parameters)

	// attributes defaults
	isVariadic := expr.Function.IsVariadic
	c := false // if we're calling a C function

	// find them attributes yo
	if expr.Function.Attrs != nil {
		c = expr.Function.Attrs.Contains("c")
	}

	if argLen < paramLen {
		s.Err(expr, "Call to `%s` has too few arguments, expects %d, have %d",
			expr.Function.Name, paramLen, argLen)
	} else if !isVariadic && argLen > paramLen {
		// we only care if it's not variadic
		s.Err(expr, "Call to `%s` has too many arguments, expects %d, have %d",
			expr.Function.Name, paramLen, argLen)
	}

	for i, arg := range expr.Arguments {
		if i >= len(expr.Function.Parameters) { // we have a variadic arg
			if !isVariadic {
				panic("woah")
			}

			if !c {
				panic("The `variadic` attribute should only be used on calls to C functions")
			}

			// varargs take type promotions. If we don't do these, the whole thing fucks up.
			switch arg.GetType().ActualType() {
			case parser.PRIMITIVE_f32:
				expr.Arguments[i] = &parser.CastExpr{
					Expr: arg,
					Type: parser.PRIMITIVE_f64,
				}
			case parser.PRIMITIVE_s8, parser.PRIMITIVE_s16:
				expr.Arguments[i] = &parser.CastExpr{
					Expr: arg,
					Type: parser.PRIMITIVE_int,
				}
			case parser.PRIMITIVE_u8, parser.PRIMITIVE_u16:
				expr.Arguments[i] = &parser.CastExpr{
					Expr: arg,
					Type: parser.PRIMITIVE_uint,
				}
			}
		} else {
			if !arg.GetType().Equals(expr.Function.Parameters[i].Variable.Type) {
				s.Err(arg, "Mismatched types in function call: `%s` and `%s`", arg.GetType().TypeName(), expr.Function.Parameters[i].Variable.Type.TypeName())
			}
		}
	}
}

func (v *TypeCheck) CheckArrayAccessExpr(s *SemanticAnalyzer, expr *parser.ArrayAccessExpr) {
	if _, ok := expr.Array.GetType().(parser.ArrayType); !ok {
		s.Err(expr, "Cannot index type `%s` as an array", expr.Array.GetType().TypeName())
	}

	if !expr.Subscript.GetType().IsIntegerType() {
		s.Err(expr, "Array subscript must be an integer type, have `%s`", expr.Subscript.GetType().TypeName())
	}
}

func (v *TypeCheck) CheckTupleAccessExpr(s *SemanticAnalyzer, expr *parser.TupleAccessExpr) {
	tupleType, ok := expr.Tuple.GetType().ActualType().(*parser.TupleType)
	if !ok {
		s.Err(expr, "Cannot index type `%s` as a tuple", expr.Tuple.GetType().TypeName())
	}

	if expr.Index >= uint64(len(tupleType.Members)) {
		s.Err(expr, "Index `%d` (element %d) is greater than size of tuple `%s`", expr.Index, expr.Index+1, tupleType.TypeName())
	}
}

func (v *TypeCheck) CheckDerefAccessExpr(s *SemanticAnalyzer, expr *parser.DerefAccessExpr) {
	if _, ok := expr.Expr.GetType().(parser.PointerType); !ok {
		s.Err(expr, "Cannot dereference expression of type `%s`", expr.Expr.GetType().TypeName())
	}
}

func (v *TypeCheck) CheckNumericLiteral(s *SemanticAnalyzer, lit *parser.NumericLiteral) {
	if !(lit.Type.IsIntegerType() || lit.Type.IsFloatingType()) {
		s.Err(lit, "Numeric literal was non-integer, non-float type: %s", lit.Type.TypeName())
	}

	if lit.IsFloat && lit.Type.IsIntegerType() {
		s.Err(lit, "Floating numeric literal has integer type: %s", lit.Type.TypeName())
	}

	if lit.Type.IsFloatingType() {
		// TODO
	} else {
		// Guaranteed to be integer type and integer literal
		var bits int

		switch lit.Type.ActualType() {
		case parser.PRIMITIVE_int, parser.PRIMITIVE_uint:
			bits = 9000 // FIXME work out proper size
		case parser.PRIMITIVE_u8, parser.PRIMITIVE_s8:
			bits = 8
		case parser.PRIMITIVE_u16, parser.PRIMITIVE_s16:
			bits = 16
		case parser.PRIMITIVE_u32, parser.PRIMITIVE_s32:
			bits = 32
		case parser.PRIMITIVE_u64, parser.PRIMITIVE_s64:
			bits = 64
		case parser.PRIMITIVE_u128, parser.PRIMITIVE_s128:
			bits = 128
		default:
			panic("wrong type here: " + lit.Type.TypeName())
		}

		/*if lit.Type.IsSigned() {
			bits -= 1
			// FIXME this will give us a warning if a number is the lowest negative it can be
			// because the `-` is a separate expression. eg:
			// x: s8 = -128; // this gives a warning even though it's correct
		}*/

		if bits < lit.IntValue.BitLen() {
			s.Warn(lit, "Integer overflows %s", lit.Type.TypeName())
		}
	}
}

func (v *TypeCheck) CheckArrayLiteral(s *SemanticAnalyzer, lit *parser.ArrayLiteral) {
	// TODO: check array length once we get around to that stuff
	arrayType, ok := lit.Type.(parser.ArrayType)
	if !ok {
		panic("Type of array literal was not `ArrayType`")
	}

	memType := arrayType.MemberType
	for _, mem := range lit.Members {
		if !mem.GetType().Equals(memType) {
			s.Err(mem, "Cannot use element of type `%s` in array of type `%s`", mem.GetType().TypeName(), memType.TypeName())
		}
	}
}

func (v *TypeCheck) CheckTupleLiteral(s *SemanticAnalyzer, lit *parser.TupleLiteral) {
	tupleType, ok := lit.Type.ActualType().(*parser.TupleType)
	if !ok {
		panic("Type of tuple literal was not `TupleType`")
	}
	memberTypes := tupleType.Members

	if len(lit.Members) != len(memberTypes) {
		s.Err(lit, "Invalid amount of entries in tuple")
	}

	for idx, mem := range lit.Members {
		if !mem.GetType().Equals(memberTypes[idx]) {
			s.Err(lit, "Cannot use component of type `%s` in tuple position of type `%s`", mem.GetType().TypeName(), memberTypes[idx])
		}
	}
}

func (v *TypeCheck) CheckStructLiteral(s *SemanticAnalyzer, lit *parser.StructLiteral) {
	structType, ok := lit.Type.ActualType().(*parser.StructType)
	if !ok {
		panic("Type of struct literal was not `StructType`")
	}

	for name, mem := range lit.Values {
		decl := structType.GetVariableDecl(name)
		if decl == nil {
			s.Err(lit, "No member named `%s` on struct of type `%s`", name, structType.TypeName())
		}

		if !mem.GetType().Equals(decl.Variable.Type) {
			s.Err(lit, "Cannot use value of type `%s` as member of `%s` with type `%s`",
				mem.GetType().TypeName(), decl.Variable.Type.TypeName(), structType.TypeName())
		}
	}
}

func (v *TypeCheck) CheckEnumLiteral(s *SemanticAnalyzer, lit *parser.EnumLiteral) {
	enumType, ok := lit.Type.ActualType().(*parser.EnumType)
	if !ok {
		panic("Type of enum literal was not `EnumType`")
	}

	memIdx := enumType.MemberIndex(lit.Member)

	if memIdx < 0 || memIdx >= len(enumType.Members) {
		s.Err(lit, "Enum `%s` has no member `%s`", lit.Type.TypeName(), lit.Member)
		return
	}
}
