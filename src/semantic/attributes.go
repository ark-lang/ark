package semantic

import (
	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/parser"
)

type AttributeCheck struct {
}

func (v *AttributeCheck) Init(s *SemanticAnalyzer)       {}
func (v *AttributeCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *AttributeCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *AttributeCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {}

func (_ AttributeCheck) Name() string { return "attribute" }

func (v *AttributeCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	switch n := n.(type) {
	case *ast.TypeDecl:
		typ := n.NamedType.Type
		switch typ.(type) {
		case ast.StructType:
			v.CheckStructType(s, typ.(ast.StructType))
		}

	case *ast.FunctionDecl:
		v.CheckFunctionDecl(s, n)
	//case *ast.TraitDecl:
	//	v.CheckTraitDecl(s, n)

	case *ast.VariableDecl:
		v.CheckVariableDecl(s, n)
	}
}

func (v *AttributeCheck) Finalize(s *SemanticAnalyzer) {

}

func (v *AttributeCheck) CheckFunctionDecl(s *SemanticAnalyzer, n *ast.FunctionDecl) {
	v.CheckAttrsDistanceFromLine(s, n.Function.Type.Attrs(), n.Pos().Line, "function", n.Function.Name)

	for _, attr := range n.Function.Type.Attrs() {
		switch attr.Key {
		case "deprecated":
		case "c":
		case "call_conv":
		case "nomangle":
		case "inline":
			switch attr.Value {
			case "always":
			case "never":
			case "maybe":
			default:
				s.Err(attr, "Invalid value `%s` for [inline] attribute", attr.Value)
			}
		default:
			s.Err(attr, "Invalid function attribute key `%s`", attr.Key)
		}
	}
}

func (v *AttributeCheck) CheckStructType(s *SemanticAnalyzer, n ast.StructType) {
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

/*func (v *AttributeCheck) CheckTraitDecl(s *SemanticAnalyzer, n *ast.TraitDecl) {
	v.CheckAttrsDistanceFromLine(s, n.Trait.Attrs(), n.Pos().Line, "type", n.Trait.TypeName())

	for _, attr := range n.Trait.Attrs() {
		if attr.Key != "deprecated" {
			s.Err(attr, "Invalid trait attribute key `%s`", attr.Key)
		}
	}
}*/

func (v *AttributeCheck) CheckVariableDecl(s *SemanticAnalyzer, n *ast.VariableDecl) {
	v.CheckAttrsDistanceFromLine(s, n.Variable.Attrs, n.Pos().Line, "variable", n.Variable.Name)

	for _, attr := range n.Variable.Attrs {
		switch attr.Key {
		case "deprecated":
			// value is optional, nothing to check
		case "nozero":
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
