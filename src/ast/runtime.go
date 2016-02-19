package ast

type RuntimeTypes struct {
	RuneType Type
	StringType Type
}

var runeType Type
var stringType Type

func SetRuntimeTypes(val *RuntimeTypes) {
	if val.RuneType == nil {
		panic("INTERNAL ERROR: Rune type missing from runtime")
	}
	runeType = val.RuneType
	
	if val.StringType == nil {
		panic("INTERNAL ERROR: String type missing from runtime")
        }
	stringType = val.StringType

	builtinScope.InsertType(runeType, true)
	builtinScope.InsertType(stringType, true)
}

