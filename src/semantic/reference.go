package semantic

import (
	"reflect"

	"github.com/ark-lang/ark/src/parser"
)

type ReferenceCheck struct {
	InFunction int
}

func (v *ReferenceCheck) Init(s *SemanticAnalyzer)       {}
func (v *ReferenceCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *ReferenceCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *ReferenceCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.FunctionDecl, *parser.LambdaExpr:
		v.InFunction--
	}
}

func (v *ReferenceCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.FunctionDecl, *parser.LambdaExpr:
		v.InFunction++
	}

	switch n := n.(type) {
	case *parser.FunctionDecl:
		v.checkFunction(s, n, n.Function)

	case *parser.LambdaExpr:
		v.checkFunction(s, n, n.Function)

	case *parser.VariableDecl:
		if v.InFunction <= 0 {
			if typeReferenceContainsReferenceType(n.Variable.Type, nil) {
				s.Err(n, "Global variable has reference-containing type `%s`", n.Variable.Type.String())
			}
		}
	}
}

func (v *ReferenceCheck) checkFunction(s *SemanticAnalyzer, loc parser.Locatable, fn *parser.Function) {
	if typeReferenceContainsReferenceType(fn.Type.Return, nil) {
		s.Err(loc, "Function has reference-containing return type `%s`", fn.Type.Return.String())
	}
}

func (v *ReferenceCheck) Finalize(s *SemanticAnalyzer) {

}

func typeReferenceContainsReferenceType(typ *parser.TypeReference, visited map[*parser.TypeReference]struct{ visited, value bool }) bool {
	if visited == nil {
		visited = make(map[*parser.TypeReference]struct{ visited, value bool })
	}

	if visited[typ].visited {
		return visited[typ].value
	}
	visited[typ] = struct{ visited, value bool }{true, false}

	if typeContainsReferenceType(typ.BaseType, visited) {
		visited[typ] = struct{ visited, value bool }{true, true}
		return true
	}

	for _, garg := range typ.GenericArguments {
		if typeReferenceContainsReferenceType(garg, visited) {
			visited[typ] = struct{ visited, value bool }{true, true}
			return true
		}
	}

	return false
}

func typeContainsReferenceType(typ parser.Type, visited map[*parser.TypeReference]struct{ visited, value bool }) bool {
	switch typ := typ.ActualType().(type) {
	case parser.ReferenceType:
		return true

	case parser.PointerType:
		return typeReferenceContainsReferenceType(typ.Addressee, visited)

	case parser.ArrayType:
		return typeReferenceContainsReferenceType(typ.MemberType, visited)

	case parser.StructType:
		for _, field := range typ.Members {
			if typeReferenceContainsReferenceType(field.Type, visited) {
				return true
			}
		}
		return false

	case parser.EnumType:
		for _, member := range typ.Members {
			if typeContainsReferenceType(member.Type, visited) {
				return true
			}
		}
		return false

	case parser.TupleType:
		for _, field := range typ.Members {
			if typeReferenceContainsReferenceType(field, visited) {
				return true
			}
		}
		return false

	case *parser.SubstitutionType, parser.PrimitiveType, parser.FunctionType:
		return false

	default:
		panic("unimplemented type: " + reflect.TypeOf(typ).String())
	}
}
