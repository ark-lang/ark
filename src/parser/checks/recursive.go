package checks

import (
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util/log"
)

type RecursiveDefinitionCheck struct {
}

func (v *RecursiveDefinitionCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *RecursiveDefinitionCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *RecursiveDefinitionCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	var typ parser.Type
	switch n.(type) {
	case *parser.EnumDecl:
		decl := n.(*parser.EnumDecl)
		typ = decl.Enum

	case *parser.StructDecl:
		decl := n.(*parser.StructDecl)
		typ = decl.Struct

		// TODO: Check tuple types once we add named types for everything

	default:
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

func isTypeRecursive(typ parser.Type) (bool, []parser.Type) {
	var check func(current parser.Type, path *[]parser.Type, traversed map[parser.Type]bool) bool
	check = func(current parser.Type, path *[]parser.Type, traversed map[parser.Type]bool) bool {
		switch current.(type) {
		case *parser.StructType, *parser.TupleType, *parser.EnumType:
			if traversed[current] {
				return true
			}
			traversed[current] = true
		}

		switch current.(type) {
		case *parser.StructType:
			st := current.(*parser.StructType)
			for _, decl := range st.Variables {
				if check(decl.Variable.Type, path, traversed) {
					*path = append(*path, decl.Variable.Type)
					return true
				}
			}

		case *parser.TupleType:

			tt := current.(*parser.TupleType)
			for _, mem := range tt.Members {
				if check(mem, path, traversed) {
					*path = append(*path, mem)
					return true
				}
			}

		case *parser.EnumType:
			et := current.(*parser.EnumType)
			for _, mem := range et.Members {
				if check(mem.Type, path, traversed) {
					*path = append(*path, mem.Type)
					return true
				}
			}

			// TODO: Add array if we ever add embedded fixed size/static arrays
		}
		return false
	}

	var path []parser.Type
	return check(typ, &path, make(map[parser.Type]bool)), path
}
