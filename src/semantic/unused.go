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
	}
}

func (v *UnusedCheck) Finalize(s *SemanticAnalyzer) {
	v.AnalyzeUsage(s)
}

func (v *UnusedCheck) AnalyzeUsage(s *SemanticAnalyzer) {
	for _, node := range v.encountered {
		switch node.(type) {
		case *parser.VariableDecl:
			decl := node.(*parser.VariableDecl)
			if !decl.Variable.Attrs.Contains("unused") && !decl.Variable.IsParameter && !decl.Variable.IsReceiver && !decl.Variable.FromStruct && v.uses[decl.Variable] == 0 {
				s.Err(decl, "Unused variable `%s`", decl.Variable.Name)
			}

		case *parser.FunctionDecl:
			decl := node.(*parser.FunctionDecl)
			if decl.Function.Name != "main" {
				continue
			}

			if !decl.Function.Type.Attrs().Contains("unused") && v.uses[decl.Function] == 0 {
				// TODO add compiler option for this?
				//s.Err(decl, "Unused function `%s`", decl.Function.Name)
			}

		}
	}
}
