package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type UseBeforeDeclareCheck struct {
	scopes []map[string]bool
	scope  map[string]bool
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

func (v *UseBeforeDeclareCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.VariableDecl:
		decl := n.(*parser.VariableDecl)
		v.scope[decl.Variable.Name] = true

	case *parser.VariableAccessExpr:
		expr := n.(*parser.VariableAccessExpr)
		if !v.scope[expr.Variable.Name] {
			s.Err(expr, "Use of variable before declaration: %s", expr.Variable.Name)
		}
	}
}
