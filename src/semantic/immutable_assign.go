package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type ImmutableAssignCheck struct {
}

func (v *ImmutableAssignCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *ImmutableAssignCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *ImmutableAssignCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {}

func (v *ImmutableAssignCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.VariableDecl:
		decl := n.(*parser.VariableDecl)
		_, isStructure := decl.Variable.Type.(*parser.StructType)

		if decl.Assignment == nil && !decl.Variable.Mutable && decl.Variable.ParentStruct == nil && !isStructure && !decl.Variable.IsParameter {
			// note the parent struct is nil!
			// as well as if the type is a structure!!
			// this is because we dont care if
			// a structure has an uninitialized value
			// likewise, we don't care if the variable is
			// something like `x: StructName`.
			s.Err(decl, "Variable `%s` is immutable, yet has no initial value", decl.Variable.Name)
		}

	case *parser.AssignStat:
		stat := n.(*parser.AssignStat)
		if !stat.Access.Mutable() {
			s.Err(stat, "Cannot assign value to immutable access")
		}

	case *parser.BinopAssignStat:
		stat := n.(*parser.BinopAssignStat)
		if !stat.Access.Mutable() {
			s.Err(stat, "Cannot assign value to immutable access")
		}

	}
}
