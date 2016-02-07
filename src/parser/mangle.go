package parser

import (
	"bytes"
	"fmt"
)

// In case we support multiple name mangling schemes
type MangleType int

const (
	MANGLE_ARK_UNSTABLE MangleType = iota
)

// TODO GenericInstance -> GenericContext

func TypeReferencesMangledName(mangleType MangleType, typs []*TypeReference, gcon *GenericContext) string {
	res := ""
	for _, typ := range typs {
		res += TypeReferenceMangledName(mangleType, typ, gcon)
	}
	return res
}

// easier than making a method for all types
func TypeReferenceMangledName(mangleType MangleType, typ *TypeReference, gcon *GenericContext) string {
	switch mangleType {
	case MANGLE_ARK_UNSTABLE:
		res := "_"

		for {
			if ptr, ok := typ.BaseType.(PointerType); ok {
				typ = ptr.Addressee
				res += "p"
			} else {
				break
			}
		}

		switch typ := typ.BaseType.(type) {
		case ArrayType:
			res += fmt.Sprintf("A%s", TypeReferenceMangledName(mangleType, typ.MemberType, gcon))

		case ReferenceType:
			var suffix string
			if typ.IsMutable {
				suffix = "M"
			} else {
				suffix = "C"
			}
			res += fmt.Sprintf("R%s%s", suffix, TypeReferenceMangledName(mangleType, typ.Referrer, gcon))

		case EnumType:
			res += fmt.Sprintf("E%d", len(typ.Members))
			for _, mem := range typ.Members {
				res += TypeReferenceMangledName(mangleType, &TypeReference{BaseType: mem.Type}, gcon)
			}

		case StructType:
			res += fmt.Sprintf("S%d", len(typ.Members))
			for _, mem := range typ.Members {
				res += TypeReferenceMangledName(mangleType, mem.Type, gcon)
			}

		case TupleType:
			res += fmt.Sprintf("T%d", len(typ.Members))
			for _, mem := range typ.Members {
				res += TypeReferenceMangledName(mangleType, mem, gcon)
			}

		case FunctionType:
			str := TypeReferencesMangledName(mangleType, typ.Parameters, gcon)

			str += TypeReferenceMangledName(mangleType, typ.Return, gcon)

			if typ.Receiver != nil {
				str = TypeReferenceMangledName(mangleType, &TypeReference{BaseType: typ.Receiver}, gcon) + str
			}

			res += fmt.Sprintf("%dFT%s", len(str), str)

		case *NamedType, PrimitiveType:
			name := typ.TypeName()
			res += fmt.Sprintf("%d%s", len(name), name)

		case InterfaceType:
			str := ""
			for _, fn := range typ.Functions {
				str += fn.MangledName(mangleType, gcon)
			}

			res += fmt.Sprintf("%dI%s", len(str), str)

		case *SubstitutionType:
			if sub := gcon.GetSubstitutionType(typ); sub != nil {
				res = TypeReferenceMangledName(mangleType, gcon.Get(&TypeReference{BaseType: typ}), gcon)
			} else {
				res = typ.Name
			}

		default:
			panic("unimplemented type mangling scheme")

		}

		gas := TypeReferencesMangledName(mangleType, typ.GenericArguments, gcon)
		if len(gas) > 0 {
			res += "GA" + gas
		}

		return res
	default:
		panic("")
	}
}

func (v Module) MangledName(typ MangleType) string {
	switch typ {
	case MANGLE_ARK_UNSTABLE:
		buf := new(bytes.Buffer)
		for _, mod := range v.Name.Parts {
			buf.WriteString("_M")
			buf.WriteString(fmt.Sprintf("%d", len(mod)))
			buf.WriteString(mod)
		}

		return buf.String()
	default:
		panic("")
	}
}

func (v Function) MangledNameWithReceiver(typ MangleType, receiver Type, gcon *GenericContext) string {
	if v.Name == "main" {
		return "main" // TODO make sure only one main function
	}

	switch typ {
	case MANGLE_ARK_UNSTABLE:
		var prefix string
		if receiver != nil {
			prefix = "m"
		} else if v.StaticReceiverType != nil {
			prefix = "s"
		}

		result := fmt.Sprintf("_%sF%d%s", prefix, len(v.Name), v.Name)
		for _, arg := range v.Parameters {
			result += TypeReferenceMangledName(typ, arg.Variable.Type, gcon)
		}

		result += TypeReferenceMangledName(typ, v.Type.Return, gcon)

		if receiver != nil {
			result = TypeReferenceMangledName(typ, &TypeReference{BaseType: receiver}, gcon) + result
		} else if v.StaticReceiverType != nil {
			result = TypeReferenceMangledName(typ, &TypeReference{BaseType: v.StaticReceiverType}, gcon) + result
		}

		result = v.ParentModule.MangledName(typ) + result

		return result
	default:
		panic("")
	}
}

func (v Function) MangledName(typ MangleType, gcon *GenericContext) string {
	return v.MangledNameWithReceiver(typ, v.Type.Receiver, gcon)
}

func (v Variable) MangledName(typ MangleType) string {
	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_V%d%s", len(v.Name), v.Name)
		if v.FromStruct {
			result = v.ParentModule.MangledName(typ) + result
		}
		return result
	default:
		panic("")
	}
}
