package semantic

import (
	"fmt"
	"github.com/ark-lang/ark/src/parser"
)

type BorrowCheck struct {
	currentLifetime *Lifetime
	swag            bool
}

type Lifetime struct {
	resources    map[string]Resource
	resourceKeys []string
}

type Resource interface {
	HasOwnership(*BorrowCheck) bool
}

type VariableResource struct {
	Variable *parser.Variable
	Owned    bool
}

type ParameterResource struct {
	Variable *parser.Variable
	Owned    bool
}

// OWNERSHIP STUFF

func (v *VariableResource) HasOwnership(b *BorrowCheck) bool {
	return v.Owned
}

func (v *ParameterResource) HasOwnership(b *BorrowCheck) bool {
	return v.Owned
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
	if v.swag && !n.Variable.IsParameter {
		if variable, ok := v.currentLifetime.resources[n.Variable.Name+"_VAR"]; ok {
			if varResource, ok := variable.(*VariableResource); ok {
				varResource.Owned = false
			}
		}
		v.currentLifetime.resources[n.Variable.Name+"_ARG"] = &ParameterResource{
			Variable: n.Variable,
			Owned:    true,
		}
		v.currentLifetime.resourceKeys = append(v.currentLifetime.resourceKeys, n.Variable.Name+"_ARG")
	} else {
		if variable, ok := v.currentLifetime.resources[n.Variable.Name+"_VAR"]; ok {
			if !variable.HasOwnership(v) {
				fmt.Println("error: use of transferred resource " + n.Variable.Name + "\n" + s.Module.File.MarkPos(n.Pos()))
			}
		}
	}
}

func (v *BorrowCheck) CheckAccessExpr(s *SemanticAnalyzer, n parser.AccessExpr) {
	switch n.(type) {
	case *parser.VariableAccessExpr:
		v.CheckVariableAccessExpr(s, n.(*parser.VariableAccessExpr))
	}
}

func (v *BorrowCheck) CheckAssignStat(s *SemanticAnalyzer, n *parser.AssignStat) {
	v.CheckAccessExpr(s, n.Access)
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
	v.currentLifetime.resourceKeys = append(v.currentLifetime.resourceKeys, n.Variable.Name+"_VAR")
	v.CheckExpr(s, n.Assignment)
}

// INHERITS

func (v *BorrowCheck) Finalize() {}

func (v *BorrowCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {}

func (v *BorrowCheck) EnterScope(s *SemanticAnalyzer) {
	v.currentLifetime = &Lifetime{
		resources: make(map[string]Resource),
	}
}

func (v *BorrowCheck) ExitScope(s *SemanticAnalyzer) {
	for idx, _ := range v.currentLifetime.resourceKeys {
		currentKey := v.currentLifetime.resourceKeys[idx]
		if value, ok := v.currentLifetime.resources[currentKey]; ok {
			if !value.HasOwnership(v) {
				fmt.Println("error: ownership has not been returned to ", currentKey)
			}
			delete(v.currentLifetime.resources, currentKey)
		}
	}
}

func (v *BorrowCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	switch n.(type) {
	case *parser.FunctionDecl:
		v.CheckFunctionDecl(s, n.(*parser.FunctionDecl))
	case *parser.VariableDecl:
		v.CheckVariableDecl(s, n.(*parser.VariableDecl))
	case *parser.CallStat:
		v.swag = true
		v.CheckExpr(s, n.(*parser.CallStat).Call)
		v.swag = false
	case parser.AccessExpr:
		v.CheckAccessExpr(s, n.(parser.AccessExpr))
	}
}
