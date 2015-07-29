package checks

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
)

type SemanticAnalyzer struct {
	Module          *parser.Module
	Function        *parser.Function // the function we're in, or nil if we aren't
	unresolvedNodes []*parser.Node
	modules         map[string]*parser.Module
	shouldExit      bool

	Checks []SemanticCheck
}

type SemanticCheck interface {
	EnterScope(s *SemanticAnalyzer)
	ExitScope(s *SemanticAnalyzer)
	Visit(*SemanticAnalyzer, parser.Node)
}

func (v *SemanticAnalyzer) Err(thing parser.Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Error("semantic", util.TEXT_RED+util.TEXT_BOLD+"Semantic error:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Errorln("semantic", v.Module.File.MarkPos(pos))

	v.shouldExit = true
}

func (v *SemanticAnalyzer) Warn(thing parser.Locatable, err string, stuff ...interface{}) {
	pos := thing.Pos()

	log.Warning("semantic", util.TEXT_YELLOW+util.TEXT_BOLD+"Semantic warning:"+util.TEXT_RESET+" [%s:%d:%d] %s\n",
		pos.Filename, pos.Line, pos.Char, fmt.Sprintf(err, stuff...))

	log.Warningln("semantic", v.Module.File.MarkPos(pos))
}

func (v *SemanticAnalyzer) Analyze(modules map[string]*parser.Module) {
	v.modules = modules
	v.shouldExit = false
	v.Checks = []SemanticCheck{
		&AttributeCheck{},
		&UnreachableCheck{},
		&DeprecatedCheck{},
		&RecursiveDefinitionCheck{},
		&TypeCheck{},
		&UnusedCheck{},
		&ImmutableAssignCheck{},
		&UseBeforeDeclareCheck{},
		&MiscCheck{},
	}

	v.EnterScope()
	v.VisitNodes(v.Module.Nodes)
	v.ExitScope()

	if v.shouldExit {
		os.Exit(util.EXIT_FAILURE_SEMANTIC)
	}
}

func (v *SemanticAnalyzer) VisitNodes(nodes []parser.Node) {
	for _, node := range nodes {
		v.Visit(node)
	}
}

func (v *SemanticAnalyzer) Visit(n parser.Node) {
	if isNil(n) {
		return
	}

	v.VisitChildren(n)
	for _, check := range v.Checks {
		check.Visit(v, n)
	}
}

func (v *SemanticAnalyzer) EnterScope() {
	for _, check := range v.Checks {
		check.EnterScope(v)
	}
}

func (v *SemanticAnalyzer) ExitScope() {
	for _, check := range v.Checks {
		check.ExitScope(v)
	}
}

func (v *SemanticAnalyzer) VisitChildren(n parser.Node) {
	if block, ok := n.(*parser.Block); ok {
		v.EnterScope()
		v.VisitNodes(block.Nodes)
		v.ExitScope()
	}

	if returnStat, ok := n.(*parser.ReturnStat); ok {
		v.Visit(returnStat.Value)
	}

	if ifStat, ok := n.(*parser.IfStat); ok {
		v.VisitNodes(exprsToNodes(ifStat.Exprs))
		v.VisitNodes(blocksToNodes(ifStat.Bodies))
		v.Visit(ifStat.Else)
	}

	if assignStat, ok := n.(*parser.AssignStat); ok {
		v.Visit(assignStat.Assignment)
		v.Visit(assignStat.Access)
	}

	if binopAssignStat, ok := n.(*parser.BinopAssignStat); ok {
		v.Visit(binopAssignStat.Assignment)
		v.Visit(binopAssignStat.Access)
	}

	if loopStat, ok := n.(*parser.LoopStat); ok {
		v.Visit(loopStat.Body)

		switch loopStat.LoopType {
		case parser.LOOP_TYPE_INFINITE:
		case parser.LOOP_TYPE_CONDITIONAL:
			v.Visit(loopStat.Condition)
		default:
			panic("invalid loop type")
		}
	}

	if matchStat, ok := n.(*parser.MatchStat); ok {
		v.Visit(matchStat.Target)

		for pattern, stmt := range matchStat.Branches {
			v.Visit(pattern)
			v.Visit(stmt)
		}
	}

	if binaryExpr, ok := n.(*parser.BinaryExpr); ok {
		v.Visit(binaryExpr.Lhand)
		v.Visit(binaryExpr.Rhand)
	}

	if arrayLiteral, ok := n.(*parser.ArrayLiteral); ok {
		v.VisitNodes(exprsToNodes(arrayLiteral.Members))
	}

	if callExpr, ok := n.(*parser.CallExpr); ok {
		// TODO: Visit callExpr.FunctionSource? once we get lambda/function types
		v.VisitNodes(exprsToNodes(callExpr.Arguments))
	}

	if arrayAccessExpr, ok := n.(*parser.ArrayAccessExpr); ok {
		v.Visit(arrayAccessExpr.Array)
		v.Visit(arrayAccessExpr.Subscript)
	}

	if sizeofExpr, ok := n.(*parser.SizeofExpr); ok {
		// TODO: Maybe visit sizeofExpr.Type at some point?
		v.Visit(sizeofExpr.Expr)
	}

	if tupleLiteral, ok := n.(*parser.TupleLiteral); ok {
		v.VisitNodes(exprsToNodes(tupleLiteral.Members))
	}

	if structLiteral, ok := n.(*parser.StructLiteral); ok {
		for _, value := range structLiteral.Values {
			v.Visit(value)
		}
	}

	if enumLiteral, ok := n.(*parser.EnumLiteral); ok {
		v.Visit(enumLiteral.TupleLiteral)
		v.Visit(enumLiteral.StructLiteral)
	}

	if structDecl, ok := n.(*parser.StructDecl); ok {
		v.EnterScope()
		for _, decl := range structDecl.Struct.Variables {
			v.Visit(decl)
		}
		v.ExitScope()
	}

	if blockStat, ok := n.(*parser.BlockStat); ok {
		v.Visit(blockStat.Block)
	}

	if callStat, ok := n.(*parser.CallStat); ok {
		v.Visit(callStat.Call)
	}

	if deferStat, ok := n.(*parser.DeferStat); ok {
		v.Visit(deferStat.Call)
	}

	if defaultStat, ok := n.(*parser.DefaultStat); ok {
		v.Visit(defaultStat.Target)
	}

	if addressOfExpr, ok := n.(*parser.AddressOfExpr); ok {
		v.Visit(addressOfExpr.Access)
	}

	if castExpr, ok := n.(*parser.CastExpr); ok {
		v.Visit(castExpr.Expr)
	}

	if unaryExpr, ok := n.(*parser.UnaryExpr); ok {
		v.Visit(unaryExpr.Expr)
	}

	if structAccessExpr, ok := n.(*parser.StructAccessExpr); ok {
		v.Visit(structAccessExpr.Struct)
	}

	if tupleAccessExpr, ok := n.(*parser.TupleAccessExpr); ok {
		v.Visit(tupleAccessExpr.Tuple)
	}

	if derefAccessExpr, ok := n.(*parser.DerefAccessExpr); ok {
		v.Visit(derefAccessExpr.Expr)
	}

	if traitDecl, ok := n.(*parser.TraitDecl); ok {
		v.EnterScope()
		for _, decl := range traitDecl.Trait.Functions {
			v.Visit(decl)
		}
		v.ExitScope()
	}

	if functionDecl, ok := n.(*parser.FunctionDecl); ok {
		v.EnterScope()
		for _, param := range functionDecl.Function.Parameters {
			v.Visit(param)
		}

		v.Function = functionDecl.Function
		v.Visit(functionDecl.Function.Body)
		v.Function = nil

		v.ExitScope()
	}

	if variableDecl, ok := n.(*parser.VariableDecl); ok {
		v.Visit(variableDecl.Assignment)
	}

	//panic("unhandled node: " + reflect.TypeOf(n).String())
}

func exprsToNodes(exprs []parser.Expr) []parser.Node {
	nodes := make([]parser.Node, len(exprs))
	for idx, expr := range exprs {
		nodes[idx] = expr
	}
	return nodes
}

func blocksToNodes(blocks []*parser.Block) []parser.Node {
	nodes := make([]parser.Node, len(blocks))
	for idx, expr := range blocks {
		nodes[idx] = expr
	}
	return nodes
}

func isNil(a interface{}) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}
