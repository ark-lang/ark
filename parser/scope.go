package parser

type Scope struct {
	Outer *Scope
	Vars  map[string]*Variable
	Types map[string]Type
	Funcs map[string]*Function
	Mods map[string]*Module
}

func newScope(outer *Scope) *Scope {
	return &Scope{
		Outer: outer,
		Vars:  make(map[string]*Variable),
		Types: make(map[string]Type),
		Funcs: make(map[string]*Function),
		Mods: make(map[string]*Module),
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

func (v *Scope) GetVariable(name string) *Variable {
	if r := v.Vars[name]; r != nil {
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

func (v *Scope) GetFunction(name string) *Function {
	if r := v.Funcs[name]; r != nil {
		return r
	} else if v.Outer != nil {
		return v.Outer.GetFunction(name)
	}
	return nil
}
