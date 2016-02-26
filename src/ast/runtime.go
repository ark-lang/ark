package ast

import (
	"github.com/ark-lang/ark/src/util/log"
)

func LoadRuntimeModule(mod *Module) {
	for name, ident := range mod.ModScope.Idents {
		if ident.Public {
			builtinScope.InsertIdent(ident.Value, name, ident.Type, ident.Public)
		}
	}
}

func runtimeMustLoadType(mod *Module, name string) Type {
	log.Debugln("runtime", "Loading runtime type: %s", name)
	ident := mod.ModScope.GetIdent(UnresolvedName{Name: name})
	if ident.Type != IDENT_TYPE {
		panic("INTERNAL ERROR: Type not defined in runtime: " + name)
	}
	return ident.Value.(Type)
}
