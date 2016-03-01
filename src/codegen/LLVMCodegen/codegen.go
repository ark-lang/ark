package LLVMCodegen

import (
	"fmt"
	"os"

	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/codegen"
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/semantic"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"

	"github.com/ark-lang/go-llvm/llvm"
)

// Assorted "constants" used in the codegen. Do not ever set them, only read them
var (
	enumTagType = llvm.IntType(32)
)

type functionAndFnGenericInstance struct {
	fn   *ast.Function
	gcon *ast.GenericContext // nil for no generics
}

func newfunctionAndFnGenericInstance(fn *ast.Function, gcon *ast.GenericContext) functionAndFnGenericInstance {
	return functionAndFnGenericInstance{
		fn:   fn,
		gcon: gcon,
	}
}

type variableAndFnGenericInstance struct {
	variable *ast.Variable
	gcon     *ast.GenericContext // nil for no generics
}

func newvariableAndFnGenericInstance(variable *ast.Variable, gcon *ast.GenericContext) variableAndFnGenericInstance {
	return variableAndFnGenericInstance{
		variable: variable,
		gcon:     gcon,
	}
}

type Codegen struct {
	// public options
	OutputName string
	OutputType codegen.OutputType
	LinkerArgs []string
	Linker     string // defaults to cc
	OptLevel   int

	// private stuff
	input   []*WrappedModule
	curFile *WrappedModule

	builders      map[functionAndFnGenericInstance]llvm.Builder      // map of functions to builders
	curLoopExits  map[functionAndFnGenericInstance][]llvm.BasicBlock // map of functions to slices of blocks, where each block is the exit block for current loops
	curLoopNexts  map[functionAndFnGenericInstance][]llvm.BasicBlock // map of functions to slices of blocks, where each block is the eval block for current loops
	curSegvBlocks map[functionAndFnGenericInstance]llvm.BasicBlock

	globalBuilder   llvm.Builder // used non-function stuff
	variableLookup  map[variableAndFnGenericInstance]llvm.Value
	namedTypeLookup map[string]llvm.Type

	declForFunction map[*ast.Function]*ast.FunctionDecl

	referenceAccess bool
	inFunctions     []functionAndFnGenericInstance

	lambdaID int

	inBlocks       map[functionAndFnGenericInstance][]*ast.Block
	blockDeferData map[*ast.Block][]*deferData // TODO make sure works with generics

	// size calculation stuff
	target        llvm.Target
	targetMachine llvm.TargetMachine
	targetData    llvm.TargetData
}

func (v *Codegen) builder() llvm.Builder {
	if !v.inFunction() {
		return v.globalBuilder
	}
	return v.builders[v.currentFunction()]
}

func (v *Codegen) pushFunction(fn functionAndFnGenericInstance) {
	v.inFunctions = append(v.inFunctions, fn)
}

func (v *Codegen) popFunction() {
	v.inFunctions = v.inFunctions[:len(v.inFunctions)-1]
}

func (v *Codegen) inFunction() bool {
	return len(v.inFunctions) > 0
}

func (v *Codegen) currentFunction() functionAndFnGenericInstance {
	return v.inFunctions[len(v.inFunctions)-1]
}

func (v *Codegen) currentLLVMFunction() llvm.Value {
	curFn := v.currentFunction()

	name := curFn.fn.MangledName(ast.MANGLE_ARK_UNSTABLE, curFn.gcon)
	if curFn.fn.Type.Attrs().Contains("nomangle") || curFn.fn.Anonymous {
		name = curFn.fn.Name
	}
	return v.curFile.LlvmModule.NamedFunction(name)
}

func (v *Codegen) pushBlock(block *ast.Block) {
	fn := v.currentFunction()
	v.inBlocks[fn] = append(v.inBlocks[fn], block)
}

func (v *Codegen) popBlock() {
	fn := v.currentFunction()
	v.inBlocks[fn] = v.inBlocks[fn][:len(v.inBlocks[fn])-1]
}

func (v *Codegen) currentBlock() *ast.Block {
	fn := v.currentFunction()
	return v.inBlocks[fn][len(v.inBlocks[fn])-1]
}

func (v *Codegen) nextLambdaID() int {
	id := v.lambdaID
	v.lambdaID++
	return id
}

type WrappedModule struct {
	*ast.Module
	LlvmModule llvm.Module
}

type deferData struct {
	stat *ast.DeferStat
	args []llvm.Value
}

func (v *Codegen) err(err string, stuff ...interface{}) {
	log.Error("codegen", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_CODEGEN)
}

func (v *Codegen) Generate(input []*ast.Module) {
	v.builders = make(map[functionAndFnGenericInstance]llvm.Builder)
	v.inBlocks = make(map[functionAndFnGenericInstance][]*ast.Block)
	v.globalBuilder = llvm.NewBuilder()
	defer v.globalBuilder.Dispose()

	v.curLoopExits = make(map[functionAndFnGenericInstance][]llvm.BasicBlock)
	v.curLoopNexts = make(map[functionAndFnGenericInstance][]llvm.BasicBlock)
	v.curSegvBlocks = make(map[functionAndFnGenericInstance]llvm.BasicBlock)

	v.declForFunction = make(map[*ast.Function]*ast.FunctionDecl)

	v.input = make([]*WrappedModule, len(input))
	for idx, mod := range input {
		v.input[idx] = &WrappedModule{Module: mod}
	}

	v.variableLookup = make(map[variableAndFnGenericInstance]llvm.Value)
	v.namedTypeLookup = make(map[string]llvm.Type)

	// initialize llvm target
	llvm.InitializeNativeTarget()
	llvm.InitializeNativeAsmPrinter()
	llvm.InitializeAllAsmParsers()

	// setup target stuff
	var err error
	v.target, err = llvm.GetTargetFromTriple(llvm.DefaultTargetTriple())
	if err != nil {
		panic(err)
	}
	v.targetMachine = v.target.CreateTargetMachine(llvm.DefaultTargetTriple(), "", "", llvm.CodeGenLevelNone, llvm.RelocDefault, llvm.CodeModelDefault)
	v.targetData = v.targetMachine.TargetData()

	passManager := llvm.NewPassManager()
	passBuilder := llvm.NewPassManagerBuilder()
	if v.OptLevel > 0 {
		passBuilder.SetOptLevel(v.OptLevel)
		passBuilder.Populate(passManager)
	}

	v.blockDeferData = make(map[*ast.Block][]*deferData)

	for _, infile := range v.input {
		log.Timed("codegenning", infile.Name.String(), func() {
			infile.LlvmModule = llvm.NewModule(infile.Name.String())
			v.curFile = infile

			for _, submod := range infile.Parts {
				v.declareDecls(submod.Nodes)

				for _, node := range submod.Nodes {
					v.genNode(node)
				}
			}

			if err := llvm.VerifyModule(infile.LlvmModule, llvm.ReturnStatusAction); err != nil {
				infile.LlvmModule.Dump()
				v.err("%s", err.Error())
			}

			passManager.Run(infile.LlvmModule)

			if log.AtLevel(log.LevelDebug) {
				infile.LlvmModule.Dump()
			}
		})
	}

	passManager.Dispose()

	log.Timed("creating binary", "", func() {
		v.createBinary()
	})

}

func (v *Codegen) recursiveGenericFunctionHelper(n *ast.FunctionDecl, access *ast.FunctionAccessExpr, gcon *ast.GenericContext, fn func(*ast.FunctionDecl, *ast.GenericContext)) {
	exit := true

	var checkgargs func(gargs []*ast.TypeReference)
	checkgargs = func(gargs []*ast.TypeReference) {
		for _, garg := range gargs {
			if _, ok := garg.BaseType.(*ast.SubstitutionType); ok {
				exit = false
				break
			}

			checkgargs(garg.GenericArguments)
		}
	}

	checkgargs(access.GenericArguments)

	if exit {
		fn(n, gcon)
		return
	}

	for _, subAccess := range access.ParentFunction.Accesses {
		newGcon := ast.NewGenericContext(subAccess.Function.Type.GenericParameters, subAccess.GenericArguments)
		newGcon.Outer = gcon

		v.recursiveGenericFunctionHelper(n, subAccess, newGcon, fn)
	}
}

func (v *Codegen) declareDecls(nodes []ast.Node) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.FunctionDecl:
			v.declForFunction[n.Function] = n
		}
	}

	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.FunctionDecl:
			if len(n.Function.Type.GenericParameters) == 0 {
				v.declareFunctionDecl(n, nil)
			} else {
				for _, access := range n.Function.Accesses {
					gcon := ast.NewGenericContext(access.Function.Type.GenericParameters, access.GenericArguments)

					v.recursiveGenericFunctionHelper(n, access, gcon, v.declareFunctionDecl)
				}
			}
		}
	}
}

var nonPublicLinkage = llvm.InternalLinkage

var callConvTypes = map[string]llvm.CallConv{
	"c":           llvm.CCallConv,
	"fast":        llvm.FastCallConv,
	"cold":        llvm.ColdCallConv,
	"x86stdcall":  llvm.X86StdcallCallConv,
	"x86fastcall": llvm.X86FastcallCallConv,
}

var inlineAttrType = map[string]llvm.Attribute{
	"always": llvm.AlwaysInlineAttribute,
	"never":  llvm.NoInlineAttribute,
	"maybe":  llvm.InlineHintAttribute,
}

func (v *Codegen) declareFunctionDecl(n *ast.FunctionDecl, gcon *ast.GenericContext) {
	mangledName := n.Function.MangledName(ast.MANGLE_ARK_UNSTABLE, gcon)
	if n.Function.Type.Attrs().Contains("nomangle") {
		mangledName = n.Function.Name
	}

	function := v.curFile.LlvmModule.NamedFunction(mangledName)
	if !function.IsNil() {
		// do nothing, only time this can happen is due to generics
	} else {
		// find them attributes yo
		attrs := n.Function.Type.Attrs()
		cBinding := attrs.Contains("c")

		// create the function type
		funcType := v.functionTypeToLLVMType(n.Function.Type, false, gcon)

		functionName := mangledName
		if cBinding {
			functionName = n.Function.Name
		}

		// add that shit
		function = llvm.AddFunction(v.curFile.LlvmModule, functionName, funcType)

		if !cBinding && !n.IsPublic() {
			function.SetLinkage(nonPublicLinkage)
		}

		if ccAttr := attrs.Get("call_conv"); ccAttr != nil {
			// TODO: move value checking to parser?
			if callConv, ok := callConvTypes[ccAttr.Value]; ok {
				function.SetFunctionCallConv(callConv)
			} else {
				v.err("undefined calling convention `%s` for function `%s` wanted", ccAttr.Value, n.Function.Name)
			}
		}

		if inlineAttr := attrs.Get("inline"); inlineAttr != nil {
			function.AddFunctionAttr(inlineAttrType[inlineAttr.Value])
		}

		/*// do some magical shit for later
		for i := 0; i < numOfParams; i++ {
			funcParam := function.Param(i)
			funcParam.SetName(n.Function.Parameters[i].Variable.MangledName(ast.MANGLE_ARK_UNSTABLE))
		}*/
	}
}

func (v *Codegen) getVariable(vari variableAndFnGenericInstance) llvm.Value {
	if value, ok := v.variableLookup[vari]; ok {
		return value
	}

	// Try with gcon == nil in case it's a global
	vari.gcon = nil
	if value, ok := v.variableLookup[vari]; ok {
		return value
	}

	if vari.variable.ParentModule != v.curFile.Module {
		value := llvm.AddGlobal(v.curFile.LlvmModule, v.typeRefToLLVMType(vari.variable.Type), vari.variable.MangledName(ast.MANGLE_ARK_UNSTABLE))
		value.SetLinkage(llvm.ExternalLinkage)
		v.variableLookup[vari] = value
		return value
	}

	panic("INTERNAL ERROR: Encountered undeclared variable: " + vari.variable.Name)
}

func (v *Codegen) genNode(n ast.Node) {
	switch n := n.(type) {
	case ast.Decl:
		v.genDecl(n)
	case ast.Expr:
		v.genExpr(n)
	case ast.Stat:
		v.genStat(n)
	case *ast.Block:
		v.genBlock(n)
	}
}

func (v *Codegen) genStat(n ast.Stat) {
	switch n := n.(type) {
	case *ast.ReturnStat:
		v.genReturnStat(n)
	case *ast.BreakStat:
		v.genBreakStat(n)
	case *ast.NextStat:
		v.genNextStat(n)
	case *ast.BlockStat:
		v.genBlockStat(n)
	case *ast.CallStat:
		v.genCallStat(n)
	case *ast.AssignStat:
		v.genAssignStat(n)
	case *ast.BinopAssignStat:
		v.genBinopAssignStat(n)
	case *ast.DestructAssignStat:
		v.genDestructAssignStat(n)
	case *ast.DestructBinopAssignStat:
		v.genDestructBinopAssignStat(n)
	case *ast.IfStat:
		v.genIfStat(n)
	case *ast.LoopStat:
		v.genLoopStat(n)
	case *ast.MatchStat:
		v.genMatchStat(n)
	case *ast.DeferStat:
		v.genDeferStat(n)
	default:
		panic("unimplemented stat")
	}
}

func (v *Codegen) genBreakStat(n *ast.BreakStat) {
	curExits := v.curLoopExits[v.currentFunction()]
	v.builder().CreateBr(curExits[len(curExits)-1])
}

func (v *Codegen) genNextStat(n *ast.NextStat) {
	curNexts := v.curLoopNexts[v.currentFunction()]
	v.builder().CreateBr(curNexts[len(curNexts)-1])
}

func (v *Codegen) genDeferStat(n *ast.DeferStat) {
	data := &deferData{
		stat: n,
		args: v.genCallArgs(n.Call),
	}

	v.blockDeferData[v.currentBlock()] = append(v.blockDeferData[v.currentBlock()], data)
}

func (v *Codegen) genRunDefers(block *ast.Block) {
	deferDat := v.blockDeferData[block]

	if len(deferDat) > 0 {
		for i := len(deferDat) - 1; i >= 0; i-- {
			v.genCallExprWithArgs(deferDat[i].stat.Call, deferDat[i].args)
		}
	}
}

func (v *Codegen) genBlock(n *ast.Block) {
	v.pushBlock(n)
	for i, x := range n.Nodes {
		v.genNode(x)

		if i == len(n.Nodes)-1 && !n.IsTerminating {
			v.genRunDefers(n)
		}
	}

	delete(v.blockDeferData, n)
	v.popBlock()
}

func (v *Codegen) genReturnStat(n *ast.ReturnStat) {
	var ret llvm.Value
	if n.Value != nil {
		ret = v.genExprAndLoadIfNeccesary(n.Value)
	}

	for i := len(v.inBlocks[v.currentFunction()]) - 1; i >= 0; i-- {
		v.genRunDefers(v.inBlocks[v.currentFunction()][i])
	}

	if n.Value == nil {
		v.builder().CreateRetVoid()
	} else {
		v.builder().CreateRet(ret)
	}
}

func (v *Codegen) genBlockStat(n *ast.BlockStat) {
	v.genBlock(n.Block)
}

func (v *Codegen) genCallStat(n *ast.CallStat) {
	v.genExpr(n.Call)
}

func (v *Codegen) genAssignStat(n *ast.AssignStat) {
	v.genAssign(n.Access, v.genExprAndLoadIfNeccesary(n.Assignment))
}

func (v *Codegen) genBinopAssignStat(n *ast.BinopAssignStat) {
	v.genBinopAssign(n.Operator, n.Access, v.genExprAndLoadIfNeccesary(n.Assignment), n.Assignment.GetType())
}

func (v *Codegen) genDestructAssignStat(n *ast.DestructAssignStat) {
	assignment := v.genExprAndLoadIfNeccesary(n.Assignment)
	for idx, acc := range n.Accesses {
		value := v.builder().CreateExtractValue(assignment, idx, "")
		v.genAssign(acc, value)
	}
}

func (v *Codegen) genDestructBinopAssignStat(n *ast.DestructBinopAssignStat) {
	assignment := v.genExprAndLoadIfNeccesary(n.Assignment)
	tt, _ := n.Assignment.GetType().BaseType.ActualType().(ast.TupleType)
	for idx, acc := range n.Accesses {
		value := v.builder().CreateExtractValue(assignment, idx, "")
		v.genBinopAssign(n.Operator, acc, value, tt.Members[idx])
	}
}

func (v *Codegen) genAssign(acc ast.AccessExpr, value llvm.Value) {
	if _, isDiscard := acc.(*ast.DiscardAccessExpr); isDiscard {
		return
	}

	access := v.genAccessGEP(acc)
	v.builder().CreateStore(value, access)
}

func (v *Codegen) genBinopAssign(op parser.BinOpType, acc ast.AccessExpr, value llvm.Value, valueType *ast.TypeReference) {
	if _, isDiscard := acc.(*ast.DiscardAccessExpr); isDiscard {
		return
	}

	storage := v.genAccessGEP(acc)
	storageValue := v.builder().CreateLoad(storage, "")

	result := v.genBinop(op, acc.GetType(), acc.GetType(), valueType, storageValue, value)
	v.builder().CreateStore(result, storage)
}

func isBreakOrNext(n ast.Node) bool {
	switch n.(type) {
	case *ast.BreakStat, *ast.NextStat:
		return true
	}
	return false
}

func (v *Codegen) genIfStat(n *ast.IfStat) {
	// Warning to all who tread here:
	// This function is complicated, but theoretically it should never need to
	// be changed again. God help the soul who has to edit this.

	if !v.inFunction() {
		panic("tried to gen if stat not in function")
	}

	statTerm := semantic.IsNodeTerminating(n)

	var end llvm.BasicBlock
	if !statTerm {
		end = llvm.AddBasicBlock(v.currentLLVMFunction(), "end")
	}

	for i, expr := range n.Exprs {
		cond := v.genExprAndLoadIfNeccesary(expr)

		ifTrue := llvm.AddBasicBlock(v.currentLLVMFunction(), "if_true")
		ifFalse := llvm.AddBasicBlock(v.currentLLVMFunction(), "if_false")

		v.builder().CreateCondBr(cond, ifTrue, ifFalse)

		v.builder().SetInsertPointAtEnd(ifTrue)
		v.genBlock(n.Bodies[i])

		if !statTerm && !n.Bodies[i].IsTerminating && !isBreakOrNext(n.Bodies[i].LastNode()) {
			v.builder().CreateBr(end)
		}

		v.builder().SetInsertPointAtEnd(ifFalse)

		if !statTerm {
			end.MoveAfter(ifFalse)
		}
	}

	if n.Else != nil {
		v.genBlock(n.Else)
	}

	if !statTerm && (n.Else == nil || (!n.Else.IsTerminating && !isBreakOrNext(n.Else.LastNode()))) {
		v.builder().CreateBr(end)
	}

	if !statTerm {
		v.builder().SetInsertPointAtEnd(end)
	}
}

func (v *Codegen) genLoopStat(n *ast.LoopStat) {
	curfn := v.currentFunction()
	afterBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "loop_exit")
	v.curLoopExits[curfn] = append(v.curLoopExits[curfn], afterBlock)

	switch n.LoopType {
	case ast.LOOP_TYPE_INFINITE:
		loopBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "loop_body")
		v.curLoopNexts[curfn] = append(v.curLoopNexts[curfn], loopBlock)
		v.builder().CreateBr(loopBlock)
		v.builder().SetInsertPointAtEnd(loopBlock)

		v.genBlock(n.Body)

		if !isBreakOrNext(n.Body.LastNode()) {
			v.builder().CreateBr(loopBlock)
		}

		v.builder().SetInsertPointAtEnd(afterBlock)
	case ast.LOOP_TYPE_CONDITIONAL:
		evalBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "loop_condeval")
		v.builder().CreateBr(evalBlock)
		v.curLoopNexts[curfn] = append(v.curLoopNexts[curfn], evalBlock)

		loopBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "loop_body")

		v.builder().SetInsertPointAtEnd(evalBlock)
		cond := v.genExprAndLoadIfNeccesary(n.Condition)
		v.builder().CreateCondBr(cond, loopBlock, afterBlock)

		v.builder().SetInsertPointAtEnd(loopBlock)
		v.genBlock(n.Body)

		if !isBreakOrNext(n.Body.LastNode()) {
			v.builder().CreateBr(evalBlock)
		}

		v.builder().SetInsertPointAtEnd(afterBlock)
	default:
		panic("invalid loop type")
	}

	v.curLoopExits[curfn] = v.curLoopExits[curfn][:len(v.curLoopExits[curfn])-1]
	v.curLoopNexts[curfn] = v.curLoopNexts[curfn][:len(v.curLoopNexts[curfn])-1]
}

func (v *Codegen) genMatchStat(n *ast.MatchStat) {
	// TODO: implement integral and string versions

	targetType := n.Target.GetType()
	switch targetType.BaseType.ActualType().(type) {
	case ast.EnumType:
		v.genEnumMatchStat(n)
	}
}

func (v *Codegen) genEnumMatchStat(n *ast.MatchStat) {
	et, ok := n.Target.GetType().BaseType.ActualType().(ast.EnumType)
	if !ok {
		panic("INTERNAL ERROR: Arrived in genEnumMatchStat with non enum type")
	}

	target := v.genExpr(n.Target)
	tag := v.genLoadIfNeccesary(n.Target, target)
	if !et.Simple {
		tag = v.builder().CreateExtractValue(tag, 0, "")
	}

	enterBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "match_enter")
	exitBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "match_exit")

	v.builder().CreateBr(enterBlock)

	var tags []int
	var blocks []llvm.BasicBlock
	var defaultBlock llvm.BasicBlock

	// TODO: Branch gen order is non-deterministic. We probably do not want that
	idx := 0
	for expr, branch := range n.Branches {
		var block llvm.BasicBlock
		if patt, ok := expr.(*ast.EnumPatternExpr); ok {
			mem, ok := et.GetMember(patt.MemberName.Name)
			if !ok {
				panic("INTERNAL ERROR: Enum match branch member was non existant")
			}

			block = llvm.AddBasicBlock(v.currentLLVMFunction(), "match_branch_"+mem.Name)

			tags = append(tags, mem.Tag)
			blocks = append(blocks, block)
		} else if _, ok := expr.(*ast.DiscardAccessExpr); ok {
			block = llvm.AddBasicBlock(v.currentLLVMFunction(), "match_branch_default")
			defaultBlock = block
		} else {
			panic("INTERNAL ERROR: Branch in enum match was not enum pattern or discard")
		}

		v.builder().SetInsertPointAtEnd(block)

		// Destructure the variables
		if patt, ok := expr.(*ast.EnumPatternExpr); ok && !et.Simple {
			memIdx := et.MemberIndex(patt.MemberName.Name)
			if memIdx == -1 {
				panic("INTERNAL ERROR: Enum match branch member was non existant")
			}

			gcon := ast.NewGenericContextFromTypeReference(n.Target.GetType())
			gcon.Outer = v.currentFunction().gcon
			value := v.genEnumUnionValue(target, et, memIdx, gcon)
			for idx, vari := range patt.Variables {
				if vari != nil {
					assign := v.builder().CreateExtractValue(value, idx, "")
					v.genVariable(false, vari, assign)
				}
			}
		}

		v.genNode(branch)

		if !semantic.IsNodeTerminating(branch) {
			v.builder().CreateBr(exitBlock)
		}

		exitBlock.MoveAfter(block)

		idx++
	}

	v.builder().SetInsertPointAtEnd(enterBlock)

	var sw llvm.Value
	if defaultBlock.IsNil() {
		sw = v.builder().CreateSwitch(tag, exitBlock, len(n.Branches))
	} else {
		sw = v.builder().CreateSwitch(tag, defaultBlock, len(n.Branches))
	}

	for idx := 0; idx < len(tags); idx++ {
		sw.AddCase(llvm.ConstInt(enumTagType, uint64(tags[idx]), false), blocks[idx])
	}

	v.builder().SetInsertPointAtEnd(exitBlock)
}

func (v *Codegen) genEnumUnionValue(enum llvm.Value, enumType ast.EnumType, memIdx int, gcon *ast.GenericContext) llvm.Value {
	enumTypeForMember := llvm.PointerType(v.llvmEnumTypeForMember(enumType, memIdx, gcon), 0)
	pointer := v.builder().CreateBitCast(enum, enumTypeForMember, "")
	value := v.builder().CreateLoad(pointer, "")
	return v.builder().CreateExtractValue(value, 1, "")
}

func (v *Codegen) genDecl(n ast.Decl) {
	switch n := n.(type) {
	case *ast.FunctionDecl:
		if len(n.Function.Type.GenericParameters) == 0 {
			v.genFunctionDecl(n, nil)
		} else {
			for _, access := range n.Function.Accesses {
				gcon := ast.NewGenericContext(access.Function.Type.GenericParameters, access.GenericArguments)

				v.recursiveGenericFunctionHelper(n, access, gcon, v.genFunctionDecl)
			}
		}
	case *ast.VariableDecl:
		v.genVariableDecl(n)
	case *ast.DestructVarDecl:
		v.genDestructVarDecl(n)
	case *ast.TypeDecl:
		// TODO nothing to gen?
	default:
		v.err("unimplemented decl found: `%s`", n.NodeName())
	}
}

func (v *Codegen) genFunctionDecl(n *ast.FunctionDecl, gcon *ast.GenericContext) {
	mangledName := n.Function.MangledName(ast.MANGLE_ARK_UNSTABLE, gcon)
	if n.Function.Type.Attrs().Contains("nomangle") {
		mangledName = n.Function.Name
	}

	function := v.curFile.LlvmModule.NamedFunction(mangledName)
	if function.IsNil() {
		//v.err("genning function `%s` doesn't exist in module", n.Function.Name)
		// hmmmm seems we just ignore this here
	} else {
		if !n.Prototype {
			if function.BasicBlocksCount() == 0 {
				v.genFunctionBody(n.Function, function, gcon)
			}
		}
	}
}

func (v *Codegen) genFunctionBody(fn *ast.Function, llvmFn llvm.Value, gcon *ast.GenericContext) {
	block := llvm.AddBasicBlock(llvmFn, "entry")

	v.pushFunction(newfunctionAndFnGenericInstance(fn, gcon))
	v.builders[v.currentFunction()] = llvm.NewBuilder()
	v.builder().SetInsertPointAtEnd(block)

	pars := fn.Parameters

	if fn.Type.Receiver != nil {
		newPars := make([]*ast.VariableDecl, len(pars)+1)
		newPars[0] = fn.Receiver
		copy(newPars[1:], pars)
		pars = newPars
	}

	for i, par := range pars {
		v.genVariable(false, par.Variable, llvmFn.Params()[i])
	}

	v.genBlock(fn.Body)
	v.builder().Dispose()
	delete(v.builders, v.currentFunction())
	delete(v.curLoopExits, v.currentFunction())
	delete(v.curLoopNexts, v.currentFunction())
	delete(v.curSegvBlocks, v.currentFunction())
	v.popFunction()
}

func (v *Codegen) genVariableDecl(n *ast.VariableDecl) {
	var value llvm.Value
	if n.Assignment != nil {
		value = v.genExprAndLoadIfNeccesary(n.Assignment)
	}
	v.genVariable(n.IsPublic(), n.Variable, value)
}

func (v *Codegen) genDestructVarDecl(n *ast.DestructVarDecl) {
	assignment := v.genExprAndLoadIfNeccesary(n.Assignment)

	for idx, vari := range n.Variables {
		if !n.ShouldDiscard[idx] {
			var value llvm.Value
			if v.inFunction() {
				value = v.builder().CreateExtractValue(assignment, idx, "")
			} else {
				value = llvm.ConstExtractValue(assignment, []uint32{uint32(idx)})
			}
			v.genVariable(n.IsPublic(), vari, value)
		}
	}
}

func (v *Codegen) genVariable(isPublic bool, vari *ast.Variable, assignment llvm.Value) {
	mangledName := vari.MangledName(ast.MANGLE_ARK_UNSTABLE)

	var varType llvm.Type
	if !assignment.IsNil() {
		varType = assignment.Type()
	} else {
		varType = v.typeRefToLLVMType(vari.Type)
	}

	if assignment.IsNil() && !vari.Attrs.Contains("nozero") {
		assignment = llvm.ConstNull(varType)
	}

	if v.inFunction() {
		alloc := v.createAlignedAlloca(varType, mangledName)
		v.variableLookup[newvariableAndFnGenericInstance(vari, v.currentFunction().gcon)] = alloc

		if !assignment.IsNil() {
			v.builder().CreateStore(assignment, alloc)
		}
	} else {
		// TODO cbindings
		cBinding := false

		value := llvm.AddGlobal(v.curFile.LlvmModule, varType, mangledName)
		v.variableLookup[newvariableAndFnGenericInstance(vari, nil)] = value

		if !cBinding && !isPublic {
			value.SetLinkage(nonPublicLinkage)
		}

		if !assignment.IsNil() {
			value.SetInitializer(assignment)
		}

		value.SetGlobalConstant(!vari.Mutable)
	}
}

func (v *Codegen) createAlignedAlloca(typ llvm.Type, name string) llvm.Value {
	funcEntry := v.currentLLVMFunction().EntryBasicBlock()

	// use this builder() for the variable alloca
	// this means all allocas go at the start of the function
	// so each variable is only allocated once
	allocBuilder := llvm.NewBuilder()
	defer allocBuilder.Dispose()

	allocBuilder.SetInsertPoint(funcEntry, funcEntry.FirstInstruction())

	align := v.targetData.ABITypeAlignment(typ)
	alloc := allocBuilder.CreateAlloca(typ, name)
	alloc.SetAlignment(align)
	return alloc
}

func (v *Codegen) genExpr(n ast.Expr) llvm.Value {
	switch n := n.(type) {
	case *ast.RuneLiteral:
		return v.genRuneLiteral(n)
	case *ast.NumericLiteral:
		return v.genNumericLiteral(n)
	case *ast.StringLiteral:
		return v.genStringLiteral(n)
	case *ast.BoolLiteral:
		return v.genBoolLiteral(n)
	case *ast.TupleLiteral:
		return v.genTupleLiteral(n)
	case *ast.CompositeLiteral:
		return v.genCompositeLiteral(n)
	case *ast.EnumLiteral:
		return v.genEnumLiteral(n)
	}

	if !v.inFunction() {
		v.err("[%s:%d:%d] Non-literal expressions in global scope are not currently supported",
			n.Pos().Filename, n.Pos().Line, n.Pos().Char)
	}

	switch n := n.(type) {
	case *ast.ReferenceToExpr:
		return v.genReferenceToExpr(n)
	case *ast.PointerToExpr:
		return v.genPointerToExpr(n)
	case *ast.BinaryExpr:
		return v.genBinaryExpr(n)
	case *ast.UnaryExpr:
		return v.genUnaryExpr(n)
	case *ast.CastExpr:
		return v.genCastExpr(n)
	case *ast.CallExpr:
		return v.genCallExpr(n)
	case *ast.VariableAccessExpr, *ast.StructAccessExpr,
		*ast.ArrayAccessExpr, *ast.DerefAccessExpr,
		*ast.FunctionAccessExpr:
		return v.genAccessExpr(n)
	case *ast.SizeofExpr:
		return v.genSizeofExpr(n)
	case *ast.ArrayLenExpr:
		return v.genArrayLenExpr(n)
	case *ast.LambdaExpr:
		return v.genLambdaExpr(n)
	default:
		log.Debug("codegen", "expr: %s\n", n)
		panic("unimplemented expr")
	}
}

func (v *Codegen) genLambdaExpr(n *ast.LambdaExpr) llvm.Value {
	typ := v.functionTypeToLLVMType(n.Function.Type, false, nil)
	mod := v.curFile.LlvmModule
	n.Function.Name = fmt.Sprintf("_lambda%d", v.nextLambdaID())
	fn := llvm.AddFunction(mod, n.Function.Name, typ)

	if len(n.Function.Type.GenericParameters) > 0 {
		panic("generic lambdas unimplemented")
	}

	v.genFunctionBody(n.Function, fn, nil)

	return fn
}

func (v *Codegen) genReferenceToExpr(n *ast.ReferenceToExpr) llvm.Value {
	return v.genAccessExpr(n.Access)
}

func (v *Codegen) genPointerToExpr(n *ast.PointerToExpr) llvm.Value {
	return v.genAccessExpr(n.Access)
}

func (v *Codegen) genLoadIfNeccesary(n ast.Expr, val llvm.Value) llvm.Value {
	if el, isEnumLit := n.(*ast.EnumLiteral); isEnumLit {
		et := el.Type.BaseType.ActualType().(ast.EnumType)
		isEnumLit = !et.Simple && v.inFunction()
		if !et.Simple && v.inFunction() {
			return v.builder().CreateLoad(val, "")
		}
	}

	if _, isAccess := n.(ast.AccessExpr); isAccess {
		return v.builder().CreateLoad(val, "")
	}
	return val
}

func (v *Codegen) genExprAndLoadIfNeccesary(n ast.Expr) llvm.Value {
	return v.genLoadIfNeccesary(n, v.genExpr(n))
}

func (v *Codegen) genAccessExpr(n ast.Expr) llvm.Value {
	if fae, ok := n.(*ast.FunctionAccessExpr); ok {
		gcon := ast.NewGenericContext(fae.Function.Type.GenericParameters, fae.GenericArguments)
		gcon.Outer = v.currentFunction().gcon

		var fnName string

		if fae.ReceiverAccess != nil {
			fnName = ast.GetMethod(gcon.Get(fae.ReceiverAccess.GetType()).BaseType, fae.Function.Name).MangledName(ast.MANGLE_ARK_UNSTABLE, gcon)
		} else {
			fnName = fae.Function.MangledName(ast.MANGLE_ARK_UNSTABLE, gcon)
		}

		if fae.Function.Type.Attrs().Contains("nomangle") {
			fnName = fae.Function.Name
		}

		cBinding := false
		if fae.Function.Type.Attrs() != nil {
			cBinding = fae.Function.Type.Attrs().Contains("c")
		}
		if cBinding {
			fnName = fae.Function.Name
		}

		fn := v.curFile.LlvmModule.NamedFunction(fnName)

		if fn.IsNil() {
			decl := &ast.FunctionDecl{Function: fae.Function, Prototype: true}
			decl.SetPublic(true)
			v.declareFunctionDecl(decl, gcon)

			if v.curFile.LlvmModule.NamedFunction(fnName).IsNil() {
				panic("how did this happen")
			}
			fn = v.curFile.LlvmModule.NamedFunction(fnName)
		}

		return fn
	}

	access := v.genAccessGEP(n)

	// To be able to deal with enum unions, the llvm type is not always the
	// actual type, so we bitcast the access gep to the right type.
	accessLlvmType := v.typeRefToLLVMType(n.GetType())
	return v.builder().CreateBitCast(access, llvm.PointerType(accessLlvmType, 0), "")
}

func (v *Codegen) genAccessGEP(n ast.Expr) llvm.Value {
	var curFngcon *ast.GenericContext
	if v.inFunction() {
		curFngcon = v.currentFunction().gcon
	}

	switch access := n.(type) {
	case *ast.VariableAccessExpr:
		vari := v.getVariable(newvariableAndFnGenericInstance(access.Variable, curFngcon))
		if vari.IsNil() {
			panic("vari was nil")
		}

		return vari

	case *ast.StructAccessExpr:
		gep := v.genAccessGEP(access.Struct)

		typ := access.Struct.GetType().BaseType.ActualType()
		index := typ.(ast.StructType).MemberIndex(access.Member)

		return v.builder().CreateStructGEP(gep, index, "")

	case *ast.ArrayAccessExpr:
		gep := v.genAccessGEP(access.Array)

		subscriptExpr := v.genExprAndLoadIfNeccesary(access.Subscript)
		subscriptTyp := access.Subscript.GetType().BaseType.ActualType().(ast.PrimitiveType)
		// Extend access width to system poiner width
		if !subscriptTyp.IsSigned() {
			subscriptExpr = v.builder().CreateZExt(subscriptExpr, v.targetData.IntPtrType(), "")
		} else {
			subscriptExpr = v.builder().CreateSExt(subscriptExpr, v.targetData.IntPtrType(), "")
		}

		if arrType, ok := access.Array.GetType().BaseType.ActualType().(ast.ArrayType); ok {
			if arrType.IsFixedLength {
				v.genBoundsCheck(llvm.ConstInt(v.primitiveTypeToLLVMType(ast.PRIMITIVE_uint), uint64(arrType.Length), false),
					subscriptExpr, access.Subscript.GetType().BaseType.IsSigned())

				return v.builder().CreateGEP(gep, []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false), subscriptExpr}, "")
			} else {
				v.genBoundsCheck(v.builder().CreateLoad(v.builder().CreateStructGEP(gep, 0, ""), ""),
					subscriptExpr, access.Subscript.GetType().BaseType.IsSigned())

				gep = v.builder().CreateStructGEP(gep, 1, "")

				load := v.builder().CreateLoad(gep, "")

				gepIndexes := []llvm.Value{subscriptExpr}
				return v.builder().CreateGEP(load, gepIndexes, "")
			}
		} else if _, ok := access.Array.GetType().BaseType.ActualType().(ast.PointerType); ok {
			gepIndexes := []llvm.Value{subscriptExpr}
			gep = v.genLoadIfNeccesary(access.Array, gep)
			return v.builder().CreateGEP(gep, gepIndexes, "")
		} else {
			panic("INTERNAL ERROR: ArrayAccessExpr type was not Array or Pointer")
		}

	case *ast.DerefAccessExpr:
		return v.genExprAndLoadIfNeccesary(access.Expr)

	default:
		panic("unhandled access type")
	}
}

func (v *Codegen) genBoundsCheck(limit llvm.Value, index llvm.Value, indexIsSigned bool) {
	var segvBlock llvm.BasicBlock
	needToSetupSegvBlock := false
	if b, ok := v.curSegvBlocks[v.currentFunction()]; ok {
		segvBlock = b
	} else {
		segvBlock = llvm.AddBasicBlock(v.currentLLVMFunction(), "boundscheck_segv")
		v.curSegvBlocks[v.currentFunction()] = segvBlock
		needToSetupSegvBlock = true
	}

	endBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "boundscheck_end")
	upperCheckBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "boundscheck_upper_block")

	tooLow := v.builder().CreateICmp(llvm.IntSGT, llvm.ConstInt(index.Type(), 0, false), index, "boundscheck_lower")
	v.builder().CreateCondBr(tooLow, segvBlock, upperCheckBlock)

	v.builder().SetInsertPointAtEnd(upperCheckBlock)

	// make sure limit and index have same width
	castedLimit := limit
	castedIndex := index
	if index.Type().IntTypeWidth() < limit.Type().IntTypeWidth() {
		if indexIsSigned {
			castedIndex = v.builder().CreateSExt(index, limit.Type(), "")
		} else {
			castedIndex = v.builder().CreateZExt(index, limit.Type(), "")
		}
	} else if index.Type().IntTypeWidth() > limit.Type().IntTypeWidth() {
		castedLimit = v.builder().CreateZExt(limit, index.Type(), "")
	}

	tooHigh := v.builder().CreateICmp(llvm.IntSLE, castedLimit, castedIndex, "boundscheck_upper")
	v.builder().CreateCondBr(tooHigh, segvBlock, endBlock)

	if needToSetupSegvBlock {
		v.builder().SetInsertPointAtEnd(segvBlock)
		v.genRaiseSegfault()
		v.builder().CreateUnreachable()
	}

	v.builder().SetInsertPointAtEnd(endBlock)
}

func (v *Codegen) genRaiseSegfault() {
	fn := v.curFile.LlvmModule.NamedFunction("raise")
	intType := v.primitiveTypeToLLVMType(ast.PRIMITIVE_int)

	if fn.IsNil() {
		fnType := llvm.FunctionType(intType, []llvm.Type{intType}, false)
		fn = llvm.AddFunction(v.curFile.LlvmModule, "raise", fnType)
	}

	v.builder().CreateCall(fn, []llvm.Value{llvm.ConstInt(intType, 11, false)}, "segfault")
}

func (v *Codegen) genBoolLiteral(n *ast.BoolLiteral) llvm.Value {
	var num uint64

	if n.Value {
		num = 1
	}

	return llvm.ConstInt(v.typeRefToLLVMType(n.GetType()), num, true)
}

func (v *Codegen) genRuneLiteral(n *ast.RuneLiteral) llvm.Value {
	return llvm.ConstInt(v.typeRefToLLVMType(n.GetType()), uint64(n.Value), true)
}

func (v *Codegen) genStringLiteral(n *ast.StringLiteral) llvm.Value {
	memberLLVMType := v.primitiveTypeToLLVMType(ast.PRIMITIVE_u8)
	length := len(n.Value)
	if n.IsCString {
		length++
	}

	var backingArrayPointer llvm.Value
	if v.inFunction() {
		// allocate backing array
		globString := v.builder().CreateGlobalStringPtr(n.Value, ".str")
		backingArray := v.createAlignedAlloca(llvm.ArrayType(memberLLVMType, length), ".stackstr")
		backingArrayPointer = v.builder().CreateBitCast(backingArray, llvm.PointerType(memberLLVMType, 0), "")
		v.genMemcpy(globString, backingArrayPointer, llvm.ConstInt(llvm.IntType(32), uint64(length), false))
	} else {
		backingArray := llvm.AddGlobal(v.curFile.LlvmModule, llvm.ArrayType(memberLLVMType, length), ".str")
		backingArray.SetLinkage(llvm.InternalLinkage)
		backingArray.SetGlobalConstant(false)
		backingArray.SetInitializer(llvm.ConstString(n.Value, n.IsCString))

		backingArrayPointer = llvm.ConstBitCast(backingArray, llvm.PointerType(memberLLVMType, 0))
	}

	if n.GetType().BaseType.ActualType().Equals(ast.ArrayOf(&ast.TypeReference{BaseType: ast.PRIMITIVE_u8}, false, 0)) {
		lengthValue := llvm.ConstInt(v.primitiveTypeToLLVMType(ast.PRIMITIVE_uint), uint64(length), false)
		structValue := llvm.Undef(v.typeRefToLLVMType(n.GetType()))
		structValue = v.builder().CreateInsertValue(structValue, lengthValue, 0, "")
		structValue = v.builder().CreateInsertValue(structValue, backingArrayPointer, 1, "")
		return structValue
	} else {
		return backingArrayPointer
	}
}

func (v *Codegen) genMemcpy(src llvm.Value, dst llvm.Value, length llvm.Value) {
	memcpyFn := v.curFile.LlvmModule.NamedFunction("llvm.memcpy.p0i8.p0i8.i32")
	if memcpyFn.IsNil() {
		//i8* <dest>, i8* <src>, i32 <len>, i32 <align>, i1 <isvolatile>
		fnType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{
			llvm.PointerType(llvm.IntType(8), 0), // dest
			llvm.PointerType(llvm.IntType(8), 0), // src
			llvm.IntType(32),                     // len
			llvm.IntType(32),                     // align
			llvm.IntType(1),                      // isvolatile
		}, false)
		memcpyFn = llvm.AddFunction(v.curFile.LlvmModule, "llvm.memcpy.p0i8.p0i8.i32", fnType)
	}
	v.builder().CreateCall(memcpyFn, []llvm.Value{
		dst, src, length,
		llvm.ConstInt(llvm.IntType(32), 1, false),
		llvm.ConstInt(llvm.IntType(1), 1, false),
	}, "")
}

func (v *Codegen) genCompositeLiteral(n *ast.CompositeLiteral) llvm.Value {
	switch n.GetType().BaseType.ActualType().(type) {
	case ast.ArrayType:
		return v.genArrayLiteral(n)
	case ast.StructType:
		return v.genStructLiteral(n)
	default:
		panic("invalid composite literal type")
	}
}

// Allocates a literal array on the stack
func (v *Codegen) genArrayLiteral(n *ast.CompositeLiteral) llvm.Value {
	arrayType := n.Type.BaseType.ActualType().(ast.ArrayType)
	arrayLLVMType := v.typeRefToLLVMType(n.Type)
	memberLLVMType := v.typeRefToLLVMType(n.Type.BaseType.ActualType().(ast.ArrayType).MemberType) // TODO works with generics?

	arrayValues := make([]llvm.Value, len(n.Values))
	for idx, mem := range n.Values {
		value := v.genExprAndLoadIfNeccesary(mem)
		if !v.inFunction() && !value.IsConstant() {
			v.err("Encountered non-constant value in global array")
		}
		arrayValues[idx] = value
	}

	lengthValue := llvm.ConstInt(v.primitiveTypeToLLVMType(ast.PRIMITIVE_uint), uint64(len(n.Values)), false)
	var backingArrayPointer, backingArray llvm.Value

	var length int
	if arrayType.IsFixedLength {
		length = arrayType.Length
	} else {
		length = len(n.Values)
	}

	if v.inFunction() {
		// allocate backing array
		backingArray = v.createAlignedAlloca(llvm.ArrayType(memberLLVMType, length), "")

		// copy the constant array to the backing array
		for idx, value := range arrayValues {
			gep := v.builder().CreateStructGEP(backingArray, idx, "")
			v.builder().CreateStore(value, gep)
		}

		backingArrayPointer = v.builder().CreateBitCast(backingArray, llvm.PointerType(memberLLVMType, 0), "")
		if arrayType.IsFixedLength {
			return v.builder().CreateLoad(backingArray, "")
		}
	} else {
		arrayInit := llvm.ConstArray(memberLLVMType, arrayValues)
		if arrayType.IsFixedLength {
			return arrayInit
		}

		backingArray = llvm.AddGlobal(v.curFile.LlvmModule, llvm.ArrayType(memberLLVMType, length), ".array")
		backingArray.SetLinkage(llvm.InternalLinkage)
		backingArray.SetGlobalConstant(false)
		backingArray.SetInitializer(arrayInit)
		backingArrayPointer = llvm.ConstBitCast(backingArray, llvm.PointerType(memberLLVMType, 0))
	}

	structValue := llvm.ConstNull(arrayLLVMType)
	structValue = v.builder().CreateInsertValue(structValue, lengthValue, 0, "")
	structValue = v.builder().CreateInsertValue(structValue, backingArrayPointer, 1, "")
	return structValue
}

func (v *Codegen) genStructLiteral(n *ast.CompositeLiteral) llvm.Value {
	structLLVMType := v.typeRefToLLVMType(n.Type)

	return v.genStructLiteralValues(n, llvm.Undef(structLLVMType))
}

func (v *Codegen) genStructLiteralValues(n *ast.CompositeLiteral, target llvm.Value) llvm.Value {
	structBaseType := n.Type.BaseType.ActualType().(ast.StructType)

	for i, value := range n.Values {
		name := n.Fields[i]
		idx := structBaseType.MemberIndex(name)

		memberValue := v.genExprAndLoadIfNeccesary(value)
		if !v.inFunction() && !memberValue.IsConstant() {
			v.err("Encountered non-constant value in global struct literal")
		}

		target = v.builder().CreateInsertValue(target, memberValue, idx, "")
	}

	return target
}

func (v *Codegen) genTupleLiteral(n *ast.TupleLiteral) llvm.Value {
	var tupleLLVMType llvm.Type

	var gcon *ast.GenericContext
	if n.ParentEnumLiteral != nil {
		gcon = ast.NewGenericContext(n.ParentEnumLiteral.Type.BaseType.ActualType().(ast.EnumType).GenericParameters,
			n.ParentEnumLiteral.Type.GenericArguments)
	}

	if v.inFunction() {
		if gcon == nil {
			gcon = v.currentFunction().gcon
		} else {
			gcon.Outer = v.currentFunction().gcon
		}
	}

	if n.ParentEnumLiteral != nil {
		tupleLLVMType = v.typeRefToLLVMTypeWithGenericContext(&ast.TypeReference{BaseType: n.GetType().BaseType, GenericArguments: n.ParentEnumLiteral.Type.GenericArguments}, gcon)
	} else {
		tupleLLVMType = v.typeToLLVMType(n.GetType().BaseType, gcon)
	}

	return v.genTupleLiteralValues(n, llvm.ConstNull(tupleLLVMType))
}

func (v *Codegen) genTupleLiteralValues(n *ast.TupleLiteral, target llvm.Value) llvm.Value {
	for idx, mem := range n.Members {
		memberValue := v.genExprAndLoadIfNeccesary(mem)

		if !v.inFunction() && !memberValue.IsConstant() {
			v.err("Encountered non-constant value in global tuple literal")
		}

		target = v.builder().CreateInsertValue(target, memberValue, idx, "")
	}
	return target
}

func (v *Codegen) genEnumLiteral(n *ast.EnumLiteral) llvm.Value {
	enumBaseType := n.Type.BaseType.ActualType().(ast.EnumType)

	gcon := ast.NewGenericContext(enumBaseType.GenericParameters,
		n.Type.GenericArguments)

	if v.inFunction() {
		if gcon == nil {
			gcon = v.currentFunction().gcon
		} else {
			gcon.Outer = v.currentFunction().gcon
		}
	}

	enumLLVMType := v.typeRefToLLVMTypeWithGenericContext(n.Type, gcon)

	memberIdx := enumBaseType.MemberIndex(n.Member)
	member := enumBaseType.Members[memberIdx]

	if enumBaseType.Simple {
		return llvm.ConstInt(enumLLVMType, uint64(member.Tag), false)
	}

	// TODO: Handle other integer size, maybe dynamic depending on max value?
	tagValue := llvm.ConstInt(enumTagType, uint64(member.Tag), false)

	memberLLVMType := v.enumMemberTypeToPaddedLLVMType(enumBaseType, memberIdx, gcon)

	enumLitType := v.llvmEnumTypeForMember(enumBaseType, memberIdx, gcon)
	enumValue := llvm.ConstNull(enumLitType)
	enumValue = v.builder().CreateInsertValue(enumValue, tagValue, 0, "")

	memberValue := llvm.ConstNull(memberLLVMType)
	if n.TupleLiteral != nil {
		memberValue = v.genTupleLiteralValues(n.TupleLiteral, memberValue)
	} else if n.CompositeLiteral != nil {
		memberValue = v.genStructLiteralValues(n.CompositeLiteral, memberValue)
	}
	enumValue = v.builder().CreateInsertValue(enumValue, memberValue, 1, "")

	if v.inFunction() {
		alloc := v.createAlignedAlloca(enumLitType, "")
		v.builder().CreateStore(enumValue, alloc)
		alloc = v.builder().CreateBitCast(alloc, llvm.PointerType(enumLLVMType, 0), "")
		return alloc
	}

	return enumValue
}

func (v *Codegen) genNumericLiteral(n *ast.NumericLiteral) llvm.Value {
	if n.GetType().BaseType.IsFloatingType() {
		return llvm.ConstFloat(v.typeRefToLLVMType(n.GetType()), n.AsFloat())
	} else {
		return llvm.ConstInt(v.typeRefToLLVMType(n.GetType()), n.AsInt(), false)
	}
}

func (v *Codegen) genLogicalBinop(n *ast.BinaryExpr) llvm.Value {
	and := n.Op == parser.BINOP_LOG_AND

	next := llvm.AddBasicBlock(v.currentLLVMFunction(), "and_next")
	exit := llvm.AddBasicBlock(v.currentLLVMFunction(), "and_exit")

	b1 := v.genExprAndLoadIfNeccesary(n.Lhand)
	first := v.builder().GetInsertBlock()
	if and {
		v.builder().CreateCondBr(b1, next, exit)
	} else {
		v.builder().CreateCondBr(b1, exit, next)
	}

	v.builder().SetInsertPointAtEnd(next)
	b2 := v.genExprAndLoadIfNeccesary(n.Rhand)
	next = v.builder().GetInsertBlock()
	v.builder().CreateBr(exit)

	v.builder().SetInsertPointAtEnd(exit)
	phi := v.builder().CreatePHI(b2.Type(), "and_phi")

	var testIncVal uint64
	if and {
		testIncVal = 0
	} else {
		testIncVal = 1
	}

	phi.AddIncoming([]llvm.Value{llvm.ConstInt(llvm.IntType(1), testIncVal, false), b2}, []llvm.BasicBlock{first, next})

	return phi
}

func (v *Codegen) genBinaryExpr(n *ast.BinaryExpr) llvm.Value {
	if n.Op.Category() == parser.OP_LOGICAL {
		return v.genLogicalBinop(n)
	}

	lhand := v.genExprAndLoadIfNeccesary(n.Lhand)
	rhand := v.genExprAndLoadIfNeccesary(n.Rhand)

	return v.genBinop(n.Op, n.GetType(), n.Lhand.GetType(), n.Rhand.GetType(), lhand, rhand)
}

func (v *Codegen) genBinop(operator parser.BinOpType, resType, lhandType, rhandType *ast.TypeReference, lhand, rhand llvm.Value) llvm.Value {
	if lhand.IsNil() || rhand.IsNil() {
		v.err("invalid binary expr")
	} else {
		switch operator {
		// Arithmetic
		case parser.BINOP_ADD:
			if resType.BaseType.IsFloatingType() {
				return v.builder().CreateFAdd(lhand, rhand, "")
			} else {
				return v.builder().CreateAdd(lhand, rhand, "")
			}
		case parser.BINOP_SUB:
			if resType.BaseType.IsFloatingType() {
				return v.builder().CreateFSub(lhand, rhand, "")
			} else {
				return v.builder().CreateSub(lhand, rhand, "")
			}
		case parser.BINOP_MUL:
			if resType.BaseType.IsFloatingType() {
				return v.builder().CreateFMul(lhand, rhand, "")
			} else {
				return v.builder().CreateMul(lhand, rhand, "")
			}
		case parser.BINOP_DIV:
			if resType.BaseType.IsFloatingType() {
				return v.builder().CreateFDiv(lhand, rhand, "")
			} else {
				if resType.BaseType.IsSigned() {
					return v.builder().CreateSDiv(lhand, rhand, "")
				} else {
					return v.builder().CreateUDiv(lhand, rhand, "")
				}
			}
		case parser.BINOP_MOD:
			if resType.BaseType.IsFloatingType() {
				return v.builder().CreateFRem(lhand, rhand, "")
			} else {
				if resType.BaseType.IsSigned() {
					return v.builder().CreateSRem(lhand, rhand, "")
				} else {
					return v.builder().CreateURem(lhand, rhand, "")
				}
			}

		// Comparison
		case parser.BINOP_GREATER, parser.BINOP_LESS, parser.BINOP_GREATER_EQ, parser.BINOP_LESS_EQ, parser.BINOP_EQ, parser.BINOP_NOT_EQ:
			if lhandType.BaseType.IsFloatingType() {
				return v.builder().CreateFCmp(comparisonOpToFloatPredicate(operator), lhand, rhand, "")
			} else {
				return v.builder().CreateICmp(comparisonOpToIntPredicate(operator, lhandType.BaseType.IsSigned()), lhand, rhand, "")
			}

		// Bitwise
		case parser.BINOP_BIT_AND:
			return v.builder().CreateAnd(lhand, rhand, "")
		case parser.BINOP_BIT_OR:
			return v.builder().CreateOr(lhand, rhand, "")
		case parser.BINOP_BIT_XOR:
			return v.builder().CreateXor(lhand, rhand, "")
		case parser.BINOP_BIT_LEFT:
			return v.builder().CreateShl(lhand, rhand, "")
		case parser.BINOP_BIT_RIGHT:
			// TODO make sure both operands are same type (create type cast here?)
			// TODO in semantic.go, make sure rhand is *unsigned* (LLVM always treats it that way)
			// TODO doc this
			if lhandType.BaseType.IsSigned() {
				return v.builder().CreateAShr(lhand, rhand, "")
			} else {
				return v.builder().CreateLShr(lhand, rhand, "")
			}

		default:
			panic("umimplented binop")
		}
	}

	panic("unreachable")
}

func comparisonOpToIntPredicate(op parser.BinOpType, signed bool) llvm.IntPredicate {
	switch op {
	case parser.BINOP_GREATER:
		if signed {
			return llvm.IntSGT
		}
		return llvm.IntUGT
	case parser.BINOP_LESS:
		if signed {
			return llvm.IntSLT
		}
		return llvm.IntULT
	case parser.BINOP_GREATER_EQ:
		if signed {
			return llvm.IntSGE
		}
		return llvm.IntUGE
	case parser.BINOP_LESS_EQ:
		if signed {
			return llvm.IntSLE
		}
		return llvm.IntULE
	case parser.BINOP_EQ:
		return llvm.IntEQ
	case parser.BINOP_NOT_EQ:
		return llvm.IntNE
	default:
		panic("shouln't get this")
	}
}

func comparisonOpToFloatPredicate(op parser.BinOpType) llvm.FloatPredicate {
	// TODO add stuff to docs about handling of QNAN
	switch op {
	case parser.BINOP_GREATER:
		return llvm.FloatOGT
	case parser.BINOP_LESS:
		return llvm.FloatOLT
	case parser.BINOP_GREATER_EQ:
		return llvm.FloatOGE
	case parser.BINOP_LESS_EQ:
		return llvm.FloatOLE
	case parser.BINOP_EQ:
		return llvm.FloatOEQ
	case parser.BINOP_NOT_EQ:
		return llvm.FloatONE
	default:
		panic("shouln't get this")
	}
}

func (v *Codegen) genUnaryExpr(n *ast.UnaryExpr) llvm.Value {
	expr := v.genExprAndLoadIfNeccesary(n.Expr)

	switch n.Op {
	case parser.UNOP_BIT_NOT, parser.UNOP_LOG_NOT:
		return v.builder().CreateNot(expr, "")
	case parser.UNOP_NEGATIVE:
		if n.Expr.GetType().BaseType.IsFloatingType() {
			return v.builder().CreateFNeg(expr, "")
		} else if n.Expr.GetType().BaseType.IsIntegerType() {
			return v.builder().CreateNeg(expr, "")
		} else {
			panic("internal: UNOP_NEGATIVE on non-numeric type")
		}

	default:
		panic("unimplimented unary op")
	}
}

func (v *Codegen) genCastExpr(n *ast.CastExpr) llvm.Value {
	if n.GetType().ActualTypesEqual(n.Expr.GetType()) {
		return v.genExprAndLoadIfNeccesary(n.Expr)
	}

	expr := v.genExprAndLoadIfNeccesary(n.Expr)
	exprBaseType := n.Expr.GetType().BaseType.ActualType()
	castBaseType := n.GetType().BaseType.ActualType()
	castLLVMType := v.typeRefToLLVMType(n.GetType())

	if ast.IsPointerOrReferenceType(exprBaseType) && castBaseType == ast.PRIMITIVE_uintptr {
		return v.builder().CreatePtrToInt(expr, castLLVMType, "")
	} else if ast.IsPointerOrReferenceType(castBaseType) && exprBaseType == ast.PRIMITIVE_uintptr {
		return v.builder().CreateIntToPtr(expr, castLLVMType, "")
	} else if ast.IsPointerOrReferenceType(castBaseType) && ast.IsPointerOrReferenceType(exprBaseType) {
		return v.builder().CreateBitCast(expr, castLLVMType, "")
	}

	if exprBaseType.IsIntegerType() {
		if castBaseType.IsIntegerType() {
			exprBits := v.typeToLLVMType(exprBaseType, nil).IntTypeWidth()
			castBits := castLLVMType.IntTypeWidth()
			if exprBits == castBits {
				return expr
			} else if exprBits > castBits {
				/*shiftConst := llvm.ConstInt(v.typeToLLVMType(exprBaseType), uint64(exprBits-castBits), false)
				shl := v.builder().CreateShl(expr, shiftConst, "")
				shr := v.builder().CreateAShr(shl, shiftConst, "")
				return v.builder().CreateTrunc(shr, vcastLLVMType, "")*/
				return v.builder().CreateTrunc(expr, castLLVMType, "") // TODO get this to work right!
			} else if exprBits < castBits {
				if exprBaseType.IsSigned() {
					return v.builder().CreateSExt(expr, castLLVMType, "") // TODO doc this
				} else {
					return v.builder().CreateZExt(expr, castLLVMType, "")
				}
			}
		} else if castBaseType.IsFloatingType() {
			if exprBaseType.IsSigned() {
				return v.builder().CreateSIToFP(expr, castLLVMType, "")
			} else {
				return v.builder().CreateUIToFP(expr, castLLVMType, "")
			}
		}
	} else if exprBaseType.IsFloatingType() {
		if castBaseType.IsIntegerType() {
			if exprBaseType.IsSigned() {
				return v.builder().CreateFPToSI(expr, castLLVMType, "")
			} else {
				return v.builder().CreateFPToUI(expr, castLLVMType, "")
			}
		} else if castBaseType.IsFloatingType() {
			exprBits := floatTypeBits(exprBaseType.(ast.PrimitiveType))
			castBits := floatTypeBits(castBaseType.(ast.PrimitiveType))
			if exprBits == castBits {
				return expr
			} else if exprBits > castBits {
				return v.builder().CreateFPTrunc(expr, castLLVMType, "") // TODO get this to work right!
			} else if exprBits < castBits {
				return v.builder().CreateFPExt(expr, castLLVMType, "")
			}
		}
	}

	panic("unimplimented typecast: " + n.String())
}

func (v *Codegen) genCallExprWithArgs(n *ast.CallExpr, args []llvm.Value) llvm.Value {
	call := v.builder().CreateCall(v.genExprAndLoadIfNeccesary(n.Function), args, "")

	attrs := n.Function.GetType().BaseType.ActualType().(ast.FunctionType).Attrs()
	if attr, ok := attrs["call_conv"]; ok {
		call.SetInstructionCallConv(callConvTypes[attr.Value])
	}

	return call
}

func (v *Codegen) genCallExpr(n *ast.CallExpr) llvm.Value {

	args := v.genCallArgs(n)
	return v.genCallExprWithArgs(n, args)
}

func (v *Codegen) genCallArgs(n *ast.CallExpr) []llvm.Value {
	numArgs := len(n.Arguments)
	if n.ReceiverAccess != nil {
		numArgs++
	}

	args := make([]llvm.Value, 0, numArgs)

	if n.ReceiverAccess != nil {
		llvmReciverAccess := v.genExprAndLoadIfNeccesary(n.ReceiverAccess)
		args = append(args, llvmReciverAccess)
	}

	for _, arg := range n.Arguments {
		llvmArg := v.genExprAndLoadIfNeccesary(arg)
		args = append(args, llvmArg)
	}

	return args
}

func (v *Codegen) genArrayLenExpr(n *ast.ArrayLenExpr) llvm.Value {
	arrType := n.Expr.GetType().BaseType.ActualType().(ast.ArrayType)
	if arrType.IsFixedLength {
		return llvm.ConstInt(v.targetData.IntPtrType(), uint64(arrType.Length), false)
	}

	if arrayLit, ok := n.Expr.(*ast.CompositeLiteral); ok {
		arrayLen := len(arrayLit.Values)

		return llvm.ConstInt(v.targetData.IntPtrType(), uint64(arrayLen), false)
	}

	gep := v.genAccessGEP(n.Expr)
	gep = v.builder().CreateLoad(v.builder().CreateStructGEP(gep, 0, ""), "")
	return gep
}

func (v *Codegen) genSizeofExpr(n *ast.SizeofExpr) llvm.Value {
	var typ llvm.Type

	if n.Expr != nil {
		typ = v.typeRefToLLVMType(n.Expr.GetType())
	} else {
		typ = v.typeRefToLLVMType(n.Type)
	}

	return llvm.ConstInt(v.targetData.IntPtrType(), v.targetData.TypeAllocSize(typ), false)
}
