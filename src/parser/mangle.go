package parser

import "fmt"

// In case we support multiple name mangline schemes
type MangleType int

const (
	MANGLE_ARK_UNSTABLE MangleType = iota // see https://github.com/ark-lang/ark/src/issues/401
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

func (v *Function) MangledName(typ MangleType) string {
	if v.Name == "main" {
		return "main" // TODO make sure only one main function
	}

	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_F_%d%s", len([]rune(v.Name)), v.Name)
		for _, arg := range v.Parameters {
			result += TypeMangledName(typ, arg.Variable.Type)
		}
		// TODO struct, module
		return result
	default:
		panic("")
	}
}

func (v *Variable) MangledName(typ MangleType) string {
	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_V_%d%s", len([]rune(v.Name)), v.Name)
		if v.ParentStruct != nil {
			result = v.ParentStruct.MangledName(typ) + result
		} else {
			// result = v.ParentModule.MangledName(typ) + result TODO
		}
		return result
	default:
		panic("")
	}
}

func (v *StructType) MangledName(typ MangleType) string {
	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_S_%d%s", len([]rune(v.Name)), v.Name)
		// TODO module
		return result
	default:
		panic("")
	}
}

func (v *TraitType) MangledName(typ MangleType) string {
	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_T_%d%s", len([]rune(v.Name)), v.Name)
		// TODO module
		return result
	default:
		panic("")
	}
}

func (v *EnumType) MangledName(typ MangleType) string {
	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_E_%d%s", len([]rune(v.Name)), v.Name)
		// TODO module
		return result
	default:
		panic("")
	}
}
