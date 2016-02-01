package LLVMCodegen

import (
	"reflect"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util/log"

	"github.com/ark-lang/go-llvm/llvm"
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

func (v *Codegen) addNamedTypeReference(n *parser.TypeReference) {
	ginst := parser.NewGenericInstanceFromTypeReference(n)

	if v.inFunction() {
		ginst.Outer = v.currentFunction().ginst
	}

	v.addNamedType(n.Type.(*parser.NamedType), ginst)
}

func (v *Codegen) addNamedType(n *parser.NamedType, ginst *parser.GenericInstance) {
	if len(n.GenericParameters) > 0 {
		return
	}

	switch typ := n.Type.ActualType().(type) {
	case parser.StructType:
		v.addStructType(typ, n.MangledName(parser.MANGLE_ARK_UNSTABLE), ginst)
	case parser.EnumType:
		v.addEnumType(typ, n.MangledName(parser.MANGLE_ARK_UNSTABLE), ginst)
	}
}

func (v *Codegen) addStructType(typ parser.StructType, name string, ginst *parser.GenericInstance) {
	if _, ok := v.namedTypeLookup[name]; ok {
		return
	}

	structure := v.curFile.LlvmModule.Context().StructCreateNamed(name)

	v.namedTypeLookup[name] = structure

	for _, field := range typ.Members {
		if named, ok := field.Type.Type.(*parser.NamedType); ok {
			v.addNamedType(named, ginst)
		}
	}

	structure.StructSetBody(v.structTypeToLLVMTypeFields(typ, ginst), typ.Attrs().Contains("packed"))
}

func (v *Codegen) addEnumType(typ parser.EnumType, name string, ginst *parser.GenericInstance) {
	if _, ok := v.namedTypeLookup[name]; ok {
		return
	}

	if typ.Simple {
		// TODO: Handle other integer size, maybe dynamic depending on max value?
		v.namedTypeLookup[name] = llvm.IntType(32)
	} else {
		enum := v.curFile.LlvmModule.Context().StructCreateNamed(name)
		v.namedTypeLookup[name] = enum

		for _, member := range typ.Members {
			if named, ok := member.Type.(*parser.NamedType); ok {
				v.addNamedType(named, ginst)
			}
		}

		enum.StructSetBody(v.enumTypeToLLVMTypeFields(typ, ginst), false)
	}
}

func (v *Codegen) typeRefToLLVMType(typ *parser.TypeReference) llvm.Type {
	return v.typeRefToLLVMTypeWithOuter(typ, nil)
}

// if outer is not nil, this function does not add the current function ginst as outer, as it assumes it is already there
func (v *Codegen) typeRefToLLVMTypeWithOuter(typ *parser.TypeReference, outer *parser.GenericInstance) llvm.Type {
	ginst := parser.NewGenericInstanceFromTypeReference(typ)

	if outer != nil {
		ginst.Outer = outer
	} else if v.inFunction() {
		ginst.Outer = v.currentFunction().ginst
	}

	switch nt := typ.Type.(type) {
	case *parser.NamedType:
		switch nt.Type.(type) {
		case parser.StructType, parser.EnumType:
			v.addNamedType(nt, ginst)
			lt := v.namedTypeLookup[parser.TypeReferenceMangledName(parser.MANGLE_ARK_UNSTABLE, typ, ginst)]
			return lt
		}
	}

	return v.typeToLLVMType(typ.Type, ginst)
}

func (v *Codegen) typeToLLVMType(typ parser.Type, ginst *parser.GenericInstance) llvm.Type {
	switch typ := typ.(type) {
	case parser.PrimitiveType:
		return v.primitiveTypeToLLVMType(typ)
	case parser.FunctionType:
		return v.functionTypeToLLVMType(typ, true, ginst)
	case parser.StructType:
		return v.structTypeToLLVMType(typ, ginst)
	case parser.PointerType:
		return llvm.PointerType(v.typeRefToLLVMTypeWithOuter(typ.Addressee, ginst), 0)
	case parser.ArrayType:
		return v.arrayTypeToLLVMType(typ, ginst)
	case parser.TupleType:
		return v.tupleTypeToLLVMType(typ, ginst)
	case parser.EnumType:
		return v.enumTypeToLLVMType(typ, ginst)
	case parser.ReferenceType:
		return llvm.PointerType(v.typeRefToLLVMTypeWithOuter(typ.Referrer, ginst), 0)
	case parser.SubstitutionType:
		if ginst == nil {
			panic("ginst == nil when getting substitution type")
		}
		if ginst.GetSubstitutionType(typ) == nil {
			panic("missing generic map entry for type " + typ.TypeName())
		}
		return v.typeRefToLLVMTypeWithOuter(ginst.GetSubstitutionType(typ), ginst)
	default:
		log.Debugln("codegen", "Type was %s (%s)", typ.TypeName(), reflect.TypeOf(typ))
		panic("Unimplemented type category in LLVM codegen")
	}
}

func (v *Codegen) tupleTypeToLLVMType(typ parser.TupleType, ginst *parser.GenericInstance) llvm.Type {
	fields := make([]llvm.Type, len(typ.Members))
	for idx, mem := range typ.Members {
		fields[idx] = v.typeRefToLLVMTypeWithOuter(mem, ginst)
	}

	return llvm.StructType(fields, false)
}

func (v *Codegen) arrayTypeToLLVMType(typ parser.ArrayType, ginst *parser.GenericInstance) llvm.Type {
	fields := []llvm.Type{v.primitiveTypeToLLVMType(parser.PRIMITIVE_uint),
		llvm.PointerType(v.typeRefToLLVMTypeWithOuter(typ.MemberType, ginst), 0)}

	return llvm.StructType(fields, false)
}

func (v *Codegen) structTypeToLLVMType(typ parser.StructType, ginst *parser.GenericInstance) llvm.Type {
	return llvm.StructType(v.structTypeToLLVMTypeFields(typ, ginst), typ.Attrs().Contains("packed"))
}

func (v *Codegen) structTypeToLLVMTypeFields(typ parser.StructType, ginst *parser.GenericInstance) []llvm.Type {
	numOfFields := len(typ.Members)
	fields := make([]llvm.Type, numOfFields)

	for i, member := range typ.Members {
		memberType := v.typeRefToLLVMTypeWithOuter(member.Type, ginst)
		fields[i] = memberType
	}

	return fields
}

func (v *Codegen) enumTypeToLLVMType(typ parser.EnumType, ginst *parser.GenericInstance) llvm.Type {
	if typ.Simple {
		// TODO: Handle other integer size, maybe dynamic depending on max value? (1 / 2)
		return llvm.IntType(32)
	}

	return llvm.StructType(v.enumTypeToLLVMTypeFields(typ, ginst), false)
}

func (v *Codegen) enumTypeToLLVMTypeFields(typ parser.EnumType, ginst *parser.GenericInstance) []llvm.Type {
	longestLength := uint64(0)
	for _, member := range typ.Members {
		memLength := v.targetData.TypeAllocSize(v.typeToLLVMType(member.Type, ginst))
		if memLength > longestLength {
			longestLength = memLength
		}
	}

	// TODO: verify no overflow
	return []llvm.Type{llvm.IntType(32), llvm.ArrayType(llvm.IntType(8), int(longestLength))}
}

func (v *Codegen) functionTypeToLLVMType(typ parser.FunctionType, ptr bool, ginst *parser.GenericInstance) llvm.Type {
	numOfParams := len(typ.Parameters)
	if typ.Receiver != nil {
		numOfParams++
	}

	params := make([]llvm.Type, 0, numOfParams)
	if typ.Receiver != nil {
		params = append(params, v.typeRefToLLVMTypeWithOuter(typ.Receiver, ginst))
	}
	for _, par := range typ.Parameters {
		params = append(params, v.typeRefToLLVMTypeWithOuter(par, ginst))
	}

	var returnType llvm.Type

	// oo theres a type, let's try figure it out
	if typ.Return != nil {
		returnType = v.typeRefToLLVMTypeWithOuter(typ.Return, ginst)
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
	case parser.PRIMITIVE_int, parser.PRIMITIVE_uint, parser.PRIMITIVE_uintptr:
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

	case parser.PRIMITIVE_bool:
		return llvm.IntType(1)
	case parser.PRIMITIVE_void:
		return llvm.VoidType()

	default:
		panic("Unimplemented primitive type in LLVM codegen")
	}
}
