package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type UnusedCheck struct {
	encountered     []interface{}
	encounteredDecl []parser.Node
	uses            map[interface{}]int
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
			v.encountered = append(v.encountered, n.Variable)
			v.encounteredDecl = append(v.encounteredDecl, n)
		}

	case *parser.DestructVarDecl:
		if !n.IsPublic() {
			for idx, vari := range n.Variables {
				if !n.ShouldDiscard[idx] {
					v.encountered = append(v.encountered, vari)
					v.encounteredDecl = append(v.encounteredDecl, n)
				}
			}
		}

	case *parser.FunctionDecl:
		if !n.IsPublic() {
			v.encountered = append(v.encountered, n.Function)
			v.encounteredDecl = append(v.encounteredDecl, n)
		}
	}

	switch n := n.(type) {
	case *parser.FunctionAccessExpr:
		v.uses[n.Function]++

	case *parser.VariableAccessExpr:
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
		case *parser.Variable:
			if !it.IsParameter && !it.IsReceiver && !it.FromStruct && v.uses[it] == 0 {
				s.Warn(decl, "Unused variable `%s`", it.Name)
			}

		case *parser.Function:
			if v.uses[it] == 0 {
				s.Warn(decl, "Unused function `%s`", it.Name)
			}
		}
	}
}
