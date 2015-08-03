package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type UnusedCheck struct {
	// TODO: Remove `Uses` variable from various ast nodes and keep count here
	encounteredScopes [][]parser.Node
	encountered       []parser.Node
	uses              map[interface{}]int
}

func (v *UnusedCheck) EnterScope(s *SemanticAnalyzer) {
	if v.uses == nil {
		v.uses = make(map[interface{}]int)
	}
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

func (v *UnusedCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {}

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
		v.uses[expr.Function]++

	case *parser.VariableAccessExpr:
		expr := n.(*parser.VariableAccessExpr)
		v.uses[expr.Variable]++

	case *parser.StructAccessExpr:
		expr := n.(*parser.StructAccessExpr)
		v.uses[expr.Variable]++
	}
}

func (v *UnusedCheck) AnalyzeUsage(s *SemanticAnalyzer) {
	for _, node := range v.encountered {
		switch node.(type) {
		case *parser.VariableDecl:
			decl := node.(*parser.VariableDecl)
			fromStruct := decl.Variable.ParentStruct != nil
			if !decl.Variable.Attrs.Contains("unused") && !decl.Variable.IsParameter && !fromStruct && v.uses[decl.Variable] == 0 {
				s.Err(decl, "Unused variable `%s`", decl.Variable.Name)
			}

		case *parser.FunctionDecl:
			decl := node.(*parser.FunctionDecl)
			if decl.Function.Name != "main" {
				continue
			}

			if !decl.Function.Attrs.Contains("unused") && v.uses[decl.Function] == 0 {
				// TODO add compiler option for this?
				//s.Err(decl, "Unused function `%s`", decl.Function.Name)
			}

		}
	}
}
