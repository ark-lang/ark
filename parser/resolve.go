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

func (v *semanticAnalyzer) resolve() {
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
}

func (v *StructDecl) resolve(sem *semanticAnalyzer, s *Scope) {

}

func (v *TraitDecl) resolve(sem *semanticAnalyzer, s *Scope) {

}

func (v *ImplDecl) resolve(sem *semanticAnalyzer, s *Scope) {

}

func (v *FunctionDecl) resolve(sem *semanticAnalyzer, s *Scope) {
	if !v.Prototype {
		v.Function.Body.resolve(sem, s)
	}
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

func (v *AccessExpr) resolve(sem *semanticAnalyzer, s *Scope) {
	if len(v.structVariableNames) > 0 {
		panic("todo")
	} else {
		v.Variable = s.GetVariable(v.variableName)
		if v.Variable == nil {
			sem.errResolve(v, v.variableName)
		}
	}

	/*for {
		structType, ok := v.Variable.Type.(*StructType)
		if !ok {
			v.err("Cannot access member of `%s`, type `%s`", v.Variable.Name, v.Variable.Type.TypeName())
		}

		memberName
		decl := structType.getVariableDecl(memberName)
		if decl == nil {
			v.err("Struct `%s` does not contain member `%s`", structType.TypeName(), memberName)
		}

		v.StructVariables = append(v.StructVariables, v.Variable)
		v.Variable = decl.Variable
	}*/
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
