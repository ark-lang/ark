package checks

import (
	"github.com/ark-lang/ark/src/parser"
)

type UnusedCheck struct {
	// TODO: Remove `Uses` variable from various ast nodes and keep count here
	encounteredScopes [][]parser.Node
	encountered       []parser.Node
}

func (v *UnusedCheck) EnterScope(s *SemanticAnalyzer) {
	if v.encountered != nil {
		v.encounteredScopes = append(v.encounteredScopes, v.encountered)
	}
	v.encountered = make([]parser.Node, 0)
}

func (v *UnusedCheck) ExitScope(s *SemanticAnalyzer) {
	v.AnalyzeUsage(s)

	v.encountered = nil
	if len(v.encounteredScopes) > 0 {
		idx := len(v.encounteredScopes) - 1
		v.encountered, v.encounteredScopes[idx] = v.encounteredScopes[idx], nil
		v.encounteredScopes = v.encounteredScopes[:idx]
	}
}

func (v *UnusedCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.VariableDecl:
		v.encountered = append(v.encountered, n)

	case *parser.FunctionDecl:
		v.encountered = append(v.encountered, n)
	}

	switch n.(type) {
	case *parser.CallExpr:
		expr := n.(*parser.CallExpr)
		expr.Function.Uses++

	case *parser.VariableAccessExpr:
		expr := n.(*parser.VariableAccessExpr)
		expr.Variable.Uses++

	case *parser.StructAccessExpr:
		expr := n.(*parser.StructAccessExpr)
		expr.Variable.Uses++
	}
}

func (v *UnusedCheck) AnalyzeUsage(s *SemanticAnalyzer) {
	for _, node := range v.encountered {
		switch node.(type) {
		case *parser.VariableDecl:
			decl := node.(*parser.VariableDecl)
			fromPrototype := decl.Variable.ParentFunction != nil && decl.Variable.ParentFunction.Prototype
			fromStruct := decl.Variable.ParentStruct != nil
			if !decl.Variable.Attrs.Contains("unused") && !fromPrototype && !fromStruct && decl.Variable.Uses == 0 {
				s.Err(decl, "Unused variable `%s`", decl.Variable.Name)
			}

		case *parser.FunctionDecl:
			decl := node.(*parser.FunctionDecl)
			if decl.Function.Name != "main" {
				continue
			}

			if !decl.Function.Attrs.Contains("unused") && decl.Function.Uses == 0 {
				// TODO add compiler option for this?
				//s.Err(decl, "Unused function `%s`", decl.Function.Name)
			}

		}
	}
}
