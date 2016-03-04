package semantic

import (
	"github.com/ark-lang/ark/src/ast"
)

type ImmutableAssignCheck struct {
}

func (_ ImmutableAssignCheck) Name() string { return "immutable assign" }

func (v *ImmutableAssignCheck) Init(s *SemanticAnalyzer)       {}
func (v *ImmutableAssignCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *ImmutableAssignCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *ImmutableAssignCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {}

func (v *ImmutableAssignCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	switch n := n.(type) {
	case *ast.AssignStat:
		if !n.Access.Mutable() {
			s.Err(n, "Cannot assign value to immutable access")
		}

	case *ast.BinopAssignStat:
		if !n.Access.Mutable() {
			s.Err(n, "Cannot assign value to immutable access")
		}

	case *ast.DestructAssignStat:
		for _, acc := range n.Accesses {
			if !acc.Mutable() {
				s.Err(acc, "Cannot assign value to immutable access")
			}
		}

	case *ast.DestructBinopAssignStat:
		for _, acc := range n.Accesses {
			if !acc.Mutable() {
				s.Err(acc, "Cannot assign value to immutable access")
			}
		}
	}
}

func (v *ImmutableAssignCheck) Finalize(s *SemanticAnalyzer) {

}
