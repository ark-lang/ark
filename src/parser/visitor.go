package parser

import (
	"reflect"
)

type Visitor interface {
	EnterScope(s *Scope)
	ExitScope(s *Scope)

	Visit(Node)
	PostVisit(Node)
}

type ASTVisitor struct {
	Visitor Visitor
}

func NewASTVisitor(visitor Visitor) *ASTVisitor {
	return &ASTVisitor{Visitor: visitor}
}

func (v *ASTVisitor) VisitModule(module *Module) {
	v.EnterScope(module.GlobalScope)
	v.VisitNodes(module.Nodes)
	v.ExitScope(module.GlobalScope)
}

func (v *ASTVisitor) VisitNodes(nodes []Node) {
	for _, node := range nodes {
		v.Visit(node)
	}
}

func (v *ASTVisitor) Visit(n Node) {
	if isNil(n) {
		return
	}

	v.Visitor.Visit(n)
	v.VisitChildren(n)
	v.Visitor.PostVisit(n)
}

func (v *ASTVisitor) EnterScope(s *Scope) {
	v.Visitor.EnterScope(s)
}

func (v *ASTVisitor) ExitScope(s *Scope) {
	v.Visitor.ExitScope(s)
}

func (v *ASTVisitor) VisitChildren(n Node) {
	if block, ok := n.(*Block); ok {
		if !block.NonScoping {
			v.EnterScope(block.scope)
		}

		v.VisitNodes(block.Nodes)

		if !block.NonScoping {
			v.ExitScope(block.scope)
		}

	} else if returnStat, ok := n.(*ReturnStat); ok {
		v.Visit(returnStat.Value)

	} else if ifStat, ok := n.(*IfStat); ok {
		v.VisitNodes(exprsToNodes(ifStat.Exprs))
		v.VisitNodes(blocksToNodes(ifStat.Bodies))
		v.Visit(ifStat.Else)

	} else if assignStat, ok := n.(*AssignStat); ok {
		v.Visit(assignStat.Assignment)
		v.Visit(assignStat.Access)

	} else if binopAssignStat, ok := n.(*BinopAssignStat); ok {
		v.Visit(binopAssignStat.Assignment)
		v.Visit(binopAssignStat.Access)

	} else if loopStat, ok := n.(*LoopStat); ok {
		v.Visit(loopStat.Body)

		switch loopStat.LoopType {
		case LOOP_TYPE_INFINITE:
		case LOOP_TYPE_CONDITIONAL:
			v.Visit(loopStat.Condition)
		default:
			panic("invalid loop type")
		}

	} else if matchStat, ok := n.(*MatchStat); ok {
		v.Visit(matchStat.Target)

		for pattern, stmt := range matchStat.Branches {
			v.Visit(pattern)
			v.Visit(stmt)
		}

	} else if binaryExpr, ok := n.(*BinaryExpr); ok {
		v.Visit(binaryExpr.Lhand)
		v.Visit(binaryExpr.Rhand)

	} else if arrayLiteral, ok := n.(*ArrayLiteral); ok {
		v.VisitNodes(exprsToNodes(arrayLiteral.Members))

	} else if callExpr, ok := n.(*CallExpr); ok {
		switch callExpr.functionSource.(type) {
		case *StructAccessExpr:
			v.Visit(callExpr.functionSource.(*StructAccessExpr).Struct)
		}

		// TODO: Visit callExpr.FunctionSource? once we get lambda/function types
		v.VisitNodes(exprsToNodes(callExpr.Arguments))
		v.Visit(callExpr.ReceiverAccess)

	} else if arrayAccessExpr, ok := n.(*ArrayAccessExpr); ok {
		v.Visit(arrayAccessExpr.Array)
		v.Visit(arrayAccessExpr.Subscript)

	} else if sizeofExpr, ok := n.(*SizeofExpr); ok {
		// TODO: Maybe visit sizeofExpr.Type at some point?
		v.Visit(sizeofExpr.Expr)

	} else if tupleLiteral, ok := n.(*TupleLiteral); ok {
		v.VisitNodes(exprsToNodes(tupleLiteral.Members))

	} else if structLiteral, ok := n.(*StructLiteral); ok {
		for _, value := range structLiteral.Values {
			v.Visit(value)
		}

	} else if enumLiteral, ok := n.(*EnumLiteral); ok {
		v.Visit(enumLiteral.TupleLiteral)
		v.Visit(enumLiteral.StructLiteral)

	} else if blockStat, ok := n.(*BlockStat); ok {
		v.Visit(blockStat.Block)

	} else if callStat, ok := n.(*CallStat); ok {
		v.Visit(callStat.Call)

	} else if deferStat, ok := n.(*DeferStat); ok {
		v.Visit(deferStat.Call)

	} else if defaultStat, ok := n.(*DefaultStat); ok {
		v.Visit(defaultStat.Target)

	} else if addressOfExpr, ok := n.(*AddressOfExpr); ok {
		v.Visit(addressOfExpr.Access)

	} else if castExpr, ok := n.(*CastExpr); ok {
		v.Visit(castExpr.Expr)

	} else if unaryExpr, ok := n.(*UnaryExpr); ok {
		v.Visit(unaryExpr.Expr)

	} else if structAccessExpr, ok := n.(*StructAccessExpr); ok {
		v.Visit(structAccessExpr.Struct)

	} else if tupleAccessExpr, ok := n.(*TupleAccessExpr); ok {
		v.Visit(tupleAccessExpr.Tuple)

	} else if derefAccessExpr, ok := n.(*DerefAccessExpr); ok {
		v.Visit(derefAccessExpr.Expr)

	} else if functionDecl, ok := n.(*FunctionDecl); ok {
		// TODO: Scope?
		v.EnterScope(nil)
		fn := functionDecl.Function

		if fn.IsMethod && !fn.IsStatic {
			v.Visit(functionDecl.Function.Receiver)
		}

		for _, param := range fn.Parameters {
			v.Visit(param)
		}

		if !functionDecl.Prototype {
			v.Visit(fn.Body)
		}
		v.ExitScope(nil)

	} else if variableDecl, ok := n.(*VariableDecl); ok {
		v.Visit(variableDecl.Assignment)

	} else if _, ok := n.(*NumericLiteral); ok {
		// noop

	} else if _, ok := n.(*StringLiteral); ok {
		// noop

	} else if _, ok := n.(*BoolLiteral); ok {
		// noop

	} else if _, ok := n.(*RuneLiteral); ok {
		// noop

	} else if _, ok := n.(*VariableAccessExpr); ok {
		// noop

	} else if _, ok := n.(*TypeDecl); ok {
		// noop

	} else if _, ok := n.(*DefaultExpr); ok {
		// noop

	} else if _, ok := n.(*DefaultMatchBranch); ok {
		// noop

	} else {
		panic("Unhandled node: " + reflect.TypeOf(n).String())
	}
}

func exprsToNodes(exprs []Expr) []Node {
	nodes := make([]Node, len(exprs))
	for idx, expr := range exprs {
		nodes[idx] = expr
	}
	return nodes
}

func blocksToNodes(blocks []*Block) []Node {
	nodes := make([]Node, len(blocks))
	for idx, expr := range blocks {
		nodes[idx] = expr
	}
	return nodes
}

func isNil(a interface{}) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}
