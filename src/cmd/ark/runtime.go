package main

import (
	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/semantic"
	"github.com/ark-lang/ark/src/util/log"
)

// TODO: Move this at a file and handle locating/specifying this file
const RuntimeSource = `
type rune u32;
type string []u8;
`

var runtimeModule *ast.Module

func LoadRuntime() {
	runtimeModule = &ast.Module{
		Name: &ast.ModuleName{
			Parts: []string{"__runtime"},
		},
		Dirpath: "__runtime",
		Parts:   make(map[string]*ast.Submodule),
	}

	sourcefile := &lexer.Sourcefile{
		Name:     "runtime",
		Path:     "runtime.ark",
		Contents: []rune(RuntimeSource),
		NewLines: []int{-1, -1},
	}
	lexer.Lex(sourcefile)

	tree, deps := parser.Parse(sourcefile)
	if len(deps) > 0 {
		panic("INTERNAL ERROR: No dependencies allowed in runtime")
	}
	runtimeModule.Trees = append(runtimeModule.Trees, tree)

	ast.Construct(runtimeModule, nil)
	ast.Resolve(runtimeModule, nil)

	for _, submod := range runtimeModule.Parts {
		ast.Infer(submod)
	}

	for _, submod := range runtimeModule.Parts {
		sem := semantic.NewSemanticAnalyzer(submod, *buildOwnership, *ignoreUnused)
		vis := ast.NewASTVisitor(sem)
		vis.VisitSubmodule(submod)
		sem.Finalize()
	}

	LoadValues()
}

var runeType ast.Type
var stringType ast.Type

func LoadValues() {
	rts := &ast.RuntimeTypes{
		RuneType: runtimeMustLoadType("rune"),
		StringType: runtimeMustLoadType("string"),
	}
	ast.SetRuntimeTypes(rts)
}

func runtimeMustLoadType(name string) ast.Type {
	log.Debugln("runtime", "Loading runtime type: %s", name)
	ident := runtimeModule.ModScope.GetIdent(ast.UnresolvedName{Name: name})
	if ident.Type != ast.IDENT_TYPE {
		panic("INTERNAL ERROR: Type not defined in runtime: " + name)
	}
	return ident.Value.(ast.Type)
}
