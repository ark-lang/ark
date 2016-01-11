package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type UnreachableCheck struct {
}

func (v *UnreachableCheck) Init(s *SemanticAnalyzer)       {}
func (v *UnreachableCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *UnreachableCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *UnreachableCheck) Visit(s *SemanticAnalyzer, n parser.Node) {}

func (v *UnreachableCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {
	switch n := n.(type) {
	case *parser.Block:
		for i, c := range n.Nodes {
			if i < len(n.Nodes)-1 && IsNodeTerminating(c) {
				s.Err(n.Nodes[i+1], "Unreachable code")
			}
		}

		if len(n.Nodes) > 0 {
			n.IsTerminating = IsNodeTerminating(n.Nodes[len(n.Nodes)-1])
		}

	case *parser.FunctionDecl:
		v.visitFunction(s, n, n.Function)

	case *parser.LambdaExpr:
		v.visitFunction(s, n, n.Function)
	}

}

func (v *UnreachableCheck) visitFunction(s *SemanticAnalyzer, loc parser.Locatable, fn *parser.Function) {
	if fn.Body != nil && !fn.Body.IsTerminating {
		if fn.Type.Return != nil && !fn.Type.Return.IsVoidType() {
			s.Err(loc, "Missing return statement")
		} else {
			fn.Body.Nodes = append(fn.Body.Nodes, &parser.ReturnStat{})
		}
	}
}

func (v *UnreachableCheck) Destroy(s *SemanticAnalyzer) {

}

// TODO account for break/continue
func IsNodeTerminating(n parser.Node) bool {
	switch n := n.(type) {
	case *parser.Block:
		return n.IsTerminating
	case *parser.ReturnStat:
		return true
	case *parser.IfStat:
		if n.Else == nil || n.Else != nil && !n.Else.IsTerminating {
			return false
		}

		for _, body := range n.Bodies {
			if !body.IsTerminating {
				return false
			}
		}

		return true
	}

	return false
}
