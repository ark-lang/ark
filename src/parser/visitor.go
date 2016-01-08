package parser

import (
	"reflect"
)

type Visitor interface {
	EnterScope(s *Scope)
	ExitScope(s *Scope)

	Visit(*Node) bool
	PostVisit(*Node)
}

type ASTVisitor struct {
	Visitor Visitor
}

func NewASTVisitor(visitor Visitor) *ASTVisitor {
	return &ASTVisitor{Visitor: visitor}
}

func (v *ASTVisitor) VisitModule(module *Module) {
	v.EnterScope(module.GlobalScope)
	module.Nodes = v.VisitNodes(module.Nodes)
	v.ExitScope(module.GlobalScope)
}

func (v *ASTVisitor) VisitNodes(nodes []Node) []Node {
	res := make([]Node, len(nodes))
	for idx, node := range nodes {
		res[idx] = v.Visit(node)
	}
	return res
}

func (v *ASTVisitor) VisitExprs(exprs []Expr) []Expr {
	res := make([]Expr, len(exprs))
	for idx, expr := range exprs {
		res[idx] = v.VisitExpr(expr)
	}
	return res
}

func (v *ASTVisitor) VisitBlocks(blocks []*Block) []*Block {
	res := make([]*Block, len(blocks))
	for idx, block := range blocks {
		res[idx] = v.VisitBlock(block)
	}
	return res
}

func (v *ASTVisitor) Visit(n Node) Node {
	if isNil(n) {
		return nil
	}

	if v.Visitor.Visit(&n) {
		v.VisitChildren(n)
		v.Visitor.PostVisit(&n)
	}

	return n
}

func (v *ASTVisitor) VisitExpr(e Expr) Expr {
	if isNil(e) {
		return nil
	}

	n := e.(Node)
	if v.Visitor.Visit(&n) {
		v.VisitChildren(n)
		v.Visitor.PostVisit(&n)
	}

	if n != nil {
		return n.(Expr)
	} else {
		return nil
	}
}

func (v *ASTVisitor) VisitBlock(b *Block) *Block {
	if isNil(b) {
		return nil
	}

	n := Node(b)
	if v.Visitor.Visit(&n) {
		v.VisitChildren(n)
		v.Visitor.PostVisit(&n)
	}

	if n != nil {
		return n.(*Block)
	} else {
		return nil
	}
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

		block.Nodes = v.VisitNodes(block.Nodes)

		if !block.NonScoping {
			v.ExitScope(block.scope)
		}

	} else if returnStat, ok := n.(*ReturnStat); ok {
		returnStat.Value = v.VisitExpr(returnStat.Value)

	} else if ifStat, ok := n.(*IfStat); ok {
		ifStat.Exprs = v.VisitExprs(ifStat.Exprs)
		ifStat.Bodies = v.VisitBlocks(ifStat.Bodies)
		ifStat.Else = v.VisitBlock(ifStat.Else)

	} else if assignStat, ok := n.(*AssignStat); ok {
		assignStat.Assignment = v.VisitExpr(assignStat.Assignment)
		assignStat.Access = v.Visit(assignStat.Access).(AccessExpr)

	} else if binopAssignStat, ok := n.(*BinopAssignStat); ok {
		binopAssignStat.Assignment = v.VisitExpr(binopAssignStat.Assignment)
		binopAssignStat.Access = v.Visit(binopAssignStat.Access).(AccessExpr)

	} else if loopStat, ok := n.(*LoopStat); ok {
		loopStat.Body = v.Visit(loopStat.Body).(*Block)

		switch loopStat.LoopType {
		case LOOP_TYPE_INFINITE:
		case LOOP_TYPE_CONDITIONAL:
			loopStat.Condition = v.VisitExpr(loopStat.Condition)
		default:
			panic("invalid loop type")
		}

	} else if matchStat, ok := n.(*MatchStat); ok {
		matchStat.Target = v.VisitExpr(matchStat.Target)

		res := make(map[Expr]Node)
		for pattern, stmt := range matchStat.Branches {
			res[v.VisitExpr(pattern)] = v.Visit(stmt)
		}
		matchStat.Branches = res

	} else if binaryExpr, ok := n.(*BinaryExpr); ok {
		binaryExpr.Lhand = v.VisitExpr(binaryExpr.Lhand)
		binaryExpr.Rhand = v.VisitExpr(binaryExpr.Rhand)

	} else if arrayLiteral, ok := n.(*ArrayLiteral); ok {
		arrayLiteral.Members = v.VisitExprs(arrayLiteral.Members)

	} else if callExpr, ok := n.(*CallExpr); ok {
		switch callExpr.functionSource.(type) {
		case *StructAccessExpr:
			callExpr.functionSource.(*StructAccessExpr).Struct =
				v.Visit(callExpr.functionSource.(*StructAccessExpr).Struct).(AccessExpr)
		}

		// TODO: Visit callExpr.FunctionSource? once we get lambda/function types
		callExpr.Arguments = v.VisitExprs(callExpr.Arguments)
		callExpr.ReceiverAccess = v.VisitExpr(callExpr.ReceiverAccess)

	} else if arrayAccessExpr, ok := n.(*ArrayAccessExpr); ok {
		arrayAccessExpr.Array = v.Visit(arrayAccessExpr.Array).(AccessExpr)
		arrayAccessExpr.Subscript = v.VisitExpr(arrayAccessExpr.Subscript)

	} else if sizeofExpr, ok := n.(*SizeofExpr); ok {
		// TODO: Maybe visit sizeofExpr.Type at some point?
		sizeofExpr.Expr = v.VisitExpr(sizeofExpr.Expr)

	} else if arrayLenExpr, ok := n.(*ArrayLenExpr); ok {
		arrayLenExpr.Expr = v.VisitExpr(arrayLenExpr.Expr)

	} else if tupleLiteral, ok := n.(*TupleLiteral); ok {
		tupleLiteral.Members = v.VisitExprs(tupleLiteral.Members)

	} else if structLiteral, ok := n.(*StructLiteral); ok {
		for key, value := range structLiteral.Values {
			structLiteral.Values[key] = v.VisitExpr(value)
		}

	} else if enumLiteral, ok := n.(*EnumLiteral); ok {
		n1 := v.Visit(enumLiteral.TupleLiteral)
		if n1 == nil {
			enumLiteral.TupleLiteral = nil
		} else {
			enumLiteral.TupleLiteral = n1.(*TupleLiteral)
		}

		n2 := v.Visit(enumLiteral.StructLiteral)
		if n2 == nil {
			enumLiteral.StructLiteral = nil
		} else {
			enumLiteral.StructLiteral = n2.(*StructLiteral)
		}

	} else if blockStat, ok := n.(*BlockStat); ok {
		blockStat.Block = v.VisitBlock(blockStat.Block)

	} else if callStat, ok := n.(*CallStat); ok {
		callStat.Call = v.Visit(callStat.Call).(*CallExpr)

	} else if deferStat, ok := n.(*DeferStat); ok {
		deferStat.Call = v.Visit(deferStat.Call).(*CallExpr)

	} else if defaultStat, ok := n.(*DefaultStat); ok {
		defaultStat.Target = v.Visit(defaultStat.Target).(AccessExpr)

	} else if addressOfExpr, ok := n.(*AddressOfExpr); ok {
		addressOfExpr.Access = v.VisitExpr(addressOfExpr.Access)

	} else if castExpr, ok := n.(*CastExpr); ok {
		castExpr.Expr = v.VisitExpr(castExpr.Expr)

	} else if unaryExpr, ok := n.(*UnaryExpr); ok {
		unaryExpr.Expr = v.VisitExpr(unaryExpr.Expr)

	} else if structAccessExpr, ok := n.(*StructAccessExpr); ok {
		structAccessExpr.Struct = v.Visit(structAccessExpr.Struct).(AccessExpr)

	} else if tupleAccessExpr, ok := n.(*TupleAccessExpr); ok {
		tupleAccessExpr.Tuple = v.Visit(tupleAccessExpr.Tuple).(AccessExpr)

	} else if derefAccessExpr, ok := n.(*DerefAccessExpr); ok {
		derefAccessExpr.Expr = v.VisitExpr(derefAccessExpr.Expr)

	} else if functionDecl, ok := n.(*FunctionDecl); ok {
		// TODO: Scope?
		v.EnterScope(nil)
		fn := functionDecl.Function

		if fn.Type.Receiver != nil {
			functionDecl.Function.Receiver = v.Visit(functionDecl.Function.Receiver).(*VariableDecl)
		}

		for idx, param := range fn.Parameters {
			fn.Parameters[idx] = v.Visit(param).(*VariableDecl)
		}

		if !functionDecl.Prototype {
			fn.Body = v.Visit(fn.Body).(*Block)
		}
		v.ExitScope(nil)

	} else if variableDecl, ok := n.(*VariableDecl); ok {
		variableDecl.Assignment = v.VisitExpr(variableDecl.Assignment)

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

	} else if _, ok := n.(*UseDecl); ok {
		// noop

	} else {
		panic("Unhandled node: " + reflect.TypeOf(n).String())
	}
}

func isNil(a interface{}) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}
