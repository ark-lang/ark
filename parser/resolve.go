package parser

// What is done:
// - variables
// - function
// What is not:
// - types
// - traits

type unresolvedName struct {
	moduleNames []string
	name        string
	modules     map[string]*Module
}

func (v unresolvedName) String() string {
	ret := ""
	for _, mod := range v.moduleNames {
		ret += mod + "::"
	}
	return ret + v.name
}

func (v *semanticAnalyzer) errResolve(thing Locatable, name unresolvedName) {
	v.err(thing, "Cannot resolve `%s`", name.String())
}

func (v *semanticAnalyzer) resolve(modules map[string]*Module) {
	v.modules = modules

	for _, node := range v.module.Nodes {
		node.resolve(v, v.module.GlobalScope)
	}
}

func (v *Block) resolve(sem *semanticAnalyzer, s *Scope) {
	for _, n := range v.Nodes {
		n.resolve(sem, v.scope)
	}
}

/**
 * Declarations
 */

func (v *VariableDecl) resolve(sem *semanticAnalyzer, s *Scope) {
	if v.Assignment != nil {
		v.Assignment.resolve(sem, s)
	}

	if v.Variable.Type != nil {
		v.Variable.Type = v.Variable.Type.resolveType(v, sem, s)
	}
}

func (v *StructDecl) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Struct = v.Struct.resolveType(v, sem, s).(*StructType)
}

func (v *EnumDecl) resolve(sem *semanticAnalyzer, s *Scope) {
	// TODO: this is a noop, right?
}

func (v *TraitDecl) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Trait = v.Trait.resolveType(v, sem, s).(*TraitType)
}

func (v *ImplDecl) resolve(sem *semanticAnalyzer, s *Scope) {
	for _, fun := range v.Functions {
		fun.resolve(sem, s)
	}
}

func (v *FunctionDecl) resolve(sem *semanticAnalyzer, s *Scope) {
	for _, param := range v.Function.Parameters {
		param.resolve(sem, s)
	}

	if v.Function.ReturnType != nil {
		v.Function.ReturnType = v.Function.ReturnType.resolveType(v, sem, s)
	}

	if !v.Prototype {
		v.Function.Body.resolve(sem, s)
	}
}

func (v *UseDecl) resolve(sem *semanticAnalyzer, s *Scope) {
	// later...
}

func (v *ModuleDecl) resolve(sem *semanticAnalyzer, s *Scope) {

}

/*
 * Statements
 */

func (v *ReturnStat) resolve(sem *semanticAnalyzer, s *Scope) {
	if v.Value != nil {
		v.Value.resolve(sem, s)
	}
}

func (v *IfStat) resolve(sem *semanticAnalyzer, s *Scope) {
	for _, expr := range v.Exprs {
		expr.resolve(sem, s)
	}

	for _, body := range v.Bodies {
		body.resolve(sem, s)
	}

	if v.Else != nil {
		v.Else.resolve(sem, s)
	}

}

func (v *BlockStat) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Block.resolve(sem, s)
}

func (v *CallStat) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Call.resolve(sem, s)
}

func (v *AssignStat) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Assignment.resolve(sem, s)

	if v.Deref != nil {
		v.Deref.resolve(sem, s)
	} else if v.Access != nil {
		v.Access.resolve(sem, s)
	}
}

func (v *LoopStat) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Body.resolve(sem, s)

	switch v.LoopType {
	case LOOP_TYPE_INFINITE:
	case LOOP_TYPE_CONDITIONAL:
		v.Condition.resolve(sem, s)
	default:
		panic("invalid loop type")
	}
}

func (v *MatchStat) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Target.resolve(sem, s)

	for pattern, stmt := range v.Branches {
		pattern.resolve(sem, s)
		stmt.resolve(sem, s)
	}
}

/*
 * Expressions
 */

func (v *IntegerLiteral) resolve(sem *semanticAnalyzer, s *Scope)  {}
func (v *FloatingLiteral) resolve(sem *semanticAnalyzer, s *Scope) {}
func (v *StringLiteral) resolve(sem *semanticAnalyzer, s *Scope)   {}
func (v *RuneLiteral) resolve(sem *semanticAnalyzer, s *Scope)     {}
func (v *BoolLiteral) resolve(sem *semanticAnalyzer, s *Scope)     {}

func (v *UnaryExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Expr.resolve(sem, s)
}

func (v *BinaryExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Lhand.resolve(sem, s)
	v.Rhand.resolve(sem, s)
}

func (v *ArrayLiteral) resolve(sem *semanticAnalyzer, s *Scope) {
	for _, mem := range v.Members {
		mem.resolve(sem, s)
	}
}

func (v *CastExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Type = v.Type.resolveType(v, sem, s)
	v.Expr.resolve(sem, s)
}

func (v *CallExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Function = s.GetFunction(v.functionName)
	if v.Function == nil {
		sem.errResolve(v, v.functionName)
	}

	for _, arg := range v.Arguments {
		arg.resolve(sem, s)
	}
}

// this whole function is so bad
// why does this even exist
// please get around to rewriting this in a way that doesn't suck
// good luck changing anything here without rewriting everything
// at least it works
// - MovingtoMars
func (v *AccessExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	// resolve the first name
	firstVar := v.Accesses[0]
	firstVar.Variable = s.GetVariable(firstVar.variableName)

	if v.Accesses[0].Variable == nil {
		sem.errResolve(v, firstVar.variableName)
	}

	// resolve everything else
	for i := 0; i < len(v.Accesses); i++ {
		switch v.Accesses[i].AccessType {
		case ACCESS_ARRAY:
			v.Accesses[i].Subscript.resolve(sem, s)

		case ACCESS_STRUCT:
			structType, ok := v.Accesses[i].Variable.Type.(*StructType)
			if !ok {
				sem.err(v, "Cannot access member of `%s`, type `%s`", v.Accesses[i].Variable.Name, v.Accesses[i].Variable.Type.TypeName())
			}

			memberName := v.Accesses[i+1].variableName.name // TODO check no mod access
			decl := structType.getVariableDecl(memberName)
			if decl == nil {
				sem.err(v, "Struct `%s` does not contain member `%s`", structType.TypeName(), memberName)
			}
			v.Accesses[i+1].Variable = decl.Variable
		case ACCESS_VARIABLE:
			// nothing to do

		case ACCESS_TUPLE:
			// nothing to do

		default:
			panic("unhandled access type")
		}
	}
}

func (v *AddressOfExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Access.resolve(sem, s)
}

func (v *DerefExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Expr.resolve(sem, s)
}

func (v *BracketExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	v.Expr.resolve(sem, s)
}

func (v *SizeofExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	if v.Expr != nil {
		v.Expr.resolve(sem, s)
	} else if v.Type != nil {
		v.Type = v.Type.resolveType(v, sem, s)
	} else {
		panic("invalid state")
	}
}

func (v *TupleLiteral) resolve(sem *semanticAnalyzer, s *Scope) {
	for _, mem := range v.Members {
		mem.resolve(sem, s)
	}
}

func (v *DefaultMatchBranch) resolve(sem *semanticAnalyzer, s *Scope) {}

/*
 * Types
 */

func (v PrimitiveType) resolveType(src Locatable, sem *semanticAnalyzer, s *Scope) Type {
	return v
}

func (v *StructType) resolveType(src Locatable, sem *semanticAnalyzer, s *Scope) Type {
	for _, vari := range v.Variables {
		vari.resolve(sem, s)
	}
	return v
}

func (v ArrayType) resolveType(src Locatable, sem *semanticAnalyzer, s *Scope) Type {
	return arrayOf(v.MemberType.resolveType(src, sem, s))
}

func (v *TraitType) resolveType(src Locatable, sem *semanticAnalyzer, s *Scope) Type {
	for _, fun := range v.Functions {
		fun.resolve(sem, s)
	}
	return v
}

func (v PointerType) resolveType(src Locatable, sem *semanticAnalyzer, s *Scope) Type {
	return pointerTo(v.Addressee.resolveType(src, sem, s))
}

func (v *TupleType) resolveType(src Locatable, sem *semanticAnalyzer, s *Scope) Type {
	for idx, mem := range v.Members {
		v.Members[idx] = mem.resolveType(src, sem, s)
	}
	return v
}

func (v *UnresolvedType) resolveType(src Locatable, sem *semanticAnalyzer, s *Scope) Type {
	typ := s.GetType(v.Name)
	if typ == nil {
		sem.err(src, "Cannot resolve `%s`", v.Name)
	}
	return typ
}
