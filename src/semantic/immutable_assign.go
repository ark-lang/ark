package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type ImmutableAssignCheck struct {
}

func (v *ImmutableAssignCheck) Init(s *SemanticAnalyzer)       {}
func (v *ImmutableAssignCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *ImmutableAssignCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *ImmutableAssignCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {}

func (v *ImmutableAssignCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n := n.(type) {
	case *parser.VariableDecl:
		_, isStructure := n.Variable.Type.BaseType.(parser.StructType)

		if n.Assignment == nil && !n.Variable.Mutable && !n.Variable.FromStruct && !isStructure && !n.Variable.IsParameter && !n.Variable.IsReceiver {
			// note the parent struct is nil!
			// as well as if the type is a structure!!
			// this is because we dont care if
			// a structure has an uninitialized value
			// likewise, we don't care if the variable is
			// something like `x: StructName`.
			s.Err(n, "Variable `%s` is immutable, yet has no initial value", n.Variable.Name)
		}

	case *parser.AssignStat:
		if !n.Access.Mutable() {
			s.Err(n, "Cannot assign value to immutable access")
		}

	case *parser.BinopAssignStat:
		if !n.Access.Mutable() {
			s.Err(n, "Cannot assign value to immutable access")
		}

	case *parser.DestructAssignStat:
		for _, acc := range n.Accesses {
			if !acc.Mutable() {
				s.Err(acc, "Cannot assign value to immutable access")
			}
		}

	case *parser.DestructBinopAssignStat:
		for _, acc := range n.Accesses {
			if !acc.Mutable() {
				s.Err(acc, "Cannot assign value to immutable access")
			}
		}
	}
}

func (v *ImmutableAssignCheck) Finalize(s *SemanticAnalyzer) {

}
