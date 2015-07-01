package parser

import (
	"fmt"
)

type Scope struct {
	Outer       *Scope
	Vars        map[string]*Variable
	Types       map[string]Type
	Funcs       map[string]*Function
	Mods        map[string]*Module
	UsedModules map[string]*Module
}

func newScope(outer *Scope) *Scope {
	return &Scope{
		Outer:       outer,
		Vars:        make(map[string]*Variable),
		Types:       make(map[string]Type),
		Funcs:       make(map[string]*Function),
		Mods:        make(map[string]*Module),
		UsedModules: make(map[string]*Module),
	}
}

func newGlobalScope() *Scope {
	s := newScope(nil)

	for i := 0; i < len(_PrimitiveType_index); i++ {
		s.InsertType(PrimitiveType(i))
	}

	return s
}

func (v *Scope) IsGlobal() bool {
	return v.Outer == nil
}

func (v *Scope) InsertType(t Type) Type {
	c := v.Types[t.TypeName()]
	if c == nil {
		v.Types[t.TypeName()] = t
	}
	return c
}

func (v *Scope) GetType(name string) Type {
	if r := v.Types[name]; r != nil {
		return r
	} else if v.Outer != nil {
		return v.Outer.GetType(name)
	}
	return nil
}

func (v *Scope) InsertVariable(t *Variable) *Variable {
	c := v.Vars[t.Name]
	if c == nil {
		v.Vars[t.Name] = t
		t.scope = v
	}
	return c
}

func (v *Scope) GetVariable(name unresolvedName) *Variable {
	// TODO modules
	if len(name.moduleNames) > 0 {
		panic("todo module access")
	}

	if r := v.Vars[name.name]; r != nil {
		return r
	} else if v.Outer != nil {
		return v.Outer.GetVariable(name)
	}
	return nil
}

func (v *Scope) InsertFunction(t *Function) *Function {
	c := v.Funcs[t.Name]
	if c == nil {
		v.Funcs[t.Name] = t
		t.scope = v
	}
	return c
}

func (v *Scope) GetFunction(name unresolvedName) *Function {
	if len(name.moduleNames) > 0 {
		moduleName := name.moduleNames[0]
		fmt.Println("finding module by the name " + moduleName)

		if module, ok := v.UsedModules[moduleName]; ok {
			fmt.Println("FOUND THE MODULE " + moduleName)

			if r := module.GlobalScope.Funcs[name.name]; r != nil {
				return r
			} else {
				fmt.Println("can't find the function in the module yo")
			}
		} else if v.Outer != nil {
			fmt.Println("FOUND THE MODULE IN OUTER! " + moduleName)

			if module, ok := v.Outer.UsedModules[moduleName]; ok {
				if r := module.GlobalScope.Funcs[name.name]; r != nil {
					fmt.Println("Found function " + name.name + " in the global scope funcs thing!")
					return r
				} else {
					fmt.Println("can't find the function in the module yo")
				}
			}
		} else {
			fmt.Println("no module " + moduleName + " found in used modules")
		}
	}

	if r := v.Funcs[name.name]; r != nil {
		return r
	} else if v.Outer != nil {
		return v.Outer.GetFunction(name)
	}
	return nil
}
