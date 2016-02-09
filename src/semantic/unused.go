package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type UnusedCheck struct {
	// TODO: Remove `Uses` variable from various ast nodes and keep count here
	encountered []parser.Node
	uses        map[interface{}]int
}

func (v *UnusedCheck) Init(s *SemanticAnalyzer) {
	v.uses = make(map[interface{}]int)
}

func (v *UnusedCheck) EnterScope(s *SemanticAnalyzer)               {}
func (v *UnusedCheck) ExitScope(s *SemanticAnalyzer)                {}
func (v *UnusedCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {}

func (v *UnusedCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n := n.(type) {
	case *parser.VariableDecl:
		if !n.IsPublic() {
			v.encountered = append(v.encountered, n)
		}

	case *parser.FunctionDecl:
		if !n.IsPublic() {
			v.encountered = append(v.encountered, n)
		}
	}

	switch n.(type) {
	case *parser.CallExpr:
		expr := n.(*parser.CallExpr)
		v.uses[expr.Function]++

	case *parser.VariableAccessExpr:
		expr := n.(*parser.VariableAccessExpr)
		v.uses[expr.Variable]++
	}
}

func (v *UnusedCheck) Finalize(s *SemanticAnalyzer) {
	v.AnalyzeUsage(s)
}

func (v *UnusedCheck) AnalyzeUsage(s *SemanticAnalyzer) {
	for _, node := range v.encountered {
		switch node := node.(type) {
		case *parser.VariableDecl:
			if !node.Variable.Attrs.Contains("unused") && !node.Variable.IsParameter && !node.Variable.IsReceiver && !node.Variable.FromStruct && v.uses[node.Variable] == 0 {
				s.Warn(node, "Unused variable `%s`", node.Variable.Name)
			}

		case *parser.FunctionDecl:
			if !node.Function.Type.Attrs().Contains("unused") && v.uses[node.Function] == 0 {
				s.Warn(node, "Unused function `%s`", node.Function.Name)
			}
		}
	}
}
