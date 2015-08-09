package semantic

import (
	"fmt"
	"github.com/ark-lang/ark/src/parser"
)

type BorrowCheck struct {
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

func (v *BorrowCheck) Finalize() {}

func (v *BorrowCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {

}

func (v *BorrowCheck) EnterScope(s *SemanticAnalyzer) {
	v.currentLifetime = &Lifetime{
		resources: make(map[string]Resource),
		borrows:   make(map[string]Borrow),
	}
}

func (v *BorrowCheck) ExitScope(s *SemanticAnalyzer) {

}

func (v *BorrowCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	fmt.Printf("")

	switch n.(type) {
	case *parser.FunctionDecl:
		v.CheckFunctionDecl(s, n.(*parser.FunctionDecl))
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
