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

	builder llvm.Builder

	OutputName string
}

func (v *Codegen) err(err string, stuff ...interface{}) {
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(2)
}

func (v *Codegen) createBitcode() (string, bool) {
	filename := v.curFile.Name + ".bc"
	if err := llvm.VerifyModule(v.curFile.Module, llvm.ReturnStatusAction); err != nil {
		fmt.Println(err)
		return "", true
	}

	fileHandle, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		v.err("Couldn't create bitcode file `%s`" + err.Error(), filename)
	}
	defer fileHandle.Close()

	if res := llvm.WriteBitcodeToFile(v.curFile.Module, fileHandle); res != nil {
		v.err("failed to write bitcode to file for " + v.curFile.Name)
	}

	return filename, false
}

func (v *Codegen) bitcodeToASM(filename string) (string, bool) {
	asmName := filename + ".s"
	toAsmCommand := "llc " + filename + " -o " + asmName
	
	if cmd := exec.Command(toAsmCommand); cmd != nil {
		v.err("failed to convert bitcode to assembly")
		return "", true
	}

	if cmd := exec.Command("rm " + filename); cmd != nil {
		v.err("failed to remove " + filename)
		return "", true
	}

	return asmName, false
}

func (v *Codegen) createBinary(file string) {
	link := "cc " + file + " -o main.out"
	if cmd := exec.Command(link); cmd != nil {
		v.err("failed to link object files")
	}
}

func (v *Codegen) Generate(input []*parser.File, verbose bool) {
	v.input = input
	v.builder = llvm.NewBuilder()

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

		dur := time.Since(t)
		if verbose {
			fmt.Printf(util.TEXT_BOLD+util.TEXT_GREEN+"Finished codegenning"+util.TEXT_RESET+" %s (%.2fms)\n",
				infile.Name, float32(dur.Nanoseconds())/1000000)
		}

		if bitcode, err := v.createBitcode(); !err {
			if asm, err := v.bitcodeToASM(bitcode); !err {
				v.createBinary(asm)
			}

		}

		// infile.Module.Dump()
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
	return v.genExpr(n.Value)
}

func (v *Codegen) genCallStat(n *parser.CallStat) llvm.Value {
	return v.genExpr(n.Call)
}

func (v *Codegen) genAssignStat(n *parser.AssignStat) llvm.Value {
	if n.Deref != nil {
		v.genExpr(n.Deref)
	} else {
		v.genExpr(n.Access)
	}
	return v.genExpr(n.Assignment)
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
	var res llvm.Value

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

	return res
}

func (v *Codegen) genFunctionDecl(n *parser.FunctionDecl) llvm.Value {
	var res llvm.Value

	function := v.curFile.Module.NamedFunction(n.Function.Name)
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
		function = llvm.AddFunction(v.curFile.Module, n.Function.Name, funcType)

		// do some magical shit for later
		for i := 0; i < numOfParams; i++ {
			funcParam := function.Param(i)
			funcParam.SetName(n.Function.Parameters[i].Variable.Name)
			// maybe store it in a hashmap somewhere?
		}

		block := llvm.AddBasicBlock(function, "entry")
		v.builder.SetInsertPointAtEnd(block)

		for _, stat := range n.Function.Body.Nodes {
			v.genNode(stat)
		}

		// loop thru block and gen statements

		return function
	}

	return res
}

func (v *Codegen) genStructDecl(n *parser.StructDecl) llvm.Value {
	var res llvm.Value

	for _, member := range n.Struct.Variables {
		v.genVariableDecl(member, false)
	}

	return res
}

func (v *Codegen) genVariableDecl(n *parser.VariableDecl, semicolon bool) llvm.Value {
	var res llvm.Value

	// if n.Variable.Mutable

	alloc := v.builder.CreateAlloca(typeToLLVMType(n.Variable.Type), n.Variable.Name)

	if n.Assignment != nil {
		if value := v.genExpr(n.Assignment); !value.IsNil() {
			v.builder.CreateStore(value, alloc)
		}
	}

	return res
}

func (v *Codegen) genExpr(n parser.Expr) llvm.Value {
	var res llvm.Value

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

	return res
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
		floating := n.Lhand.GetType().IsFloatingType() || n.Rhand.GetType().IsFloatingType()

		switch n.Op {
		case parser.BINOP_ADD:
			if floating {
				return v.builder.CreateFAdd(lhand, rhand, "tmp")
			} else {
				return v.builder.CreateAdd(lhand, rhand, "tmp")
			}
		case parser.BINOP_SUB:
			if floating {
				return v.builder.CreateFSub(lhand, rhand, "tmp")
			} else {
				return v.builder.CreateSub(lhand, rhand, "tmp")
			}
		case parser.BINOP_MUL:
			if floating {
				return v.builder.CreateFMul(lhand, rhand, "tmp")
			} else {
				return v.builder.CreateMul(lhand, rhand, "tmp")
			}
		case parser.BINOP_DIV:
			if floating {
				return v.builder.CreateFDiv(lhand, rhand, "tmp")
			} else {
				//? ??
				return v.builder.CreateUDiv(lhand, rhand, "tmp")
			}
		}
	}

	return res
}

func (v *Codegen) genUnaryExpr(n *parser.UnaryExpr) llvm.Value {
	var res llvm.Value

	return res
}

func (v *Codegen) genCastExpr(n *parser.CastExpr) llvm.Value {
	var res llvm.Value

	return res
}

func (v *Codegen) genCallExpr(n *parser.CallExpr) llvm.Value {
	function := v.curFile.Module.NamedFunction(n.Function.Name)
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
	var res llvm.Value

	return res
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
