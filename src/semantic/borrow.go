package semantic

/**

	This is the semantic check for the borrow checker,
	_and_ the move semantics.

	For now the borrow checker will apply move semantics
	to things on the stack as well as things allocated
	on the heap. Note that this will change once we have
	some kind of library for allocating memory...

**/

import (
	"fmt"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
)

var lifetimeIndex rune = '0'

type BorrowCheck struct {
	lifetimes       map[string]*Lifetime
	currentLifetime *Lifetime

	// state stuff for more context
	callExpr   *parser.CallExpr
	addrofExpr *parser.AddressOfExpr
}

type Lifetime struct {
	// it's important that the hashmap can be iterated
	// in the correct order, so we add the hashmap keys
	// here and iterate over this instead.
	Outer *Lifetime

	resourceKeys []string
	resources    map[string]*Resource

	borrowKeys []string
	borrows    map[string]*Borrow

	name string
}

type Resource struct {
	Name      string
	Mutable   bool
	Ownership bool
	Borrowed  bool
}

type Borrow struct {
	Lifetime *Lifetime
	Name     string
	Mutable  bool
}

func (v *BorrowCheck) CheckAddrofExpr(s *SemanticAnalyzer, n *parser.AddressOfExpr) {
	v.addrofExpr = n
}

func (v *BorrowCheck) CheckVariableAccessExpr(s *SemanticAnalyzer, n *parser.VariableAccessExpr) {
	// mangled name for resource checking
	mangledName := n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE) + "_" + v.currentLifetime.name

	// some weird checks to see if we're passing an argument
	if v.callExpr != nil && v.addrofExpr == nil && !n.Variable.IsParameter {
		if variable, ok := v.currentLifetime.resources[mangledName]; ok {
			if !variable.Ownership {
				s.Err(n, "use of moved value %s", n.Variable.Name)
			} else if variable.Borrowed {
				s.Err(n, "use of borrowed value %s", n.Variable.Name)
			} else {
				fmt.Println("resource ", n.Variable.Name, " has been moved")
				variable.Ownership = false
			}
		}
	} else if v.addrofExpr != nil && v.callExpr == nil {
		// this is the variable that is being borrowed!
		// note this only borrows for actual references
		// not stuff like a function call
		if variable, ok := v.currentLifetime.resources[mangledName]; ok {
			variable.Borrowed = true
		}
	} else {
		// creates a new resource
		if variable, ok := v.currentLifetime.resources[mangledName]; ok {
			if !variable.Ownership {
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
	switch n.Variable.Type.(type) {
	case parser.MutableReferenceType:
		v.currentLifetime.addBorrow(&Borrow{
			Name:     n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE),
			Mutable:  true,
			Lifetime: v.currentLifetime,
		})
	case parser.ConstantReferenceType:
		v.currentLifetime.addBorrow(&Borrow{
			Name:     n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE),
			Mutable:  false,
			Lifetime: v.currentLifetime,
		})
	default:
		v.currentLifetime.addResource(&Resource{
			Name:      n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE),
			Mutable:   n.Variable.Mutable,
			Ownership: true,
		})
	}
}

func (v *BorrowCheck) Finalize() {

}

func (v *BorrowCheck) PostVisit(s *SemanticAnalyzer, n parser.Node) {
	if _, ok := n.(*parser.CallStat); ok {
		v.callExpr = nil
	}

	if _, ok := n.(*parser.AddressOfExpr); ok {
		v.addrofExpr = nil
	}

	// cleanup lifetimes
	if _, ok := n.(*parser.FunctionDecl); ok {
		v.destroyLifetime(s)
	}

	if _, ok := n.(*parser.BlockStat); ok {
		v.destroyLifetime(s)
	}
}

func (v *BorrowCheck) EnterScope(s *SemanticAnalyzer) {

}

func (v *BorrowCheck) ExitScope(s *SemanticAnalyzer) {

}

func (v *BorrowCheck) Destroy(s *SemanticAnalyzer) {
	// cleanup our static lifetime
	v.destroyLifetime(s)
}

func (l *Lifetime) addBorrow(b *Borrow) {
	mangledName := b.Name + "_" + l.name
	l.borrows[mangledName] = b
	l.borrowKeys = append(l.borrowKeys, mangledName)
	b.Name = mangledName
	fmt.Println("added borrow " + mangledName + " to lifetime " + l.name)
}

func (l *Lifetime) addResource(r *Resource) {
	mangledName := r.Name + "_" + l.name
	l.resources[mangledName] = r
	l.resourceKeys = append(l.resourceKeys, mangledName)
	r.Name = mangledName
	fmt.Println("added resource " + mangledName + " to lifetime " + l.name)
}

func (v *BorrowCheck) createLifetime(s *SemanticAnalyzer) {
	lifetimeName := string(lifetimeIndex)

	// hack for readability sakes
	// will change 0 to static
	// then set to a' for the initial
	// lifetimes
	if lifetimeIndex == '0' {
		lifetimeName = "static"
		lifetimeIndex = 'a'
	}

	temp := v.currentLifetime
	v.currentLifetime = &Lifetime{
		resources: make(map[string]*Resource),
		borrows:   make(map[string]*Borrow),
		name:      util.Bold("'" + lifetimeName),
		Outer:     temp,
	}

	v.lifetimes[v.currentLifetime.name] = v.currentLifetime
	lifetimeIndex++

	fmt.Println("\ncreated lifetime " + v.currentLifetime.name)
}

func (v *BorrowCheck) destroyLifetime(s *SemanticAnalyzer) {
	temp := v.currentLifetime.Outer
	for _, key := range v.currentLifetime.borrowKeys {
		if borrow, ok := v.currentLifetime.borrows[key]; ok {
			fmt.Println("removing borrow " + borrow.Name + " from lifetime " + v.currentLifetime.name)
			delete(v.currentLifetime.borrows, key)
		}
	}

	for _, key := range v.currentLifetime.resourceKeys {
		if res, ok := v.currentLifetime.resources[key]; ok {
			// this is a hack, and I don't think this will
			// scale well or at all, but we'll cross that
			// bridge when it fucks me over...
			res.Borrowed = false

			fmt.Println("removing resource " + res.Name + " from lifetime " + v.currentLifetime.name)
			delete(v.currentLifetime.resources, key)
		}
	}

	fmt.Println("cleaning up lifetime " + v.currentLifetime.name + "\n")
	delete(v.lifetimes, v.currentLifetime.name)

	// outer exists, switch to that...
	if temp != nil {
		v.currentLifetime = temp
	}
}

func (v *BorrowCheck) Init(s *SemanticAnalyzer) {
	v.lifetimes = make(map[string]*Lifetime)

	// this is our global 'static lifetime
	v.createLifetime(s)
}

func (v *BorrowCheck) Visit(s *SemanticAnalyzer, n parser.Node) {
	fmt.Printf("") // dummy so I can keep fmt in for now...

	switch n := n.(type) {
	case *parser.FunctionDecl:
		v.createLifetime(s)
		v.CheckFunctionDecl(s, n)

	case *parser.BlockStat:
		v.createLifetime(s)

	case *parser.VariableDecl:
		v.CheckVariableDecl(s, n)

	case *parser.CallStat:
		v.callExpr = n.Call
		v.CheckCallStat(s, n)

	case *parser.AssignStat:
		v.CheckAssignStat(s, n)

	case *parser.CallExpr:
		v.CheckCallExpr(s, n)

	case *parser.AddressOfExpr:
		v.CheckAddrofExpr(s, n)

	case *parser.VariableAccessExpr:
		v.CheckVariableAccessExpr(s, n)
	}
}
