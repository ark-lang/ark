package semantic

import (
	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/util"
)

// TODO handle match/switch, if we need to

type BreakAndNextCheck struct {
	nestedLoopCount map[*ast.Function]int
	functions       []*ast.Function
}

func (v *BreakAndNextCheck) Init(s *SemanticAnalyzer) {
	v.nestedLoopCount = make(map[*ast.Function]int)
}

func (v *BreakAndNextCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *BreakAndNextCheck) ExitScope(s *SemanticAnalyzer)  {}
func (v *BreakAndNextCheck) Finalize(s *SemanticAnalyzer)   {}

func (v *BreakAndNextCheck) Visit(s *SemanticAnalyzer, n ast.Node) {
	switch n := n.(type) {
	case *ast.NextStat, *ast.BreakStat:
		if v.nestedLoopCount[v.functions[len(v.functions)-1]] == 0 {
			s.Err(n, "%s must be in a loop", util.CapitalizeFirst(n.NodeName()))
		}

	case *ast.LoopStat:
		v.nestedLoopCount[v.functions[len(v.functions)-1]]++

	case *ast.FunctionDecl:
		v.functions = append(v.functions, n.Function)
	case *ast.LambdaExpr:
		v.functions = append(v.functions, n.Function)
	}
}

func (v *BreakAndNextCheck) PostVisit(s *SemanticAnalyzer, n ast.Node) {
	switch n := n.(type) {
	case *ast.Block:
		for i, c := range n.Nodes {
			if i < len(n.Nodes)-1 && isBreakOrNext(c) {
				s.Err(n.Nodes[i+1], "Unreachable code")
			}
		}

	case *ast.LoopStat:
		v.nestedLoopCount[v.functions[len(v.functions)-1]]--
	case *ast.FunctionDecl:
		v.functions = v.functions[:len(v.functions)-1]
		delete(v.nestedLoopCount, n.Function)
	case *ast.LambdaExpr:
		v.functions = v.functions[:len(v.functions)-1]
		delete(v.nestedLoopCount, n.Function)
	}
}

func isBreakOrNext(n ast.Node) bool {
	switch n.(type) {
	case *ast.BreakStat, *ast.NextStat:
		return true
	}
	return false
}
