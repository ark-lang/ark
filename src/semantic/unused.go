package semantic

import (
	"github.com/ark-lang/ark/src/ast"
)

type UnusedCheck struct {
	encountered     []interface{}
	encounteredDecl []ast.Node
	uses            map[interface{}]int
}

func (_ UnusedCheck) Name() string { return "unused" }

func (v *UnusedCheck) Init(s *SemanticAnalyzer) {
	v.uses = make(map[interface{}]int)
	v.encountered = nil
	v.encounteredDecl = nil
}

func (v *UnusedCheck) EnterScope(s *SemanticAnalyzer)            {}
func (v *UnusedCheck) ExitScope(s *SemanticAnalyzer)             {}
func (v *UnusedCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {}

func (v *UnusedCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	switch n := n.(type) {
	case *ast.VariableDecl:
		if !n.IsPublic() {
			v.encountered = append(v.encountered, n.Variable)
			v.encounteredDecl = append(v.encounteredDecl, n)
		}

	case *ast.DestructVarDecl:
		if !n.IsPublic() {
			for idx, vari := range n.Variables {
				if !n.ShouldDiscard[idx] {
					v.encountered = append(v.encountered, vari)
					v.encounteredDecl = append(v.encounteredDecl, n)
				}
			}
		}

	case *ast.FunctionDecl:
		if !n.IsPublic() {
			v.encountered = append(v.encountered, n.Function)
			v.encounteredDecl = append(v.encounteredDecl, n)
		}
	}

	switch n := n.(type) {
	case *ast.FunctionAccessExpr:
		v.uses[n.Function]++

	case *ast.VariableAccessExpr:
		v.uses[n.Variable]++
	}
}

func (v *UnusedCheck) Finalize(s *SemanticAnalyzer) {
	v.AnalyzeUsage(s)
}

func (v *UnusedCheck) AnalyzeUsage(s *SemanticAnalyzer) {
	for idx, it := range v.encountered {
		decl := v.encounteredDecl[idx]
		switch it := it.(type) {
		case *ast.Variable:
			if !it.IsParameter && !it.IsReceiver && !it.FromStruct && v.uses[it] == 0 {
				s.Warn(decl, "Unused variable `%s`", it.Name)
			}

		case *ast.Function:
			if v.uses[it] == 0 {
				s.Warn(decl, "Unused function `%s`", it.Name)
			}
		}
	}
}
