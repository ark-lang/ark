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
	for _, attr := range v.Attrs {
		switch attr.Key {
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
	for _, attr := range v.Attrs {
		switch attr.Key {
		case "packed":
			if attr.Value != "" {
				s.err("Struct attribute `%s` doesn't expect value", attr.Key)
			}
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
	for _, attr := range v.Attrs {
		switch attr.Key {
		default:
			s.err("Invalid function attribute key `%s`", attr.Key)
		}
	}
}

func (v *ReturnStat) analyze(s *semanticAnalyzer) {}

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
	case UNOP_DEREF:
	default:
		panic("whoops")
	}
}

func (v *BinaryExpr) analyze(s *semanticAnalyzer) {}

func (v *StringLiteral) analyze(s *semanticAnalyzer) {}

func (v *IntegerLiteral) analyze(s *semanticAnalyzer) {}

func (v *FloatingLiteral) analyze(s *semanticAnalyzer) {}

func (v *RuneLiteral) analyze(s *semanticAnalyzer) {}
