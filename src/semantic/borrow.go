package semantic

import (
	"fmt"
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
)

var lifetimeIndex rune = 'a'

type BorrowCheck struct {
	lifetimes       map[string]*Lifetime
	currentLifetime *Lifetime
}

type Lifetime struct {
	// it's important that the hashmap can be iterated
	// in the correct order, so we add the hashmap keys
	// here and iterate over this instead.
	resourceKeys []string
	resources    map[string]Resource

	borrowKeys []string
	borrows    map[string]Borrow

	name string
}

type Resource struct {
}

type Borrow struct {
}

func (v *BorrowCheck) CheckAddrofExpr(s *SemanticAnalyzer, n *parser.AddressOfExpr) {

}

func (v *BorrowCheck) CheckVariableAccessExpr(s *SemanticAnalyzer, n *parser.VariableAccessExpr) {

}

func (v *BorrowCheck) CheckFunctionDecl(s *SemanticAnalyzer, n *parser.FunctionDecl) {

}

func (v *BorrowCheck) CheckAssignStat(s *SemanticAnalyzer, n *parser.AssignStat) {

}

func (v *BorrowCheck) CheckCallExpr(s *SemanticAnalyzer, n *parser.CallExpr) {

}

func (v *BorrowCheck) CheckCallStat(s *SemanticAnalyzer, n *parser.CallStat) {

}

func (v *BorrowCheck) CheckVariableDecl(s *SemanticAnalyzer, n *parser.VariableDecl) {

}

func (v *BorrowCheck) Finalize() {

}

func (v *BorrowCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {

}

func (v *BorrowCheck) EnterScope(s *SemanticAnalyzer) {

}

func (v *BorrowCheck) ExitScope(s *SemanticAnalyzer) {

}

func (v *BorrowCheck) createLifetime(s *SemanticAnalyzer) {
	v.currentLifetime = &Lifetime{
		resources: make(map[string]Resource),
		borrows:   make(map[string]Borrow),
		name:      util.Bold("" + string(lifetimeIndex) + ""),
	}
	v.lifetimes[v.currentLifetime.name] = v.currentLifetime
	lifetimeIndex++
	fmt.Println("created lifetime for new scope " + v.currentLifetime.name)
}

func (v *BorrowCheck) destroyLifetime(s *SemanticAnalyzer) {
	delete(v.lifetimes, v.currentLifetime.name)
	fmt.Println("cleaning up lifetime " + v.currentLifetime.name)
}

func (v *BorrowCheck) Init(s *SemanticAnalyzer) {
	v.lifetimes = make(map[string]*Lifetime)
}

func (v *BorrowCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	fmt.Printf("")

	switch n.(type) {
	case *parser.FunctionDecl:
		v.createLifetime(s)
		v.CheckFunctionDecl(s, n.(*parser.FunctionDecl))
		v.destroyLifetime(s)

	case *parser.VariableDecl:
		v.CheckVariableDecl(s, n.(*parser.VariableDecl))

	case *parser.CallStat:
		v.CheckCallStat(s, n.(*parser.CallStat))
	case *parser.AssignStat:
		v.CheckAssignStat(s, n.(*parser.AssignStat))

	case *parser.CallExpr:
		v.CheckCallExpr(s, n.(*parser.CallExpr))

	case *parser.AddressOfExpr:
		v.CheckAddrofExpr(s, n.(*parser.AddressOfExpr))

	case *parser.VariableAccessExpr:
		v.CheckVariableAccessExpr(s, n.(*parser.VariableAccessExpr))
	}
}
