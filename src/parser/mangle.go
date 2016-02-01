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

// easier than making a method for all types
func TypeReferenceMangledName(mangleType MangleType, typ *TypeReference, ginst *GenericInstance) string {
	switch mangleType {
	case MANGLE_ARK_UNSTABLE:
		res := "_"

		for {
			if ptr, ok := typ.Type.(PointerType); ok {
				typ = ptr.Addressee
				res += "p"
			} else {
				break
			}
		}

		switch typ := typ.Type.(type) {
		case ArrayType:
			res += fmt.Sprintf("A%s", TypeReferenceMangledName(mangleType, typ.MemberType, ginst))

		case ReferenceType:
			var suffix string
			if typ.IsMutable {
				suffix = "M"
			} else {
				suffix = "C"
			}
			res += fmt.Sprintf("R%s%s", suffix, TypeReferenceMangledName(mangleType, typ.Referrer, ginst))

		case EnumType:
			res += fmt.Sprintf("E%d", len(typ.Members))
			for _, mem := range typ.Members {
				res += TypeReferenceMangledName(mangleType, &TypeReference{Type: mem.Type}, ginst)
			}

		case StructType:
			res += fmt.Sprintf("S%d", len(typ.Members))
			for _, mem := range typ.Members {
				res += TypeReferenceMangledName(mangleType, mem.Type, ginst)
			}

		case TupleType:
			res += fmt.Sprintf("T%d", len(typ.Members))
			for _, mem := range typ.Members {
				res += TypeReferenceMangledName(mangleType, mem, ginst)
			}

		case FunctionType:
			str := ""
			for _, par := range typ.Parameters {
				str += TypeReferenceMangledName(mangleType, par, ginst)
			}

			str += TypeReferenceMangledName(mangleType, typ.Return, ginst)

			if typ.Receiver != nil {
				str = TypeReferenceMangledName(mangleType, typ.Receiver, ginst) + str
			}

			res += fmt.Sprintf("%dFT%s", len(str), str)

		case *NamedType, PrimitiveType:
			name := typ.TypeName()
			return res + fmt.Sprintf("%d%s", len(name), name)

		case InterfaceType:
			str := ""
			for _, fn := range typ.Functions {
				str += fn.MangledName(mangleType, ginst)
			}

			res += fmt.Sprintf("%dI%s", len(str), str)

		case *SubstitutionType:
			return TypeReferenceMangledName(mangleType, ginst.Get(&TypeReference{Type: typ}), ginst)

		default:
			panic("unimplemented type mangling scheme")

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

func (v Function) MangledName(typ MangleType, ginst *GenericInstance) string {
	if v.Name == "main" {
		return "main" // TODO make sure only one main function
	}

	switch typ {
	case MANGLE_ARK_UNSTABLE:
		var prefix string
		if v.Type.Receiver != nil {
			prefix = "m"
		} else if v.StaticReceiverType != nil {
			prefix = "s"
		}

		result := fmt.Sprintf("_%sF%d%s", prefix, len(v.Name), v.Name)
		for _, arg := range v.Parameters {
			result += TypeReferenceMangledName(typ, arg.Variable.Type, ginst)
		}

		result += TypeReferenceMangledName(typ, v.Type.Return, ginst)

		if v.Type.Receiver != nil {
			result = TypeReferenceMangledName(typ, v.Receiver.Variable.Type, ginst) + result
		} else if v.StaticReceiverType != nil {
			result = TypeReferenceMangledName(typ, &TypeReference{Type: v.StaticReceiverType}, ginst) + result
		}

		result = v.ParentModule.MangledName(typ) + result

		return result
	default:
		panic("")
	}
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

func (v NamedType) MangledName(typ MangleType) string {
	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_N%d%s", len(v.Name), v.Name)

		result = v.ParentModule.MangledName(typ) + result

		return result
	default:
		panic("")
	}
}
