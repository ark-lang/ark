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
	Name      string
	Mutable   bool
	Ownership bool
}

type Borrow struct {
	Lifetime *Lifetime
	Name     string
	Mutable  bool
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
	if _, ok := n.Variable.Type.(*parser.MutableReferenceType); ok {

	} else if _, ok := n.Variable.Type.(*parser.ConstantReferenceType); ok {

	} else {
		v.currentLifetime.addResource(Resource{
			Name:      n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE),
			Mutable:   n.Variable.Mutable,
			Ownership: true,
		})
	}
}

func (v *BorrowCheck) Finalize() {

}

func (v *BorrowCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {
	// destroy the lifetime at the end of its scope.
	if _, ok := n.(*parser.FunctionDecl); ok {
		v.destroyLifetime(s)
	}
}

func (v *BorrowCheck) EnterScope(s *SemanticAnalyzer) {

}

func (v *BorrowCheck) ExitScope(s *SemanticAnalyzer) {

}

func (l *Lifetime) addResource(r Resource) {
	mangledName := r.Name + "_" + l.name
	l.resources[mangledName] = r
	l.resourceKeys = append(l.resourceKeys, mangledName)
	fmt.Println("added resource " + mangledName + " to lifetime " + l.name)
}

func (v *BorrowCheck) createLifetime(s *SemanticAnalyzer) {
	v.currentLifetime = &Lifetime{
		resources: make(map[string]Resource),
		borrows:   make(map[string]Borrow),
		name:      util.Bold("'" + string(lifetimeIndex) + ""),
	}
	v.lifetimes[v.currentLifetime.name] = v.currentLifetime
	lifetimeIndex++
	fmt.Println("created lifetime " + v.currentLifetime.name)
}

func (v *BorrowCheck) destroyLifetime(s *SemanticAnalyzer) {
	for _, key := range v.currentLifetime.resourceKeys {
		if res, ok := v.currentLifetime.resources[key]; ok {
			fmt.Println("removing resource " + res.Name + " from lifetime " + v.currentLifetime.name)
			delete(v.currentLifetime.resources, key)
		}
	}
	delete(v.lifetimes, v.currentLifetime.name)
	fmt.Println("cleaning up lifetime " + v.currentLifetime.name)
}

func (v *BorrowCheck) Init(s *SemanticAnalyzer) {
	v.lifetimes = make(map[string]*Lifetime)
}

func (v *BorrowCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	fmt.Printf("") // dummy so I can keep fmt in for now...

	switch n.(type) {
	case *parser.FunctionDecl:
		v.createLifetime(s)
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
