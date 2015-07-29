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
	switch n.(type) {
	case *parser.EnumDecl:
		decl := n.(*parser.EnumDecl)
		usedNames := make(map[string]bool)
		usedTags := make(map[int]bool)
		for _, mem := range decl.Enum.Members {
			if usedNames[mem.Name] {
				s.Err(decl, "Duplicate member name `%s`", mem.Name)
			}
			usedNames[mem.Name] = true

			if usedTags[mem.Tag] {
				s.Err(decl, "Duplciate enum tag `%d` on member `%s`", mem.Tag, mem.Name)
			}
			usedTags[mem.Tag] = true
		}
	}

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
