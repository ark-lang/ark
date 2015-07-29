package parser

import "fmt"

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
				return res + fmt.Sprintf("%d%s", len([]rune(typ.TypeName())), typ.TypeName())
			}
		}
	default:
		panic("")
	}
}

func (v *Module) MangledName(typ MangleType) string {
	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_M%d%s", len(v.Name), v.Name)

		// TODO parent module

		return result
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
		result := fmt.Sprintf("_F%d%s", len(v.Name), v.Name)
		for _, arg := range v.Parameters {
			result += TypeMangledName(typ, arg.Variable.Type)
		}

		// TODO check parent struct

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
		if v.ParentStruct == nil {
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

func (v *TraitType) MangledName(typ MangleType) string {
	return ""
}

func (v *EnumType) MangledName(typ MangleType) string {
	return ""
}
