package semantic

import (
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
)

type MiscCheck struct {
	InFunction int
}

func (v *MiscCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *MiscCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *MiscCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	if _, ok := n.(*parser.FunctionDecl); ok {
		v.InFunction++
	}

	if v.InFunction <= 0 {
		switch n.(type) {
		case *parser.ReturnStat:
			s.Err(n, "%s must be in function", util.CapitalizeFirst(n.NodeName()))
		}
	} else {
		switch n.(type) {
		case *parser.TypeDecl:
			s.Err(n, "%s must not be in function", util.CapitalizeFirst(n.NodeName()))
		}
	}
}

func (v *MiscCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {
	if _, ok := n.(*parser.FunctionDecl); ok {
		v.InFunction--
	}
}
