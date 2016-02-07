package LLVMCodegen

import (
	"fmt"
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

func (v *Codegen) addNamedTypeReference(n *parser.TypeReference, gcon *parser.GenericContext) {
	if v.inFunction() {
		gcon.Outer = v.currentFunction().gcon
	}

	v.addNamedType(n.BaseType.(*parser.NamedType), parser.TypeReferenceMangledName(parser.MANGLE_ARK_UNSTABLE, n, gcon), gcon)
}

func (v *Codegen) addNamedType(n *parser.NamedType, name string, gcon *parser.GenericContext) {
	switch actualtyp := n.Type.ActualType().(type) {
	case parser.StructType:
		v.addStructType(actualtyp, name, gcon)
	case parser.EnumType:
		v.addEnumType(actualtyp, name, gcon)
	}
}

func (v *Codegen) addStructType(typ parser.StructType, name string, gcon *parser.GenericContext) {
	if _, ok := v.namedTypeLookup[name]; ok {
		return
	}

	structure := v.curFile.LlvmModule.Context().StructCreateNamed(name)

	v.namedTypeLookup[name] = structure

	for _, field := range typ.Members {
		if _, ok := field.Type.BaseType.(*parser.NamedType); ok {
			v.addNamedTypeReference(field.Type, gcon)
		}
	}

	structure.StructSetBody(v.structTypeToLLVMTypeFields(typ, gcon), typ.Attrs().Contains("packed"))
}

func (v *Codegen) addEnumType(typ parser.EnumType, name string, gcon *parser.GenericContext) {
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
				v.addNamedType(named, parser.TypeReferenceMangledName(parser.MANGLE_ARK_UNSTABLE, &parser.TypeReference{BaseType: member.Type}, gcon), gcon)
			}
		}

		enum.StructSetBody(v.enumTypeToLLVMTypeFields(typ, gcon), false)
	}
}

func (v *Codegen) typeRefToLLVMType(typ *parser.TypeReference) llvm.Type {
	return v.typeRefToLLVMTypeWithOuter(typ, nil)
}

func (v *Codegen) typeRefToLLVMTypeWithGenericContext(typ *parser.TypeReference, gcon *parser.GenericContext) llvm.Type {
	switch nt := typ.BaseType.(type) {
	case *parser.NamedType:
		switch nt.ActualType().(type) {
		case parser.StructType, parser.EnumType:
			v.addNamedTypeReference(typ, gcon)
			lt := v.namedTypeLookup[parser.TypeReferenceMangledName(parser.MANGLE_ARK_UNSTABLE, typ, gcon)]
			return lt

		default:
			return v.typeToLLVMType(nt.Type, gcon)
		}
	}

	return v.typeToLLVMType(typ.BaseType, gcon)
}

// if outer is not nil, this function does not add the current function gcon as outer, as it assumes it is already there
func (v *Codegen) typeRefToLLVMTypeWithOuter(typ *parser.TypeReference, outer *parser.GenericContext) llvm.Type {
	gcon := parser.NewGenericContextFromTypeReference(typ)

	if outer != nil {
		gcon.Outer = outer
	} else if v.inFunction() {
		gcon.Outer = v.currentFunction().gcon
	}

	return v.typeRefToLLVMTypeWithGenericContext(typ, gcon)
}

func (v *Codegen) typeToLLVMType(typ parser.Type, gcon *parser.GenericContext) llvm.Type {
	switch typ := typ.(type) {
	case parser.PrimitiveType:
		return v.primitiveTypeToLLVMType(typ)
	case parser.FunctionType:
		return v.functionTypeToLLVMType(typ, true, gcon)
	case parser.StructType:
		return v.structTypeToLLVMType(typ, gcon)
	case parser.PointerType:
		return llvm.PointerType(v.typeRefToLLVMTypeWithOuter(typ.Addressee, gcon), 0)
	case parser.ArrayType:
		return v.arrayTypeToLLVMType(typ, gcon)
	case parser.TupleType:
		return v.tupleTypeToLLVMType(typ, gcon)
	case parser.EnumType:
		return v.enumTypeToLLVMType(typ, gcon)
	case parser.ReferenceType:
		return llvm.PointerType(v.typeRefToLLVMTypeWithOuter(typ.Referrer, gcon), 0)
	case *parser.NamedType:
		switch typ.Type.(type) {
		case parser.StructType, parser.EnumType:
			// If something seems wrong with the code and thie problems seems to come from here,
			// make sure the type doesn't need generics arguments as well.
			// This here ignores them.

			name := parser.TypeReferenceMangledName(parser.MANGLE_ARK_UNSTABLE, &parser.TypeReference{BaseType: typ}, gcon)
			v.addNamedType(typ, name, gcon)
			lt := v.namedTypeLookup[name]

			return lt

		default:
			return v.typeToLLVMType(typ.Type, gcon)
		}
	case *parser.SubstitutionType:
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

func (v *Codegen) tupleTypeToLLVMType(typ parser.TupleType, gcon *parser.GenericContext) llvm.Type {
	fields := make([]llvm.Type, len(typ.Members))
	for idx, mem := range typ.Members {
		fields[idx] = v.typeRefToLLVMTypeWithOuter(mem, gcon)
	}

	return llvm.StructType(fields, false)
}

func (v *Codegen) arrayTypeToLLVMType(typ parser.ArrayType, gcon *parser.GenericContext) llvm.Type {
	memType := v.typeRefToLLVMTypeWithOuter(typ.MemberType, gcon)

	if typ.IsFixedLength {
		return llvm.ArrayType(memType, typ.Length)
	} else {
		fields := []llvm.Type{v.primitiveTypeToLLVMType(parser.PRIMITIVE_uint), llvm.PointerType(memType, 0)}
		return llvm.StructType(fields, false)
	}
}

func (v *Codegen) structTypeToLLVMType(typ parser.StructType, gcon *parser.GenericContext) llvm.Type {
	return llvm.StructType(v.structTypeToLLVMTypeFields(typ, gcon), typ.Attrs().Contains("packed"))
}

func (v *Codegen) structTypeToLLVMTypeFields(typ parser.StructType, gcon *parser.GenericContext) []llvm.Type {
	numOfFields := len(typ.Members)
	fields := make([]llvm.Type, numOfFields)

	for i, member := range typ.Members {
		memberType := v.typeRefToLLVMTypeWithOuter(member.Type, gcon)
		fields[i] = memberType
	}

	return fields
}

func (v *Codegen) enumTypeToLLVMType(typ parser.EnumType, gcon *parser.GenericContext) llvm.Type {
	if typ.Simple {
		// TODO: Handle other integer size, maybe dynamic depending on max value? (1 / 2)
		return llvm.IntType(32)
	}

	return llvm.StructType(v.enumTypeToLLVMTypeFields(typ, gcon), false)
}

func (v *Codegen) enumTypeToLLVMTypeFields(typ parser.EnumType, gcon *parser.GenericContext) []llvm.Type {
	longestLength := uint64(0)
	for _, member := range typ.Members {
		memLength := v.targetData.TypeAllocSize(v.typeToLLVMType(member.Type, gcon))
		if memLength > longestLength {
			longestLength = memLength
		}
	}

	// TODO: verify no overflow
	return []llvm.Type{llvm.IntType(32), llvm.ArrayType(llvm.IntType(8), int(longestLength))}
}

func (v *Codegen) functionTypeToLLVMType(typ parser.FunctionType, ptr bool, gcon *parser.GenericContext) llvm.Type {
	numOfParams := len(typ.Parameters)
	if typ.Receiver != nil {
		numOfParams++
	}

	params := make([]llvm.Type, 0, numOfParams)
	if typ.Receiver != nil {
		params = append(params, v.typeToLLVMType(typ.Receiver, gcon))
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
