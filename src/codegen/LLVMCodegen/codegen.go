package LLVMCodegen

import (
	"C"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"

	"llvm.org/llvm/bindings/go/llvm"
)
import (
	"github.com/ark-lang/ark/src/util/log"
)

const intSize = int(unsafe.Sizeof(C.int(0)))

type Codegen struct {
	input   []*parser.Module
	curFile *parser.Module

	builder                        llvm.Builder
	variableLookup                 map[*parser.Variable]llvm.Value
	structLookup_UseHelperFunction map[*parser.StructType]llvm.Type // use getStructDecl

	OutputName   string
	OutputType   OutputType
	CompilerArgs []string
	Compiler     string // defaults to cc
	LinkerArgs   []string
	Linker       string // defaults to cc

	modules map[string]*parser.Module

	inFunction      bool
	currentFunction llvm.Value

	currentBlock   *parser.Block
	blockDeferData map[*parser.Block][]*deferData

	// dirty thing for global arrays
	arrayIndex int

	// size calculation stuff
	target        llvm.Target
	targetMachine llvm.TargetMachine
	targetData    llvm.TargetData
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

func (v *Codegen) Generate(input []*parser.Module, modules map[string]*parser.Module) {
	v.input = input
	v.builder = llvm.NewBuilder()
	v.variableLookup = make(map[*parser.Variable]llvm.Value)
	v.structLookup_UseHelperFunction = make(map[*parser.StructType]llvm.Type)

	// initialize llvm target
	llvm.InitializeNativeTarget()

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
	passBuilder.SetOptLevel(3)
	//passBuilder.Populate(passManager) //leave this off until the compiler is better

	v.modules = make(map[string]*parser.Module)
	v.blockDeferData = make(map[*parser.Block][]*deferData)

	for _, infile := range input {
		log.Verboseln("codegen", util.TEXT_BOLD+util.TEXT_GREEN+"Started codegenning "+util.TEXT_RESET+infile.Name)
		t := time.Now()

		infile.Module = llvm.NewModule(infile.Name)
		v.curFile = infile

		v.modules[v.curFile.Name] = v.curFile

		v.declareDecls(infile.Nodes)

		for _, node := range infile.Nodes {
			v.genNode(node)
		}

		if err := llvm.VerifyModule(infile.Module, llvm.ReturnStatusAction); err != nil {
			infile.Module.Dump()
			v.err("%s", err.Error())
		}

		passManager.Run(infile.Module)

		if log.AtLevel(log.LevelVerbose) {
			infile.Module.Dump()
		}

		dur := time.Since(t)
		log.Verbose("codegen", util.TEXT_BOLD+util.TEXT_GREEN+"Finished codegenning "+util.TEXT_RESET+infile.Name+" (%.2fms)\n",
			float32(dur.Nanoseconds())/1000000)
	}

	passManager.Dispose()

	v.createBinary()
}

func (v *Codegen) declareDecls(nodes []parser.Node) {
	for _, node := range nodes {
		if n, ok := node.(parser.Decl); ok {
			switch n.(type) {
			case *parser.StructDecl:
				v.declareStructDecl(n.(*parser.StructDecl))
			}
		}
	}

	for _, node := range nodes {
		if n, ok := node.(parser.Decl); ok {
			switch n.(type) {
			case *parser.FunctionDecl:
				v.declareFunctionDecl(n.(*parser.FunctionDecl))
			}
		}
	}
}

func (v *Codegen) declareStructDecl(n *parser.StructDecl) {
	v.addStructType(n.Struct)
}

func (v *Codegen) addStructType(typ *parser.StructType) {
	if _, ok := v.structLookup_UseHelperFunction[typ]; ok {
		return
	}

	for _, field := range typ.Variables {
		if struc, ok := field.Variable.Type.(*parser.StructType); ok {
			v.addStructType(struc) // TODO check recursive loop
		}
	}

	numOfFields := len(typ.Variables)
	fields := make([]llvm.Type, numOfFields)
	packed := false

	for i, member := range typ.Variables {
		memberType := v.typeToLLVMType(member.Variable.Type)
		fields[i] = memberType
	}

	structure := llvm.StructType(fields, packed)
	llvm.AddGlobal(v.curFile.Module, structure, typ.MangledName(parser.MANGLE_ARK_UNSTABLE))
	v.structLookup_UseHelperFunction[typ] = structure

}

func (v *Codegen) declareFunctionDecl(n *parser.FunctionDecl) {
	mangledName := n.Function.MangledName(parser.MANGLE_ARK_UNSTABLE)
	function := v.curFile.Module.NamedFunction(mangledName)
	if !function.IsNil() {
		v.err("function `%s` already exists in module", n.Function.Name)
	} else {
		numOfParams := len(n.Function.Parameters)
		params := make([]llvm.Type, numOfParams)
		for i, par := range n.Function.Parameters {
			params[i] = v.typeToLLVMType(par.Variable.Type)
		}

		// attributes defaults
		cBinding := false

		// find them attributes yo
		if n.Function.Attrs != nil {
			cBinding = n.Function.Attrs.Contains("c")
		}

		// assume it's void
		funcTypeRaw := llvm.VoidType()

		// oo theres a type, let's try figure it out
		if n.Function.ReturnType != nil {
			funcTypeRaw = v.typeToLLVMType(n.Function.ReturnType)
		}

		// create the function type
		funcType := llvm.FunctionType(funcTypeRaw, params, n.Function.IsVariadic)

		functionName := mangledName
		if cBinding {
			functionName = n.Function.Name
		}

		// add that shit
		function = llvm.AddFunction(v.curFile.Module, functionName, funcType)

		// do some magical shit for later
		for i := 0; i < numOfParams; i++ {
			funcParam := function.Param(i)
			funcParam.SetName(n.Function.Parameters[i].Variable.MangledName(parser.MANGLE_ARK_UNSTABLE))
		}

		if cBinding {
			function.SetFunctionCallConv(llvm.CCallConv)
		} else {
			function.SetFunctionCallConv(llvm.FastCallConv)
		}
	}
}

func (v *Codegen) genNode(n parser.Node) {
	switch n.(type) {
	case parser.Decl:
		v.genDecl(n.(parser.Decl))
	case parser.Expr:
		v.genExpr(n.(parser.Expr))
	case parser.Stat:
		v.genStat(n.(parser.Stat))
	case *parser.Block:
		v.genBlock(n.(*parser.Block))
	}
}

func (v *Codegen) genStat(n parser.Stat) {
	switch n.(type) {
	case *parser.ReturnStat:
		v.genReturnStat(n.(*parser.ReturnStat))
	case *parser.BlockStat:
		v.genBlockStat(n.(*parser.BlockStat))
	case *parser.CallStat:
		v.genCallStat(n.(*parser.CallStat))
	case *parser.AssignStat:
		v.genAssignStat(n.(*parser.AssignStat))
	case *parser.IfStat:
		v.genIfStat(n.(*parser.IfStat))
	case *parser.LoopStat:
		v.genLoopStat(n.(*parser.LoopStat))
	case *parser.MatchStat:
		v.genMatchStat(n.(*parser.MatchStat))
	case *parser.DeferStat:
		v.genDeferStat(n.(*parser.DeferStat))
	default:
		panic("unimplimented stat")
	}
}

func (v *Codegen) genDeferStat(n *parser.DeferStat) {
	data := &deferData{
		stat: n,
	}

	v.blockDeferData[v.currentBlock] = append(v.blockDeferData[v.currentBlock], data)

	for _, arg := range n.Call.Arguments {
		data.args = append(data.args, v.genExpr(arg))
	}
}

func (v *Codegen) genBlock(n *parser.Block) {
	for i, x := range n.Nodes {
		v.currentBlock = n // set it on every iteration to overide sub-blocks

		if i == len(n.Nodes)-1 && !n.IsTerminating {
			fmt.Println("true")
		}

		v.genNode(x)

		if (i == len(n.Nodes)-1 && !n.IsTerminating) || (i == len(n.Nodes)-2 && n.IsTerminating) {
			// print all the defer stats
			deferDat := v.blockDeferData[n]

			if len(deferDat) > 0 {
				for i := len(deferDat) - 1; i >= 0; i-- {
					v.genCallExprWithArgs(deferDat[i].stat.Call, deferDat[i].args)
				}
			}

			delete(v.blockDeferData, n)
		}
	}

	v.currentBlock = nil

}

func (v *Codegen) genReturnStat(n *parser.ReturnStat) {
	if n.Value == nil {
		v.builder.CreateRetVoid()
	} else {
		v.builder.CreateRet(v.genExpr(n.Value))
	}
}

func (v *Codegen) genBlockStat(n *parser.BlockStat) {
	v.genBlock(n.Block)
}

func (v *Codegen) genCallStat(n *parser.CallStat) {
	v.genExpr(n.Call)
}

func (v *Codegen) genAssignStat(n *parser.AssignStat) {
	v.builder.CreateStore(v.genExpr(n.Assignment), v.genAccessGEP(n.Access))
}

func (v *Codegen) genIfStat(n *parser.IfStat) {
	// Warning to all who tread here:
	// This function is complicated, but theoretically it should never need to
	// be changed again. God help the soul who has to edit this.

	if !v.inFunction {
		panic("tried to gen if stat not in function")
	}

	statTerm := parser.IsNodeTerminating(n)

	var end llvm.BasicBlock
	if !statTerm {
		end = llvm.AddBasicBlock(v.currentFunction, "end")
	}

	for i, expr := range n.Exprs {
		cond := v.genExpr(expr)

		ifTrue := llvm.AddBasicBlock(v.currentFunction, "if_true")
		ifFalse := llvm.AddBasicBlock(v.currentFunction, "if_false")

		v.builder.CreateCondBr(cond, ifTrue, ifFalse)

		v.builder.SetInsertPointAtEnd(ifTrue)
		v.genBlock(n.Bodies[i])

		if !statTerm && !n.Bodies[i].IsTerminating {
			v.builder.CreateBr(end)
		}

		v.builder.SetInsertPointAtEnd(ifFalse)

		if !statTerm {
			end.MoveAfter(ifFalse)
		}
	}

	if n.Else != nil {
		v.genBlock(n.Else)
	}

	if !statTerm && (n.Else == nil || !n.Else.IsTerminating) {
		v.builder.CreateBr(end)
	}

	if !statTerm {
		v.builder.SetInsertPointAtEnd(end)
	}
}

func (v *Codegen) genLoopStat(n *parser.LoopStat) {
	switch n.LoopType {
	case parser.LOOP_TYPE_INFINITE:
		loopBlock := llvm.AddBasicBlock(v.currentFunction, "")
		v.builder.CreateBr(loopBlock)
		v.builder.SetInsertPointAtEnd(loopBlock)

		v.genBlock(n.Body)

		v.builder.CreateBr(loopBlock)
		afterBlock := llvm.AddBasicBlock(v.currentFunction, "")
		v.builder.SetInsertPointAtEnd(afterBlock)
	case parser.LOOP_TYPE_CONDITIONAL:
		evalBlock := llvm.AddBasicBlock(v.currentFunction, "")
		v.builder.CreateBr(evalBlock)

		loopBlock := llvm.AddBasicBlock(v.currentFunction, "")
		afterBlock := llvm.AddBasicBlock(v.currentFunction, "")

		v.builder.SetInsertPointAtEnd(evalBlock)
		cond := v.genExpr(n.Condition)
		v.builder.CreateCondBr(cond, loopBlock, afterBlock)

		v.builder.SetInsertPointAtEnd(loopBlock)
		v.genBlock(n.Body)
		v.builder.CreateBr(evalBlock)

		v.builder.SetInsertPointAtEnd(afterBlock)
	default:
		panic("invalid loop type")
	}
}

func (v *Codegen) genMatchStat(n *parser.MatchStat) {
	// TODO: implement
}

func (v *Codegen) genDecl(n parser.Decl) {
	switch n.(type) {
	case *parser.FunctionDecl:
		v.genFunctionDecl(n.(*parser.FunctionDecl))
	case *parser.UseDecl:
		v.genUseDecl(n.(*parser.UseDecl))
	case *parser.StructDecl:
		//return v.genStructDecl(n.(*parser.StructDecl)) not used
	case *parser.TraitDecl:
		// nothing to gen
	case *parser.ImplDecl:
		v.genImplDecl(n.(*parser.ImplDecl))
	case *parser.EnumDecl:
		// todo
	case *parser.VariableDecl:
		v.genVariableDecl(n.(*parser.VariableDecl), true)
	default:
		v.err("unimplimented decl found: `%s`", n.NodeName())
	}
}

func (v *Codegen) genUseDecl(n *parser.UseDecl) {
	// later
}

func (v *Codegen) genFunctionDecl(n *parser.FunctionDecl) llvm.Value {
	var res llvm.Value

	mangledName := n.Function.MangledName(parser.MANGLE_ARK_UNSTABLE)
	function := v.curFile.Module.NamedFunction(mangledName)
	if function.IsNil() {
		//v.err("genning function `%s` doesn't exist in module", n.Function.Name)
		// hmmmm seems we just ignore this here
	} else {

		if !n.Prototype {
			block := llvm.AddBasicBlock(function, "entry")
			v.builder.SetInsertPointAtEnd(block)

			for i, par := range n.Function.Parameters {
				alloc := v.builder.CreateAlloca(v.typeToLLVMType(par.Variable.Type), par.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE))
				v.variableLookup[par.Variable] = alloc

				v.builder.CreateStore(function.Params()[i], alloc)
			}

			v.inFunction = true
			v.currentFunction = function
			v.genBlock(n.Function.Body)
			v.inFunction = false
		}
	}

	return res
}

func (c *Codegen) genImplDecl(n *parser.ImplDecl) llvm.Value {
	var res llvm.Value

	/*for _, fun := range n.Functions {}*/

	return res
}

func (v *Codegen) genVariableDecl(n *parser.VariableDecl, semicolon bool) llvm.Value {
	var res llvm.Value

	if v.inFunction {
		mangledName := n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE)

		funcEntry := v.currentFunction.EntryBasicBlock()

		// use this builder for the variable alloca
		// this means all allocas go at the start of the function
		// so each variable is only allocated once
		allocBuilder := llvm.NewBuilder()

		if funcEntry == v.builder.GetInsertBlock() {
			allocBuilder.SetInsertPointAtEnd(funcEntry)
		} else {
			allocBuilder.SetInsertPointBefore(funcEntry.LastInstruction())
		}

		alloc := allocBuilder.CreateAlloca(v.typeToLLVMType(n.Variable.Type), mangledName)

		// set allocated memory to zero
		fn := v.curFile.Module.NamedFunction("llvm.memset.p0i8.i32")
		if fn.IsNil() {
			fnType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{llvm.PointerType(llvm.IntType(8), 0), llvm.IntType(8), llvm.IntType(32), llvm.IntType(32), llvm.IntType(1)}, false)
			fn = llvm.AddFunction(v.curFile.Module, "llvm.memset.p0i8.i32", fnType)
		}

		// cast alloc to byte array
		castAlloc := allocBuilder.CreateBitCast(alloc, llvm.PointerType(llvm.IntType(8), 0), "")

		// get type length
		size := v.targetData.TypeAllocSize(v.typeToLLVMType(n.Variable.Type))
		sizeValue := llvm.ConstInt(llvm.IntType(32), size, false)

		// call memset intrinsic
		allocBuilder.CreateCall(fn, []llvm.Value{castAlloc, llvm.ConstInt(llvm.IntType(8), 0, false), sizeValue, llvm.ConstInt(llvm.IntType(32), 0, false), llvm.ConstInt(llvm.IntType(1), 0, false)}, "")

		allocBuilder.Dispose()

		v.variableLookup[n.Variable] = alloc

		if n.Assignment != nil {
			if value := v.genExpr(n.Assignment); !value.IsNil() {
				v.builder.CreateStore(value, alloc)
			}
		}
	} else {
		mangledName := n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE)
		varType := v.typeToLLVMType(n.Variable.Type)
		value := llvm.AddGlobal(v.curFile.Module, varType, mangledName)
		value.SetLinkage(llvm.InternalLinkage)
		value.SetGlobalConstant(!n.Variable.Mutable)
		if n.Assignment != nil {
			value.SetInitializer(v.genExpr(n.Assignment))
		}
		v.variableLookup[n.Variable] = value
	}

	return res
}

func (v *Codegen) genExpr(n parser.Expr) llvm.Value {
	switch n.(type) {
	case *parser.AddressOfExpr:
		return v.genAddressOfExpr(n.(*parser.AddressOfExpr))
	case *parser.RuneLiteral:
		return v.genRuneLiteral(n.(*parser.RuneLiteral))
	case *parser.NumericLiteral:
		return v.genNumericLiteral(n.(*parser.NumericLiteral))
	case *parser.StringLiteral:
		return v.genStringLiteral(n.(*parser.StringLiteral))
	case *parser.BoolLiteral:
		return v.genBoolLiteral(n.(*parser.BoolLiteral))
	case *parser.TupleLiteral:
		return v.genTupleLiteral(n.(*parser.TupleLiteral))
	case *parser.ArrayLiteral:
		return v.genArrayLiteral(n.(*parser.ArrayLiteral))
	case *parser.BinaryExpr:
		return v.genBinaryExpr(n.(*parser.BinaryExpr))
	case *parser.UnaryExpr:
		return v.genUnaryExpr(n.(*parser.UnaryExpr))
	case *parser.CastExpr:
		return v.genCastExpr(n.(*parser.CastExpr))
	case *parser.CallExpr:
		return v.genCallExpr(n.(*parser.CallExpr))
	case *parser.VariableAccessExpr, *parser.StructAccessExpr, *parser.ArrayAccessExpr, *parser.TupleAccessExpr, *parser.DerefAccessExpr:
		return v.genAccessExpr(n)
	case *parser.SizeofExpr:
		return v.genSizeofExpr(n.(*parser.SizeofExpr))
	default:
		log.Debug("codegen", "expr: %s\n", n)
		panic("unimplemented expr")
	}
}

func (v *Codegen) genAddressOfExpr(n *parser.AddressOfExpr) llvm.Value {
	return v.genAccessGEP(n.Access)
}

func (v *Codegen) genAccessExpr(n parser.Expr) llvm.Value {
	return v.builder.CreateLoad(v.genAccessGEP(n), "")
}

func (v *Codegen) genAccessGEP(n parser.Expr) llvm.Value {
	switch n.(type) {
	case *parser.VariableAccessExpr:
		vae := n.(*parser.VariableAccessExpr)
		return v.builder.CreateGEP(v.variableLookup[vae.Variable], []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false)}, "")

	case *parser.StructAccessExpr:
		sae := n.(*parser.StructAccessExpr)

		gep := v.genAccessGEP(sae.Struct)
		index := sae.Struct.GetType().(*parser.StructType).VariableIndex(sae.Variable)
		return v.builder.CreateGEP(gep, []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false), llvm.ConstInt(llvm.Int32Type(), uint64(index), false)}, "")

	case *parser.ArrayAccessExpr:
		aae := n.(*parser.ArrayAccessExpr)

		gep := v.genAccessGEP(aae.Array)
		subscriptExpr := v.genExpr(aae.Subscript)

		sizeGepIndexes := []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false), llvm.ConstInt(llvm.Int32Type(), 0, false)}
		v.genBoundsCheck(v.builder.CreateLoad(v.builder.CreateGEP(gep, sizeGepIndexes, ""), ""), subscriptExpr, aae.Subscript.GetType())

		gepIndexes := []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false), llvm.ConstInt(llvm.Int32Type(), 1, false)}
		gep = v.builder.CreateGEP(gep, gepIndexes, "")

		load := v.builder.CreateLoad(gep, "")

		gepIndexes = []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false), subscriptExpr}
		return v.builder.CreateGEP(load, gepIndexes, "")

	case *parser.TupleAccessExpr:
		tae := n.(*parser.TupleAccessExpr)

		gep := v.genAccessGEP(tae.Tuple)
		index := tae.Index

		return v.builder.CreateGEP(gep, []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false), llvm.ConstInt(llvm.Int32Type(), index, false)}, "")

	case *parser.DerefAccessExpr:
		dae := n.(*parser.DerefAccessExpr)

		return v.genExpr(dae.Expr)

	default:
		panic("unhandled access type")
	}
}

func (v *Codegen) genBoundsCheck(limit llvm.Value, index llvm.Value, indexType parser.Type) {
	segvBlock := llvm.AddBasicBlock(v.currentFunction, "boundscheck_segv")
	endBlock := llvm.AddBasicBlock(v.currentFunction, "boundscheck_end")
	upperCheckBlock := llvm.AddBasicBlock(v.currentFunction, "boundscheck_upper_block")

	tooLow := v.builder.CreateICmp(llvm.IntSGT, llvm.ConstInt(index.Type(), 0, false), index, "boundscheck_lower")
	v.builder.CreateCondBr(tooLow, segvBlock, upperCheckBlock)

	v.builder.SetInsertPointAtEnd(upperCheckBlock)

	// make sure limit and index have same width
	castedLimit := limit
	castedIndex := index
	if index.Type().IntTypeWidth() < limit.Type().IntTypeWidth() {
		if indexType.IsSigned() {
			castedIndex = v.builder.CreateSExt(index, limit.Type(), "")
		} else {
			castedIndex = v.builder.CreateZExt(index, limit.Type(), "")
		}
	} else if index.Type().IntTypeWidth() > limit.Type().IntTypeWidth() {
		castedLimit = v.builder.CreateZExt(limit, index.Type(), "")
	}

	tooHigh := v.builder.CreateICmp(llvm.IntSLE, castedLimit, castedIndex, "boundscheck_upper")
	v.builder.CreateCondBr(tooHigh, segvBlock, endBlock)

	v.builder.SetInsertPointAtEnd(segvBlock)
	v.genRaiseSegfault()
	v.builder.CreateUnreachable()

	v.builder.SetInsertPointAtEnd(endBlock)
}

func (v *Codegen) genRaiseSegfault() {
	fn := v.curFile.Module.NamedFunction("raise")
	intType := v.typeToLLVMType(parser.PRIMITIVE_int)

	if fn.IsNil() {
		fnType := llvm.FunctionType(intType, []llvm.Type{intType}, false)
		fn = llvm.AddFunction(v.curFile.Module, "raise", fnType)
	}

	v.builder.CreateCall(fn, []llvm.Value{llvm.ConstInt(intType, 11, false)}, "segfault")
}

func (v *Codegen) genBoolLiteral(n *parser.BoolLiteral) llvm.Value {
	var num uint64

	if n.Value {
		num = 1
	}

	return llvm.ConstInt(v.typeToLLVMType(n.GetType()), num, true)
}

func (v *Codegen) genRuneLiteral(n *parser.RuneLiteral) llvm.Value {
	return llvm.ConstInt(v.typeToLLVMType(n.GetType()), uint64(n.Value), true)
}

// Allocates a literal array on the stack
func (v *Codegen) genArrayLiteral(n *parser.ArrayLiteral) llvm.Value {
	arrayLLVMType := v.typeToLLVMType(n.Type)
	memberLLVMType := v.typeToLLVMType(n.Type.(parser.ArrayType).MemberType)

	if v.inFunction {
		// allocate backing array
		arrAlloca := v.builder.CreateAlloca(llvm.ArrayType(memberLLVMType, len(n.Members)), "")

		// copy the constant array to the backing array
		for idx, value := range n.Members {
			gep := v.builder.CreateGEP(arrAlloca, []llvm.Value{llvm.ConstInt(llvm.IntType(32), 0, false), llvm.ConstInt(llvm.IntType(32), uint64(idx), false)}, "")
			value := v.genExpr(value)
			v.builder.CreateStore(value, gep)
		}

		// allocate struct
		structAlloca := v.builder.CreateAlloca(arrayLLVMType, "")

		// set the length of the array
		lenGEP := v.builder.CreateGEP(structAlloca, []llvm.Value{llvm.ConstInt(llvm.IntType(32), 0, false), llvm.ConstInt(llvm.IntType(32), 0, false)}, "")
		v.builder.CreateStore(llvm.ConstInt(llvm.IntType(32), uint64(len(n.Members)), false), lenGEP)

		// set the array pointer to the backing array we allocated
		arrGEP := v.builder.CreateGEP(structAlloca, []llvm.Value{llvm.ConstInt(llvm.IntType(32), 0, false), llvm.ConstInt(llvm.IntType(32), 1, false)}, "")
		v.builder.CreateStore(v.builder.CreateBitCast(arrAlloca, llvm.PointerType(llvm.ArrayType(memberLLVMType, 0), 0), ""), arrGEP)

		return v.builder.CreateLoad(structAlloca, "")
	} else {
		backName := fmt.Sprintf("_globarr_back_%d", v.arrayIndex)
		v.arrayIndex++

		backGlob := llvm.AddGlobal(v.curFile.Module, llvm.ArrayType(memberLLVMType, len(n.Members)), backName)
		backGlob.SetLinkage(llvm.InternalLinkage)
		backGlob.SetGlobalConstant(false)

		arrConstVals := make([]llvm.Value, len(n.Members))
		for idx, mem := range n.Members {
			value := v.genExpr(mem)
			if !value.IsConstant() {
				v.err("Encountered non-constant value in global array")
			}
			arrConstVals[idx] = v.genExpr(mem)
		}
		backGlob.SetInitializer(llvm.ConstArray(memberLLVMType, arrConstVals))

		lengthVal := llvm.ConstInt(llvm.IntType(32), uint64(len(n.Members)), false)
		backRef := llvm.ConstBitCast(backGlob, llvm.PointerType(llvm.ArrayType(memberLLVMType, 0), 0))

		return llvm.ConstStruct([]llvm.Value{lengthVal, backRef}, false)
	}
}

func (v *Codegen) genTupleLiteral(n *parser.TupleLiteral) llvm.Value {
	// TODO: Is this optimal?
	// TODO: This probably doesn't work with global variables
	var values []llvm.Value
	for _, mem := range n.Members {
		values = append(values, v.genExpr(mem))
	}

	return llvm.ConstStruct(values, false)
}

func (v *Codegen) genNumericLiteral(n *parser.NumericLiteral) llvm.Value {
	if n.Type.IsFloatingType() {
		return llvm.ConstFloat(v.typeToLLVMType(n.Type), n.AsFloat())
	} else {
		return llvm.ConstInt(v.typeToLLVMType(n.Type), n.AsInt(), false)
	}
}

func (v *Codegen) genStringLiteral(n *parser.StringLiteral) llvm.Value {
	return v.builder.CreateGlobalStringPtr(n.Value, "str")
}

func (v *Codegen) genBinaryExpr(n *parser.BinaryExpr) llvm.Value {
	var res llvm.Value

	lhand := v.genExpr(n.Lhand)
	rhand := v.genExpr(n.Rhand)

	if lhand.IsNil() || rhand.IsNil() {
		v.err("invalid binary expr")
	} else {
		switch n.Op {
		// Arithmetic
		case parser.BINOP_ADD:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFAdd(lhand, rhand, "")
			} else {
				return v.builder.CreateAdd(lhand, rhand, "")
			}
		case parser.BINOP_SUB:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFSub(lhand, rhand, "")
			} else {
				return v.builder.CreateSub(lhand, rhand, "")
			}
		case parser.BINOP_MUL:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFMul(lhand, rhand, "")
			} else {
				return v.builder.CreateMul(lhand, rhand, "")
			}
		case parser.BINOP_DIV:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFDiv(lhand, rhand, "")
			} else {
				if n.GetType().(parser.PrimitiveType).IsSigned() {
					return v.builder.CreateSDiv(lhand, rhand, "")
				} else {
					return v.builder.CreateUDiv(lhand, rhand, "")
				}
			}
		case parser.BINOP_MOD:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFRem(lhand, rhand, "")
			} else {
				if n.GetType().(parser.PrimitiveType).IsSigned() {
					return v.builder.CreateSRem(lhand, rhand, "")
				} else {
					return v.builder.CreateURem(lhand, rhand, "")
				}
			}

		// Comparison
		case parser.BINOP_GREATER, parser.BINOP_LESS, parser.BINOP_GREATER_EQ, parser.BINOP_LESS_EQ, parser.BINOP_EQ, parser.BINOP_NOT_EQ:
			if n.Lhand.GetType().IsFloatingType() {
				return v.builder.CreateFCmp(comparisonOpToFloatPredicate(n.Op), lhand, rhand, "")
			} else {
				return v.builder.CreateICmp(comparisonOpToIntPredicate(n.Op, n.Lhand.GetType().IsSigned()), lhand, rhand, "")
			}

		// Bitwise
		case parser.BINOP_BIT_AND:
			return v.builder.CreateAnd(lhand, rhand, "")
		case parser.BINOP_BIT_OR:
			return v.builder.CreateOr(lhand, rhand, "")
		case parser.BINOP_BIT_XOR:
			return v.builder.CreateXor(lhand, rhand, "")
		case parser.BINOP_BIT_LEFT:
			return v.builder.CreateShl(lhand, rhand, "")
		case parser.BINOP_BIT_RIGHT:
			// TODO make sure both operands are same type (create type cast here?)
			// TODO in semantic.go, make sure rhand is *unsigned* (LLVM always treats it that way)
			// TODO doc this
			if n.Lhand.GetType().IsSigned() {
				return v.builder.CreateAShr(lhand, rhand, "")
			} else {
				return v.builder.CreateLShr(lhand, rhand, "")
			}

		// Logical
		case parser.BINOP_LOG_AND:
			return v.builder.CreateAnd(lhand, rhand, "")
		case parser.BINOP_LOG_OR:
			return v.builder.CreateOr(lhand, rhand, "")

		default:
			panic("umimplented binop")
		}
	}

	return res
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
		return v.builder.CreateNot(expr, "")
	case parser.UNOP_NEGATIVE:
		return v.builder.CreateNeg(expr, "")
	default:
		panic("unimplimented unary op")
	}
}

func (v *Codegen) genCastExpr(n *parser.CastExpr) llvm.Value {
	if n.GetType() == n.Expr.GetType() {
		return v.genExpr(n.Expr)
	}

	exprType := n.Expr.GetType()
	castType := n.GetType()

	if exprType.IsIntegerType() || exprType == parser.PRIMITIVE_rune {
		if _, ok := castType.(parser.PointerType); ok {
			// TODO: This might not be right in all cases
			return v.builder.CreateIntToPtr(v.genExpr(n.Expr), v.typeToLLVMType(castType), "")
		} else if castType.IsIntegerType() || castType == parser.PRIMITIVE_rune {
			exprBits := v.typeToLLVMType(exprType).IntTypeWidth()
			castBits := v.typeToLLVMType(castType).IntTypeWidth()
			if exprBits == castBits {
				return v.genExpr(n.Expr)
			} else if exprBits > castBits {
				/*shiftConst := llvm.ConstInt(v.typeToLLVMType(exprType), uint64(exprBits-castBits), false)
				shl := v.builder.CreateShl(v.genExpr(n.Expr), shiftConst, "")
				shr := v.builder.CreateAShr(shl, shiftConst, "")
				return v.builder.CreateTrunc(shr, v.typeToLLVMType(castType), "")*/
				return v.builder.CreateTrunc(v.genExpr(n.Expr), v.typeToLLVMType(castType), "") // TODO get this to work right!
			} else if exprBits < castBits {
				if exprType.IsSigned() {
					return v.builder.CreateSExt(v.genExpr(n.Expr), v.typeToLLVMType(castType), "") // TODO doc this
				} else {
					return v.builder.CreateZExt(v.genExpr(n.Expr), v.typeToLLVMType(castType), "")
				}
			}
		} else if castType.IsFloatingType() {
			if exprType.IsSigned() {
				return v.builder.CreateSIToFP(v.genExpr(n.Expr), v.typeToLLVMType(castType), "")
			} else {
				return v.builder.CreateUIToFP(v.genExpr(n.Expr), v.typeToLLVMType(castType), "")
			}
		}
	} else if exprType.IsFloatingType() {
		if castType.IsIntegerType() || castType == parser.PRIMITIVE_rune {
			if exprType.IsSigned() {
				return v.builder.CreateFPToSI(v.genExpr(n.Expr), v.typeToLLVMType(castType), "")
			} else {
				return v.builder.CreateFPToUI(v.genExpr(n.Expr), v.typeToLLVMType(castType), "")
			}
		} else if castType.IsFloatingType() {
			exprBits := floatTypeBits(exprType.(parser.PrimitiveType))
			castBits := floatTypeBits(castType.(parser.PrimitiveType))
			if exprBits == castBits {
				return v.genExpr(n.Expr)
			} else if exprBits > castBits {
				return v.builder.CreateFPTrunc(v.genExpr(n.Expr), v.typeToLLVMType(castType), "") // TODO get this to work right!
			} else if exprBits < castBits {
				return v.builder.CreateFPExt(v.genExpr(n.Expr), v.typeToLLVMType(castType), "")
			}
		}
	}

	panic("unimplimented typecast")
}

func floatTypeBits(ty parser.PrimitiveType) int {
	switch ty {
	case parser.PRIMITIVE_f32:
		return 32
	case parser.PRIMITIVE_f64:
		return 64
	case parser.PRIMITIVE_f128:
		return 128
	default:
		panic("this isn't a float type")
	}
}

func (v *Codegen) genCallExprWithArgs(n *parser.CallExpr, args []llvm.Value) llvm.Value {
	// todo dont use attributes, use that C:: shit
	cBinding := false
	if n.Function.Attrs != nil {
		cBinding = n.Function.Attrs.Contains("c")
	}

	// eww
	mangledName := n.Function.MangledName(parser.MANGLE_ARK_UNSTABLE)
	functionName := mangledName
	if cBinding {
		functionName = n.Function.Name
	}

	function := v.curFile.Module.NamedFunction(functionName)
	if function.IsNil() {
		v.declareFunctionDecl(&parser.FunctionDecl{Function: n.Function, Prototype: true})

		if v.curFile.Module.NamedFunction(functionName).IsNil() {
			panic("how did this happen")
		}
	}

	call := v.builder.CreateCall(function, args, "")

	call.SetInstructionCallConv(function.FunctionCallConv())

	return call
}

func (v *Codegen) genCallExpr(n *parser.CallExpr) llvm.Value {
	args := make([]llvm.Value, 0, len(n.Arguments))
	for _, arg := range n.Arguments {
		args = append(args, v.genExpr(arg))
	}

	return v.genCallExprWithArgs(n, args)
}

func (v *Codegen) genSizeofExpr(n *parser.SizeofExpr) llvm.Value {
	if n.Expr != nil {
		gep := v.builder.CreateGEP(llvm.ConstNull(llvm.PointerType(v.typeToLLVMType(n.Expr.GetType()), 0)), []llvm.Value{llvm.ConstInt(llvm.IntType(32), 1, false)}, "")
		return v.builder.CreatePtrToInt(gep, v.typeToLLVMType(n.GetType()), "sizeof")
	} else {
		// we have a type
		panic("can't do this yet")
	}
}

func (v *Codegen) typeToLLVMType(typ parser.Type) llvm.Type {
	switch typ.(type) {
	case parser.PrimitiveType:
		return primitiveTypeToLLVMType(typ.(parser.PrimitiveType))
	case *parser.StructType:
		return v.structTypeToLLVMType(typ.(*parser.StructType))
	case parser.PointerType:
		return llvm.PointerType(v.typeToLLVMType(typ.(parser.PointerType).Addressee), 0)
	case parser.ArrayType:
		return v.arrayTypeToLLVMType(typ.(parser.ArrayType))
	case *parser.TupleType:
		return v.tupleTypeToLLVMType(typ.(*parser.TupleType))
	default:
		panic("Unimplemented type category in LLVM codegen")
	}
}

func (v *Codegen) tupleTypeToLLVMType(typ *parser.TupleType) llvm.Type {
	// TODO: Maybe move to lookup table like struct
	var fields []llvm.Type
	for _, mem := range typ.Members {
		fields = append(fields, v.typeToLLVMType(mem))
	}

	return llvm.StructType(fields, false)
}

func (v *Codegen) arrayTypeToLLVMType(typ parser.ArrayType) llvm.Type {
	fields := []llvm.Type{llvm.IntType(32), llvm.PointerType(llvm.ArrayType(v.typeToLLVMType(typ.MemberType), 0), 0)}

	return llvm.StructType(fields, false)
}

func (v *Codegen) structTypeToLLVMType(typ *parser.StructType) llvm.Type {
	if val, ok := v.structLookup_UseHelperFunction[typ]; ok {
		return val
	} else {
		panic("why no struct type entry")
	}
}

func primitiveTypeToLLVMType(typ parser.PrimitiveType) llvm.Type {
	switch typ {
	case parser.PRIMITIVE_int, parser.PRIMITIVE_uint:
		return llvm.IntType(intSize * 8)
	case parser.PRIMITIVE_s8, parser.PRIMITIVE_u8:
		return llvm.IntType(8)
	case parser.PRIMITIVE_s16, parser.PRIMITIVE_u16:
		return llvm.IntType(16)
	case parser.PRIMITIVE_s32, parser.PRIMITIVE_u32:
		return llvm.IntType(32)
	case parser.PRIMITIVE_s64, parser.PRIMITIVE_u64:
		return llvm.IntType(64)
	case parser.PRIMITIVE_i128, parser.PRIMITIVE_u128:
		return llvm.IntType(128)

	case parser.PRIMITIVE_f32:
		return llvm.FloatType()
	case parser.PRIMITIVE_f64:
		return llvm.DoubleType()
	case parser.PRIMITIVE_f128:
		return llvm.FP128Type()

	case parser.PRIMITIVE_rune: // runes are signed 32-bit int
		return llvm.IntType(32)
	case parser.PRIMITIVE_bool:
		return llvm.IntType(1)
	case parser.PRIMITIVE_str:
		return llvm.PointerType(llvm.IntType(8), 0)

	case parser.PRIMITIVE_void:
		return llvm.VoidType()

	default:
		panic("Unimplemented primitive type in LLVM codegen")
	}
}
