package parser

import (
	"fmt"
	"os"

	"github.com/ark-lang/ark/util"
)

// IMPORTANT NOTE for setTypeHint():
// When implementing this function for an Expr, only set the Expr's Type if
// you are on a lowest-level Expr, ie. a literal. That means, if you Expr
// contains a pointer to another Expr(s), simple pass the type hint along to that
// Expr(s) then return.

type semanticAnalyzer struct {
	file     *File
	function *Function // the function we're in, or nil if we aren't
}

func (v *semanticAnalyzer) err(thing Locatable, err string, stuff ...interface{}) {
	line, char := thing.Pos()
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Semantic error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		v.file.Name, line, char, fmt.Sprintf(err, stuff...))
	os.Exit(2)
}

func (v *semanticAnalyzer) warn(thing Locatable, err string, stuff ...interface{}) {
	line, char := thing.Pos()
	fmt.Printf(util.TEXT_YELLOW+util.TEXT_BOLD+"Semantic warning:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		v.file.Name, line, char, fmt.Sprintf(err, stuff...))
}

func (v *semanticAnalyzer) warnDeprecated(thing Locatable, typ, name, message string) {
	mess := fmt.Sprintf("Access of deprecated %s `%s`", typ, name)
	if message == "" {
		v.warn(thing, mess)
	} else {
		v.warn(thing, mess+": "+message)
	}
}

func (v *semanticAnalyzer) checkDuplicateAttrs(attrs []*Attr) {
	encountered := make(map[string]bool)
	for _, attr := range attrs {
		if encountered[attr.Key] {
			v.err(attr, "Duplicate attribute `%s`", attr.Key)
		}
		encountered[attr.Key] = true
	}
}

func (v *semanticAnalyzer) analyze() {
	for _, node := range v.file.Nodes {
		node.analyze(v)
	}
}

func (v *Block) analyze(s *semanticAnalyzer) {
	for _, n := range v.Nodes {
		n.analyze(s)
	}
}

func (v *Function) analyze(s *semanticAnalyzer) {
	// make sure there are no illegal attributes
	s.checkDuplicateAttrs(v.Attrs)
	for _, attr := range v.Attrs {
		switch attr.Key {
		case "deprecated":
			// value is optional, nothing to check
		case "c":
			// idk yet who cares
		case "variadic":
			// idk yet who cares
		default:
			s.err(attr, "Invalid function attribute key `%s`", attr.Key)
		}
	}

	s.function = v
	if v.Body != nil {
		v.Body.analyze(s)
	}
	s.function = nil
}

func (v *StructType) analyze(s *semanticAnalyzer) {
	// make sure there are no illegal attributes
	s.checkDuplicateAttrs(v.attrs)
	for _, attr := range v.Attrs() {
		switch attr.Key {
		case "packed":
			if attr.Value != "" {
				s.err(attr, "Struct attribute `%s` doesn't expect value", attr.Key)
			}
		case "deprecated":
			// value is optional, nothing to check
		default:
			s.err(attr, "Invalid struct attribute key `%s`", attr.Key)
		}
	}

	for _, decl := range v.Variables {
		decl.analyze(s)
	}
}

func (v *Variable) analyze(s *semanticAnalyzer) {
	// make sure there are no illegal attributes
	s.checkDuplicateAttrs(v.Attrs)
	for _, attr := range v.Attrs {
		switch attr.Key {
		case "deprecated":
			// value is optional, nothing to check
		default:
			s.err(attr, "Invalid variable attribute key `%s`", attr.Key)
		}
	}
}

/**
 * Declarations
 */

func (v *VariableDecl) analyze(s *semanticAnalyzer) {
	v.Variable.analyze(s)
	if v.Assignment != nil {
		v.Assignment.setTypeHint(v.Variable.Type)
		v.Assignment.analyze(s)

		if v.Variable.Type == nil { // type is inferred
			v.Variable.Type = v.Assignment.GetType()
		} else if v.Variable.Type != v.Assignment.GetType() {
			s.err(v, "Cannot assign expression of type `%s` to variable of type `%s`",
				v.Assignment.GetType().TypeName(), v.Variable.Type.TypeName())
		}
	}

	if dep := getAttr(v.Variable.Type.Attrs(), "deprecated"); dep != nil {
		s.warnDeprecated(v, "type", v.Variable.Type.TypeName(), dep.Value)
	}

	s.checkAttrsDistanceFromLine(v.Variable.Attrs, v.lineNumber, "variable", v.Variable.Name)
}

func (v *StructDecl) analyze(s *semanticAnalyzer) {
	v.Struct.analyze(s)
	s.checkAttrsDistanceFromLine(v.Struct.Attrs(), v.lineNumber, "type", v.Struct.TypeName())
}

func (v *FunctionDecl) analyze(s *semanticAnalyzer) {
	v.Function.analyze(s)
	s.checkAttrsDistanceFromLine(v.Function.Attrs, v.lineNumber, "function", v.Function.Name)
}

/*
 * Statements
 */

func (v *ReturnStat) analyze(s *semanticAnalyzer) {
	if s.function == nil {
		s.err(v, "Return statement must be in a function")
	}

	if v.Value == nil {
		if s.function.ReturnType != nil {
			s.err(v.Value, "Cannot return void from function `%s` of type `%s`",
				s.function.Name, s.function.ReturnType.TypeName())
		}
	} else {
		v.Value.setTypeHint(s.function.ReturnType)
		v.Value.analyze(s)
		if v.Value.GetType() != s.function.ReturnType {
			s.err(v.Value, "Cannot return expression of type `%s` from function `%s` of type `%s`",
				v.Value.GetType().TypeName(), s.function.Name, s.function.ReturnType.TypeName())
		}
	}
}

func (v *IfStat) analyze(s *semanticAnalyzer) {
	for _, expr := range v.Exprs {
		expr.setTypeHint(PRIMITIVE_bool)
		expr.analyze(s)
		if expr.GetType() != PRIMITIVE_bool {
			s.err(expr, "If condition must be of type `bool`")
		}
	}

	for _, body := range v.Bodies {
		body.analyze(s)
	}

	if v.Else != nil {
		v.Else.analyze(s)
	}

}

// CallStat

func (v *CallStat) analyze(s *semanticAnalyzer) {
	v.Call.analyze(s)
}

// AssignStat

func (v *AssignStat) analyze(s *semanticAnalyzer) {
	if (v.Deref == nil) == (v.Access == nil) { // make sure aren't both not null or null
		panic("oh no")
	}

	var lhs Expr
	if v.Deref != nil {
		lhs = v.Deref
	} else if v.Access != nil {
		var variable *Variable
		if len(v.Access.StructVariables) > 0 {
			variable = v.Access.StructVariables[0]
		} else {
			variable = v.Access.Variable
		}
		if !variable.Mutable {
			s.err(v, "Cannot assign value to immutable variable `%s`", variable.Name)
		}
		lhs = v.Access
	}

	v.Assignment.setTypeHint(lhs.GetType())
	v.Assignment.analyze(s)
	lhs.analyze(s)
	if lhs.GetType() != v.Assignment.GetType() {
		s.err(v, "Mismatched types: `%s` and `%s`", lhs.GetType().TypeName(), v.Assignment.GetType().TypeName())
	}
}

// LoopStat

func (v *LoopStat) analyze(s *semanticAnalyzer) {
	v.Body.analyze(s)

	switch v.LoopType {
	case LOOP_TYPE_INFINITE:
	case LOOP_TYPE_CONDITIONAL:
		v.Condition.setTypeHint(PRIMITIVE_bool)
		v.Condition.analyze(s)
	default:
		panic("invalid loop type")
	}
}

/*
 * Expressions
 */

// UnaryExpr

func (v *UnaryExpr) analyze(s *semanticAnalyzer) {
	v.Expr.analyze(s)

	switch v.Op {
	case UNOP_LOG_NOT:
		if v.Expr.GetType() == PRIMITIVE_bool {
			v.Type = PRIMITIVE_bool
		} else {
			s.err(v, "Used logical not on non-bool")
		}
	case UNOP_BIT_NOT:
		if v.Expr.GetType().IsIntegerType() || v.Expr.GetType().IsFloatingType() {
			v.Type = v.Expr.GetType()
		} else {
			s.err(v, "Used bitwise not on non-numeric type")
		}
	case UNOP_NEGATIVE:
		if v.Expr.GetType().IsIntegerType() || v.Expr.GetType().IsFloatingType() {
			v.Type = v.Expr.GetType()
		} else {
			s.err(v, "Used negative on non-numeric type")
		}
	case UNOP_ADDRESS:
		if _, ok := v.Expr.(*AccessExpr); !ok {
			s.err(v, "Cannot take address of non-variable")
		}
		v.Type = pointerTo(v.Expr.GetType())
	default:
		panic("whoops")
	}
}

func (v *UnaryExpr) setTypeHint(t Type) {
	switch v.Op {
	case UNOP_LOG_NOT:
		v.Expr.setTypeHint(PRIMITIVE_bool)
	case UNOP_BIT_NOT, UNOP_NEGATIVE:
		v.Expr.setTypeHint(t)
	case UNOP_ADDRESS:
	default:
		panic("whoops")
	}
}

// BinaryExpr

func (v *BinaryExpr) analyze(s *semanticAnalyzer) {
	v.Lhand.analyze(s)
	v.Rhand.analyze(s)

	switch v.Op {
	case BINOP_ADD, BINOP_SUB, BINOP_MUL, BINOP_DIV, BINOP_MOD,
		BINOP_GREATER, BINOP_LESS, BINOP_GREATER_EQ, BINOP_LESS_EQ, BINOP_EQ, BINOP_NOT_EQ,
		BINOP_BIT_AND, BINOP_BIT_OR, BINOP_BIT_XOR:
		if v.Lhand.GetType() != v.Rhand.GetType() {
			s.err(v, "Operands for binary operator `%s` must have the same type, have `%s` and `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName(), v.Rhand.GetType().TypeName())
		} else if lht := v.Lhand.GetType(); !(lht.IsIntegerType() || lht.IsFloatingType() || lht.LevelsOfIndirection() > 0) {
			s.err(v, "Operands for binary operator `%s` must be numeric or pointers, have `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName())
		} else {
			switch v.Op.Category() {
			case OP_ARITHMETIC:
				v.Type = v.Lhand.GetType()
			case OP_COMPARISON:
				v.Type = PRIMITIVE_bool
			default:
				panic("shouldn't happen ever, one of the devs really fucked up, sorry!")
			}
		}

	case BINOP_BIT_LEFT, BINOP_BIT_RIGHT:
		if lht := v.Lhand.GetType(); !(lht.IsFloatingType() || lht.IsIntegerType() || lht.LevelsOfIndirection() > 0) {
			s.err(v.Lhand, "Left-hand operand for bitshift operator `%s` must be numeric or a pointer, have `%s`",
				v.Op.OpString(), lht.TypeName())
		} else if !v.Rhand.GetType().IsIntegerType() {
			s.err(v.Rhand, "Right-hand operatnd for bitshift operator `%s` must be an integer, have `%s`",
				v.Op.OpString(), v.Rhand.GetType().TypeName())
		} else {
			v.Type = lht
		}

	case BINOP_LOG_AND, BINOP_LOG_OR:
		if v.Lhand.GetType() != PRIMITIVE_bool || v.Rhand.GetType() != PRIMITIVE_bool {
			s.err(v, "Operands for logical operator `%s` must have the same type, have `%s` and `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName(), v.Rhand.GetType().TypeName())
		} else {
			v.Type = PRIMITIVE_bool
		}

	default:
		panic("unimplemented bin operation")
	}
}

func (v *BinaryExpr) setTypeHint(t Type) {
	switch v.Op.Category() {
	case OP_ARITHMETIC, OP_BITWISE:
		if v.Op == BINOP_BIT_LEFT || v.Op == BINOP_BIT_RIGHT {
			v.Rhand.setTypeHint(nil)
			v.Lhand.setTypeHint(t)
			return
		}
		if t == nil {
			if v.Lhand.GetType() == nil && v.Rhand.GetType() != nil {
				v.Lhand.setTypeHint(v.Rhand.GetType())
				return
			} else if v.Rhand.GetType() == nil && v.Lhand.GetType() != nil {
				v.Rhand.setTypeHint(v.Lhand.GetType())
				return
			}
		}
		v.Lhand.setTypeHint(t)
		v.Rhand.setTypeHint(t)
	case OP_COMPARISON:
		if v.Lhand.GetType() == nil && v.Rhand.GetType() != nil {
			v.Lhand.setTypeHint(v.Rhand.GetType())
		} else if v.Rhand.GetType() == nil && v.Lhand.GetType() != nil {
			v.Rhand.setTypeHint(v.Lhand.GetType())
		} else {
			v.Lhand.setTypeHint(nil)
			v.Rhand.setTypeHint(nil)
		}
	case OP_LOGICAL:
		v.Lhand.setTypeHint(PRIMITIVE_bool)
		v.Rhand.setTypeHint(PRIMITIVE_bool)
	default:
		panic("missing opcategory")
	}
}

// IntegerLiteral

func (v *IntegerLiteral) analyze(s *semanticAnalyzer) {}

func (v *IntegerLiteral) setTypeHint(t Type) {
	switch t {
	case PRIMITIVE_int, PRIMITIVE_uint,
		PRIMITIVE_i8, PRIMITIVE_i16, PRIMITIVE_i32, PRIMITIVE_i64, PRIMITIVE_i128,
		PRIMITIVE_u8, PRIMITIVE_u16, PRIMITIVE_u32, PRIMITIVE_u64, PRIMITIVE_u128:
		v.Type = t
	default:
		v.Type = PRIMITIVE_int // TODO check overflow
	}
}

// FloatingLiteral

func (v *FloatingLiteral) analyze(s *semanticAnalyzer) {}

func (v *FloatingLiteral) setTypeHint(t Type) {
	switch t {
	case PRIMITIVE_f64, PRIMITIVE_f32, PRIMITIVE_f128:
		v.Type = t
	default:
		v.Type = PRIMITIVE_f32
	}
}

// StringLiteral

func (v *StringLiteral) analyze(s *semanticAnalyzer) {}
func (v *StringLiteral) setTypeHint(t Type)          {}

// RuneLiteral

func (v *RuneLiteral) analyze(s *semanticAnalyzer) {}
func (v *RuneLiteral) setTypeHint(t Type)          {}

// CastExpr

func (v *CastExpr) analyze(s *semanticAnalyzer) {
	v.Expr.setTypeHint(nil)
	v.Expr.analyze(s)
	if v.Type == v.Expr.GetType() {
		s.warn(v, "Casting expression of type `%s` to the same type",
			v.Type.TypeName())
	} else if !v.Expr.GetType().CanCastTo(v.Type) {
		s.err(v, "Cannot cast expression of type `%s` to type `%s`",
			v.Expr.GetType().TypeName(), v.Type.TypeName())
	}
}

func (v *CastExpr) setTypeHint(t Type) {}

// CallExpr

func (v *CallExpr) analyze(s *semanticAnalyzer) {
	argLen := len(v.Arguments)
	paramLen := len(v.Function.Parameters)

	// attributes defaults
	isVariadic := false

	// find them attributes yo
	if v.Function.Attrs != nil {
		attributes := v.Function.Attrs

		// todo hashmap or some shit
		for _, attr := range attributes {
			switch attr.Key {
			case "variadic":
				isVariadic = true
			default:
				// do nothing
			}
		}
	}

	if argLen < paramLen {
		s.err(v, "Call to `%s` has too few arguments, expects %d, have %d",
			v.Function.Name, paramLen, argLen)
	} else if !isVariadic && argLen > paramLen {
		// we only care if it's not variadic
		s.err(v, "Call to `%s` has too many arguments, expects %d, have %d",
			v.Function.Name, paramLen, argLen)
	}

	for i, arg := range v.Arguments {
		// this fucks up when something is variadic
		// presumably because it has to infer the type
		// but there is no type to look at?
		// @MovingToMars
		arg.setTypeHint(v.Function.Parameters[i].Variable.Type)
		arg.analyze(s)
	}

	if dep := getAttr(v.Function.Attrs, "deprecated"); dep != nil {
		s.warnDeprecated(v, "function", v.Function.Name, dep.Value)
	}
}

func (v *CallExpr) setTypeHint(t Type) {}

// AccessExpr

func (v *AccessExpr) analyze(s *semanticAnalyzer) {
	for _, variable := range append(v.StructVariables, v.Variable) {
		if dep := getAttr(variable.Attrs, "deprecated"); dep != nil {
			s.warnDeprecated(v, "variable", variable.Name, dep.Value)
		}
	}
}

func (v *AccessExpr) setTypeHint(t Type) {}

// DerefExpr

func (v *DerefExpr) analyze(s *semanticAnalyzer) {
	if ptr, ok := v.Expr.GetType().(PointerType); !ok {
		s.err(v, "Cannot dereference expression of type `%s`", v.Expr.GetType().TypeName())
	} else {
		v.Type = ptr.Addressee
	}
}

func (v *DerefExpr) setTypeHint(t Type) {
	v.Expr.setTypeHint(pointerTo(t))
}

// BracketExpr

func (v *BracketExpr) analyze(s *semanticAnalyzer) {
	v.Expr.analyze(s)
}

func (v *BracketExpr) setTypeHint(t Type) {
	v.Expr.setTypeHint(t)
}
