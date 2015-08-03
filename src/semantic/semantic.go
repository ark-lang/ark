package semantic

import (
	"fmt"
	"os"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
)

type SemanticAnalyzer struct {
	Module          *parser.Module
	unresolvedNodes []*parser.Node
	shouldExit      bool

	Checks []SemanticCheck
}

type SemanticCheck interface {
	EnterScope(s *SemanticAnalyzer)
	ExitScope(s *SemanticAnalyzer)
	Visit(*SemanticAnalyzer, parser.Node)
	PostVisit(*SemanticAnalyzer, parser.Node)
}

func (v *SemanticAnalyzer) Err(thing parser.Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Error("semantic", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Errorln("semantic", v.Module.File.MarkPos(pos))

	v.shouldExit = true
}

func (v *SemanticAnalyzer) Warn(thing parser.Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Warning("semantic", util.TEXT_YELLOW+util.TEXT_BOLD+"warning:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Warningln("semantic", v.Module.File.MarkPos(pos))
}

func NewSemanticAnalyzer(module *parser.Module, useOwnership bool) *SemanticAnalyzer {
	res := &SemanticAnalyzer{}
	res.shouldExit = false
	res.Module = module
	res.Checks = []SemanticCheck{
		&AttributeCheck{},
		&UnreachableCheck{},
		&DeprecatedCheck{},
		&RecursiveDefinitionCheck{},
		&TypeCheck{},
		&UnusedCheck{},
		&ImmutableAssignCheck{},
		&UseBeforeDeclareCheck{},
		&MiscCheck{},
	}

	if useOwnership {
		fmt.Println("warning: ownership is enabled, things may break")
		res.Checks = append(res.Checks, &BorrowCheck{})
	}

	return res
}

func (v *SemanticAnalyzer) Finalize() {
	if v.shouldExit {
		os.Exit(util.EXIT_FAILURE_SEMANTIC)
	}
}

func (v *SemanticAnalyzer) Visit(n parser.Node) {
	for _, check := range v.Checks {
		check.Visit(v, n)
	}
}

func (v *SemanticAnalyzer) PostVisit(n parser.Node) {
	for _, check := range v.Checks {
		check.PostVisit(v, n)
	}
}

func (v *SemanticAnalyzer) EnterScope(s *parser.Scope) {
	for _, check := range v.Checks {
		check.EnterScope(v)
	}
}

func (v *SemanticAnalyzer) ExitScope(s *parser.Scope) {
	for _, check := range v.Checks {
		check.ExitScope(v)
	}
}
