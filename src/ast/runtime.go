package ast

import (
	"github.com/ark-lang/ark/src/util/log"
)

var runeType Type
var stringType Type

func LoadRuntimeModule(mod *Module) {
	runeType = runtimeMustLoadType(mod, "rune")
	stringType = runtimeMustLoadType(mod, "string")

	builtinScope.InsertType(runeType, true)
	builtinScope.InsertType(stringType, true)
}

func runtimeMustLoadType(mod *Module, name string) Type {
	log.Debugln("runtime", "Loading runtime type: %s", name)
	ident := mod.ModScope.GetIdent(UnresolvedName{Name: name})
	if ident.Type != IDENT_TYPE {
		panic("INTERNAL ERROR: Type not defined in runtime: " + name)
	}
	return ident.Value.(Type)
}
