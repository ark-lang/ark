package semantic

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

func (v *DeprecatedCheck) checkTypeReference(s *SemanticAnalyzer, loc parser.Locatable, typref *parser.TypeReference) {
	if dep := typref.Type.Attrs().Get("deprecated"); dep != nil {
		v.WarnDeprecated(s, loc, "type", typref.Type.TypeName(), dep.Value)
	}

	for _, garg := range typref.GenericArguments {
		v.checkTypeReference(s, loc, garg)
	}
}

func (v *DeprecatedCheck) Init(s *SemanticAnalyzer)       {}
func (v *DeprecatedCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *DeprecatedCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *DeprecatedCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {}

func (v *DeprecatedCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n := n.(type) {
	case *parser.VariableDecl:
		v.checkTypeReference(s, n, n.Variable.Type)

	case *parser.CallExpr:
		/*if dep := n.Function.Type.Attrs().Get("deprecated"); dep != nil {
			v.WarnDeprecated(s, n, "function", n.Function.Name, dep.Value)
		}*/ // TODO

	case *parser.VariableAccessExpr:
		if dep := n.Variable.Attrs.Get("deprecated"); dep != nil {
			v.WarnDeprecated(s, n, "variable", n.Variable.Name, dep.Value)
		}
	}
}

func (v *DeprecatedCheck) Finalize(s *SemanticAnalyzer) {

}
