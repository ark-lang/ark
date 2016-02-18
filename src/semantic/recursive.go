package semantic

import (
	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/util/log"
)

type RecursiveDefinitionCheck struct {
}

func (v *RecursiveDefinitionCheck) Init(s *SemanticAnalyzer)       {}
func (v *RecursiveDefinitionCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *RecursiveDefinitionCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *RecursiveDefinitionCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {}

func (v *RecursiveDefinitionCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	var typ ast.Type

	if typeDecl, ok := n.(*ast.TypeDecl); ok {
		typ = typeDecl.NamedType
	} else {
		return
	}

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
			traversed[current] = true
		}

		switch current.(type) {
		case ast.StructType:
			st := current.(ast.StructType)
			for _, mem := range st.Members {
				if check(mem.Type.BaseType, path, traversed) {
					*path = append(*path, mem.Type.BaseType)
					return true
				}
			}

		case ast.TupleType:
			tt := current.(ast.TupleType)
			for _, mem := range tt.Members {
				if check(mem.BaseType, path, traversed) {
					*path = append(*path, mem.BaseType)
					return true
				}
			}

		case ast.EnumType:
			et := current.(ast.EnumType)
			for _, mem := range et.Members {
				if check(mem.Type, path, traversed) {
					*path = append(*path, mem.Type)
					return true
				}
			}

		case *ast.NamedType:
			nt := current.(*ast.NamedType)
			if check(nt.Type, path, traversed) {
				*path = append(*path, nt.Type)
				return true
			}

			// TODO: Add array if we ever add embedded fixed size/static arrays
		}
		return false
	}

	var path []ast.Type
	return check(typ, &path, make(map[ast.Type]bool)), path
}
