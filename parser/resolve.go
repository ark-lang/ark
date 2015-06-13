package parser

func (v *semanticAnalyzer) resolveAll() {
	for _, n := range v.unresolvedNodes {
		switch n.(type) {
		case *CallExpr:
			v.resolveCallExpr(n.(*CallExpr))
		case *CastExpr:
			v.resolveCastExpr(n.(*CastExpr))
		case *VariableDecl:
			v.resolveVariableDecl(n.(*VariableDecl))
		case *FunctionDecl:
			v.resolveFunctionDecl(n.(*FunctionDecl))
		}
	}
}

func (v *semanticAnalyzer) resolveCallExpr(n *CallExpr) {
	n.Function = v.module.GlobalScope.GetFunction(n.functionName)

	if n.Function == nil {
		v.err(n, "Unresolved function `%s`", n.functionName)
	}
}

func (v *semanticAnalyzer) resolveCastExpr(n *CastExpr) {
	n.Type = getTypeFromString(v.module.GlobalScope, n.typeName)

	if n.Type == nil {
		v.err(n, "Unresolved type `%s`", n.typeName)
	}
}

func (v *semanticAnalyzer) resolveVariableDecl(n *VariableDecl) {
	n.Variable.Type = getTypeFromString(v.module.GlobalScope, n.Variable.typeName)

	if n.Variable.Type == nil {
		v.err(n, "Unresolved type `%s`", n.Variable.typeName)
	}
}

func (v *semanticAnalyzer) resolveFunctionDecl(n *FunctionDecl) {
	n.Function.ReturnType = getTypeFromString(v.module.GlobalScope, n.Function.returnTypeName)

	if n.Function.ReturnType == nil {
		v.err(n, "Unresolved type `%s`", n.Function.returnTypeName)
	}
}
