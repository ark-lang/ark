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
	module          *Module
	function        *Function // the function we're in, or nil if we aren't
	unresolvedNodes []Node
	modules         map[string]*Module
}

func (v *semanticAnalyzer) err(thing Locatable, err string, stuff ...interface{}) {
	filename, line, char := thing.Pos()
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Semantic error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		filename, line, char, fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_SEMANTIC)
}

func (v *semanticAnalyzer) warn(thing Locatable, err string, stuff ...interface{}) {
	filename, line, char := thing.Pos()
	fmt.Printf(util.TEXT_YELLOW+util.TEXT_BOLD+"Semantic warning:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		filename, line, char, fmt.Sprintf(err, stuff...))
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

func (v *semanticAnalyzer) analyze(modules map[string]*Module) {
	v.modules = modules

	// pass modules to resolve
	v.resolve(modules)

	for _, node := range v.module.Nodes {
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

func (v *ModuleDecl) analyze(s *semanticAnalyzer) {

}

func (v *EnumDecl) analyze(s *semanticAnalyzer) {
	// here we infer the enum integer type from given typed expressions, if any
	var enumValueType Type
	for _, member := range v.Body {
		if member.Value != nil && member.Value.GetType() != nil {
			if !member.Value.GetType().IsIntegerType() {
				s.err(v, "Enum member value must be an integer type, have `%s`", member.Value.GetType().TypeName())
			} else {
				if enumValueType == nil {
					enumValueType = member.Value.GetType()
				} else {
					s.err(v, "Enum values have conflicting types: `%s` and `%s`", enumValueType.TypeName(), member.Value.GetType().TypeName())
				}
			}
		}
	}

	// no expressions to infer from, just default to int
	if enumValueType == nil {
		enumValueType = PRIMITIVE_int
	}

	// holds current value
	index := uint64(0)

	for _, member := range v.Body {
		if member.Value == nil {
			member.Value = &IntegerLiteral{Value: index, Type: enumValueType}
			member.Value.analyze(s)
			index++
		} else {
			member.Value.setTypeHint(enumValueType)
			member.Value.analyze(s)

			if member.Value.GetType() != enumValueType {
				s.err(v, "Incompatible types in enum: `%s` and `%s`", member.Value.GetType().TypeName(), enumValueType.TypeName())
			}

			index, err := evaluateEnumExpr(member.Value)
			if err != nil {
				s.err(v, "Enum value `%s` must be constant", member.Name)
			}
			index++
		}
	}
}

func evaluateEnumExpr(expr Expr) (uint64, error) {
	return 0, nil // TODO
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

func (v *TraitType) analyze(s *semanticAnalyzer) {
	// make sure there are no illegal attributes
	s.checkDuplicateAttrs(v.attrs)
	for _, attr := range v.Attrs() {
		if attr.Key != "deprecated" {
			s.err(attr, "Invalid trait attribute key `%s`", attr.Key)
		}
	}

	for _, decl := range v.Functions {
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

func (v *TraitDecl) analyze(s *semanticAnalyzer) {
	v.Trait.analyze(s)
	s.checkAttrsDistanceFromLine(v.Trait.Attrs(), v.lineNumber, "type", v.Trait.TypeName())
}

func (v *ImplDecl) analyze(s *semanticAnalyzer) {
	// TODO
}

func (v *UseDecl) analyze(s *semanticAnalyzer) {

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
		if s.function.ReturnType == nil {
			s.err(v.Value, "Cannot return expression from void function")
		} else {
			v.Value.setTypeHint(s.function.ReturnType)
			v.Value.analyze(s)
			if v.Value.GetType() != s.function.ReturnType {
				s.err(v.Value, "Cannot return expression of type `%s` from function `%s` of type `%s`",
					v.Value.GetType().TypeName(), s.function.Name, s.function.ReturnType.TypeName())
			}
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

// BlockStat

func (v *BlockStat) analyze(s *semanticAnalyzer) {
	v.Block.analyze(s)
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
		//var variable *Variable

		/*if len(v.Access.StructVariables) > 0 { TODO
			variable = v.Access.StructVariables[0]
		} else {
			variable = v.Access.Variable
		}*/

		/*if !variable.Mutable {
			s.err(v, "Cannot assign value to immutable variable `%s`", variable.Name)
		}*/
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

// MatchStat

func (v *MatchStat) analyze(s *semanticAnalyzer) {
	v.Target.analyze(s)

	for pattern, stmt := range v.Branches {
		pattern.analyze(s)
		stmt.analyze(s)
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
	default:
		panic("whoops")
	}
}

// BinaryExpr

func (v *BinaryExpr) analyze(s *semanticAnalyzer) {
	v.Lhand.analyze(s)
	v.Rhand.analyze(s)

	switch v.Op {
	case BINOP_EQ, BINOP_NOT_EQ:
		if v.Lhand.GetType() != v.Rhand.GetType() {
			s.err(v, "Operands for binary operator `%s` must have the same type, have `%s` and `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName(), v.Rhand.GetType().TypeName())
		} else if lht := v.Lhand.GetType(); !(lht == PRIMITIVE_bool || lht == PRIMITIVE_rune || lht.IsIntegerType() || lht.IsFloatingType() || lht.LevelsOfIndirection() > 0) {
			s.err(v, "Operands for binary operator `%s` must be numeric, or pointers or booleans, have `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName())
		} else {
			v.Type = PRIMITIVE_bool
		}

	case BINOP_ADD, BINOP_SUB, BINOP_MUL, BINOP_DIV, BINOP_MOD,
		BINOP_GREATER, BINOP_LESS, BINOP_GREATER_EQ, BINOP_LESS_EQ,
		BINOP_BIT_AND, BINOP_BIT_OR, BINOP_BIT_XOR:
		if v.Lhand.GetType() != v.Rhand.GetType() {
			s.err(v, "Operands for binary operator `%s` must have the same type, have `%s` and `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName(), v.Rhand.GetType().TypeName())
		} else if lht := v.Lhand.GetType(); !(lht == PRIMITIVE_rune || lht.IsIntegerType() || lht.IsFloatingType() || lht.LevelsOfIndirection() > 0) {
			s.err(v, "Operands for binary operator `%s` must be numeric or pointers, have `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName())
		} else {
			switch v.Op.Category() {
			case OP_ARITHMETIC:
				v.Type = v.Lhand.GetType()
			case OP_COMPARISON:
				v.Type = PRIMITIVE_bool
			default:
				s.err(v, "invalid operands specified `%s`", v.Op.String())
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
		PRIMITIVE_s8, PRIMITIVE_s16, PRIMITIVE_s32, PRIMITIVE_s64, PRIMITIVE_i128,
		PRIMITIVE_u8, PRIMITIVE_u16, PRIMITIVE_u32, PRIMITIVE_u64, PRIMITIVE_u128:
		v.Type = t
	default:
		v.Type = PRIMITIVE_int // TODO check overflow
	}
}

// FloatingLiteral

func (v *FloatingLiteral) analyze(s *semanticAnalyzer) {}

func (v *FloatingLiteral) setTypeHint(t Type) {
	if v.Type != nil {
		// we've already figured out the type from a suffix
		return
	}

	switch t {
	case PRIMITIVE_f64, PRIMITIVE_f32, PRIMITIVE_f128:
		v.Type = t
	default:
		v.Type = PRIMITIVE_f64
	}
}

// StringLiteral

func (v *StringLiteral) analyze(s *semanticAnalyzer) {}
func (v *StringLiteral) setTypeHint(t Type)          {}

// RuneLiteral

func (v *RuneLiteral) analyze(s *semanticAnalyzer) {}
func (v *RuneLiteral) setTypeHint(t Type)          {}

// BoolLiteral

func (v *BoolLiteral) analyze(s *semanticAnalyzer) {}
func (v *BoolLiteral) setTypeHint(t Type)          {}

// ArrayLiteral

func (v *ArrayLiteral) analyze(s *semanticAnalyzer) {
	// TODO make type inferring stuff actually work well

	var memType Type // type of each member of the array

	if v.Type == nil {
		memType = nil
	} else {
		arrayType, ok := v.Type.(ArrayType)
		if !ok {
			s.err(v, "Invalid type")
		}
		memType = arrayType.MemberType
	}

	for _, mem := range v.Members {
		mem.setTypeHint(memType)
	}

	for _, mem := range v.Members {
		mem.analyze(s)
	}

	if v.Type == nil {
		// now we work out the type of the whole array be checking the member types
		for i := 1; i < len(v.Members); i++ {
			if v.Members[i-1].GetType() != v.Members[i].GetType() {
				s.err(v, "Array member type mismatch: `%s` and `%s`", v.Members[i-1].GetType().TypeName(), v.Members[i].GetType().TypeName())
			}
		}

		if v.Members[0].GetType() == nil {
			s.err(v, "Couldn't infer type of array members") // don't think this can ever happen
		}
		v.Type = arrayOf(v.Members[0].GetType())
	} else {
		for _, mem := range v.Members {
			if mem.GetType() != memType {
				s.err(v, "Cannot use element of type `%s` in array of type `%s`", mem.GetType().TypeName(), memType.TypeName())
			}
		}
	}
}

func (v *ArrayLiteral) setTypeHint(t Type) {
	v.Type = t
}

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
	isVariadic := v.Function.IsVariadic
	c := false // if we're calling a C function

	// find them attributes yo
	if v.Function.Attrs != nil {
		attributes := v.Function.Attrs

		// todo hashmap or some shit
		for _, attr := range attributes {
			switch attr.Key {
			case "c":
				c = true
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
		if i >= len(v.Function.Parameters) { // we have a variadic arg
			if !isVariadic {
				panic("woah")
			}
			arg.setTypeHint(nil)
			arg.analyze(s)

			if !c {
				panic("The `variadic` attribute should only be used on calls to C functions")
			}

			// varargs take type promotions. If we don't do these, the whole thing fucks up.
			switch arg.GetType() {
			case PRIMITIVE_f32:
				v.Arguments[i] = &CastExpr{
					Expr: arg,
					Type: PRIMITIVE_f64,
				}
			case PRIMITIVE_s8, PRIMITIVE_s16:
				v.Arguments[i] = &CastExpr{
					Expr: arg,
					Type: PRIMITIVE_int,
				}
			case PRIMITIVE_u8, PRIMITIVE_u16:
				v.Arguments[i] = &CastExpr{
					Expr: arg,
					Type: PRIMITIVE_uint,
				}
			}
		} else {
			arg.setTypeHint(v.Function.Parameters[i].Variable.Type)
			arg.analyze(s)

			if arg.GetType() != v.Function.Parameters[i].Variable.Type {
				s.err(arg, "Mismatched types in function call: `%s` and `%s`", arg.GetType(), v.Function.Parameters[i].Variable.Type)
			}
		}
	}

	if dep := getAttr(v.Function.Attrs, "deprecated"); dep != nil {
		s.warnDeprecated(v, "function", v.Function.Name, dep.Value)
	}
}

func (v *CallExpr) setTypeHint(t Type) {}

// AccessExpr

func (v *AccessExpr) analyze(s *semanticAnalyzer) {
	for _, access := range v.Accesses {
		if dep := getAttr(access.Variable.Attrs, "deprecated"); dep != nil {
			s.warnDeprecated(v, "variable", access.Variable.Name, dep.Value)
		}

		if access.AccessType == ACCESS_ARRAY {
			access.Subscript.setTypeHint(PRIMITIVE_int)
			access.Subscript.analyze(s)

			if !access.Subscript.GetType().IsIntegerType() {
				s.err(v, "Array subscript must be an integer type, have `%s`", access.Subscript.GetType().TypeName())
			}
		}
	}
}

func (v *AccessExpr) setTypeHint(t Type) {}

// AddressOfExpr

func (v *AddressOfExpr) analyze(s *semanticAnalyzer) {
	v.Access.analyze(s)
}

func (v *AddressOfExpr) setTypeHint(t Type) {}

// DerefExpr

func (v *DerefExpr) analyze(s *semanticAnalyzer) {
	v.Expr.analyze(s)
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

// SizeofExpr

func (v *SizeofExpr) analyze(s *semanticAnalyzer) {
	if v.Expr != nil {
		v.Expr.setTypeHint(nil)
		v.Expr.analyze(s)
	}
}

func (v *SizeofExpr) setTypeHint(t Type) {
	//v.Expr.setTypeHint(t)
}

// TupleLiteral

func (v *TupleLiteral) analyze(s *semanticAnalyzer) {
	// cba do it later
}

func (v *TupleLiteral) setTypeHint(t Type) {
	v.Type = t
}

// DefaultMatchBranch

func (v *DefaultMatchBranch) analyze(s *semanticAnalyzer) {

}

func (v *DefaultMatchBranch) setTypeHint(t Type) {

}
