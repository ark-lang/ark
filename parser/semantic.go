package parser

import (
	"fmt"
	"os"

	"github.com/ark-lang/ark-go/util"
)

type semanticAnalyzer struct {
	file *File
}

func (v *semanticAnalyzer) err(err string, stuff ...interface{}) {
	/*fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Semantic error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
	v.peek(0).Filename, v.peek(0).LineNumber, v.peek(0).CharNumber, fmt.Sprintf(err, stuff...))*/
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Semantic error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(2)
}

func (v *semanticAnalyzer) analyze() {
	for _, node := range v.file.nodes {
		node.analyze(v)
	}
}

func (v *Block) analyze(s *semanticAnalyzer) {
	for _, n := range v.Nodes {
		n.analyze(s)
	}
}

func (v *VariableDecl) analyze(s *semanticAnalyzer) {
	v.Variable.analyze(s)
	v.Assignment.analyze(s)
}

func (v *Variable) analyze(s *semanticAnalyzer) {
	// make sure there are no illegal attributes
	s.checkDuplicateAttrs(v.Attrs)
	for _, attr := range v.Attrs {
		switch attr.Key {
		case "deprecated":
			// value is optional, nothing to check
		default:
			s.err("Invalid variable attribute key `%s`", attr.Key)
		}
	}
}

func (v *StructDecl) analyze(s *semanticAnalyzer) {
	v.Struct.analyze(s)
}

func (v *StructType) analyze(s *semanticAnalyzer) {
	// make sure there are no illegal attributes
	s.checkDuplicateAttrs(v.Attrs)
	for _, attr := range v.Attrs {
		switch attr.Key {
		case "packed":
			if attr.Value != "" {
				s.err("Struct attribute `%s` doesn't expect value", attr.Key)
			}
		case "deprecated":
			// value is optional, nothing to check
		default:
			s.err("Invalid struct attribute key `%s`", attr.Key)
		}
	}
}

func (v *FunctionDecl) analyze(s *semanticAnalyzer) {
	v.Function.analyze(s)
}

func (v *Function) analyze(s *semanticAnalyzer) {
	// make sure there are no illegal attributes
	s.checkDuplicateAttrs(v.Attrs)
	for _, attr := range v.Attrs {
		switch attr.Key {
		case "deprecated":
			// value is optional, nothing to check
		default:
			s.err("Invalid function attribute key `%s`", attr.Key)
		}
	}
}

func (v *semanticAnalyzer) checkDuplicateAttrs(attrs []*Attr) {
	encountered := make(map[string]bool)
	for _, attr := range attrs {
		if encountered[attr.Key] {
			v.err("Duplicate attribute `%s`", attr.Key)
		}
		encountered[attr.Key] = true
	}
}

func (v *ReturnStat) analyze(s *semanticAnalyzer) {
	v.Value.analyze(s)
	// TODO check return value same as function type
}

func (v *UnaryExpr) analyze(s *semanticAnalyzer) {
	v.Expr.analyze(s)

	switch v.Op {
	case UNOP_LOG_NOT:
		if v.Expr.GetType() == PRIMITIVE_bool {
			v.Type = PRIMITIVE_bool
		} else {
			s.err("Used logical not on non-bool")
		}
	case UNOP_BIT_NOT:
		if v.Expr.GetType().IsIntegerType() || v.Expr.GetType().IsFloatingType() {
			v.Type = v.Expr.GetType()
		} else {
			s.err("Used bitwise not on non-numeric type")
		}
	case UNOP_ADDRESS:
		v.Type = pointerTo(v.Expr.GetType())
		// TODO make sure v.Expr is a variable! (can't take address of a literal)
	case UNOP_DEREF:
		if ptr, ok := v.Expr.GetType().(PointerType); ok {
			v.Type = ptr.Addressee
		} else {
			s.err("Used dereference operator on non-pointer")
		}
	default:
		panic("whoops")
	}
}

func (v *BinaryExpr) analyze(s *semanticAnalyzer) {
	v.Lhand.analyze(s)
	v.Rhand.analyze(s)

	switch v.Op {
	case BINOP_ADD, BINOP_SUB, BINOP_MUL, BINOP_DIV, BINOP_MOD,
		BINOP_GREATER, BINOP_LESS, BINOP_GREATER_EQ, BINOP_LESS_EQ, BINOP_EQ, BINOP_NOT_EQ,
		BINOP_BIT_AND, BINOP_BIT_OR, BINOP_BIT_XOR:
		if v.Lhand.GetType() != v.Rhand.GetType() {
			s.err("Operands for binary operator `%s` must have the same type, have `%s` and `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName(), v.Rhand.GetType().TypeName())
		} else if lht := v.Lhand.GetType(); !(lht.IsIntegerType() || lht.IsFloatingType() || lht.LevelsOfIndirection() > 0) {
			s.err("Operands for binary operator `%s` must be numeric or pointers, have `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName())
		} else {
			switch v.Op.Category() {
			case OP_ARITHMETIC:
				v.Type = v.Lhand.GetType()
			case OP_COMPARISON:
				v.Type = PRIMITIVE_bool
			default:
				panic("shouldn't happenen ever")
			}
		}

	case BINOP_DOT: // TODO

	case BINOP_BIT_LEFT, BINOP_BIT_RIGHT:
		if lht := v.Lhand.GetType(); !(lht.IsFloatingType() || lht.IsIntegerType() || lht.LevelsOfIndirection() > 0) {
			s.err("Left-hand operand for bitshift operator `%s` must be numeric or a pointer, have `%s`",
				v.Op.OpString(), lht.TypeName())
		} else if !v.Rhand.GetType().IsIntegerType() {
			s.err("Right-hand operatnd for bitshift operator `%s` must be an integer, have `%s`",
				v.Op.OpString(), v.Rhand.GetType().TypeName())
		} else {
			v.Type = lht
		}

	case BINOP_LOG_AND, BINOP_LOG_OR:
		if v.Lhand.GetType() != PRIMITIVE_bool || v.Rhand.GetType() != PRIMITIVE_bool {
			s.err("Operands for logical operator `%s` must have the same type, have `%s` and `%s`",
				v.Op.OpString(), v.Lhand.GetType().TypeName(), v.Rhand.GetType().TypeName())
		} else {
			v.Type = PRIMITIVE_bool
		}

	case BINOP_ASSIGN:

	default:
		panic("unimplemented bin operation")
	}
}

func (v *StringLiteral) analyze(s *semanticAnalyzer) {}

func (v *IntegerLiteral) analyze(s *semanticAnalyzer) {}

func (v *FloatingLiteral) analyze(s *semanticAnalyzer) {}

func (v *RuneLiteral) analyze(s *semanticAnalyzer) {}
