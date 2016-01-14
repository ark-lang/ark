package semantic

import "github.com/ark-lang/ark/src/parser"

// TODO handle match/switch, if we need to

type BreakAndNextCheck struct {
}

func (v *BreakAndNextCheck) Init(s *SemanticAnalyzer)                     {}
func (v *BreakAndNextCheck) EnterScope(s *SemanticAnalyzer)               {}
func (v *BreakAndNextCheck) ExitScope(s *SemanticAnalyzer)                {}
func (v *BreakAndNextCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {}
func (v *BreakAndNextCheck) Destroy(s *SemanticAnalyzer)                  {}

func (v *BreakAndNextCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n := n.(type) {
	case *parser.Block:
		for i, c := range n.Nodes {
			if i < len(n.Nodes)-1 && isBreakOrNext(c) {
				s.Err(n.Nodes[i+1], "Unreachable code")
			}
		}
	}
}

func isBreakOrNext(n parser.Node) bool {
	switch n.(type) {
	case *parser.BreakStat, *parser.NextStat:
		return true
	}
	return false
}
