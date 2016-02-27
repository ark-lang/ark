package semantic

import (
	"github.com/ark-lang/ark/src/ast"
)

type UseBeforeDeclareCheck struct {
	scopes []map[string]bool
	scope  map[string]bool
}

func (_ UseBeforeDeclareCheck) Name() string { return "use before declare" }

func (v *UseBeforeDeclareCheck) Init(s *SemanticAnalyzer) {
	v.scopes = nil
}

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

func (v *UseBeforeDeclareCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {}

func (v *UseBeforeDeclareCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	switch n := n.(type) {
	case *ast.VariableDecl:
		v.scope[n.Variable.Name] = true

	case *ast.DestructVarDecl:
		for idx, vari := range n.Variables {
			if !n.ShouldDiscard[idx] {
				v.scope[vari.Name] = true
			}
		}

	case *ast.EnumPatternExpr:
		for _, vari := range n.Variables {
			if vari != nil {
				v.scope[vari.Name] = true
			}
		}

	case *ast.VariableAccessExpr:
		if !v.scope[n.Variable.Name] && n.Variable.ParentModule == s.Submodule.Parent {
			s.Err(n, "Use of variable before declaration: %s", n.Variable.Name)
		}
	}
}

func (v *UseBeforeDeclareCheck) Finalize(s *SemanticAnalyzer) {

}
