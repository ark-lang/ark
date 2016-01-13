package LLVMCodegen

import (
	"reflect"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util/log"
	"llvm.org/llvm/bindings/go/llvm"
)

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

func (v *Codegen) typeToLLVMType(typ parser.Type) llvm.Type {
	switch typ := typ.(type) {
	case parser.PrimitiveType:
		return v.primitiveTypeToLLVMType(typ)
	case parser.FunctionType:
		return v.functionTypeToLLVMType(typ, true)
	case parser.StructType:
		return v.structTypeToLLVMType(typ)
	case parser.PointerType:
		return llvm.PointerType(v.typeToLLVMType(typ.Addressee), 0)
	case parser.ArrayType:
		return v.arrayTypeToLLVMType(typ)
	case parser.TupleType:
		return v.tupleTypeToLLVMType(typ)
	case parser.EnumType:
		return v.enumTypeToLLVMType(typ)
	case *parser.NamedType:
		nt := typ
		switch nt.Type.(type) {
		case parser.StructType, parser.EnumType:
			v.addNamedType(nt)
			lt := v.namedTypeLookup[nt.MangledName(parser.MANGLE_ARK_UNSTABLE)]
			return lt

		default:
			return v.typeToLLVMType(nt.Type)
		}
	case parser.MutableReferenceType:
		return llvm.PointerType(v.typeToLLVMType(typ.Referrer), 0)
	case parser.ConstantReferenceType:
		return llvm.PointerType(v.typeToLLVMType(typ.Referrer), 0)
	default:
		log.Debugln("codegen", "Type was %s (%s)", typ.TypeName(), reflect.TypeOf(typ))
		panic("Unimplemented type category in LLVM codegen")
	}
}

func (v *Codegen) tupleTypeToLLVMType(typ parser.TupleType) llvm.Type {
	fields := make([]llvm.Type, len(typ.Members))
	for idx, mem := range typ.Members {
		fields[idx] = v.typeToLLVMType(mem)
	}

	return llvm.StructType(fields, false)
}

func (v *Codegen) arrayTypeToLLVMType(typ parser.ArrayType) llvm.Type {
	fields := []llvm.Type{v.typeToLLVMType(parser.PRIMITIVE_uint),
		llvm.PointerType(llvm.ArrayType(v.typeToLLVMType(typ.MemberType), 0), 0)}

	return llvm.StructType(fields, false)
}

func (v *Codegen) structTypeToLLVMType(typ parser.StructType) llvm.Type {
	return llvm.StructType(v.structTypeToLLVMTypeFields(typ), typ.Attrs().Contains("packed"))
}

func (v *Codegen) structTypeToLLVMTypeFields(typ parser.StructType) []llvm.Type {
	numOfFields := len(typ.Variables)
	fields := make([]llvm.Type, numOfFields)

	for i, member := range typ.Variables {
		memberType := v.typeToLLVMType(member.Variable.Type)
		fields[i] = memberType
	}

	return fields
}

func (v *Codegen) enumTypeToLLVMType(typ parser.EnumType) llvm.Type {
	if typ.Simple {
		// TODO: Handle other integer size, maybe dynamic depending on max value? (1 / 2)
		return llvm.IntType(32)
	}

	return llvm.StructType(v.enumTypeToLLVMTypeFields(typ), false)
}

func (v *Codegen) enumTypeToLLVMTypeFields(typ parser.EnumType) []llvm.Type {
	longestLength := uint64(0)
	for _, member := range typ.Members {
		memLength := v.targetData.TypeAllocSize(v.typeToLLVMType(member.Type))
		if memLength > longestLength {
			longestLength = memLength
		}
	}

	// TODO: verify no overflow
	return []llvm.Type{llvm.IntType(32), llvm.ArrayType(llvm.IntType(8), int(longestLength))}
}

func (v *Codegen) functionTypeToLLVMType(typ parser.FunctionType, ptr bool) llvm.Type {
	numOfParams := len(typ.Parameters)
	if typ.Receiver != nil {
		numOfParams++
	}

	params := make([]llvm.Type, 0, numOfParams)
	if typ.Receiver != nil {
		params = append(params, v.typeToLLVMType(typ.Receiver))
	}
	for _, par := range typ.Parameters {
		params = append(params, v.typeToLLVMType(par))
	}

	var returnType llvm.Type

	// oo theres a type, let's try figure it out
	if typ.Return != nil {
		returnType = v.typeToLLVMType(typ.Return)
	} else {
		returnType = llvm.VoidType()
	}

	// create the function type
	funcType := llvm.FunctionType(returnType, params, typ.IsVariadic)

	if ptr {
		funcType = llvm.PointerType(funcType, 0)
	}

	return funcType
}

func (v *Codegen) primitiveTypeToLLVMType(typ parser.PrimitiveType) llvm.Type {
	switch typ {
	case parser.PRIMITIVE_int, parser.PRIMITIVE_uint:
		return v.targetData.IntPtrType()

	case parser.PRIMITIVE_s8, parser.PRIMITIVE_u8:
		return llvm.IntType(8)
	case parser.PRIMITIVE_s16, parser.PRIMITIVE_u16:
		return llvm.IntType(16)
	case parser.PRIMITIVE_s32, parser.PRIMITIVE_u32:
		return llvm.IntType(32)
	case parser.PRIMITIVE_s64, parser.PRIMITIVE_u64:
		return llvm.IntType(64)
	case parser.PRIMITIVE_s128, parser.PRIMITIVE_u128:
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
