package checks

import (
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
)

type MiscCheck struct {
}

func (v *MiscCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *MiscCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *MiscCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	if s.Function == nil {
		switch n.(type) {
		case *parser.ReturnStat:
			s.Err(n, "%s must be in function", util.CapitalizeFirst(n.NodeName()))
		}
	} else {
		switch n.(type) {
		case *parser.TypeDecl:
			s.Err(n, "%s must be in function", util.CapitalizeFirst(n.NodeName()))
		}
	}
}
