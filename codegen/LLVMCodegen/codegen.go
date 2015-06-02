package LLVMCodegen

import (
	"C"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/ark-lang/ark-go/parser"
	"github.com/ark-lang/ark-go/util"

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

		infile.Module.Dump()
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

	if n.Variable.Mutable {
		// do mut stuff here
	}

	if n.Assignment != nil {
		v.genExpr(n.Assignment)
	}

	return res
}

func (v *Codegen) genExpr(n parser.Expr) llvm.Value {
	var res llvm.Value

	switch n.(type) {
	case *parser.RuneLiteral:
		v.genRuneLiteral(n.(*parser.RuneLiteral))
	case *parser.IntegerLiteral:
		v.genIntegerLiteral(n.(*parser.IntegerLiteral))
	case *parser.FloatingLiteral:
		v.genFloatingLiteral(n.(*parser.FloatingLiteral))
	case *parser.StringLiteral:
		v.genStringLiteral(n.(*parser.StringLiteral))
	case *parser.BinaryExpr:
		v.genBinaryExpr(n.(*parser.BinaryExpr))
	case *parser.UnaryExpr:
		v.genUnaryExpr(n.(*parser.UnaryExpr))
	case *parser.CastExpr:
		v.genCastExpr(n.(*parser.CastExpr))
	case *parser.CallExpr:
		v.genCallExpr(n.(*parser.CallExpr))
	case *parser.AccessExpr:
		v.genAccessExpr(n.(*parser.AccessExpr))
	case *parser.DerefExpr:
		v.genDerefExpr(n.(*parser.DerefExpr))
	case *parser.BracketExpr:
		v.genBracketExpr(n.(*parser.BracketExpr))
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
	var res llvm.Value

	return res
}

func (v *Codegen) genFloatingLiteral(n *parser.FloatingLiteral) llvm.Value {
	var res llvm.Value

	return res
}

func (v *Codegen) genStringLiteral(n *parser.StringLiteral) llvm.Value {
	var res llvm.Value

	return res
}

func (v *Codegen) genBinaryExpr(n *parser.BinaryExpr) llvm.Value {
	var res llvm.Value

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
		panic("not sure how this works yet")

	default:
		panic("Unimplemented primitive type in LLVM codegen")
	}
}
