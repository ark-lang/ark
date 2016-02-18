package ast

import (
	"fmt"
	"os"
	"strings"

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
	Type   IdentType
	Value  interface{}
	Public bool
	Scope  *Scope
}

type Scope struct {
	Outer       *Scope
	Idents      map[string]*Ident
	Module      *Module   // module this scope belongs to, nil if builtin
	Function    *Function // function this scope is inside, nil if global/builtin/etc
	UsedModules map[string]*Module
}

func newScope(outer *Scope, mod *Module, fn *Function) *Scope {
	return &Scope{
		Outer:       outer,
		Idents:      make(map[string]*Ident),
		UsedModules: make(map[string]*Module),
		Module:      mod,
		Function:    fn,
	}
}

var builtinScope *Scope

func init() {
	builtinScope = newScope(nil, nil, nil)

	for i := 0; i < len(_PrimitiveType_index); i++ {
		builtinScope.InsertType(PrimitiveType(i), true)
	}
}

func NewGlobalScope(mod *Module) *Scope {
	s := newScope(builtinScope, mod, nil)

	return s
}

func NewCScope(mod *Module) *Scope {
	s := newScope(nil, mod, nil)
	s.InsertType(&NamedType{Name: "uint", Type: PRIMITIVE_u32}, true)
	s.InsertType(&NamedType{Name: "int", Type: PRIMITIVE_s32}, true)
	s.InsertType(&NamedType{Name: "void", Type: PRIMITIVE_u8}, true)
	return s
}

func (v *Scope) err(err string, stuff ...interface{}) {
	// TODO: which log tag
	// TODO: These errors are unacceptably shitty
	log.Error("parser", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_PARSE)
}

func (v *Scope) InsertIdent(value interface{}, name string, typ IdentType, public bool) *Ident {
	c := v.Idents[name]
	if c == nil {
		v.Idents[name] = &Ident{
			Type:   typ,
			Value:  value,
			Public: public,
			Scope:  v,
		}
	}
	return c
}

func (v *Scope) InsertType(t Type, public bool) *Ident {
	if sub, ok := t.(*SubstitutionType); ok {
		return v.InsertIdent(t, sub.Name, IDENT_TYPE, public)
	}
	return v.InsertIdent(t, t.TypeName(), IDENT_TYPE, public)
}

func (v *Scope) InsertVariable(t *Variable, public bool) *Ident {
	return v.InsertIdent(t, t.Name, IDENT_VARIABLE, public)
}

func (v *Scope) InsertFunction(t *Function, public bool) *Ident {
	return v.InsertIdent(t, t.Name, IDENT_FUNCTION, public)
}

func (v *Scope) UseModule(t *Module) {
	v.UsedModules[t.Name.Last()] = t
}

func (v *Scope) UseModuleAs(t *Module, name string) {
	v.UsedModules[name] = t
}

func (v *Scope) GetIdent(name UnresolvedName) *Ident {
	scope := v

	for idx, modname := range name.ModuleNames {
		module, ok := scope.UsedModules[modname]
		for !ok && scope.Outer != nil {
			scope = scope.Outer
			module, ok = scope.UsedModules[modname]
		}

		if !ok {
			// Check for the case of the static method
			if idx == len(name.ModuleNames)-1 {
				lastName, method := name.Split()
				if r := v.GetIdent(lastName); r != nil && r.Type == IDENT_TYPE {
					typ := r.Value.(Type)
					if nt, ok := typ.(*NamedType); ok {
						fn := nt.GetStaticMethod(method)
						if fn != nil {
							return &Ident{IDENT_FUNCTION, fn, true, scope}
						}
					}
				}
			}

			return nil
		}

		scope = module.ModScope
	}

	if r := scope.Idents[name.Name]; r != nil {
		return r
	} else if r := scope.UsedModules[name.Name]; r != nil {
		return &Ident{IDENT_MODULE, r, true, v}
	} else if v.Outer != nil {
		return v.Outer.GetIdent(name)
	}

	return nil
}

func (v *Scope) Dump(depth int) {
	indent := strings.Repeat(" ", depth)

	if depth == 0 {
		log.Debug("parser", indent)
		log.Debugln("parser", "This scope:")
	}

	for name, ident := range v.Idents {
		log.Debug("parser", indent)
		log.Debugln("parser", " %s (%s)", name, ident.Type)
	}

	if v.Outer != nil {
		log.Debug("parser", indent)
		log.Debugln("parser", "Parent scope:")
		v.Outer.Dump(depth + 1)
	}

}
