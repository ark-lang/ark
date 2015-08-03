package semantic

import (
	"github.com/ark-lang/ark/src/parser"
)

type AttributeCheck struct {
}

func (v *AttributeCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *AttributeCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *AttributeCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {}

func (v *AttributeCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.TypeDecl:
		typ := n.(*parser.TypeDecl).NamedType.Type
		switch typ.(type) {
		case *parser.StructType:
			v.CheckStructType(s, typ.(*parser.StructType))
		}

	case *parser.FunctionDecl:
		v.CheckFunctionDecl(s, n.(*parser.FunctionDecl))
	//case *parser.TraitDecl:
	//	v.CheckTraitDecl(s, n.(*parser.TraitDecl))

	case *parser.VariableDecl:
		v.CheckVariableDecl(s, n.(*parser.VariableDecl))
	}
}

func (v *AttributeCheck) CheckFunctionDecl(s *SemanticAnalyzer, n *parser.FunctionDecl) {
	v.CheckAttrsDistanceFromLine(s, n.Function.Attrs, n.Pos().Line, "function", n.Function.Name)

	for _, attr := range n.Function.Attrs {
		switch attr.Key {
		case "deprecated":
		case "unused":
		case "c":
		default:
			s.Err(attr, "Invalid function attribute key `%s`", attr.Key)
		}
	}
}

func (v *AttributeCheck) CheckStructType(s *SemanticAnalyzer, n *parser.StructType) {
	for _, attr := range n.Attrs() {
		switch attr.Key {
		case "packed":
			if attr.Value != "" {
				s.Err(attr, "Struct attribute `%s` doesn't expect value", attr.Key)
			}
		case "deprecated":
			// value is optional, nothing to check
		default:
			s.Err(attr, "Invalid struct attribute key `%s`", attr.Key)
		}
	}
}

/*func (v *AttributeCheck) CheckTraitDecl(s *SemanticAnalyzer, n *parser.TraitDecl) {
	v.CheckAttrsDistanceFromLine(s, n.Trait.Attrs(), n.Pos().Line, "type", n.Trait.TypeName())

	for _, attr := range n.Trait.Attrs() {
		if attr.Key != "deprecated" {
			s.Err(attr, "Invalid trait attribute key `%s`", attr.Key)
		}
	}
}*/

func (v *AttributeCheck) CheckVariableDecl(s *SemanticAnalyzer, n *parser.VariableDecl) {
	v.CheckAttrsDistanceFromLine(s, n.Variable.Attrs, n.Pos().Line, "variable", n.Variable.Name)

	for _, attr := range n.Variable.Attrs {
		switch attr.Key {
		case "deprecated":
			// value is optional, nothing to check
		case "unused":
		default:
			s.Err(attr, "Invalid variable attribute key `%s`", attr.Key)
		}
	}
}

func (v *AttributeCheck) CheckAttrsDistanceFromLine(s *SemanticAnalyzer, attrs parser.AttrGroup, line int, declType, declName string) {
	// Turn map into a list sorted by line number
	var sorted []*parser.Attr
	for _, attr := range attrs {
		index := 0
		for idx, innerAttr := range sorted {
			if attr.Pos().Line >= innerAttr.Pos().Line {
				index = idx
			}
		}

		sorted = append(sorted, nil)
		copy(sorted[index+1:], sorted[index:])
		sorted[index] = attr
	}

	for i := len(sorted) - 1; i >= 0; i-- {
		if sorted[i].Pos().Line < line-1 {
			// mute warnings from attribute blocks
			if !sorted[i].FromBlock {
				s.Warn(sorted[i], "Gap of %d lines between declaration of %s `%s` and `%s` attribute", line-sorted[i].Pos().Line, declType, declName, sorted[i].Key)
			}
		}
		line = sorted[i].Pos().Line
	}
}
