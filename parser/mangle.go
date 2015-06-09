package parser

import "fmt"

// In case we support multiple name mangline schemes
type MangleType int

const (
	MANGLE_ARK_UNSTABLE MangleType = iota // see https://github.com/ark-lang/ark/issues/401
)

func (v *Function) MangledName(typ MangleType) string {
	if v.Name == "main" {
		return "main" // TODO make sure only one main function
	}

	switch typ {
	case MANGLE_ARK_UNSTABLE:
		result := fmt.Sprintf("_F_%d%s", len([]rune(v.Name)), v.Name)
		for _, arg := range v.Parameters {
			result += fmt.Sprintf("_%d%s", len([]rune(arg.Variable.Type.TypeName())), arg.Variable.Type.TypeName())
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
