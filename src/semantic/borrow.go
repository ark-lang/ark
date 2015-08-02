package semantic

import (
	"fmt"
	"github.com/ark-lang/ark/src/parser"
)

type BorrowCheck struct {
	currentLifetime *Lifetime
}

type Lifetime struct {
	resources map[string]Resource
}

type Resource interface {
	HasOwnership(*BorrowCheck)
}

type VariableResource struct {
	Variable *parser.Variable
	Owned    bool
}

// OWNERSHIP STUFF

func (v *VariableResource) HasOwnership(b *BorrowCheck) {
	fmt.Println("Does ", v.Variable.Name, " have ownership? ", v.Owned)
}

// CHECKS

func (v *BorrowCheck) CheckFunctionDecl(s *SemanticAnalyzer, n *parser.FunctionDecl) {
	// for the prototypes
	if n.Function.Body == nil {
		return
	}

	for _, node := range n.Function.Body.Nodes {
		v.Visit(s, node)
	}
}

func (v *BorrowCheck) CheckExpr(s *SemanticAnalyzer, n parser.Expr) {
	if variableAccessExpr, ok := n.(*parser.VariableAccessExpr); ok {
		v.CheckVariableAccessExpr(s, variableAccessExpr)
	}

	if callExpr, ok := n.(*parser.CallExpr); ok {
		v.CheckCallExpr(s, callExpr)
	}
}

func (v *BorrowCheck) CheckVariableAccessExpr(s *SemanticAnalyzer, n *parser.VariableAccessExpr) {
	// it's an argument
	if n.Variable.IsArgument {
		if variable, ok := v.currentLifetime.resources[n.Variable.Name+"_VAR"]; ok {
			variable.(*VariableResource).Owned = false
		}
		v.currentLifetime.resources[n.Variable.Name+"_ARG"] = &VariableResource{
			Variable: n.Variable,
			Owned:    true,
		}
	}
}

func (v *BorrowCheck) CheckCallExpr(s *SemanticAnalyzer, n *parser.CallExpr) {
	for _, arg := range n.Arguments {
		v.CheckExpr(s, arg)
	}
}

func (v *BorrowCheck) CheckVariableDecl(s *SemanticAnalyzer, n *parser.VariableDecl) {
	v.currentLifetime.resources[n.Variable.Name+"_VAR"] = &VariableResource{
		Variable: n.Variable,
		Owned:    true,
	}
	v.CheckExpr(s, n.Assignment)
}

// INHERITS

func (v *BorrowCheck) EnterScope(s *SemanticAnalyzer) {
	v.currentLifetime = &Lifetime{
		resources: make(map[string]Resource),
	}
}

func (v *BorrowCheck) ExitScope(s *SemanticAnalyzer) {
	for _, value := range v.currentLifetime.resources {
		value.HasOwnership(v)
	}
}

func (v *BorrowCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.FunctionDecl:
		v.CheckFunctionDecl(s, n.(*parser.FunctionDecl))
	case *parser.VariableDecl:
		v.CheckVariableDecl(s, n.(*parser.VariableDecl))
	case *parser.CallStat:
		v.CheckExpr(s, n.(*parser.CallStat).Call)
	}
}
