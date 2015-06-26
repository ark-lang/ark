package parser

import "github.com/ark-lang/ark/util"

type Type interface {
	TypeName() string
	RawType() Type            // type disregarding pointers
	LevelsOfIndirection() int // number of pointers you have to go through to get to the actual type
	IsIntegerType() bool      // true for all int types
	IsFloatingType() bool     // true for all floating-point types
	IsSigned() bool           // true for all signed integer types
	CanCastTo(Type) bool      // true if the receiver can be typecast to the parameter
	Attrs() []*Attr           // fetches the attributes associated with the type
}

//go:generate stringer -type=PrimitiveType
type PrimitiveType int

const (
	PRIMITIVE_s8 PrimitiveType = iota
	PRIMITIVE_s16
	PRIMITIVE_s32
	PRIMITIVE_s64
	PRIMITIVE_i128

	PRIMITIVE_u8
	PRIMITIVE_u16
	PRIMITIVE_u32
	PRIMITIVE_u64
	PRIMITIVE_u128

	PRIMITIVE_f32
	PRIMITIVE_f64
	PRIMITIVE_f128

	PRIMITIVE_str
	PRIMITIVE_rune

	PRIMITIVE_int
	PRIMITIVE_uint

	PRIMITIVE_bool

	PRIMITIVE_void
)

func (v PrimitiveType) IsIntegerType() bool {
	switch v {
	case PRIMITIVE_s8, PRIMITIVE_s16, PRIMITIVE_s32, PRIMITIVE_s64, PRIMITIVE_i128,
		PRIMITIVE_u8, PRIMITIVE_u16, PRIMITIVE_u32, PRIMITIVE_u64, PRIMITIVE_u128,
		PRIMITIVE_int, PRIMITIVE_uint:
		return true
	default:
		return false
	}
}

func (v PrimitiveType) IsFloatingType() bool {
	switch v {
	case PRIMITIVE_f32, PRIMITIVE_f64, PRIMITIVE_f128:
		return true
	default:
		return false
	}
}

func (v PrimitiveType) IsSigned() bool {
	switch v {
	case PRIMITIVE_s8, PRIMITIVE_s16, PRIMITIVE_s32, PRIMITIVE_s64, PRIMITIVE_i128, PRIMITIVE_int:
		return true
	default:
		return false
	}
}

func (v PrimitiveType) TypeName() string {
	return v.String()[10:]
}

func (v PrimitiveType) RawType() Type {
	return v
}

func (v PrimitiveType) LevelsOfIndirection() int {
	return 0
}

func (v PrimitiveType) CanCastTo(t Type) bool {
	return (v.IsIntegerType() || v.IsFloatingType() || v == PRIMITIVE_rune) &&
		(t.IsFloatingType() || t.IsIntegerType() || t == PRIMITIVE_rune)
}

func (v PrimitiveType) Attrs() []*Attr {
	return nil
}

// StructType

type StructType struct {
	Name      string
	Variables []*VariableDecl
	attrs     []*Attr
}

func (v *StructType) String() string {
	result := "(" + util.Blue("StructType") + ": "
	for _, attr := range v.attrs {
		result += attr.String() + " "
	}
	result += v.Name + "\n"
	for _, decl := range v.Variables {
		result += "\t" + decl.String() + "\n"
	}
	return result + util.Magenta(" <"+v.MangledName(MANGLE_ARK_UNSTABLE)+"> ") + ")"
}

func (v *StructType) TypeName() string {
	return v.Name
}

func (v *StructType) RawType() Type {
	return v
}

func (v *StructType) IsSigned() bool {
	return false
}

func (v *StructType) LevelsOfIndirection() int {
	return 0
}

func (v *StructType) IsIntegerType() bool {
	return false
}

func (v *StructType) IsFloatingType() bool {
	return false
}

func (v *StructType) CanCastTo(t Type) bool {
	return false
}

func (v *StructType) getVariableDecl(s string) *VariableDecl {
	for _, decl := range v.Variables {
		if decl.Variable.Name == s {
			return decl
		}
	}
	return nil
}

func (v *StructType) addVariableDecl(decl *VariableDecl) {
	v.Variables = append(v.Variables, decl)
	decl.Variable.ParentStruct = v
}

func (v *StructType) VariableIndex(d *Variable) int {
	for i, decl := range v.Variables {
		if decl.Variable == d {
			return i
		}
	}
	return -1
}

func (v *StructType) Attrs() []*Attr {
	return v.attrs
}

// ArrayType

type ArrayType struct {
	MemberType Type
	attrs      []*Attr
}

// IMPORTANT:
// Use this method to make ArrayTypes. Most importantly, NEVER take the
// address of a ArrayType! ie, using &ArrayType{}. This would make two
// ArrayTypes to the same PrimitiveType inequal, when they should be equal.
func arrayOf(t Type) ArrayType {
	return ArrayType{MemberType: t}
}

func (v ArrayType) String() string {
	result := "(" + util.Blue("ArrayType") + ": "
	for _, attr := range v.attrs {
		result += attr.String() + " "
	}
	return result + v.TypeName() + ")" //+ util.Magenta(" <"+v.MangledName(MANGLE_ARK_UNSTABLE)+"> ") + ")"
}

func (v ArrayType) TypeName() string {
	return "[]" + v.MemberType.TypeName()
}

func (v ArrayType) RawType() Type {
	return v
}

func (v ArrayType) IsSigned() bool {
	return false
}

func (v ArrayType) LevelsOfIndirection() int {
	return 0
}

func (v ArrayType) IsIntegerType() bool {
	return false
}

func (v ArrayType) IsFloatingType() bool {
	return false
}

func (v ArrayType) CanCastTo(t Type) bool {
	return false
}

func (v ArrayType) Attrs() []*Attr {
	return v.attrs
}

// TraitType

type TraitType struct {
	Name      string
	Functions []*FunctionDecl
	attrs     []*Attr
}

func (v *TraitType) String() string {
	result := "(" + util.Blue("TraitType") + ": "
	for _, attr := range v.attrs {
		result += attr.String() + " "
	}
	result += v.Name + "\n"
	for _, decl := range v.Functions {
		result += "\t" + decl.String() + "\n"
	}
	return result + util.Magenta(" <"+v.MangledName(MANGLE_ARK_UNSTABLE)+"> ") + ")"
}

func (v *TraitType) TypeName() string {
	return v.Name
}

func (v *TraitType) RawType() Type {
	return v
}

func (v *TraitType) IsSigned() bool {
	return false
}

func (v *TraitType) LevelsOfIndirection() int {
	return 0
}

func (v *TraitType) IsIntegerType() bool {
	return false
}

func (v *TraitType) IsFloatingType() bool {
	return false
}

func (v *TraitType) CanCastTo(t Type) bool {
	return false
}

func (v *TraitType) getFunctionDecl(s string) *FunctionDecl {
	for _, decl := range v.Functions {
		if decl.Function.Name == s {
			return decl
		}
	}
	return nil
}

func (v *TraitType) addFunctionDecl(decl *FunctionDecl) {
	v.Functions = append(v.Functions, decl)
}

func (v *TraitType) Attrs() []*Attr {
	return v.attrs
}

// PointerType

type PointerType struct {
	Addressee Type
}

// IMPORTANT:
// Use this method to make PointerTypes. Most importantly, NEVER take the
// address of a PointerType! ie, using &PointerType{}. This would make two
// PointerTypes to the same PrimitiveType inequal, when they should be equal.
func pointerTo(t Type) PointerType {
	return PointerType{Addressee: t}
}

func (v PointerType) TypeName() string {
	return "^" + v.Addressee.TypeName()
}

func (v PointerType) RawType() Type {
	return v.Addressee.RawType()
}

func (v PointerType) LevelsOfIndirection() int {
	return v.Addressee.LevelsOfIndirection() + 1
}

func (v PointerType) IsIntegerType() bool {
	return false
}

func (v PointerType) IsFloatingType() bool {
	return false
}

func (v PointerType) CanCastTo(t Type) bool {
	return false
}

func (v PointerType) Attrs() []*Attr {
	return nil
}

func (v PointerType) IsSigned() bool {
	return false
}
