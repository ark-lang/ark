package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type UseBeforeDeclareCheck struct {
	scopes []map[string]bool
	scope  map[string]bool
}

func (v *UseBeforeDeclareCheck) Init(s *SemanticAnalyzer) {}
func (v *UseBeforeDeclareCheck) EnterScope(s *SemanticAnalyzer) {
	lastScope := v.scope
	if v.scope != nil {
		v.scopes = append(v.scopes, v.scope)
	}

	v.scope = make(map[string]bool)
	for name, declared := range lastScope {
		v.scope[name] = declared
	}
}
func (v *UseBeforeDeclareCheck) ExitScope(s *SemanticAnalyzer) {
	v.scope = nil
	if len(v.scopes) > 0 {
		idx := len(v.scopes) - 1
		v.scope, v.scopes[idx] = v.scopes[idx], nil
		v.scopes = v.scopes[:idx]
	}
}

func (v *UseBeforeDeclareCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {}

func (v *UseBeforeDeclareCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n := n.(type) {
	case *parser.VariableDecl:
		v.scope[n.Variable.Name] = true

	case *parser.DestructVarDecl:
		for idx, vari := range n.Variables {
			if !n.ShouldDiscard[idx] {
				v.scope[vari.Name] = true
			}
		}

	case *parser.VariableAccessExpr:
		if !v.scope[n.Variable.Name] && n.Variable.ParentModule == s.Submodule.Parent {
			s.Err(n, "Use of variable before declaration: %s", n.Variable.Name)
		}
	}
}

func (v *UseBeforeDeclareCheck) Finalize(s *SemanticAnalyzer) {

}
