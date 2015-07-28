package checks

import (
	"github.com/ark-lang/ark/src/parser"
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

	case *parser.ReturnStat:
		stat := n.(*parser.ReturnStat)
		if s.Function == nil {
			s.Err(stat, "Return statement must be in a function")
		}
	}
}
