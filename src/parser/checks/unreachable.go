package checks

import (
	"github.com/ark-lang/ark/src/parser"
)

type UnreachableCheck struct {
}

func (v *UnreachableCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *UnreachableCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *UnreachableCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.Block:
		block := n.(*parser.Block)

		for i, c := range block.Nodes {
			if i < len(block.Nodes)-1 && IsNodeTerminating(c) {
				s.Err(block.Nodes[i+1], "Unreachable code")
			}
		}

		if len(block.Nodes) > 0 {
			block.IsTerminating = IsNodeTerminating(block.Nodes[len(block.Nodes)-1])
		}

	case *parser.FunctionDecl:
		decl := n.(*parser.FunctionDecl)
		if !decl.Prototype && !decl.Function.Body.IsTerminating {
			if decl.Function.ReturnType != nil && decl.Function.ReturnType != parser.PRIMITIVE_void {
				s.Err(decl, "Missing return statement")
			} else {
				decl.Function.Body.Nodes = append(decl.Function.Body.Nodes, &parser.ReturnStat{})
			}
		}

	default:
		return
	}

}

func IsNodeTerminating(n parser.Node) bool {
	if block, ok := n.(*parser.Block); ok {
		return block.IsTerminating
	} else if _, ok := n.(*parser.ReturnStat); ok {
		return true
	} else if ifStat, ok := n.(*parser.IfStat); ok {
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
