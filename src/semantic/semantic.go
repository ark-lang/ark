package semantic

import (
	"fmt"
	"os"

	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
)

type SemanticAnalyzer struct {
	Module          *ast.Module
	Submodule       *ast.Submodule
	unresolvedNodes []*ast.Node
	shouldExit      bool

	Check SemanticCheck
}

type SemanticCheck interface {
	Init(s *SemanticAnalyzer)
	EnterScope(s *SemanticAnalyzer)
	ExitScope(s *SemanticAnalyzer)
	Visit(*SemanticAnalyzer, ast.Node)
	PostVisit(*SemanticAnalyzer, ast.Node)
	Finalize(*SemanticAnalyzer)
	Name() string
}

func (v *SemanticAnalyzer) Err(thing ast.Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Error("semantic", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Errorln("semantic", v.Submodule.File.MarkPos(pos))

	v.shouldExit = true
}

func (v *SemanticAnalyzer) Warn(thing ast.Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Warning("semantic", util.TEXT_YELLOW+util.TEXT_BOLD+"warning:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Warningln("semantic", v.Submodule.File.MarkPos(pos))
}

func SemCheck(module *ast.Module, ignoreUnused bool) {
	checks := []SemanticCheck{
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
		checks = append(checks, &UnusedCheck{})
	}

	for _, check := range checks {
		log.Timed("analysis pass", check.Name(), func() {
			for _, submod := range module.Parts {
				log.Timed("checking submodule", module.Name.String()+"/"+submod.File.Name, func() {
					res := &SemanticAnalyzer{
						Module:    module,
						Submodule: submod,
						Check:     check,
					}
					res.Init()

					vis := ast.NewASTVisitor(res)
					vis.VisitSubmodule(submod)

					res.Finalize()
				})

			}
		})
	}
}

// the initial check for a semantic pass
// this will be called _once_ and should be
// used to initialize things, etc...
func (v *SemanticAnalyzer) Init() {
	v.Check.Init(v)
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
	v.Check.Finalize(v)

	if v.shouldExit {
		os.Exit(util.EXIT_FAILURE_SEMANTIC)
	}
}

func (v *SemanticAnalyzer) Visit(n *ast.Node) bool {
	v.Check.Visit(v, *n)

	// NOTE: The following means that if we encountered an error we will not
	// analyze further down the AST. This should hinder some panics with
	// relation to invalid data.
	return !v.shouldExit
}

func (v *SemanticAnalyzer) PostVisit(n *ast.Node) {
	v.Check.PostVisit(v, *n)
}

func (v *SemanticAnalyzer) EnterScope() {
	v.Check.EnterScope(v)
}

func (v *SemanticAnalyzer) ExitScope() {
	v.Check.ExitScope(v)
}
