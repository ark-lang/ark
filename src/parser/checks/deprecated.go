package checks

import (
	"fmt"
	"github.com/ark-lang/ark/src/parser"
)

type DeprecatedCheck struct {
}

func (v *DeprecatedCheck) WarnDeprecated(s *SemanticAnalyzer, thing parser.Locatable, typ, name, message string) {
	mess := fmt.Sprintf("Access of deprecated %s `%s`", typ, name)
	if message == "" {
		s.Warn(thing, mess)
	} else {
		s.Warn(thing, mess+": "+message)
	}
}

func (v *DeprecatedCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *DeprecatedCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *DeprecatedCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.VariableDecl:
		decl := n.(*parser.VariableDecl)
		if dep := decl.Variable.Type.Attrs().Get("deprecated"); dep != nil {
			v.WarnDeprecated(s, decl, "type", decl.Variable.Type.TypeName(), dep.Value)
		}

	case *parser.CallExpr:
		expr := n.(*parser.CallExpr)
		if dep := expr.Function.Attrs.Get("deprecated"); dep != nil {
			v.WarnDeprecated(s, expr, "function", expr.Function.Name, dep.Value)
		}

	case *parser.VariableAccessExpr:
		expr := n.(*parser.VariableAccessExpr)
		if dep := expr.Variable.Attrs.Get("deprecated"); dep != nil {
			v.WarnDeprecated(s, expr, "variable", expr.Variable.Name, dep.Value)
		}

	case *parser.StructAccessExpr:
		expr := n.(*parser.StructAccessExpr)
		if dep := expr.Variable.Attrs.Get("deprecated"); dep != nil {
			v.WarnDeprecated(s, expr, "variable", expr.Variable.Name, dep.Value)
		}
	}
}
