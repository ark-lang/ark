package LLVMCodegen

import (
	"fmt"
	"os"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/semantic"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"

	"github.com/ark-lang/go-llvm/llvm"
)

type functionAndFnGenericInstance struct {
	fn   *parser.Function
	gcon *parser.GenericContext // nil for no generics
}

func newfunctionAndFnGenericInstance(fn *parser.Function, gcon *parser.GenericContext) functionAndFnGenericInstance {
	return functionAndFnGenericInstance{
		fn:   fn,
		gcon: gcon,
	}
}

type variableAndFnGenericInstance struct {
	variable *parser.Variable
	gcon     *parser.GenericContext // nil for no generics
}

func newvariableAndFnGenericInstance(variable *parser.Variable, gcon *parser.GenericContext) variableAndFnGenericInstance {
	return variableAndFnGenericInstance{
		variable: variable,
		gcon:     gcon,
	}
}

type Codegen struct {
	// public options
	OutputName string
	OutputType OutputType
	LinkerArgs []string
	Linker     string // defaults to cc
	OptLevel   int

	// private stuff
	input   []*WrappedModule
	curFile *WrappedModule

	builders     map[functionAndFnGenericInstance]llvm.Builder      // map of functions to builders
	curLoopExits map[functionAndFnGenericInstance][]llvm.BasicBlock // map of functions to slices of blocks, where each block is the exit block for current loops
	curLoopNexts map[functionAndFnGenericInstance][]llvm.BasicBlock // map of functions to slices of blocks, where each block is the eval block for current loops

	globalBuilder   llvm.Builder // used non-function stuff
	variableLookup  map[variableAndFnGenericInstance]llvm.Value
	namedTypeLookup map[string]llvm.Type

	declForFunction map[*parser.Function]*parser.FunctionDecl

	referenceAccess bool
	inFunctions     []functionAndFnGenericInstance

	lambdaID int

	inBlocks       map[functionAndFnGenericInstance][]*parser.Block
	blockDeferData map[*parser.Block][]*deferData // TODO make sure works with generics

	// dirty thing for global arrays
	arrayIndex int

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

	name := curFn.fn.MangledName(parser.MANGLE_ARK_UNSTABLE, curFn.gcon)
	if curFn.fn.Type.Attrs().Contains("nomangle") {
		name = curFn.fn.Name
	}
	return v.curFile.LlvmModule.NamedFunction(name)
}

func (v *Codegen) pushBlock(block *parser.Block) {
	fn := v.currentFunction()
	v.inBlocks[fn] = append(v.inBlocks[fn], block)
}

func (v *Codegen) popBlock() {
	fn := v.currentFunction()
	v.inBlocks[fn] = v.inBlocks[fn][:len(v.inBlocks[fn])-1]
}

func (v *Codegen) currentBlock() *parser.Block {
	fn := v.currentFunction()
	return v.inBlocks[fn][len(v.inBlocks[fn])-1]
}

func (v *Codegen) nextLambdaID() int {
	id := v.lambdaID
	v.lambdaID++
	return id
}

type WrappedModule struct {
	*parser.Module
	LlvmModule llvm.Module
}

type deferData struct {
	stat *parser.DeferStat
	args []llvm.Value
}

func (v *Codegen) err(err string, stuff ...interface{}) {
	log.Error("codegen", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_CODEGEN)
}

func (v *Codegen) Generate(input []*parser.Module) {
	v.builders = make(map[functionAndFnGenericInstance]llvm.Builder)
	v.inBlocks = make(map[functionAndFnGenericInstance][]*parser.Block)
	v.globalBuilder = llvm.NewBuilder()
	defer v.globalBuilder.Dispose()

	v.curLoopExits = make(map[functionAndFnGenericInstance][]llvm.BasicBlock)
	v.curLoopNexts = make(map[functionAndFnGenericInstance][]llvm.BasicBlock)

	v.declForFunction = make(map[*parser.Function]*parser.FunctionDecl)

	v.input = make([]*WrappedModule, len(input))
	for idx, mod := range input {
		v.input[idx] = &WrappedModule{Module: mod}
	}

	v.variableLookup = make(map[variableAndFnGenericInstance]llvm.Value)
	v.namedTypeLookup = make(map[string]llvm.Type)

	// initialize llvm target
	llvm.InitializeNativeTarget()
	llvm.InitializeNativeAsmPrinter()

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

	v.blockDeferData = make(map[*parser.Block][]*deferData)

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

func (v *Codegen) recursiveGenericFunctionHelper(n *parser.FunctionDecl, access *parser.FunctionAccessExpr, gcon *parser.GenericContext, fn func(*parser.FunctionDecl, *parser.GenericContext)) {
	exit := true
	for _, garg := range access.GenericArguments {
		if _, ok := garg.BaseType.(*parser.SubstitutionType); ok {
			exit = false
			break
		}
	}

	if exit {
		fn(n, gcon)
		return
	}

	for _, subAccess := range access.ParentFunction.Accesses {
		newGcon := parser.NewGenericContext(subAccess.Function.Type.GenericParameters, subAccess.GenericArguments)
		newGcon.Outer = gcon

		v.recursiveGenericFunctionHelper(n, subAccess, newGcon, fn)
	}
}

func (v *Codegen) declareDecls(nodes []parser.Node) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *parser.FunctionDecl:
			v.declForFunction[n.Function] = n
		}
	}

	for _, node := range nodes {
		switch n := node.(type) {
		case *parser.FunctionDecl:
			if len(n.Function.Type.GenericParameters) == 0 {
				v.declareFunctionDecl(n, nil)
			} else {
				for _, access := range n.Function.Accesses {
					gcon := parser.NewGenericContext(access.Function.Type.GenericParameters, access.GenericArguments)

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

func (v *Codegen) declareFunctionDecl(n *parser.FunctionDecl, gcon *parser.GenericContext) {
	mangledName := n.Function.MangledName(parser.MANGLE_ARK_UNSTABLE, gcon)
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
			funcParam.SetName(n.Function.Parameters[i].Variable.MangledName(parser.MANGLE_ARK_UNSTABLE))
		}*/
	}
}

func (v *Codegen) getVariable(vari variableAndFnGenericInstance) llvm.Value {
	if value, ok := v.variableLookup[vari]; ok {
		return value
	}

	if vari.variable.ParentModule != v.curFile.Module {
		value := llvm.AddGlobal(v.curFile.LlvmModule, v.typeRefToLLVMType(vari.variable.Type), vari.variable.MangledName(parser.MANGLE_ARK_UNSTABLE))
		value.SetLinkage(llvm.ExternalLinkage)
		v.variableLookup[vari] = value
		return value
	}

	v.err("Encountered undeclared variable `%s` in same modules", vari.variable.Name)
	return llvm.Value{}
}

func (v *Codegen) genNode(n parser.Node) {
	switch n := n.(type) {
	case parser.Decl:
		v.genDecl(n)
	case parser.Expr:
		v.genExpr(n)
	case parser.Stat:
		v.genStat(n)
	case *parser.Block:
		v.genBlock(n)
	}
}

func (v *Codegen) genStat(n parser.Stat) {
	switch n := n.(type) {
	case *parser.ReturnStat:
		v.genReturnStat(n)
	case *parser.BreakStat:
		v.genBreakStat(n)
	case *parser.NextStat:
		v.genNextStat(n)
	case *parser.BlockStat:
		v.genBlockStat(n)
	case *parser.CallStat:
		v.genCallStat(n)
	case *parser.AssignStat:
		v.genAssignStat(n)
	case *parser.BinopAssignStat:
		v.genBinopAssignStat(n)
	case *parser.IfStat:
		v.genIfStat(n)
	case *parser.LoopStat:
		v.genLoopStat(n)
	case *parser.MatchStat:
		v.genMatchStat(n)
	case *parser.DeferStat:
		v.genDeferStat(n)
	default:
		panic("unimplemented stat")
	}
}

func (v *Codegen) genBreakStat(n *parser.BreakStat) {
	curExits := v.curLoopExits[v.currentFunction()]
	v.builder().CreateBr(curExits[len(curExits)-1])
}

func (v *Codegen) genNextStat(n *parser.NextStat) {
	curNexts := v.curLoopNexts[v.currentFunction()]
	v.builder().CreateBr(curNexts[len(curNexts)-1])
}

func (v *Codegen) genDeferStat(n *parser.DeferStat) {
	data := &deferData{
		stat: n,
	}

	v.blockDeferData[v.currentBlock()] = append(v.blockDeferData[v.currentBlock()], data)

	for _, arg := range n.Call.Arguments {
		data.args = append(data.args, v.genExpr(arg))
	}
}

func (v *Codegen) genRunDefers(block *parser.Block) {
	deferDat := v.blockDeferData[block]

	if len(deferDat) > 0 {
		for i := len(deferDat) - 1; i >= 0; i-- {
			v.genCallExprWithArgs(deferDat[i].stat.Call, deferDat[i].args)
		}
	}
}

func (v *Codegen) genBlock(n *parser.Block) {
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

func (v *Codegen) genReturnStat(n *parser.ReturnStat) {
	var ret llvm.Value
	if n.Value != nil {
		ret = v.genExpr(n.Value)
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

func (v *Codegen) genBlockStat(n *parser.BlockStat) {
	v.genBlock(n.Block)
}

func (v *Codegen) genCallStat(n *parser.CallStat) {
	v.genExpr(n.Call)
}

func (v *Codegen) genAssignStat(n *parser.AssignStat) {
	v.builder().CreateStore(v.genExpr(n.Assignment), v.genAccessGEP(n.Access))
}

func (v *Codegen) genBinopAssignStat(n *parser.BinopAssignStat) {
	storage := v.genAccessGEP(n.Access)

	storageValue := v.builder().CreateLoad(storage, "")
	assignmentValue := v.genExpr(n.Assignment)

	value := v.genBinop(n.Operator, n.Access.GetType(), n.Access.GetType(), n.Assignment.GetType(), storageValue, assignmentValue)
	v.builder().CreateStore(value, storage)
}

func isBreakOrNext(n parser.Node) bool {
	switch n.(type) {
	case *parser.BreakStat, *parser.NextStat:
		return true
	}
	return false
}

func (v *Codegen) genIfStat(n *parser.IfStat) {
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
		cond := v.genExpr(expr)

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

func (v *Codegen) genLoopStat(n *parser.LoopStat) {
	curfn := v.currentFunction()
	afterBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "loop_exit")
	v.curLoopExits[curfn] = append(v.curLoopExits[curfn], afterBlock)

	switch n.LoopType {
	case parser.LOOP_TYPE_INFINITE:
		loopBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "loop_body")
		v.curLoopNexts[curfn] = append(v.curLoopNexts[curfn], loopBlock)
		v.builder().CreateBr(loopBlock)
		v.builder().SetInsertPointAtEnd(loopBlock)

		v.genBlock(n.Body)

		if !isBreakOrNext(n.Body.LastNode()) {
			v.builder().CreateBr(loopBlock)
		}

		v.builder().SetInsertPointAtEnd(afterBlock)
	case parser.LOOP_TYPE_CONDITIONAL:
		evalBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "loop_condeval")
		v.builder().CreateBr(evalBlock)
		v.curLoopNexts[curfn] = append(v.curLoopNexts[curfn], evalBlock)

		loopBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "loop_body")

		v.builder().SetInsertPointAtEnd(evalBlock)
		cond := v.genExpr(n.Condition)
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

func (v *Codegen) genMatchStat(n *parser.MatchStat) {
	// TODO: implement
}

func (v *Codegen) genDecl(n parser.Decl) {
	switch n := n.(type) {
	case *parser.FunctionDecl:
		if len(n.Function.Type.GenericParameters) == 0 {
			v.genFunctionDecl(n, nil)
		} else {
			for _, access := range n.Function.Accesses {
				gcon := parser.NewGenericContext(access.Function.Type.GenericParameters, access.GenericArguments)

				v.recursiveGenericFunctionHelper(n, access, gcon, v.genFunctionDecl)
			}
		}
	case *parser.VariableDecl:
		v.genVariableDecl(n, true)
	case *parser.TypeDecl:
		// TODO nothing to gen?
	default:
		v.err("unimplemented decl found: `%s`", n.NodeName())
	}
}

func (v *Codegen) genFunctionDecl(n *parser.FunctionDecl, gcon *parser.GenericContext) {
	mangledName := n.Function.MangledName(parser.MANGLE_ARK_UNSTABLE, gcon)
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

func (v *Codegen) genFunctionBody(fn *parser.Function, llvmFn llvm.Value, gcon *parser.GenericContext) {
	block := llvm.AddBasicBlock(llvmFn, "entry")

	v.pushFunction(newfunctionAndFnGenericInstance(fn, gcon))
	v.builders[v.currentFunction()] = llvm.NewBuilder()
	v.builder().SetInsertPointAtEnd(block)

	pars := fn.Parameters

	if fn.Type.Receiver != nil {
		newPars := make([]*parser.VariableDecl, len(pars)+1)
		newPars[0] = fn.Receiver
		copy(newPars[1:], pars)
		pars = newPars
	}

	for i, par := range pars {
		alloc := v.builder().CreateAlloca(v.typeRefToLLVMType(par.Variable.Type), par.Variable.Name)
		v.variableLookup[newvariableAndFnGenericInstance(par.Variable, gcon)] = alloc

		v.builder().CreateStore(llvmFn.Params()[i], alloc)
	}

	v.genBlock(fn.Body)
	v.builder().Dispose()
	delete(v.builders, v.currentFunction())
	delete(v.curLoopExits, v.currentFunction())
	delete(v.curLoopNexts, v.currentFunction())
	v.popFunction()
}

func (v *Codegen) genVariableDecl(n *parser.VariableDecl, semicolon bool) llvm.Value {
	var res llvm.Value

	if v.inFunction() {
		mangledName := n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE)

		funcEntry := v.currentLLVMFunction().EntryBasicBlock()

		// use this builder() for the variable alloca
		// this means all allocas go at the start of the function
		// so each variable is only allocated once
		allocBuilder := llvm.NewBuilder()

		if funcEntry == v.builder().GetInsertBlock() {
			allocBuilder.SetInsertPointAtEnd(funcEntry)
		} else {
			allocBuilder.SetInsertPointBefore(funcEntry.LastInstruction())
		}

		varType := v.typeRefToLLVMType(n.Variable.Type)
		alloc := allocBuilder.CreateAlloca(varType, mangledName)

		allocBuilder.Dispose()

		v.variableLookup[newvariableAndFnGenericInstance(n.Variable, v.currentFunction().gcon)] = alloc

		if n.Assignment != nil {
			if value := v.genExpr(n.Assignment); !value.IsNil() {
				v.builder().CreateStore(value, alloc)
			}
		} else if !n.Variable.Attrs.Contains("nozero") {
			v.builder().CreateStore(llvm.ConstNull(varType), alloc)
		}
	} else {
		// TODO cbindings
		cBinding := false

		mangledName := n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE)
		varType := v.typeRefToLLVMType(n.Variable.Type)

		value := llvm.AddGlobal(v.curFile.LlvmModule, varType, mangledName)
		// TODO: External by default to export everything, change once we get access specifiers

		if !cBinding && !n.IsPublic() {
			value.SetLinkage(nonPublicLinkage)
		}
		value.SetGlobalConstant(!n.Variable.Mutable)
		if n.Assignment != nil {
			value.SetInitializer(v.genExpr(n.Assignment))
		} else {
			value.SetInitializer(llvm.ConstNull(varType))
		}
		v.variableLookup[newvariableAndFnGenericInstance(n.Variable, nil)] = value
	}

	return res
}

func (v *Codegen) genExpr(n parser.Expr) llvm.Value {
	switch n := n.(type) {
	case *parser.ReferenceToExpr:
		return v.genReferenceToExpr(n)
	case *parser.PointerToExpr:
		return v.genPointerToExpr(n)
	case *parser.RuneLiteral:
		return v.genRuneLiteral(n)
	case *parser.NumericLiteral:
		return v.genNumericLiteral(n)
	case *parser.StringLiteral:
		return v.genStringLiteral(n)
	case *parser.BoolLiteral:
		return v.genBoolLiteral(n)
	case *parser.TupleLiteral:
		return v.genTupleLiteral(n)
	case *parser.CompositeLiteral:
		return v.genCompositeLiteral(n)
	case *parser.EnumLiteral:
		return v.genEnumLiteral(n)
	case *parser.BinaryExpr:
		return v.genBinaryExpr(n)
	case *parser.UnaryExpr:
		return v.genUnaryExpr(n)
	case *parser.CastExpr:
		return v.genCastExpr(n)
	case *parser.CallExpr:
		return v.genCallExpr(n)
	case *parser.VariableAccessExpr, *parser.StructAccessExpr,
		*parser.ArrayAccessExpr, *parser.DerefAccessExpr,
		*parser.FunctionAccessExpr:
		return v.genAccessExpr(n)
	case *parser.SizeofExpr:
		return v.genSizeofExpr(n)
	case *parser.ArrayLenExpr:
		return v.genArrayLenExpr(n)
	case *parser.LambdaExpr:
		return v.genLambdaExpr(n)
	default:
		log.Debug("codegen", "expr: %s\n", n)
		panic("unimplemented expr")
	}
}

func (v *Codegen) genLambdaExpr(n *parser.LambdaExpr) llvm.Value {
	typ := v.functionTypeToLLVMType(n.Function.Type, false, nil)
	mod := v.curFile.LlvmModule
	fn := llvm.AddFunction(mod, fmt.Sprintf("_Lambda%d", v.nextLambdaID()), typ)

	if len(n.Function.Type.GenericParameters) > 0 {
		panic("generic lambdas unimplemented")
	}

	v.genFunctionBody(n.Function, fn, nil)

	return fn
}

func (v *Codegen) genReferenceToExpr(n *parser.ReferenceToExpr) llvm.Value {
	return v.genAccessGEP(n.Access)
}

func (v *Codegen) genPointerToExpr(n *parser.PointerToExpr) llvm.Value {
	return v.genAccessGEP(n.Access)
}

func (v *Codegen) genAccessExpr(n parser.Expr) llvm.Value {
	if fae, ok := n.(*parser.FunctionAccessExpr); ok {
		gcon := parser.NewGenericContext(fae.Function.Type.GenericParameters, fae.GenericArguments)
		gcon.Outer = v.currentFunction().gcon

		var fnName string
		if fae.ReceiverAccess != nil {
			fnName = fae.Function.MangledNameWithReceiver(parser.MANGLE_ARK_UNSTABLE, fae.ReceiverAccess.GetType().BaseType, gcon)
		} else {
			fnName = fae.Function.MangledName(parser.MANGLE_ARK_UNSTABLE, gcon)
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
			decl := &parser.FunctionDecl{Function: fae.Function, Prototype: true}
			decl.SetPublic(true)
			v.declareFunctionDecl(decl, gcon)

			if v.curFile.LlvmModule.NamedFunction(fnName).IsNil() {
				panic("how did this happen")
			}
			fn = v.curFile.LlvmModule.NamedFunction(fnName)
		}

		return fn
	}

	return v.builder().CreateLoad(v.genAccessGEP(n), "")
}

func (v *Codegen) genAccessGEP(n parser.Expr) llvm.Value {
	var curFngcon *parser.GenericContext
	if v.inFunction() {
		curFngcon = v.currentFunction().gcon
	}

	switch access := n.(type) {
	case *parser.VariableAccessExpr:
		varType := v.getVariable(newvariableAndFnGenericInstance(access.Variable, curFngcon))
		if varType.IsNil() {
			panic("varType was nil")
		}
		gep := v.builder().CreateGEP(varType, []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false)}, "")

		return gep

	case *parser.StructAccessExpr:
		gep := v.genAccessGEP(access.Struct)

		typ := access.Struct.GetType().BaseType.ActualType()

		index := typ.(parser.StructType).MemberIndex(access.Member)
		return v.builder().CreateStructGEP(gep, index, "")

	case *parser.ArrayAccessExpr:
		gep := v.genAccessGEP(access.Array)

		subscriptExpr := v.genExpr(access.Subscript)
		subsTyp := access.Subscript.GetType().BaseType.ActualType().(parser.PrimitiveType)
		// Extend access width to system poiner width
		if !subsTyp.IsSigned() {
			subscriptExpr = v.builder().CreateZExt(subscriptExpr, v.targetData.IntPtrType(), "")
		} else {
			subscriptExpr = v.builder().CreateSExt(subscriptExpr, v.targetData.IntPtrType(), "")
		}

		v.genBoundsCheck(v.builder().CreateLoad(v.builder().CreateStructGEP(gep, 0, ""), ""),
			subscriptExpr, access.Subscript.GetType().BaseType.IsSigned())

		gep = v.builder().CreateStructGEP(gep, 1, "")

		load := v.builder().CreateLoad(gep, "")

		gepIndexes := []llvm.Value{subscriptExpr}
		return v.builder().CreateGEP(load, gepIndexes, "")

	case *parser.DerefAccessExpr:
		return v.genExpr(access.Expr)

	default:
		panic("unhandled access type")
	}
}

func (v *Codegen) genBoundsCheck(limit llvm.Value, index llvm.Value, indexIsSigned bool) {
	segvBlock := llvm.AddBasicBlock(v.currentLLVMFunction(), "boundscheck_segv")
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

	v.builder().SetInsertPointAtEnd(segvBlock)
	v.genRaiseSegfault()
	v.builder().CreateUnreachable()

	v.builder().SetInsertPointAtEnd(endBlock)
}

func (v *Codegen) genRaiseSegfault() {
	fn := v.curFile.LlvmModule.NamedFunction("raise")
	intType := v.primitiveTypeToLLVMType(parser.PRIMITIVE_int)

	if fn.IsNil() {
		fnType := llvm.FunctionType(intType, []llvm.Type{intType}, false)
		fn = llvm.AddFunction(v.curFile.LlvmModule, "raise", fnType)
	}

	v.builder().CreateCall(fn, []llvm.Value{llvm.ConstInt(intType, 11, false)}, "segfault")
}

func (v *Codegen) genBoolLiteral(n *parser.BoolLiteral) llvm.Value {
	var num uint64

	if n.Value {
		num = 1
	}

	return llvm.ConstInt(v.typeRefToLLVMType(n.GetType()), num, true)
}

func (v *Codegen) genRuneLiteral(n *parser.RuneLiteral) llvm.Value {
	return llvm.ConstInt(v.typeRefToLLVMType(n.GetType()), uint64(n.Value), true)
}

func (v *Codegen) genStringLiteral(n *parser.StringLiteral) llvm.Value {
	memberLLVMType := v.primitiveTypeToLLVMType(parser.PRIMITIVE_u8)
	nullTerm := n.IsCString
	length := len(n.Value)
	if nullTerm {
		length++
	}

	var backingArrayPointer llvm.Value

	if v.inFunction() {
		// allocate backing array
		backingArray := v.builder().CreateAlloca(llvm.ArrayType(memberLLVMType, length), "stackstr")
		v.builder().CreateStore(llvm.ConstString(n.Value, nullTerm), backingArray)

		backingArrayPointer = v.builder().CreateBitCast(backingArray, llvm.PointerType(memberLLVMType, 0), "")
	} else {
		backName := fmt.Sprintf("_globarr_back_%d", v.arrayIndex)
		v.arrayIndex++

		backingArray := llvm.AddGlobal(v.curFile.LlvmModule, llvm.ArrayType(memberLLVMType, length), backName)
		backingArray.SetLinkage(llvm.InternalLinkage)
		backingArray.SetGlobalConstant(false)
		backingArray.SetInitializer(llvm.ConstString(n.Value, nullTerm))

		backingArrayPointer = llvm.ConstBitCast(backingArray, llvm.PointerType(memberLLVMType, 0))
	}

	if n.GetType().BaseType.ActualType().Equals(parser.ArrayOf(&parser.TypeReference{BaseType: parser.PRIMITIVE_u8})) {
		lengthValue := llvm.ConstInt(v.primitiveTypeToLLVMType(parser.PRIMITIVE_uint), uint64(length), false)
		structValue := llvm.Undef(v.typeRefToLLVMType(n.GetType()))
		structValue = v.builder().CreateInsertValue(structValue, lengthValue, 0, "")
		structValue = v.builder().CreateInsertValue(structValue, backingArrayPointer, 1, "")
		return structValue
	} else {
		return backingArrayPointer
	}
}

func (v *Codegen) genCompositeLiteral(n *parser.CompositeLiteral) llvm.Value {
	switch n.GetType().BaseType.ActualType().(type) {
	case parser.ArrayType:
		return v.genArrayLiteral(n)
	case parser.StructType:
		return v.genStructLiteral(n)
	default:
		panic("invalid composite literal type")
	}
}

// Allocates a literal array on the stack
func (v *Codegen) genArrayLiteral(n *parser.CompositeLiteral) llvm.Value {
	arrayLLVMType := v.typeRefToLLVMType(n.Type)
	memberLLVMType := v.typeRefToLLVMType(n.Type.BaseType.ActualType().(parser.ArrayType).MemberType) // TODO works with generics?

	arrayValues := make([]llvm.Value, len(n.Values))
	for idx, mem := range n.Values {
		value := v.genExpr(mem)
		if !v.inFunction() && !value.IsConstant() {
			v.err("Encountered non-constant value in global array")
		}
		arrayValues[idx] = value
	}

	lengthValue := llvm.ConstInt(v.primitiveTypeToLLVMType(parser.PRIMITIVE_uint), uint64(len(n.Values)), false)
	var backingArrayPointer llvm.Value

	if v.inFunction() {
		// allocate backing array
		backingArray := v.builder().CreateAlloca(llvm.ArrayType(memberLLVMType, len(n.Values)), "")

		// copy the constant array to the backing array
		for idx, value := range arrayValues {
			gep := v.builder().CreateStructGEP(backingArray, idx, "")
			v.builder().CreateStore(value, gep)
		}

		backingArrayPointer = v.builder().CreateBitCast(backingArray, llvm.PointerType(memberLLVMType, 0), "")
	} else {
		backName := fmt.Sprintf("_globarr_back_%d", v.arrayIndex)
		v.arrayIndex++

		backingArray := llvm.AddGlobal(v.curFile.LlvmModule, llvm.ArrayType(memberLLVMType, len(n.Values)), backName)
		backingArray.SetLinkage(llvm.InternalLinkage)
		backingArray.SetGlobalConstant(false)
		backingArray.SetInitializer(llvm.ConstArray(memberLLVMType, arrayValues))

		backingArrayPointer = llvm.ConstBitCast(backingArray, llvm.PointerType(memberLLVMType, 0))
	}

	structValue := llvm.Undef(arrayLLVMType)
	structValue = v.builder().CreateInsertValue(structValue, lengthValue, 0, "")
	structValue = v.builder().CreateInsertValue(structValue, backingArrayPointer, 1, "")
	return structValue
}

func (v *Codegen) genStructLiteral(n *parser.CompositeLiteral) llvm.Value {
	structBaseType := n.Type.BaseType.ActualType().(parser.StructType)
	structLLVMType := v.typeRefToLLVMType(n.Type)

	structValue := llvm.Undef(structLLVMType)

	for i, value := range n.Values {
		name := n.Fields[i]
		idx := structBaseType.MemberIndex(name)

		memberValue := v.genExpr(value)
		if !v.inFunction() && !memberValue.IsConstant() {
			v.err("Encountered non-constant value in global struct literal")
		}

		structValue = v.builder().CreateInsertValue(structValue, v.genExpr(value), idx, "")
	}

	return structValue
}

func (v *Codegen) genTupleLiteral(n *parser.TupleLiteral) llvm.Value {
	var tupleLLVMType llvm.Type

	var gcon *parser.GenericContext
	if n.ParentEnumLiteral != nil {
		gcon = parser.NewGenericContext(n.ParentEnumLiteral.Type.BaseType.ActualType().(parser.EnumType).GenericParameters,
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
		tupleLLVMType = v.typeRefToLLVMTypeWithGenericContext(&parser.TypeReference{BaseType: n.GetType().BaseType, GenericArguments: n.ParentEnumLiteral.Type.GenericArguments}, gcon)
	} else {
		tupleLLVMType = v.typeToLLVMType(n.GetType().BaseType, gcon)
	}

	tupleValue := llvm.Undef(tupleLLVMType)
	for idx, mem := range n.Members {
		memberValue := v.genExpr(mem)

		if !v.inFunction() && !memberValue.IsConstant() {
			v.err("Encountered non-constant value in global tuple literal")
		}

		tupleValue = v.builder().CreateInsertValue(tupleValue, memberValue, idx, "")
	}

	return tupleValue
}

func (v *Codegen) genEnumLiteral(n *parser.EnumLiteral) llvm.Value {
	enumBaseType := n.Type.BaseType.ActualType().(parser.EnumType)

	gcon := parser.NewGenericContext(enumBaseType.GenericParameters,
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
	tagValue := llvm.ConstInt(llvm.IntType(32), uint64(member.Tag), false)

	enumValue := llvm.Undef(enumLLVMType)
	enumValue = v.builder().CreateInsertValue(enumValue, tagValue, 0, "")

	memberLLVMType := v.typeToLLVMType(member.Type, gcon)
	// memberLLVMType := v.typeRefToLLVMTypeWithGenericContext(&parser.TypeReference{BaseType: member.Type, GenericArguments: n.Type.GenericArguments}, gcon)

	var memberValue llvm.Value
	if n.TupleLiteral != nil {
		memberValue = v.genTupleLiteral(n.TupleLiteral)
	} else if n.CompositeLiteral != nil {
		memberValue = v.genCompositeLiteral(n.CompositeLiteral)
	}

	if v.inFunction() {
		alloc := v.builder().CreateAlloca(enumLLVMType, "")

		tagGep := v.builder().CreateStructGEP(alloc, 0, "")
		v.builder().CreateStore(tagValue, tagGep)

		if !memberValue.IsNil() {
			dataGep := v.builder().CreateStructGEP(alloc, 1, "")

			dataGep = v.builder().CreateBitCast(dataGep, llvm.PointerType(memberLLVMType, 0), "")

			v.builder().CreateStore(memberValue, dataGep)
		}

		return v.builder().CreateLoad(alloc, "")
	} else {
		panic("unimplemented: global enum literal")
	}
}

func (v *Codegen) genNumericLiteral(n *parser.NumericLiteral) llvm.Value {
	if n.GetType().BaseType.IsFloatingType() {
		return llvm.ConstFloat(v.typeRefToLLVMType(n.GetType()), n.AsFloat())
	} else {
		return llvm.ConstInt(v.typeRefToLLVMType(n.GetType()), n.AsInt(), false)
	}
}

func (v *Codegen) genLogicalBinop(n *parser.BinaryExpr) llvm.Value {
	and := n.Op == parser.BINOP_LOG_AND

	next := llvm.AddBasicBlock(v.currentLLVMFunction(), "and_next")
	exit := llvm.AddBasicBlock(v.currentLLVMFunction(), "and_exit")

	b1 := v.genExpr(n.Lhand)
	first := v.builder().GetInsertBlock()
	if and {
		v.builder().CreateCondBr(b1, next, exit)
	} else {
		v.builder().CreateCondBr(b1, exit, next)
	}

	v.builder().SetInsertPointAtEnd(next)
	b2 := v.genExpr(n.Rhand)
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

func (v *Codegen) genBinaryExpr(n *parser.BinaryExpr) llvm.Value {
	if n.Op.Category() == parser.OP_LOGICAL {
		return v.genLogicalBinop(n)
	}

	lhand := v.genExpr(n.Lhand)
	rhand := v.genExpr(n.Rhand)

	return v.genBinop(n.Op, n.GetType(), n.Lhand.GetType(), n.Rhand.GetType(), lhand, rhand)
}

func (v *Codegen) genBinop(operator parser.BinOpType, resType, lhandType, rhandType *parser.TypeReference, lhand, rhand llvm.Value) llvm.Value {
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

func (v *Codegen) genUnaryExpr(n *parser.UnaryExpr) llvm.Value {
	expr := v.genExpr(n.Expr)

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

func (v *Codegen) genCastExpr(n *parser.CastExpr) llvm.Value {
	if n.GetType().ActualTypesEqual(n.Expr.GetType()) {
		return v.genExpr(n.Expr)
	}

	expr := v.genExpr(n.Expr)
	exprBaseType := n.Expr.GetType().BaseType.ActualType()
	castBaseType := n.GetType().BaseType.ActualType()
	castLLVMType := v.typeRefToLLVMType(n.GetType())

	if parser.IsPointerOrReferenceType(exprBaseType) && castBaseType == parser.PRIMITIVE_uintptr {
		return v.builder().CreatePtrToInt(expr, castLLVMType, "")
	} else if parser.IsPointerOrReferenceType(castBaseType) && exprBaseType == parser.PRIMITIVE_uintptr {
		return v.builder().CreateIntToPtr(expr, castLLVMType, "")
	} else if parser.IsPointerOrReferenceType(castBaseType) && parser.IsPointerOrReferenceType(exprBaseType) {
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
			exprBits := floatTypeBits(exprBaseType.(parser.PrimitiveType))
			castBits := floatTypeBits(exprBaseType.(parser.PrimitiveType))
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

func (v *Codegen) genCallExprWithArgs(n *parser.CallExpr, args []llvm.Value) llvm.Value {
	call := v.builder().CreateCall(v.genAccessExpr(n.Function), args, "")

	attrs := n.Function.GetType().BaseType.(parser.FunctionType).Attrs()
	if attr, ok := attrs["call_conv"]; ok {
		call.SetInstructionCallConv(callConvTypes[attr.Value])
	}

	return call
}

func (v *Codegen) genCallExpr(n *parser.CallExpr) llvm.Value {
	numArgs := len(n.Arguments)
	if n.ReceiverAccess != nil {
		numArgs++
	}

	args := make([]llvm.Value, 0, numArgs)

	if n.ReceiverAccess != nil {
		args = append(args, v.genExpr(n.ReceiverAccess))
	}

	for _, arg := range n.Arguments {
		args = append(args, v.genExpr(arg))
	}

	return v.genCallExprWithArgs(n, args)
}

func (v *Codegen) genArrayLenExpr(n *parser.ArrayLenExpr) llvm.Value {
	if arrayLit, ok := n.Expr.(*parser.CompositeLiteral); ok {
		arrayLen := len(arrayLit.Values)

		return llvm.ConstInt(llvm.IntType(64), uint64(arrayLen), false)
	}

	gep := v.genAccessGEP(n.Expr)
	gep = v.builder().CreateLoad(v.builder().CreateStructGEP(gep, 0, ""), "")
	return gep
}

func (v *Codegen) genSizeofExpr(n *parser.SizeofExpr) llvm.Value {
	var typ llvm.Type

	if n.Expr != nil {
		typ = v.typeRefToLLVMType(n.Expr.GetType())
	} else {
		typ = v.typeRefToLLVMType(n.Type)
	}

	return llvm.ConstInt(v.targetData.IntPtrType(), v.targetData.TypeAllocSize(typ), false)
}
