package semantic

import (
	"fmt"
	"os"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
)

type SemanticAnalyzer struct {
	Submodule       *parser.Submodule
	unresolvedNodes []*parser.Node
	shouldExit      bool

	Checks []SemanticCheck
}

type SemanticCheck interface {
	Init(s *SemanticAnalyzer)
	EnterScope(s *SemanticAnalyzer)
	ExitScope(s *SemanticAnalyzer)
	Visit(*SemanticAnalyzer, parser.Node)
	PostVisit(*SemanticAnalyzer, parser.Node)
	Finalize(*SemanticAnalyzer)
}

func (v *SemanticAnalyzer) Err(thing parser.Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Error("semantic", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Errorln("semantic", v.Submodule.File.MarkPos(pos))

	v.shouldExit = true
}

func (v *SemanticAnalyzer) Warn(thing parser.Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Warning("semantic", util.TEXT_YELLOW+util.TEXT_BOLD+"warning:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Warningln("semantic", v.Submodule.File.MarkPos(pos))
}

func NewSemanticAnalyzer(module *parser.Submodule, useOwnership bool, ignoreUnused bool) *SemanticAnalyzer {
	res := &SemanticAnalyzer{}
	res.shouldExit = false
	res.Submodule = module
	res.Checks = []SemanticCheck{
		&AttributeCheck{},
		&UnreachableCheck{},
		&BreakAndNextCheck{},
		&DeprecatedCheck{},
		&RecursiveDefinitionCheck{},
		&TypeCheck{},
		&ImmutableAssignCheck{},
		&UseBeforeDeclareCheck{},
		&MiscCheck{},
		&ReferenceCheck{},
	}

	if !ignoreUnused {
		res.Checks = append(res.Checks, &UnusedCheck{})
	}

	res.Init()

	return res
}

// the initial check for a semantic pass
// this will be called _once_ and should be
// used to initialize things, etc...
func (v *SemanticAnalyzer) Init() {
	for _, check := range v.Checks {
		check.Init(v)
	}
}

// Finalize is called after all checks have been run, and should be used for
// cleaning up and any checks that depend on having completely traversed the
// syntax tree.
func (v *SemanticAnalyzer) Finalize() {
	// If we already encountered an error, exit now
	if v.shouldExit {
		os.Exit(util.EXIT_FAILURE_SEMANTIC)
	}

	// destroy stuff before finalisation
	for _, check := range v.Checks {
		check.Finalize(v)
	}

	if v.shouldExit {
		os.Exit(util.EXIT_FAILURE_SEMANTIC)
	}
}

func (v *SemanticAnalyzer) Visit(n *parser.Node) bool {
	for _, check := range v.Checks {
		check.Visit(v, *n)
	}

	// NOTE: The following means that if we encountered an error we will not
	// analyze further down the AST. This should hinder some panics with
	// relation to invalid data.
	// Should the need arise we can further propagate this bool as a return
	// value from SemanticCheck.Visit(). For this to work properly we might
	// need to loop over checks as the outer loop, instead of the inner loop.
	return !v.shouldExit
}

func (v *SemanticAnalyzer) PostVisit(n *parser.Node) {
	for _, check := range v.Checks {
		check.PostVisit(v, *n)
	}
}

func (v *SemanticAnalyzer) EnterScope() {
	for _, check := range v.Checks {
		check.EnterScope(v)
	}
}

func (v *SemanticAnalyzer) ExitScope() {
	for _, check := range v.Checks {
		check.ExitScope(v)
	}
}
