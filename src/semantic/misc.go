package semantic

import (
	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/util"
)

type MiscCheck struct {
	InFunction int
}

func (_ MiscCheck) Name() string { return "misc" }

func (v *MiscCheck) Init(s *SemanticAnalyzer) {
	v.InFunction = 0
}

func (v *MiscCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *MiscCheck) ExitScope(s *SemanticAnalyzer)  {}

func (v *MiscCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	switch n.(type) {
	case *ast.FunctionDecl, *ast.LambdaExpr:
		v.InFunction++
	}

	if v.InFunction <= 0 {
		switch n.(type) {
		case *ast.ReturnStat:
			s.Err(n, "%s must be in function", util.CapitalizeFirst(n.NodeName()))
		}
	} else {
		switch n.(type) {
		case *ast.TypeDecl:
			s.Err(n, "%s must not be in function", util.CapitalizeFirst(n.NodeName()))

		case *ast.FunctionDecl:
			if v.InFunction > 1 {
				s.Err(n, "%s must not be in function", util.CapitalizeFirst(n.NodeName()))
			}
		}
	}
}

func (v *MiscCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {
	switch n.(type) {
	case *ast.FunctionDecl, *ast.LambdaExpr:
		v.InFunction--
	}
}

func (v *MiscCheck) Finalize(s *SemanticAnalyzer) {

}
