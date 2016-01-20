package parser

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
	Type  IdentType
	Value interface{}
}

type Scope struct {
	Outer       *Scope
	Idents      map[string]*Ident
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
var stringType = &NamedType{
	Name: "string",
	Type: ArrayOf(PRIMITIVE_u8),
}

func init() {
	builtinScope = newScope(nil)

	for i := 0; i < len(_PrimitiveType_index); i++ {
		builtinScope.InsertType(PrimitiveType(i))
	}

	builtinScope.InsertType(stringType)
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
	os.Exit(util.EXIT_FAILURE_PARSE)
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

func (v *Scope) UseModule(t *Module) {
	v.UsedModules[t.Name.Last()] = t
}

func (v *Scope) UseModuleAs(t *Module, name string) {
	v.UsedModules[name] = t
}

func (v *Scope) GetIdent(name UnresolvedName) *Ident {
	scope := v

	for _, modname := range name.ModuleNames {
		if module, ok := scope.UsedModules[modname]; ok {
			//TODO: We might need to do somethign with UseScope here
			scope = module.ModScope
		} else if scope.Outer != nil {
			if module, ok := scope.Outer.UsedModules[modname]; ok {
				scope = module.ModScope
			}
		} else {
			return nil
		}
	}

	if r := scope.Idents[name.Name]; r != nil {
		return r
	} else if r := scope.UsedModules[name.Name]; r != nil {
		return &Ident{IDENT_MODULE, r}
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
