package LLVMCodegen

import (
	"fmt"
	"reflect"

	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/ark/src/util/log"

	"github.com/ark-lang/go-llvm/llvm"
)

func floatTypeBits(ty ast.PrimitiveType) int {
	switch ty {
	case ast.PRIMITIVE_f32:
		return 32
	case ast.PRIMITIVE_f64:
		return 64
	case ast.PRIMITIVE_f128:
		return 128
	default:
		panic("this isn't a float type")
	}
}

func (v *Codegen) addNamedTypeReference(n *ast.TypeReference, gcon *ast.GenericContext) {
	if v.inFunction() {
		gcon.Outer = v.currentFunction().gcon
	}

	v.addNamedType(n.BaseType.(*ast.NamedType), ast.TypeReferenceMangledName(ast.MANGLE_ARK_UNSTABLE, n, gcon), gcon)
}

func (v *Codegen) addNamedType(n *ast.NamedType, name string, gcon *ast.GenericContext) {
	switch actualtyp := n.Type.ActualType().(type) {
	case ast.StructType:
		v.addStructType(actualtyp, name, gcon)
	case ast.EnumType:
		v.addEnumType(actualtyp, name, gcon)
	}
}

func (v *Codegen) addStructType(typ ast.StructType, name string, gcon *ast.GenericContext) {
	if _, ok := v.namedTypeLookup[name]; ok {
		return
	}

	structure := v.curFile.LlvmModule.Context().StructCreateNamed(name)

	v.namedTypeLookup[name] = structure

	for _, field := range typ.Members {
		if _, ok := field.Type.BaseType.(*ast.NamedType); ok {
			v.addNamedTypeReference(field.Type, gcon)
		}
	}

	structure.StructSetBody(v.structTypeToLLVMTypeFields(typ, gcon), typ.Attrs().Contains("packed"))
}

func (v *Codegen) addEnumType(typ ast.EnumType, name string, gcon *ast.GenericContext) {
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
			if named, ok := member.Type.(*ast.NamedType); ok {
				v.addNamedType(named, ast.TypeReferenceMangledName(ast.MANGLE_ARK_UNSTABLE, &ast.TypeReference{BaseType: member.Type}, gcon), gcon)
			}
		}

		enum.StructSetBody(v.enumTypeToLLVMTypeFields(typ, gcon), false)
	}
}

func (v *Codegen) typeRefToLLVMType(typ *ast.TypeReference) llvm.Type {
	return v.typeRefToLLVMTypeWithOuter(typ, nil)
}

func (v *Codegen) typeRefToLLVMTypeWithGenericContext(typ *ast.TypeReference, gcon *ast.GenericContext) llvm.Type {
	switch nt := typ.BaseType.(type) {
	case *ast.NamedType:
		switch nt.ActualType().(type) {
		case ast.StructType, ast.EnumType:
			v.addNamedTypeReference(typ, gcon)
			lt := v.namedTypeLookup[ast.TypeReferenceMangledName(ast.MANGLE_ARK_UNSTABLE, typ, gcon)]
			return lt

		default:
			return v.typeToLLVMType(nt.Type, gcon)
		}
	}

	return v.typeToLLVMType(typ.BaseType, gcon)
}

// if outer is not nil, this function does not add the current function gcon as outer, as it assumes it is already there
func (v *Codegen) typeRefToLLVMTypeWithOuter(typ *ast.TypeReference, outer *ast.GenericContext) llvm.Type {
	gcon := ast.NewGenericContextFromTypeReference(typ)

	if outer != nil {
		gcon.Outer = outer
	} else if v.inFunction() {
		gcon.Outer = v.currentFunction().gcon
	}

	return v.typeRefToLLVMTypeWithGenericContext(typ, gcon)
}

func (v *Codegen) typeToLLVMType(typ ast.Type, gcon *ast.GenericContext) llvm.Type {
	switch typ := typ.(type) {
	case ast.PrimitiveType:
		return v.primitiveTypeToLLVMType(typ)
	case ast.FunctionType:
		return v.functionTypeToLLVMType(typ, true, gcon)
	case ast.StructType:
		return v.structTypeToLLVMType(typ, gcon)
	case ast.PointerType:
		return llvm.PointerType(v.typeRefToLLVMTypeWithOuter(typ.Addressee, gcon), 0)
	case ast.ArrayType:
		return v.arrayTypeToLLVMType(typ, gcon)
	case ast.TupleType:
		return v.tupleTypeToLLVMType(typ, gcon)
	case ast.EnumType:
		return v.enumTypeToLLVMType(typ, gcon)
	case ast.ReferenceType:
		return llvm.PointerType(v.typeRefToLLVMTypeWithOuter(typ.Referrer, gcon), 0)
	case *ast.NamedType:
		switch typ.Type.(type) {
		case ast.StructType, ast.EnumType:
			// If something seems wrong with the code and thie problems seems to come from here,
			// make sure the type doesn't need generics arguments as well.
			// This here ignores them.

			name := ast.TypeReferenceMangledName(ast.MANGLE_ARK_UNSTABLE, &ast.TypeReference{BaseType: typ}, gcon)
			v.addNamedType(typ, name, gcon)
			lt := v.namedTypeLookup[name]

			return lt

		default:
			return v.typeToLLVMType(typ.Type, gcon)
		}
	case *ast.SubstitutionType:
		if gcon == nil {
			panic("gcon == nil when getting substitution type")
		}
		if gcon.GetSubstitutionType(typ) == nil {
			fmt.Println(gcon)
			panic("missing generic map entry for type " + typ.TypeName() + " " + fmt.Sprintf("(%p)", typ))
		}
		return v.typeRefToLLVMTypeWithOuter(gcon.GetSubstitutionType(typ), gcon)
	default:
		log.Debugln("codegen", "Type was %s (%s)", typ.TypeName(), reflect.TypeOf(typ))
		panic("Unimplemented type category in LLVM codegen")
	}
}

func (v *Codegen) tupleTypeToLLVMType(typ ast.TupleType, gcon *ast.GenericContext) llvm.Type {
	fields := make([]llvm.Type, len(typ.Members))
	for idx, mem := range typ.Members {
		fields[idx] = v.typeRefToLLVMTypeWithOuter(mem, gcon)
	}

	return llvm.StructType(fields, false)
}

func (v *Codegen) arrayTypeToLLVMType(typ ast.ArrayType, gcon *ast.GenericContext) llvm.Type {
	memType := v.typeRefToLLVMTypeWithOuter(typ.MemberType, gcon)

	if typ.IsFixedLength {
		return llvm.ArrayType(memType, typ.Length)
	} else {
		fields := []llvm.Type{v.primitiveTypeToLLVMType(ast.PRIMITIVE_uint), llvm.PointerType(memType, 0)}
		return llvm.StructType(fields, false)
	}
}

func (v *Codegen) structTypeToLLVMType(typ ast.StructType, gcon *ast.GenericContext) llvm.Type {
	return llvm.StructType(v.structTypeToLLVMTypeFields(typ, gcon), typ.Attrs().Contains("packed"))
}

func (v *Codegen) structTypeToLLVMTypeFields(typ ast.StructType, gcon *ast.GenericContext) []llvm.Type {
	numOfFields := len(typ.Members)
	fields := make([]llvm.Type, numOfFields)

	for i, member := range typ.Members {
		memberType := v.typeRefToLLVMTypeWithOuter(member.Type, gcon)
		fields[i] = memberType
	}

	return fields
}

func (v *Codegen) enumTypeToLLVMType(typ ast.EnumType, gcon *ast.GenericContext) llvm.Type {
	if typ.Simple {
		// TODO: Handle other integer size, maybe dynamic depending on max value? (1 / 2)
		return llvm.IntType(32)
	}

	return llvm.StructType(v.enumTypeToLLVMTypeFields(typ, gcon), false)
}

func (v *Codegen) enumTypeToLLVMTypeFields(typ ast.EnumType, gcon *ast.GenericContext) []llvm.Type {
	longestLength := uint64(0)
	for _, member := range typ.Members {
		memLength := v.targetData.TypeAllocSize(v.typeToLLVMType(member.Type, gcon))
		if memLength > longestLength {
			longestLength = memLength
		}
	}

	// TODO: verify no overflow
	return []llvm.Type{enumTagType, llvm.ArrayType(llvm.IntType(8), int(longestLength))}
}

func (v *Codegen) enumMemberTypeToPaddedLLVMType(enumType ast.EnumType, memberIdx int, gcon *ast.GenericContext) llvm.Type {
	longestLength := uint64(0)
	for _, member := range enumType.Members {
		memLength := v.targetData.TypeAllocSize(v.typeToLLVMType(member.Type, gcon))
		if memLength > longestLength {
			longestLength = memLength
		}
	}

	member := enumType.Members[memberIdx]
	actualType := v.typeToLLVMType(member.Type, gcon)
	amountPad := longestLength - v.targetData.TypeAllocSize(actualType)

	if amountPad > 0 {
		types := make([]llvm.Type, actualType.StructElementTypesCount()+1)
		copy(types, actualType.StructElementTypes())
		types[len(types)-1] = llvm.ArrayType(llvm.IntType(8), int(amountPad))
		return llvm.StructType(types, false)
	} else {
		return actualType
	}
}

func (v *Codegen) llvmEnumTypeForMember(enumType ast.EnumType, memberIdx int, gcon *ast.GenericContext) llvm.Type {
	return llvm.StructType([]llvm.Type{enumTagType, v.enumMemberTypeToPaddedLLVMType(enumType, memberIdx, gcon)}, false)
}

func (v *Codegen) functionTypeToLLVMType(typ ast.FunctionType, ptr bool, gcon *ast.GenericContext) llvm.Type {
	numOfParams := len(typ.Parameters)
	if typ.Receiver != nil {
		numOfParams++
	}

	params := make([]llvm.Type, 0, numOfParams)
	if typ.Receiver != nil {
		params = append(params, v.typeRefToLLVMTypeWithOuter(typ.Receiver, gcon))
	}
	for _, par := range typ.Parameters {
		params = append(params, v.typeRefToLLVMTypeWithOuter(par, gcon))
	}

	var returnType llvm.Type

	// oo theres a type, let's try figure it out
	if typ.Return != nil {
		returnType = v.typeRefToLLVMTypeWithOuter(typ.Return, gcon)
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

func (v *Codegen) primitiveTypeToLLVMType(typ ast.PrimitiveType) llvm.Type {
	switch typ {
	case ast.PRIMITIVE_int, ast.PRIMITIVE_uint, ast.PRIMITIVE_uintptr:
		return v.targetData.IntPtrType()

	case ast.PRIMITIVE_s8, ast.PRIMITIVE_u8:
		return llvm.IntType(8)
	case ast.PRIMITIVE_s16, ast.PRIMITIVE_u16:
		return llvm.IntType(16)
	case ast.PRIMITIVE_s32, ast.PRIMITIVE_u32:
		return llvm.IntType(32)
	case ast.PRIMITIVE_s64, ast.PRIMITIVE_u64:
		return llvm.IntType(64)
	case ast.PRIMITIVE_s128, ast.PRIMITIVE_u128:
		return llvm.IntType(128)

	case ast.PRIMITIVE_f32:
		return llvm.FloatType()
	case ast.PRIMITIVE_f64:
		return llvm.DoubleType()
	case ast.PRIMITIVE_f128:
		return llvm.FP128Type()

	case ast.PRIMITIVE_bool:
		return llvm.IntType(1)
	case ast.PRIMITIVE_void:
		return llvm.VoidType()

	default:
		panic("Unimplemented primitive type in LLVM codegen")
	}
}
