package semantic

import (
	"github.com/ark-lang/ark/src/ast"
)

type UnreachableCheck struct {
}

func (_ UnreachableCheck) Name() string { return "unreachable" }

func (v *UnreachableCheck) Init(s *SemanticAnalyzer)       {}
func (v *UnreachableCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *UnreachableCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *UnreachableCheck) Visit(s *SemanticAnalyzer, n ast.Node) {}

func (v *UnreachableCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {
	switch n := n.(type) {
	case *ast.Block:
		for i, c := range n.Nodes {
			if i < len(n.Nodes)-1 && IsNodeTerminating(c) {
				s.Err(n.Nodes[i+1], "Unreachable code")
			}
		}

		if len(n.Nodes) > 0 {
			n.IsTerminating = IsNodeTerminating(n.Nodes[len(n.Nodes)-1])
		}

	case *ast.FunctionDecl:
		v.visitFunction(s, n, n.Function)

	case *ast.LambdaExpr:
		v.visitFunction(s, n, n.Function)
	}

}

func (v *UnreachableCheck) visitFunction(s *SemanticAnalyzer, loc ast.Locatable, fn *ast.Function) {
	if fn.Body != nil && !fn.Body.IsTerminating {
		if fn.Type.Return != nil && !fn.Type.Return.BaseType.ActualType().IsVoidType() {
			s.Err(loc, "Missing return statement")
		} else {
			fn.Body.Nodes = append(fn.Body.Nodes, &ast.ReturnStat{})
			fn.Body.IsTerminating = true
		}
	}
}

func (v *UnreachableCheck) Finalize(s *SemanticAnalyzer) {

}

func IsNodeTerminating(n ast.Node) bool {
	switch n := n.(type) {
	case *ast.Block:
		return n.IsTerminating
	case *ast.ReturnStat:
		return true
	case *ast.IfStat:
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
