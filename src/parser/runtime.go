package parser

import (
	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util/log"
)

// TODO: Move this at a file and handle locating/specifying this file
// NOTE: Code specified in the runtime MUST be guaranteed semantically correct,
// as running semantic tests on it is a pain in the butt.
const RuntimeSource = `
type rune u32;
type string []u8;
`

var runtimeModule *Module

func LoadRuntime() {
	runtimeModule = &Module{
		Name: &ModuleName{
			Parts: []string{"__runtime"},
		},
		Dirpath: "__runtime",
		Parts:   make(map[string]*Submodule),
	}

	sourcefile := &lexer.Sourcefile{
		Name:     "runtime",
		Path:     "runtime.ark",
		Contents: []rune(RuntimeSource),
		NewLines: []int{-1, -1},
	}
	lexer.Lex(sourcefile)

	tree, deps := Parse(sourcefile)
	if len(deps) > 0 {
		panic("INTERNAL ERROR: No dependencies allowed in runtime")
	}
	runtimeModule.Trees = append(runtimeModule.Trees, tree)

	Construct(runtimeModule, nil)
	Resolve(runtimeModule, nil)

	for _, submod := range runtimeModule.Parts {
		Infer(submod)
	}

	LoadValues()
}

var runeType Type
var stringType Type

func LoadValues() {
	runeType = runtimeMustLoadType("rune")
	stringType = runtimeMustLoadType("string")

	builtinScope.InsertType(runeType, true)
	builtinScope.InsertType(stringType, true)
}

func runtimeMustLoadType(name string) Type {
	log.Debugln("runtime", "Loading runtime type: %s", name)
	ident := runtimeModule.ModScope.GetIdent(UnresolvedName{Name: name})
	if ident.Type != IDENT_TYPE {
		panic("INTERNAL ERROR: Type not defined in runtime: " + name)
	}
	return ident.Value.(Type)
}
