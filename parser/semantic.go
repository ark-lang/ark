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

type SemanticAnalyzer struct {
	Module          *Module
	function        *Function // the function we're in, or nil if we aren't
	unresolvedNodes []Node
	modules         map[string]*Module
	shouldExit      bool
}

func (v *SemanticAnalyzer) err(thing Locatable, err string, stuff ...interface{}) {
	filename, line, char := thing.Pos()
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Semantic error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		filename, line, char, fmt.Sprintf(err, stuff...))
	v.shouldExit = true
}

func (v *SemanticAnalyzer) warn(thing Locatable, err string, stuff ...interface{}) {
	filename, line, char := thing.Pos()
	fmt.Printf(util.TEXT_YELLOW+util.TEXT_BOLD+"Semantic warning:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		filename, line, char, fmt.Sprintf(err, stuff...))
}

func (v *SemanticAnalyzer) warnDeprecated(thing Locatable, typ, name, message string) {
	mess := fmt.Sprintf("Access of deprecated %s `%s`", typ, name)
	if message == "" {
		v.warn(thing, mess)
	} else {
		v.warn(thing, mess+": "+message)
	}
}

func (v *SemanticAnalyzer) analyzeUsage(nodes []Node) {
	for _, node := range nodes {
		if variable, ok := node.(*VariableDecl); ok {
			if !variable.Variable.Attrs.Contains("unused") {
				if variable.Variable.Uses == 0 {
					v.err(variable, "unused variable `%s`", variable.Variable.Name)
				}
			}
		} else if function, ok := node.(*FunctionDecl); ok {
			if !function.Function.Attrs.Contains("unused") {
				if function.Function.Name != "main" && function.Function.Uses == 0 {
					//v.err(function, "unused function `%s`", function.Function.Name)
					// TODO add compiler option for this?
				}
			}
			if function.Function.Body != nil {
				v.analyzeUsage(function.Function.Body.Nodes)
			}
		} else if impl, ok := node.(*ImplDecl); ok {
			for _, function := range impl.Functions {
				if !function.Function.Attrs.Contains("unused") {
					if function.Function.Name != "main" && function.Function.Uses == 0 {
						v.err(function, "unused function `%s`", function.Function.Name)
					}
				}
				if function.Function.Body != nil {
					v.analyzeUsage(function.Function.Body.Nodes)
				}
			}
		}
	}
}

func (v *SemanticAnalyzer) Analyze(modules map[string]*Module) {
	v.modules = modules
	v.shouldExit = false

	for _, node := range v.Module.Nodes {
		node.analyze(v)
	}

	// once we're done analyzing everything
	// check for unused stuff
	v.analyzeUsage(v.Module.Nodes)

	if v.shouldExit {
		os.Exit(util.EXIT_FAILURE_SEMANTIC)
	}
}

func (v *Block) analyze(s *SemanticAnalyzer) {
	for i, n := range v.Nodes {
		n.analyze(s)

		if i < len(v.Nodes)-1 && IsNodeTerminating(n) {
			s.err(v.Nodes[i+1], "Unreachable code")
		}
	}

	if len(v.Nodes) > 0 {
		v.IsTerminating = IsNodeTerminating(v.Nodes[len(v.Nodes)-1])
	}
}

func IsNodeTerminating(n Node) bool {
	if block, ok := n.(*Block); ok {
		return block.IsTerminating
	} else if _, ok := n.(*ReturnStat); ok {
		return true
	} else if ifStat, ok := n.(*IfStat); ok {
		if ifStat.Else == nil || ifStat.Else != nil && !ifStat.Else.IsTerminating {
			return false
		}

		for _, body := range ifStat.Bodies {
			if !body.IsTerminating {
				return false
			}
		}

		if ifStat.Else != nil && !ifStat.Else.IsTerminating {
			return false
		}

		return true
	}

	return false
}

func (v *Function) analyze(s *SemanticAnalyzer) {
	// make sure there are no illegal attributes
	for _, attr := range v.Attrs {
		switch attr.Key {
		case "deprecated":
		case "unused":
		case "c":
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

func (v *ModuleDecl) analyze(s *SemanticAnalyzer) {

}

func (v *EnumDecl) analyze(s *SemanticAnalyzer) {
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

func (v *StructType) analyze(s *SemanticAnalyzer) {
	// make sure there are no illegal attributes
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

func (v *TraitType) analyze(s *SemanticAnalyzer) {
	// make sure there are no illegal attributes
	for _, attr := range v.Attrs() {
		if attr.Key != "deprecated" {
			s.err(attr, "Invalid trait attribute key `%s`", attr.Key)
		}
	}

	for _, decl := range v.Functions {
		decl.analyze(s)
	}
}

func (v *Variable) analyze(s *SemanticAnalyzer) {
	// make sure there are no illegal attributes
	for _, attr := range v.Attrs {
		switch attr.Key {
		case "deprecated":
			// value is optional, nothing to check
		case "unused":
		default:
			s.err(attr, "Invalid variable attribute key `%s`", attr.Key)
		}
	}
}

/**
 * Declarations
 */

func (v *VariableDecl) analyze(s *SemanticAnalyzer) {
	v.Variable.analyze(s)

	_, isStructure := v.Variable.Type.(*StructType)

	if v.Assignment != nil {
		v.Assignment.setTypeHint(v.Variable.Type)
		v.Assignment.analyze(s)

		if v.Variable.Type == nil { // type is inferred
			v.Variable.Type = v.Assignment.GetType()
		} else if !v.Variable.Type.Equals(v.Assignment.GetType()) {
			s.err(v, "Cannot assign expression of type `%s` to variable of type `%s`",
				v.Assignment.GetType().TypeName(), v.Variable.Type.TypeName())
		}
	} else if v.Assignment == nil && !v.Variable.Mutable && v.Variable.ParentStruct == nil && !isStructure {
		// note the parent struct is nil!
		// as well as if the type is a structure!!
		// this is because we dont care if
		// a structure has an uninitialized value
		// likewise, we don't care if the variable is
		// something like `x: StructName`.
		s.err(v, "Variable `%s` is immutable, yet has no initial value", v.Variable.Name)
	}

	if dep := v.Variable.Type.Attrs().Get("deprecated"); dep != nil {
		s.warnDeprecated(v, "type", v.Variable.Type.TypeName(), dep.Value)
	}

	s.checkAttrsDistanceFromLine(v.Variable.Attrs, v.lineNumber, "variable", v.Variable.Name)
}

func (v *StructDecl) analyze(s *SemanticAnalyzer) {
	v.Struct.analyze(s)
	s.checkAttrsDistanceFromLine(v.Struct.Attrs(), v.lineNumber, "type", v.Struct.TypeName())
}

func (v *TraitDecl) analyze(s *SemanticAnalyzer) {
	v.Trait.analyze(s)
	s.checkAttrsDistanceFromLine(v.Trait.Attrs(), v.lineNumber, "type", v.Trait.TypeName())
}

func (v *ImplDecl) analyze(s *SemanticAnalyzer) {
	// TODO
}

func (v *UseDecl) analyze(s *SemanticAnalyzer) {

}

func (v *FunctionDecl) analyze(s *SemanticAnalyzer) {
	v.Function.analyze(s)

	if !v.Prototype && !v.Function.Body.IsTerminating {
		if v.Function.ReturnType != nil && v.Function.ReturnType != PRIMITIVE_void {
			s.err(v, "Missing return statement")
		} else {
			v.Function.Body.Nodes = append(v.Function.Body.Nodes, &ReturnStat{})
		}
	}

	s.checkAttrsDistanceFromLine(v.Function.Attrs, v.lineNumber, "function", v.Function.Name)
}

/*
 * Statements
 */

func (v *ReturnStat) analyze(s *SemanticAnalyzer) {
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
			if !v.Value.GetType().Equals(s.function.ReturnType) {
				s.err(v.Value, "Cannot return expression of type `%s` from function `%s` of type `%s`",
					v.Value.GetType().TypeName(), s.function.Name, s.function.ReturnType.TypeName())
			}
		}
	}
}

func (v *IfStat) analyze(s *SemanticAnalyzer) {
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

func (v *BlockStat) analyze(s *SemanticAnalyzer) {
	v.Block.analyze(s)
}

// CallStat

func (v *CallStat) analyze(s *SemanticAnalyzer) {
	v.Call.analyze(s)
}

// AssignStat

func (v *AssignStat) analyze(s *SemanticAnalyzer) {
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

func (v *LoopStat) analyze(s *SemanticAnalyzer) {
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

func (v *MatchStat) analyze(s *SemanticAnalyzer) {
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

func (v *UnaryExpr) analyze(s *SemanticAnalyzer) {
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

func (v *BinaryExpr) analyze(s *SemanticAnalyzer) {
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

func (v *IntegerLiteral) analyze(s *SemanticAnalyzer) {}

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

func (v *FloatingLiteral) analyze(s *SemanticAnalyzer) {}

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

func (v *StringLiteral) analyze(s *SemanticAnalyzer) {}
func (v *StringLiteral) setTypeHint(t Type)          {}

// RuneLiteral

func (v *RuneLiteral) analyze(s *SemanticAnalyzer) {}
func (v *RuneLiteral) setTypeHint(t Type)          {}

// BoolLiteral

func (v *BoolLiteral) analyze(s *SemanticAnalyzer) {}
func (v *BoolLiteral) setTypeHint(t Type)          {}

// ArrayLiteral

func (v *ArrayLiteral) analyze(s *SemanticAnalyzer) {
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
			if !mem.GetType().Equals(memType) {
				s.err(v, "Cannot use element of type `%s` in array of type `%s`", mem.GetType().TypeName(), memType.TypeName())
			}
		}
	}
}

func (v *ArrayLiteral) setTypeHint(t Type) {
	v.Type = t
}

// CastExpr

func (v *CastExpr) analyze(s *SemanticAnalyzer) {
	v.Expr.setTypeHint(nil)
	v.Expr.analyze(s)
	if v.Type.Equals(v.Expr.GetType()) {
		s.warn(v, "Casting expression of type `%s` to the same type",
			v.Type.TypeName())
	} else if !v.Expr.GetType().CanCastTo(v.Type) {
		s.err(v, "Cannot cast expression of type `%s` to type `%s`",
			v.Expr.GetType().TypeName(), v.Type.TypeName())
	}
}

func (v *CastExpr) setTypeHint(t Type) {}

// CallExpr

func (v *CallExpr) analyze(s *SemanticAnalyzer) {
	argLen := len(v.Arguments)
	paramLen := len(v.Function.Parameters)

	// attributes defaults
	isVariadic := v.Function.IsVariadic
	c := false // if we're calling a C function

	// find them attributes yo
	if v.Function.Attrs != nil {
		c = v.Function.Attrs.Contains("c")
	}

	if argLen < paramLen {
		s.err(v, "Call to `%s` has too few arguments, expects %d, have %d",
			v.Function.Name, paramLen, argLen)
	} else if !isVariadic && argLen > paramLen {
		// we only care if it's not variadic
		s.err(v, "Call to `%s` has too many arguments, expects %d, have %d",
			v.Function.Name, paramLen, argLen)
	}

	v.Function.Uses++

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

	if dep := v.Function.Attrs.Get("deprecated"); dep != nil {
		s.warnDeprecated(v, "function", v.Function.Name, dep.Value)
	}
}

func (v *CallExpr) setTypeHint(t Type) {}

// AccessExpr

func (v *AccessExpr) analyze(s *SemanticAnalyzer) {
	for _, access := range v.Accesses {
		if dep := access.Variable.Attrs.Get("deprecated"); dep != nil {
			s.warnDeprecated(v, "variable", access.Variable.Name, dep.Value)
		}

		if access.Variable != nil {
			access.Variable.Uses++
		}

		if access.AccessType == ACCESS_ARRAY {
			access.Subscript.setTypeHint(PRIMITIVE_int)
			access.Subscript.analyze(s)

			if !access.Subscript.GetType().IsIntegerType() {
				s.err(v, "Array subscript must be an integer type, have `%s`", access.Subscript.GetType().TypeName())
			}
		}

		if access.AccessType == ACCESS_TUPLE {
			tupleType, ok := access.Variable.Type.(*TupleType)
			if !ok {
				s.err(v, "Cannot index type `%s` as a tuple", access.Variable.Type.TypeName())
			}

			if access.Index >= uint64(len(tupleType.Members)) {
				s.err(v, "Index `%d` (element %d) is greater than size of tuple `%s`", access.Index, access.Index+1, tupleType.TypeName())
			}
		}
	}
}

func (v *AccessExpr) setTypeHint(t Type) {}

// AddressOfExpr

func (v *AddressOfExpr) analyze(s *SemanticAnalyzer) {
	v.Access.analyze(s)
}

func (v *AddressOfExpr) setTypeHint(t Type) {}

// DerefExpr

func (v *DerefExpr) analyze(s *SemanticAnalyzer) {
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

// SizeofExpr

func (v *SizeofExpr) analyze(s *SemanticAnalyzer) {
	if v.Expr != nil {
		v.Expr.setTypeHint(nil)
		v.Expr.analyze(s)
	}
}

func (v *SizeofExpr) setTypeHint(t Type) {
	//v.Expr.setTypeHint(t)
}

// TupleLiteral

func (v *TupleLiteral) analyze(s *SemanticAnalyzer) {
	var memberTypes []Type

	tupleType, ok := v.Type.(*TupleType)
	if ok {
		memberTypes = tupleType.Members
	}

	if len(v.Members) == len(memberTypes) {
		for idx, mem := range memberTypes {
			v.Members[idx].setTypeHint(mem)
		}
	} else {
		for _, mem := range v.Members {
			mem.setTypeHint(nil)
		}
	}

	for _, mem := range v.Members {
		mem.analyze(s)
	}

	if v.Type == nil {
		var members []Type
		for _, mem := range v.Members {
			if mem.GetType() == nil {
				s.err(mem, "Couldn't infer type of tuple component")
			}
			members = append(members, mem.GetType())
		}

		v.Type = tupleOf(members...)
	} else {
		if len(v.Members) != len(memberTypes) {
			s.err(v, "Invalid amount of entries in tuple")
		}

		for idx, mem := range v.Members {
			if !mem.GetType().Equals(memberTypes[idx]) {
				s.err(v, "Cannot use component of type `%s` in tuple position of type `%s`", mem.GetType().TypeName(), memberTypes[idx])
			}
		}
	}
}

func (v *TupleLiteral) setTypeHint(t Type) {
	typ, ok := t.(*TupleType)
	if ok {
		v.Type = typ
	}
}

// DefaultMatchBranch

func (v *DefaultMatchBranch) analyze(s *SemanticAnalyzer) {

}

func (v *DefaultMatchBranch) setTypeHint(t Type) {

}
