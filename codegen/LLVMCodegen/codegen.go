package LLVMCodegen

import (
	"C"
	"unsafe"

	"github.com/ark-lang/ark-go/parser"

	"llvm.org/llvm/bindings/go/llvm"
)

const intSize = int(unsafe.Sizeof(C.int(0)))

type Codegen struct {
	OutputName string
}

func (v *Codegen) Generate(input []*parser.File) {

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
	panic("TODO")
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
