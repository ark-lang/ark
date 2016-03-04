package semantic

import (
	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/util/log"
)

type RecursiveDefinitionCheck struct {
}

func (_ RecursiveDefinitionCheck) Name() string { return "recursive definition" }

func (v *RecursiveDefinitionCheck) Init(s *SemanticAnalyzer)       {}
func (v *RecursiveDefinitionCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *RecursiveDefinitionCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *RecursiveDefinitionCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {}

func (v *RecursiveDefinitionCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	if typeDecl, ok := n.(*ast.TypeDecl); ok {
		typ := typeDecl.NamedType
		if ok, path := isTypeRecursive(typ); ok {
			s.Err(n, "Encountered recursive type definition")

			log.Errorln("semantic", "Path taken:")
			for _, typ := range path {
				log.Error("semantic", typ.TypeName())
				log.Error("semantic", " <- ")
			}
			log.Error("semantic", "%s\n\n", typ.TypeName())
		}
	}
}

func (v *RecursiveDefinitionCheck) Finalize(s *SemanticAnalyzer) {

}

func isTypeRecursive(typ ast.Type) (bool, []ast.Type) {
	typ = typ.ActualType()

	var check func(current ast.Type, path *[]ast.Type, traversed map[ast.Type]bool) bool
	check = func(current ast.Type, path *[]ast.Type, traversed map[ast.Type]bool) bool {
		switch current.(type) {
		case *ast.NamedType:
			if traversed[current] {
				return true
			}
		}

		switch typ := current.(type) {
		case ast.StructType:
			for _, mem := range typ.Members {
				if check(mem.Type.BaseType, path, traversed) {
					*path = append(*path, mem.Type.BaseType)
					return true
				}
			}

		case ast.TupleType:
			for _, mem := range typ.Members {
				if check(mem.BaseType, path, traversed) {
					*path = append(*path, mem.BaseType)
					return true
				}
			}

		case ast.EnumType:
			for _, mem := range typ.Members {
				if check(mem.Type, path, traversed) {
					*path = append(*path, mem.Type)
					return true
				}
			}

		case *ast.NamedType:
			traversed[current] = true
			if check(typ.Type, path, traversed) {
				*path = append(*path, typ.Type)
				return true
			}
			traversed[current] = false

		case ast.ArrayType:
			if typ.IsFixedLength && check(typ.MemberType.BaseType, path, traversed) {
				*path = append(*path, typ.MemberType.BaseType)
				return true
			}
		}
		return false
	}

	var path []ast.Type
	return check(typ, &path, make(map[ast.Type]bool)), path
}
