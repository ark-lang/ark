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

	UsedModules map[string]*ModuleLookup
}

func newScope(outer *Scope) *Scope {
	return &Scope{
		Outer:       outer,
		Idents:      make(map[string]*Ident),
		UsedModules: make(map[string]*ModuleLookup),
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

func NewCScope() *Scope {
	s := newScope(nil)
	s.InsertType(&NamedType{Name: "uint", Type: PRIMITIVE_u32})
	s.InsertType(&NamedType{Name: "int", Type: PRIMITIVE_s32})
	s.InsertType(&NamedType{Name: "void", Type: PRIMITIVE_u8})
	return s
}

func (v *Scope) err(err string, stuff ...interface{}) {
	// TODO: which log tag
	// TODO: These errors are unacceptably shitty
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
	scope := v

	for _, modname := range name.moduleNames {
		if module, ok := scope.UsedModules[modname]; ok {
			scope = module.Module.GlobalScope
		} else if scope.Outer != nil {
			if module, ok := scope.Outer.UsedModules[modname]; ok {
				scope = module.Module.GlobalScope
			}
		} else {
			return nil
		}
	}

	if r := scope.Idents[name.name]; r != nil {
		return r
	} else if v.Outer != nil {
		return v.Outer.GetIdent(name)
	}
	return nil
}
