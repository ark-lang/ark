package semantic

import (
	"fmt"
	"github.com/ark-lang/ark/src/parser"
)

type BorrowCheck struct {
	currentLifetime    *Lifetime
	checkingCallExpr   bool
	checkingAddrofExpr *parser.AddressOfExpr
}

type Lifetime struct {
	// it's important that the hashmap can be iterated
	// in the correct order, so we add the hashmap keys
	// here and iterate over this instead.
	resourceKeys []string
	resources    map[string]Resource
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

func (v *VariableResource) HasOwnership(b *BorrowCheck) bool {
	return v.Owned
}

func (v *ParameterResource) HasOwnership(b *BorrowCheck) bool {
	return v.Owned
}

func (v *BorrowCheck) CheckAddrofExpr(s *SemanticAnalyzer, n *parser.AddressOfExpr) {
	v.checkingAddrofExpr = n
}

func (v *BorrowCheck) CheckVariableAccessExpr(s *SemanticAnalyzer, n *parser.VariableAccessExpr) {
	// really messy, but basically if this
	// is not nil, we're checking variable access
	// inside of an addr of
	if v.checkingAddrofExpr != nil {
		addrof := v.checkingAddrofExpr
		if addrof.Mutable && !n.Variable.Mutable {
			s.Err(n, "cannot create a mutable reference to an immutable variable `%s`", n.Variable.Name)
		}
	}

	// it's an argument
	if v.checkingCallExpr && !n.Variable.IsParameter && v.checkingAddrofExpr == nil {
		// variable exists in current lifetime
		if variable, ok := v.currentLifetime.resources[n.Variable.Name+"_VAR"]; ok {
			// set ownership to false
			if varResource, ok := variable.(*VariableResource); ok {
				if !varResource.HasOwnership(v) {
					s.Err(n, "use of moved value %s", n.Variable.Name)
				}
				varResource.Owned = false
			}
		}

		// parameter has ownership
		v.currentLifetime.resources[n.Variable.Name+"_ARG"] = &ParameterResource{
			Variable: n.Variable,
			Owned:    true,
		}
		v.currentLifetime.resourceKeys = append(v.currentLifetime.resourceKeys, n.Variable.Name+"_ARG")
	} else {
		if variable, ok := v.currentLifetime.resources[n.Variable.Name+"_VAR"]; ok {
			if !variable.HasOwnership(v) {
				s.Err(n, "use of moved value %s", n.Variable.Name)
			}
		}
	}
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
	v.currentLifetime.resources[n.Variable.Name+"_VAR"] = &VariableResource{
		Variable: n.Variable,
		Owned:    true,
	}
	v.currentLifetime.resourceKeys = append(v.currentLifetime.resourceKeys, n.Variable.Name+"_VAR")
}

func (v *BorrowCheck) Finalize() {}

func (v *BorrowCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {
	if _, ok := n.(*parser.CallStat); ok {
		v.checkingCallExpr = false
	}
	if _, ok := n.(*parser.AddressOfExpr); ok {
		v.checkingAddrofExpr = nil
	}
}

func (v *BorrowCheck) EnterScope(s *SemanticAnalyzer) {
	v.currentLifetime = &Lifetime{
		resources: make(map[string]Resource),
	}
}

func (v *BorrowCheck) ExitScope(s *SemanticAnalyzer) {
	for idx, _ := range v.currentLifetime.resourceKeys {
		currentKey := v.currentLifetime.resourceKeys[idx]
		if value, ok := v.currentLifetime.resources[currentKey]; ok {
			if value.HasOwnership(v) {
				// todo check if it's heap allocated and
				// destroy it accordingly?
			}
			delete(v.currentLifetime.resources, currentKey)
		}
	}
}

func (v *BorrowCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	fmt.Printf("")

	switch n.(type) {
	case *parser.FunctionDecl:
		v.CheckFunctionDecl(s, n.(*parser.FunctionDecl))
	case *parser.VariableDecl:
		v.CheckVariableDecl(s, n.(*parser.VariableDecl))

	case *parser.CallStat:
		v.checkingCallExpr = true
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
