package parser

type Scope struct {
	Outer *Scope
	Vars map[string]*Variable
	Types map[string]Type
}

func newScope(outer *Scope) *Scope {
	return &Scope {
		Outer: outer,
		Vars: make(map[string]*Variable),
		Types: make(map[string]Type),
	}
}

func newGlobalScope() *Scope {
	s := newScope(nil)
	
	for i := 0; i < len(_PrimitiveType_index); i++ {
		s.InsertType(PrimitiveType(i))
	}
	
	return s
}

func (v *Scope) InsertType(t Type) Type {
	c := v.Types[t.GetTypeName()]
	if c == nil {
		v.Types[t.GetTypeName()] = t
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
