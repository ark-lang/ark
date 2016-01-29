package semantic

import (
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
)

// TODO handle match/switch, if we need to

type BreakAndNextCheck struct {
	nestedLoopCount map[*parser.Function]int
	functions       []*parser.Function
}

func (v *BreakAndNextCheck) Init(s *SemanticAnalyzer) {
	v.nestedLoopCount = make(map[*parser.Function]int)
}

func (v *BreakAndNextCheck) EnterScope(s *SemanticAnalyzer) {}
func (v *BreakAndNextCheck) ExitScope(s *SemanticAnalyzer)  {}
func (v *BreakAndNextCheck) Finalize(s *SemanticAnalyzer)   {}

func (v *BreakAndNextCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n := n.(type) {
	case *parser.NextStat, *parser.BreakStat:
		if v.nestedLoopCount[v.functions[len(v.functions)-1]] == 0 {
			s.Err(n, "%s must be in a loop", util.CapitalizeFirst(n.NodeName()))
		}

	case *parser.LoopStat:
		v.nestedLoopCount[v.functions[len(v.functions)-1]]++

	case *parser.FunctionDecl:
		v.functions = append(v.functions, n.Function)
	case *parser.LambdaExpr:
		v.functions = append(v.functions, n.Function)
	}
}

func (v *BreakAndNextCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {
	switch n := n.(type) {
	case *parser.Block:
		for i, c := range n.Nodes {
			if i < len(n.Nodes)-1 && isBreakOrNext(c) {
				s.Err(n.Nodes[i+1], "Unreachable code")
			}
		}

	case *parser.LoopStat:
		v.nestedLoopCount[v.functions[len(v.functions)-1]]--
	case *parser.FunctionDecl:
		v.functions = v.functions[:len(v.functions)-1]
		delete(v.nestedLoopCount, n.Function)
	case *parser.LambdaExpr:
		v.functions = v.functions[:len(v.functions)-1]
		delete(v.nestedLoopCount, n.Function)
	}
}

func isBreakOrNext(n parser.Node) bool {
	switch n.(type) {
	case *parser.BreakStat, *parser.NextStat:
		return true
	}
	return false
}
