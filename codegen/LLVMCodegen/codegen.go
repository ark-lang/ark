package LLVMCodegen

import (
	"C"
	"fmt"
	"os"
	"os/exec"
	"time"
	"unsafe"

	"github.com/ark-lang/ark/parser"
	"github.com/ark-lang/ark/util"

	"llvm.org/llvm/bindings/go/llvm"
)

const intSize = int(unsafe.Sizeof(C.int(0)))

type Codegen struct {
	input   []*parser.Module
	curFile *parser.Module

	builder                        llvm.Builder
	variableLookup                 map[*parser.Variable]llvm.Value
	structLookup_UseHelperFunction map[*parser.StructType]llvm.Type // use getStructDecl

	OutputName string
	OutputAsm  bool
	CCArgs     []string

	modules map[string]*parser.Module

	inFunction      bool
	currentFunction llvm.Value
}

func (v *Codegen) err(err string, stuff ...interface{}) {
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_CODEGEN)
}

func (v *Codegen) createBitcode(file *parser.Module) string {
	filename := file.Name + ".bc"

	fileHandle, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		v.err("Couldn't create bitcode file "+filename+": `%s`", err.Error())
	}
	defer fileHandle.Close()

	if err := llvm.WriteBitcodeToFile(file.Module, fileHandle); err != nil {
		v.err("failed to write bitcode to file for "+file.Name+": `%s`", err.Error())
	}

	return filename
}

func (v *Codegen) bitcodeToASM(filename string) string {
	asmName := filename + ".s"
	cmd := exec.Command("llc", filename, "-o", filename+".s")
	if out, err := cmd.CombinedOutput(); err != nil {
		v.err("Failed to convert bitcode to assembly: `%s`\n%s", err.Error(), string(out))
	}

	return asmName
}

func (v *Codegen) createBinary() {
	linkArgs := append(v.CCArgs, "-fno-PIE", "-nodefaultlibs", "-lc")
	asmFiles := []string{}

	for _, file := range v.input {
		name := v.createBitcode(file)
		asmName := v.bitcodeToASM(name)
		asmFiles = append(asmFiles, asmName)

		if err := os.Remove(name); err != nil {
			v.err("Failed to remove "+name+": `%s`", err.Error())
		}

		linkArgs = append(linkArgs, asmName)
	}

	if v.OutputAsm {
		return
	}

	if v.OutputName == "" {
		panic("OutputName is empty")
	}
	linkArgs = append(linkArgs, "-o", v.OutputName)

	cmd := exec.Command("cc", linkArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		v.err("failed to link object files: `%s`\n%s", err.Error(), string(out))
	}

	for _, asmFile := range asmFiles {
		os.Remove(asmFile)
	}
}

func (v *Codegen) Generate(input []*parser.Module, modules map[string]*parser.Module, verbose bool) {
	v.input = input
	v.builder = llvm.NewBuilder()
	v.variableLookup = make(map[*parser.Variable]llvm.Value)
	v.structLookup_UseHelperFunction = make(map[*parser.StructType]llvm.Type)

	if verbose {
		fmt.Println(util.TEXT_BOLD + util.TEXT_GREEN + "Started codegenning" + util.TEXT_RESET)
	}
	t := time.Now()

	passManager := llvm.NewPassManager()
	passBuilder := llvm.NewPassManagerBuilder()
	passBuilder.SetOptLevel(3)
	//passBuilder.Populate(passManager) //leave this off until the compiler is better

	v.modules = make(map[string]*parser.Module)

	for _, infile := range input {
		infile.Module = llvm.NewModule(infile.Name)
		v.curFile = infile

		fmt.Println("adding module " + v.curFile.Name)
		v.modules[v.curFile.Name] = v.curFile

		v.declareDecls(infile.Nodes)

		for _, node := range infile.Nodes {
			v.genNode(node)
		}

		if verbose {
			infile.Module.Dump()
		}

		if err := llvm.VerifyModule(infile.Module, llvm.ReturnStatusAction); err != nil {
			v.err("%s", err.Error())
		}

		passManager.Run(infile.Module)
	}

	passManager.Dispose()

	v.createBinary()

	dur := time.Since(t)
	if verbose {
		fmt.Printf(util.TEXT_BOLD+util.TEXT_GREEN+"Finished codegenning"+util.TEXT_RESET+" (%.2fms)\n",
			float32(dur.Nanoseconds())/1000000)
	}
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
			attributes := n.Function.Attrs

			// todo hashmap or some shit
			for _, attr := range attributes {
				switch attr.Key {
				case "c":
					cBinding = true
				default:
					// do nothing
				}
			}
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
	default:
		panic("unimplimented stat")
	}
}

func (v *Codegen) genBlock(n *parser.Block) {
	for _, x := range n.Nodes {
		v.genNode(x)
	}
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
	if n.Access != nil {
		v.builder.CreateStore(v.genExpr(n.Assignment), v.genAddressOfExpr(&parser.AddressOfExpr{Access: n.Access})) // TODO do this better
	} else {
		v.genDerefExpr(n.Deref)
	}
}

func (v *Codegen) genIfStat(n *parser.IfStat) {
	if !v.inFunction {
		panic("tried to gen if stat not in function")
	}

	end := llvm.AddBasicBlock(v.currentFunction, "")

	for i, expr := range n.Exprs {
		cond := v.genExpr(expr)

		ifTrue := llvm.AddBasicBlock(v.currentFunction, "")
		ifFalse := llvm.AddBasicBlock(v.currentFunction, "")

		v.builder.CreateCondBr(cond, ifTrue, ifFalse)

		v.builder.SetInsertPointAtEnd(ifTrue)
		for _, node := range n.Bodies[i].Nodes {
			v.genNode(node)
		}

		if len(n.Bodies[i].Nodes) == 0 {
			v.builder.CreateBr(end)
		} else {
			switch n.Bodies[i].Nodes[len(n.Bodies[i].Nodes)-1].(type) {
			case *parser.ReturnStat: // do nothing
			default:
				v.builder.CreateBr(end)
			}
		}

		v.builder.SetInsertPointAtEnd(ifFalse)
		end.MoveAfter(ifFalse)
	}

	if n.Else != nil {
		for _, node := range n.Else.Nodes {
			v.genNode(node)
		}
	}

	v.builder.CreateBr(end)

	v.builder.SetInsertPointAtEnd(end)
}

func (v *Codegen) genLoopStat(n *parser.LoopStat) {
	switch n.LoopType {
	case parser.LOOP_TYPE_INFINITE:
		loopBlock := llvm.AddBasicBlock(v.currentFunction, "")
		v.builder.CreateBr(loopBlock)
		v.builder.SetInsertPointAtEnd(loopBlock)

		for _, node := range n.Body.Nodes {
			v.genNode(node)
		}

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
		for _, node := range n.Body.Nodes {
			v.genNode(node)
		}
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
		// TODO
	case *parser.EnumDecl:
		// todo
	case *parser.VariableDecl:
		v.genVariableDecl(n.(*parser.VariableDecl), true)
	case *parser.ModuleDecl:
		// TODO!!
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
			for _, stat := range n.Function.Body.Nodes {
				v.genNode(stat)
			}
			v.inFunction = false

			retType := llvm.VoidType()
			if n.Function.ReturnType != nil {
				retType = v.typeToLLVMType(n.Function.ReturnType)
			}

			// function returns void, lets return void
			// unless its a prototype obviously...
			if retType == llvm.VoidType() && !n.Prototype {
				v.builder.CreateRetVoid()
			}
		}
	}

	return res
}

func (v *Codegen) genVariableDecl(n *parser.VariableDecl, semicolon bool) llvm.Value {
	var res llvm.Value

	// if n.Variable.Mutable

	// TODO global vars

	mangledName := n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE)
	alloc := v.builder.CreateAlloca(v.typeToLLVMType(n.Variable.Type), mangledName)
	v.variableLookup[n.Variable] = alloc

	if n.Assignment != nil {
		if value := v.genExpr(n.Assignment); !value.IsNil() {
			v.builder.CreateStore(value, alloc)
		}
	}

	return res
}

func (v *Codegen) genExpr(n parser.Expr) llvm.Value {
	switch n.(type) {
	case *parser.AddressOfExpr:
		return v.genAddressOfExpr(n.(*parser.AddressOfExpr))
	case *parser.RuneLiteral:
		return v.genRuneLiteral(n.(*parser.RuneLiteral))
	case *parser.IntegerLiteral:
		return v.genIntegerLiteral(n.(*parser.IntegerLiteral))
	case *parser.FloatingLiteral:
		return v.genFloatingLiteral(n.(*parser.FloatingLiteral))
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
	case *parser.AccessExpr:
		return v.genAccessExpr(n.(*parser.AccessExpr))
	case *parser.DerefExpr:
		return v.genDerefExpr(n.(*parser.DerefExpr))
	case *parser.BracketExpr:
		return v.genBracketExpr(n.(*parser.BracketExpr))
	case *parser.SizeofExpr:
		return v.genSizeofExpr(n.(*parser.SizeofExpr))
	default:
		panic("unimplemented expr")
	}
}

func (v *Codegen) genAddressOfExpr(n *parser.AddressOfExpr) llvm.Value {
	gep := v.builder.CreateGEP(v.variableLookup[n.Access.Accesses[0].Variable], []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false)}, "")

	for i := 0; i < len(n.Access.Accesses); i++ {

		switch n.Access.Accesses[i].AccessType {
		case parser.ACCESS_ARRAY:
			gepIndexes := []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false), llvm.ConstInt(llvm.Int32Type(), 1, false)}
			gep = v.builder.CreateGEP(gep, gepIndexes, "")

			load := v.builder.CreateLoad(gep, "")

			// TODO check that access is in bounds!

			gepIndexes = []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false), v.genExpr(n.Access.Accesses[i].Subscript)}
			gep = v.builder.CreateGEP(load, gepIndexes, "")

		case parser.ACCESS_VARIABLE:
			// nothing to do

		case parser.ACCESS_STRUCT:
			index := n.Access.Accesses[i].Variable.Type.(*parser.StructType).VariableIndex(n.Access.Accesses[i+1].Variable)
			gep = v.builder.CreateGEP(gep, []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 0, false), llvm.ConstInt(llvm.Int32Type(), uint64(index), false)}, "")

		default:
			panic("")
		}
	}

	return gep
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
	memberLLVMType := v.typeToLLVMType(n.Type.(parser.ArrayType).MemberType)

	// allocate backing array
	arrAlloca := v.builder.CreateArrayAlloca(llvm.ArrayType(memberLLVMType, len(n.Members)), llvm.ConstInt(llvm.IntType(32), uint64(len(n.Members)), false), "")

	// allocate the array object
	structAlloca := v.builder.CreateAlloca(v.typeToLLVMType(n.Type), "")

	// set the length of the array
	lenGEP := v.builder.CreateGEP(structAlloca, []llvm.Value{llvm.ConstInt(llvm.IntType(32), 0, false), llvm.ConstInt(llvm.IntType(32), 0, false)}, "")
	v.builder.CreateStore(llvm.ConstInt(llvm.IntType(32), uint64(len(n.Members)), false), lenGEP)

	// set the array pointer to the backing array we allocated
	arrGEP := v.builder.CreateGEP(structAlloca, []llvm.Value{llvm.ConstInt(llvm.IntType(32), 0, false), llvm.ConstInt(llvm.IntType(32), 1, false)}, "")
	v.builder.CreateStore(v.builder.CreateBitCast(arrAlloca, llvm.PointerType(llvm.ArrayType(memberLLVMType, 0), 0), ""), arrGEP)

	// copy the constant array to the backing array
	arrConstVals := make([]llvm.Value, 0, len(n.Members))
	for _, mem := range n.Members {
		arrConstVals = append(arrConstVals, v.genExpr(mem))
	}
	arrConst := llvm.ConstArray(llvm.ArrayType(memberLLVMType, len(n.Members)), arrConstVals)
	v.builder.CreateStore(arrConst, arrAlloca)

	return v.builder.CreateLoad(structAlloca, "")
}

func (v *Codegen) genTupleLiteral(n *parser.TupleLiteral) llvm.Value {
	panic("not done yet")
}

func (v *Codegen) genIntegerLiteral(n *parser.IntegerLiteral) llvm.Value {
	return llvm.ConstInt(v.typeToLLVMType(n.Type), n.Value, false)
}

func (v *Codegen) genFloatingLiteral(n *parser.FloatingLiteral) llvm.Value {
	return llvm.ConstFloat(v.typeToLLVMType(n.Type), n.Value)
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
			if n.GetType().IsFloatingType() {
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
		if castType.IsIntegerType() || castType == parser.PRIMITIVE_rune {
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

func (v *Codegen) genCallExpr(n *parser.CallExpr) llvm.Value {
	// todo dont use attributes, use that C:: shit
	cBinding := false
	for _, attr := range n.Function.Attrs {
		switch attr.Key {
		case "c":
			cBinding = true
		default:
			// whatever
		}
	}

	// eww
	mangledName := n.Function.MangledName(parser.MANGLE_ARK_UNSTABLE)
	functionName := mangledName
	if cBinding {
		functionName = n.Function.Name
	}

swag:
	function := v.curFile.Module.NamedFunction(functionName)
	if function.IsNil() {
		v.declareFunctionDecl(&parser.FunctionDecl{Function: n.Function, Prototype: true})
		goto swag
	}

	numOfArguments := len(n.Arguments)
	args := make([]llvm.Value, numOfArguments)
	for i, arg := range n.Arguments {
		args[i] = v.genExpr(arg)
	}

	call := v.builder.CreateCall(function, args, "")

	call.SetInstructionCallConv(function.FunctionCallConv())

	return call
}

func (v *Codegen) genAccessExpr(n *parser.AccessExpr) llvm.Value {
	return v.builder.CreateLoad(v.genAddressOfExpr(&parser.AddressOfExpr{Access: n}), "") // TODO do this better
}

func (v *Codegen) genDerefExpr(n *parser.DerefExpr) llvm.Value {
	return v.builder.CreateLoad(v.genExpr(n.Expr), "")
}

func (v *Codegen) genBracketExpr(n *parser.BracketExpr) llvm.Value {
	return v.genExpr(n.Expr)
}

func (v *Codegen) genSizeofExpr(n *parser.SizeofExpr) llvm.Value {
	if n.Expr != nil {
		gep := v.builder.CreateGEP(llvm.ConstNull(llvm.PointerType(v.typeToLLVMType(n.Expr.GetType()), 0)), []llvm.Value{llvm.ConstInt(llvm.Int32Type(), 1, false)}, "")
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
	default:
		panic("Unimplemented type category in LLVM codegen")
	}
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

	default:
		panic("Unimplemented primitive type in LLVM codegen")
	}
}
