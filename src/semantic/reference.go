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
			if typeReferenceContainsReferenceType(n.Variable.Type) {
				s.Err(n, "Global variable has reference-containing type `%s`", n.Variable.Type.String())
			}
		}
	}
}

func (v *ReferenceCheck) checkFunction(s *SemanticAnalyzer, loc parser.Locatable, fn *parser.Function) {
	if typeReferenceContainsReferenceType(fn.Type.Return) {
		s.Err(loc, "Function has reference-containing return type `%s`", fn.Type.Return.String())
	}
}

func (v *ReferenceCheck) Finalize(s *SemanticAnalyzer) {

}

func typeReferenceContainsReferenceType(typ *parser.TypeReference) bool {
	if typeContainsReferenceType(typ.BaseType) {
		return true
	}

	for _, garg := range typ.GenericArguments {
		if typeReferenceContainsReferenceType(garg) {
			return true
		}
	}

	return false
}

func typeContainsReferenceType(typ parser.Type) bool {
	switch typ := typ.ActualType().(type) {
	case parser.ReferenceType:
		return true

	case parser.PointerType:
		return typeReferenceContainsReferenceType(typ.Addressee)

	case parser.ArrayType:
		return typeReferenceContainsReferenceType(typ.MemberType)

	case parser.StructType:
		for _, field := range typ.Members {
			if typeReferenceContainsReferenceType(field.Type) {
				return true
			}
		}
		return false

	case parser.EnumType:
		for _, member := range typ.Members {
			if typeContainsReferenceType(member.Type) {
				return true
			}
		}
		return false

	case parser.TupleType:
		for _, field := range typ.Members {
			if typeReferenceContainsReferenceType(field) {
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
