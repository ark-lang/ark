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
	input   []*parser.File
	curFile *parser.File

	builder        llvm.Builder
	variableLookup map[*parser.Variable]llvm.Value

	OutputName string
}

func (v *Codegen) err(err string, stuff ...interface{}) {
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(2)
}

func (v *Codegen) createBitcode(file *parser.File) string {
	filename := file.Name + ".bc"
	if err := llvm.VerifyModule(file.Module, llvm.ReturnStatusAction); err != nil {
		v.err(err.Error())
	}

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
	linkArgs := []string{}
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

	if v.OutputName == "" {
		panic("OutputName is empty")
	}
	linkArgs = append(linkArgs, "-o", v.OutputName)

	cmd := exec.Command("cc", linkArgs...)
	if err := cmd.Run(); err != nil {
		v.err("failed to link object files: `%s`", err.Error())
	}

	for _, asmFile := range asmFiles {
		os.Remove(asmFile)
	}
}

func (v *Codegen) Generate(input []*parser.File, verbose bool) {
	v.input = input
	v.builder = llvm.NewBuilder()
	v.variableLookup = make(map[*parser.Variable]llvm.Value)

	for _, infile := range input {
		infile.Module = llvm.NewModule(infile.Name)
		v.curFile = infile

		if verbose {
			fmt.Println(util.TEXT_BOLD+util.TEXT_GREEN+"Started codegenning"+util.TEXT_RESET, infile.Name)
		}
		t := time.Now()

		for _, node := range infile.Nodes {
			v.genNode(node)
		}

		infile.Module.Dump()
		v.createBinary()

		dur := time.Since(t)
		if verbose {
			fmt.Printf(util.TEXT_BOLD+util.TEXT_GREEN+"Finished codegenning"+util.TEXT_RESET+" %s (%.2fms)\n",
				infile.Name, float32(dur.Nanoseconds())/1000000)
		}
	}
}

func (v *Codegen) genNode(n parser.Node) llvm.Value {
	var res llvm.Value

	switch n.(type) {
	case parser.Decl:
		return v.genDecl(n.(parser.Decl))
	case parser.Expr:
		return v.genExpr(n.(parser.Expr))
	case parser.Stat:
		return v.genStat(n.(parser.Stat))
	}

	return res
}

func (v *Codegen) genStat(n parser.Stat) llvm.Value {
	switch n.(type) {
	case *parser.ReturnStat:
		return v.genReturnStat(n.(*parser.ReturnStat))
	case *parser.CallStat:
		return v.genCallStat(n.(*parser.CallStat))
	case *parser.AssignStat:
		return v.genAssignStat(n.(*parser.AssignStat))
	case *parser.IfStat:
		return v.genIfStat(n.(*parser.IfStat))
	case *parser.LoopStat:
		return v.genLoopStat(n.(*parser.LoopStat))
	default:
		panic("unimplimented stat")
	}
}

func (v *Codegen) genReturnStat(n *parser.ReturnStat) llvm.Value {
	var res llvm.Value
	if n.Value == nil {
		res = v.builder.CreateRetVoid()
	} else {
		res = v.builder.CreateRet(v.genExpr(n.Value))
	}
	return res
}

func (v *Codegen) genCallStat(n *parser.CallStat) llvm.Value {
	return v.genExpr(n.Call)
}

func (v *Codegen) genAssignStat(n *parser.AssignStat) llvm.Value {

	alloca := v.variableLookup[n.Access.Variable]
	store := v.builder.CreateStore(v.genExpr(n.Assignment), alloca)
	return store
}

func (v *Codegen) genIfStat(n *parser.IfStat) llvm.Value {
	var res llvm.Value

	for _, expr := range n.Exprs {
		v.genExpr(expr)
	}

	return res
}

func (v *Codegen) genLoopStat(n *parser.LoopStat) llvm.Value {
	var res llvm.Value

	switch n.LoopType {
	case parser.LOOP_TYPE_INFINITE:
	case parser.LOOP_TYPE_CONDITIONAL:
		v.genExpr(n.Condition)
	default:
		panic("invalid loop type")
	}

	return res
}

func (v *Codegen) genDecl(n parser.Decl) llvm.Value {
	switch n.(type) {
	case *parser.FunctionDecl:
		return v.genFunctionDecl(n.(*parser.FunctionDecl))
	case *parser.StructDecl:
		return v.genStructDecl(n.(*parser.StructDecl))
	case *parser.VariableDecl:
		return v.genVariableDecl(n.(*parser.VariableDecl), true)
	default:
		panic("unimplimented decl")
	}
}

func (v *Codegen) genFunctionDecl(n *parser.FunctionDecl) llvm.Value {
	var res llvm.Value

	mangledName := n.Function.MangledName(parser.MANGLE_ARK_UNSTABLE)
	function := v.curFile.Module.NamedFunction(mangledName)
	if !function.IsNil() {
		v.err("function `%s` already exists in module", n.Function.Name)
		return res
	} else {
		numOfParams := len(n.Function.Parameters)
		params := make([]llvm.Type, numOfParams)
		for i, par := range n.Function.Parameters {
			params[i] = typeToLLVMType(par.Variable.Type)
		}

		// assume it's void
		funcTypeRaw := llvm.VoidType()

		// oo theres a type, let's try figure it out
		if n.Function.ReturnType != nil {
			funcTypeRaw = typeToLLVMType(n.Function.ReturnType)
		}

		// create the function type
		funcType := llvm.FunctionType(funcTypeRaw, params, false)

		// add that shit
		function = llvm.AddFunction(v.curFile.Module, mangledName, funcType)

		// do some magical shit for later
		for i := 0; i < numOfParams; i++ {
			funcParam := function.Param(i)
			funcParam.SetName(n.Function.Parameters[i].Variable.MangledName(parser.MANGLE_ARK_UNSTABLE))
			// maybe store it in a hashmap somewhere?
		}

		block := llvm.AddBasicBlock(function, "entry")
		v.builder.SetInsertPointAtEnd(block)

		for _, stat := range n.Function.Body.Nodes {
			v.genNode(stat)
		}

		// function returns void, lets return void...
		if funcTypeRaw == llvm.VoidType() {
			v.builder.CreateRetVoid()
		}

		// loop thru block and gen statements

		return function
	}
}

func (v *Codegen) genStructDecl(n *parser.StructDecl) llvm.Value {
	numOfFields := len(n.Struct.Variables)
	fields := make([]llvm.Type, numOfFields)
	packed := false

	for i, member := range n.Struct.Variables {
		memberType := typeToLLVMType(member.Variable.Type)
		fields[i] = memberType
	}

	structure := llvm.StructType(fields, packed)
	result := llvm.AddGlobal(v.curFile.Module, structure, n.Struct.MangledName(parser.MANGLE_ARK_UNSTABLE))
	return result
}

func (v *Codegen) genVariableDecl(n *parser.VariableDecl, semicolon bool) llvm.Value {
	var res llvm.Value

	// if n.Variable.Mutable

	mangledName := n.Variable.MangledName(parser.MANGLE_ARK_UNSTABLE)
	alloc := v.builder.CreateAlloca(typeToLLVMType(n.Variable.Type), mangledName)
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
	case *parser.RuneLiteral:
		return v.genRuneLiteral(n.(*parser.RuneLiteral))
	case *parser.IntegerLiteral:
		return v.genIntegerLiteral(n.(*parser.IntegerLiteral))
	case *parser.FloatingLiteral:
		return v.genFloatingLiteral(n.(*parser.FloatingLiteral))
	case *parser.StringLiteral:
		return v.genStringLiteral(n.(*parser.StringLiteral))
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
	default:
		panic("unimplemented expr")
	}
}

func (v *Codegen) genRuneLiteral(n *parser.RuneLiteral) llvm.Value {
	var res llvm.Value
	return res
}

func (v *Codegen) genIntegerLiteral(n *parser.IntegerLiteral) llvm.Value {
	return llvm.ConstInt(typeToLLVMType(n.Type), n.Value, false)
}

func (v *Codegen) genFloatingLiteral(n *parser.FloatingLiteral) llvm.Value {
	var res llvm.Value

	return res
}

func (v *Codegen) genStringLiteral(n *parser.StringLiteral) llvm.Value {
	val := llvm.ConstString(n.Value, true)
	str := llvm.AddGlobal(v.curFile.Module, val.Type(), "")
	str.SetLinkage(llvm.InternalLinkage)
	str.SetGlobalConstant(true)
	str.SetInitializer(val)
	str = llvm.ConstBitCast(str, llvm.PointerType(llvm.IntType(32), 0))
	global := llvm.AddGlobal(v.curFile.Module, val.Type(), "")
	global.SetGlobalConstant(true)
	global.SetLinkage(llvm.InternalLinkage)
	global.SetInitializer(str)
	return global
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
				return v.builder.CreateFAdd(lhand, rhand, "tmp")
			} else {
				return v.builder.CreateAdd(lhand, rhand, "tmp")
			}
		case parser.BINOP_SUB:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFSub(lhand, rhand, "tmp")
			} else {
				return v.builder.CreateSub(lhand, rhand, "tmp")
			}
		case parser.BINOP_MUL:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFMul(lhand, rhand, "tmp")
			} else {
				return v.builder.CreateMul(lhand, rhand, "tmp")
			}
		case parser.BINOP_DIV:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFDiv(lhand, rhand, "tmp")
			} else {
				if n.GetType().(parser.PrimitiveType).IsSigned() {
					return v.builder.CreateSDiv(lhand, rhand, "tmp")
				} else {
					return v.builder.CreateUDiv(lhand, rhand, "tmp")
				}
			}
		case parser.BINOP_MOD:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFRem(lhand, rhand, "tmp")
			} else {
				if n.GetType().(parser.PrimitiveType).IsSigned() {
					return v.builder.CreateSRem(lhand, rhand, "tmp")
				} else {
					return v.builder.CreateURem(lhand, rhand, "tmp")
				}
			}

		// Comparison
		case parser.BINOP_GREATER, parser.BINOP_LESS, parser.BINOP_GREATER_EQ, parser.BINOP_LESS_EQ, parser.BINOP_EQ, parser.BINOP_NOT_EQ:
			if n.GetType().IsFloatingType() {
				return v.builder.CreateFCmp(comparisonOpToFloatPredicate(n.Op), lhand, rhand, "tmp")
			} else {
				return v.builder.CreateICmp(comparisonOpToIntPredicate(n.Op, n.Lhand.GetType().(parser.PrimitiveType).IsSigned()), lhand, rhand, "tmp")
			}

		// Bitwise
		case parser.BINOP_BIT_AND:
			return v.builder.CreateAnd(lhand, rhand, "tmp")
		case parser.BINOP_BIT_OR:
			return v.builder.CreateOr(lhand, rhand, "tmp")
		case parser.BINOP_BIT_XOR:
			return v.builder.CreateXor(lhand, rhand, "tmp")
		case parser.BINOP_BIT_LEFT:
			return v.builder.CreateShl(lhand, rhand, "tmp")
		case parser.BINOP_BIT_RIGHT:
			// TODO make sure both operands are same type (create type cast here?)
			// TODO in semantic.go, make sure rhand is *unsigned* (LLVM always treats it that way)
			// TODO logical shift right?
			return v.builder.CreateAShr(lhand, rhand, "tmp")

		// Logical
		case parser.BINOP_LOG_AND:
			return v.builder.CreateAnd(lhand, rhand, "tmp")
		case parser.BINOP_LOG_OR:
			return v.builder.CreateOr(lhand, rhand, "tmp")

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
		return v.builder.CreateNot(expr, "tmp")
	case parser.UNOP_ADDRESS:
		panic("hmmm")
	default:
		panic("unimplimented unary op")
	}
}

func (v *Codegen) genCastExpr(n *parser.CastExpr) llvm.Value {
	var res llvm.Value

	return res
}

func (v *Codegen) genCallExpr(n *parser.CallExpr) llvm.Value {
	mangledName := n.Function.MangledName(parser.MANGLE_ARK_UNSTABLE)
	function := v.curFile.Module.NamedFunction(mangledName)
	if function.IsNil() {
		v.err("function does not exist in current module")
	}

	numOfArguments := len(n.Arguments)
	if function.ParamsCount() != numOfArguments {
		v.err("invalid amount of arguments given to function %s", n.Function.Name)
	}

	args := make([]llvm.Value, numOfArguments)
	for i, arg := range n.Arguments {
		args[i] = v.genExpr(arg)
	}

	return v.builder.CreateCall(function, args, "")
}

func (v *Codegen) genAccessExpr(n *parser.AccessExpr) llvm.Value {
	var res llvm.Value

	return res
}

func (v *Codegen) genDerefExpr(n *parser.DerefExpr) llvm.Value {
	var res llvm.Value

	return res
}

func (v *Codegen) genBracketExpr(n *parser.BracketExpr) llvm.Value {
	return v.genExpr(n.Expr)
}

func typeToLLVMType(typ parser.Type) llvm.Type {
	switch typ.(type) {
	case parser.PrimitiveType:
		return primitiveTypeToLLVMType(typ.(parser.PrimitiveType))
	case *parser.StructType:
		return structTypeToLLVMType(typ.(*parser.StructType))
	case parser.PointerType:
		return llvm.PointerType(typeToLLVMType(typ.(parser.PointerType).Addressee), 0)
	default:
		panic("Unimplemented type category in LLVM codegen")
	}
}

func structTypeToLLVMType(typ *parser.StructType) llvm.Type {
	return llvm.IntType(1)
}

func primitiveTypeToLLVMType(typ parser.PrimitiveType) llvm.Type {
	switch typ {
	case parser.PRIMITIVE_int, parser.PRIMITIVE_uint:
		return llvm.IntType(intSize * 8)
	case parser.PRIMITIVE_i8, parser.PRIMITIVE_u8:
		return llvm.IntType(8)
	case parser.PRIMITIVE_i16, parser.PRIMITIVE_u16:
		return llvm.IntType(16)
	case parser.PRIMITIVE_i32, parser.PRIMITIVE_u32:
		return llvm.IntType(32)
	case parser.PRIMITIVE_i64, parser.PRIMITIVE_u64:
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
		// maybe this?
		return llvm.PointerType(llvm.IntType(32), 0)

	default:
		panic("Unimplemented primitive type in LLVM codegen")
	}
}
