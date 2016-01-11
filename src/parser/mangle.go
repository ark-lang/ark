package parser

import (
	"bytes"
	"fmt"
)

// In case we support multiple name mangling schemes
type MangleType int

const (
	MANGLE_ARK_UNSTABLE MangleType = iota // see https://github.com/ark-lang/ark-rfcs/issues/3
)

// easier then making a method for all types
func TypeMangledName(mangleType MangleType, typ Type) string {
	switch mangleType {
	case MANGLE_ARK_UNSTABLE:
		res := "_"

		for {
			if ptr, ok := typ.(PointerType); ok {
				typ = ptr.Addressee
				res += "p"
			} else {
				break
			}
		}

		switch typ.(type) {
		case ArrayType:
			at := typ.(ArrayType)
			res += fmt.Sprintf("A%s", TypeMangledName(mangleType, at.MemberType))

		case ConstantReferenceType:
			crt := typ.(ConstantReferenceType)
			res += fmt.Sprintf("RC%s", TypeMangledName(mangleType, crt.Referrer))

		case MutableReferenceType:
			mrt := typ.(MutableReferenceType)
			res += fmt.Sprintf("RM%s", TypeMangledName(mangleType, mrt.Referrer))

		case EnumType:
			et := typ.(EnumType)
			res += fmt.Sprintf("E%d", len(et.Members))
			for _, mem := range et.Members {
				res += TypeMangledName(mangleType, mem.Type)
			}

		case StructType:
			st := typ.(StructType)
			res += fmt.Sprintf("S%d", len(st.Variables))
			for _, decl := range st.Variables {
				res += TypeMangledName(mangleType, decl.Variable.Type)
			}

		case TupleType:
			tt := typ.(TupleType)
			res += fmt.Sprintf("T%d", len(tt.Members))
			for _, mem := range tt.Members {
				res += TypeMangledName(mangleType, mem)
			}

		default:
			name := typ.TypeName()
			return res + fmt.Sprintf("%d%s", len(name), name)

		}

		return res
	default:
		panic("")
	}
}

func (v *Module) MangledName(typ MangleType) string {
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

func (v *Function) MangledName(typ MangleType) string {
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
			result += TypeMangledName(typ, arg.Variable.Type)
		}

		if v.Type.Receiver != nil {
			result = TypeMangledName(typ, v.Receiver.Variable.Type) + result
		} else if v.StaticReceiverType != nil {
			result = TypeMangledName(typ, v.StaticReceiverType) + result
		}

		result = v.ParentModule.MangledName(typ) + result

		return result
	default:
		panic("")
	}
}

func (v *Variable) MangledName(typ MangleType) string {
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

func (v *NamedType) MangledName(typ MangleType) string {
	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_N%d%s", len(v.Name), v.Name)

		result = v.ParentModule.MangledName(typ) + result

		return result
	default:
		panic("")
	}
}
