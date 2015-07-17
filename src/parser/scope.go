package parser

import (
	"fmt"
	"os"

	"github.com/ark-lang/ark/src/util/log"

	"github.com/ark-lang/ark/src/util"
)

type IdentType int

const (
	IDENT_VARIABLE IdentType = iota
	IDENT_TYPE
	IDENT_FUNCTION
	IDENT_MODULE
)

func (v IdentType) String() string {
	switch v {
	case IDENT_FUNCTION:
		return "function"
	case IDENT_MODULE:
		return "module"
	case IDENT_TYPE:
		return "type"
	case IDENT_VARIABLE:
		return "variable"
	default:
		panic("unimplemented ident type")
	}
}

type Ident struct {
	Type  IdentType
	Value interface{}
}

type Scope struct {
	Outer  *Scope
	Idents map[string]*Ident

	UsedModules map[string]*Module
}

func newScope(outer *Scope) *Scope {
	return &Scope{
		Outer:       outer,
		Idents:      make(map[string]*Ident),
		UsedModules: make(map[string]*Module),
	}
}

var builtinScope *Scope

func init() {
	builtinScope = newScope(nil)

	for i := 0; i < len(_PrimitiveType_index); i++ {
		builtinScope.InsertType(PrimitiveType(i))
	}
}

func NewGlobalScope() *Scope {
	s := newScope(builtinScope)

	return s
}

func (v *Scope) err(err string, stuff ...interface{}) {
	// TODO: which log tag
	log.Error("parser", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_CODEGEN)
}

func (v *Scope) InsertIdent(value interface{}, name string, typ IdentType) *Ident {
	c := v.Idents[name]
	if c == nil {
		v.Idents[name] = &Ident{
			Type:  typ,
			Value: value,
		}
	}
	return c
}

func (v *Scope) InsertType(t Type) *Ident {
	return v.InsertIdent(t, t.TypeName(), IDENT_TYPE)
}

func (v *Scope) InsertVariable(t *Variable) *Ident {
	return v.InsertIdent(t, t.Name, IDENT_VARIABLE)
}

func (v *Scope) InsertFunction(t *Function) *Ident {
	return v.InsertIdent(t, t.Name, IDENT_FUNCTION)
}

func (v *Scope) GetIdent(name unresolvedName) *Ident {
	if len(name.moduleNames) > 0 {
		moduleName := name.moduleNames[0]

		if module, ok := v.UsedModules[moduleName]; ok {
			if r := module.GlobalScope.Idents[name.name]; r != nil {
				return r
			} else {
				v.err("could not find function `%s` in module `%s`", name.name, moduleName)
			}
		} else if v.Outer != nil {
			if module, ok := v.Outer.UsedModules[moduleName]; ok {
				if r := module.GlobalScope.Idents[name.name]; r != nil {
					return r
				} else {
					v.err("could not find function `%s` in module `%s`", name.name, moduleName)
				}
			}
		} else {
			v.err("could not find `" + moduleName + "`, are you sure it's being used in this module?\n\n    `use " + moduleName + ";`\n")
		}
	}

	if r := v.Idents[name.name]; r != nil {
		return r
	} else if v.Outer != nil {
		return v.Outer.GetIdent(name)
	}
	return nil
}
