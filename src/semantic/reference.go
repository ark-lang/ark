package semantic

import (
	"reflect"

	"github.com/ark-lang/ark/src/ast"
)

type ReferenceCheck struct {
	InFunction int
}

func (_ ReferenceCheck) Name() string { return "reference" }

func (v *ReferenceCheck) Init(s *SemanticAnalyzer)       {}
func (v *ReferenceCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *ReferenceCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *ReferenceCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {
	switch n.(type) {
	case *ast.FunctionDecl, *ast.LambdaExpr:
		v.InFunction--
	}
}

func (v *ReferenceCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	switch n.(type) {
	case *ast.FunctionDecl, *ast.LambdaExpr:
		v.InFunction++
	}

	switch n := n.(type) {
	case *ast.FunctionDecl:
		v.checkFunction(s, n, n.Function)

	case *ast.LambdaExpr:
		v.checkFunction(s, n, n.Function)

	case *ast.VariableDecl:
		if v.InFunction <= 0 {
			if typeReferenceContainsReferenceType(n.Variable.Type, nil) {
				s.Err(n, "Global variable has reference-containing type `%s`", n.Variable.Type.String())
			}
		}
	}
}

func (v *ReferenceCheck) checkFunction(s *SemanticAnalyzer, loc ast.Locatable, fn *ast.Function) {
	if typeReferenceContainsReferenceType(fn.Type.Return, nil) {
		s.Err(loc, "Function has reference-containing return type `%s`", fn.Type.Return.String())
	}
}

func (v *ReferenceCheck) Finalize(s *SemanticAnalyzer) {

}

func typeReferenceContainsReferenceType(typ *ast.TypeReference, visited map[*ast.TypeReference]struct{ visited, value bool }) bool {
	if visited == nil {
		visited = make(map[*ast.TypeReference]struct{ visited, value bool })
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

func typeContainsReferenceType(typ ast.Type, visited map[*ast.TypeReference]struct{ visited, value bool }) bool {
	switch typ := typ.ActualType().(type) {
	case ast.ReferenceType:
		return true

	case ast.PointerType:
		return typeReferenceContainsReferenceType(typ.Addressee, visited)

	case ast.ArrayType:
		return typeReferenceContainsReferenceType(typ.MemberType, visited)

	case ast.StructType:
		for _, field := range typ.Members {
			if typeReferenceContainsReferenceType(field.Type, visited) {
				return true
			}
		}
		return false

	case ast.EnumType:
		for _, member := range typ.Members {
			if typeContainsReferenceType(member.Type, visited) {
				return true
			}
		}
		return false

	case ast.TupleType:
		for _, field := range typ.Members {
			if typeReferenceContainsReferenceType(field, visited) {
				return true
			}
		}
		return false

	case *ast.SubstitutionType, ast.PrimitiveType, ast.FunctionType:
		return false

	default:
		panic("unimplemented type: " + reflect.TypeOf(typ).String())
	}
}
