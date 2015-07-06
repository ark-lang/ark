package parser

import (
	"github.com/ark-lang/ark/util"
)

type Type interface {
	TypeName() string
	RawType() Type            // type disregarding pointers
	LevelsOfIndirection() int // number of pointers you have to go through to get to the actual type
	IsIntegerType() bool      // true for all int types
	IsFloatingType() bool     // true for all floating-point types
	IsSigned() bool           // true for all signed integer types
	CanCastTo(Type) bool      // true if the receiver can be typecast to the parameter
	Attrs() AttrGroup         // fetches the attributes associated with the type
	Equals(Type) bool         // compares whether two types are equal

	// TODO: Should this be here?
	resolveType(Locatable, *Resolver, *Scope) Type
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

func (v PrimitiveType) Attrs() AttrGroup {
	return nil
}

func (v PrimitiveType) Equals(t Type) bool {
	other, ok := t.(PrimitiveType)
	if !ok {
		return false
	}
	return v == other
}

// StructType

type StructType struct {
	Name      string
	Variables []*VariableDecl
	attrs     AttrGroup
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

func (v *StructType) Attrs() AttrGroup {
	return v.attrs
}

func (v *StructType) Equals(t Type) bool {
	// TODO: Check for struct equality
	panic("please implement the rest of this, if we ever need it")

	other, ok := t.(*StructType)
	if !ok {
		return false
	}

	if v.Name != other.Name {
		return false
	}

	if !v.Attrs().Equals(other.Attrs()) {
		return false
	}

	// TODO: Check struct members
	return true
}

// ArrayType

type ArrayType struct {
	MemberType Type
	attrs      AttrGroup
}

// IMPORTANT:
// Using this function is no longer important, just make sure to use
// .Equals() to compare two types.
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

func (v ArrayType) Attrs() AttrGroup {
	return v.attrs
}

func (v ArrayType) Equals(t Type) bool {
	other, ok := t.(ArrayType)
	if !ok {
		return false
	}

	if !v.Attrs().Equals(other.Attrs()) {
		return false
	}

	if !v.MemberType.Equals(other.MemberType) {
		return false
	}

	return true
}

// TraitType

type TraitType struct {
	Name      string
	Functions []*FunctionDecl
	attrs     AttrGroup
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

func (v *TraitType) Attrs() AttrGroup {
	return v.attrs
}

func (v *TraitType) Equals(t Type) bool {
	// TODO: Check for trait equality
	panic("please implement the rest of this, if we ever need it")

	other, ok := t.(*TraitType)
	if !ok {
		return false
	}

	if v.Name != other.Name {
		return false
	}

	if !v.Attrs().Equals(other.Attrs()) {
		return false
	}

	// TODO: Check trait function types
	return true
}

// PointerType

type PointerType struct {
	Addressee Type
}

// IMPORTANT:
// Using this function is no longer important, just make sure to use
// .Equals() to compare two types.
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

func (v PointerType) Attrs() AttrGroup {
	return nil
}

func (v PointerType) IsSigned() bool {
	return false
}

func (v PointerType) Equals(t Type) bool {
	other, ok := t.(PointerType)
	if !ok {
		return false
	}
	return v == other
}

// TupleType

func tupleOf(types ...Type) Type {
	if len(types) == 1 {
		return types[0]
	}
	return &TupleType{Members: types}
}

type TupleType struct {
	Members []Type
}

func (v *TupleType) String() string {
	result := "(" + util.Blue("TupleType") + ": "
	for _, mem := range v.Members {
		result += "\t" + mem.TypeName() + "\n"
	}
	return result + ")"
}

func (v *TupleType) TypeName() string {
	result := "|"
	for idx, mem := range v.Members {
		result += mem.TypeName()

		// if we are not at the last component
		if idx < len(v.Members)-1 {
			result += ", "
		}
	}
	result += "|"
	return result
}

func (v *TupleType) RawType() Type {
	return v
}

func (v *TupleType) IsSigned() bool {
	return false
}

func (v *TupleType) LevelsOfIndirection() int {
	return 0
}

func (v *TupleType) IsIntegerType() bool {
	return false
}

func (v *TupleType) IsFloatingType() bool {
	return false
}

func (v *TupleType) CanCastTo(t Type) bool {
	return false
}

func (v *TupleType) addMember(decl Type) {
	v.Members = append(v.Members, decl)
}

func (v *TupleType) Attrs() AttrGroup {
	return nil
}

func (v *TupleType) Equals(t Type) bool {
	other, ok := t.(*TupleType)
	if !ok {
		return false
	}

	if len(v.Members) != len(other.Members) {
		return false
	}

	for idx, mem := range v.Members {
		if mem != other.Members[idx] {
			return false
		}
	}

	return true
}

type UnresolvedType struct {
	Name string
}

func (v *UnresolvedType) String() string {
	return "(" + util.Blue("UnresolvedType") + ": " + v.Name + ")"
}

func (v *UnresolvedType) TypeName() string {
	return v.Name
}

func (v *UnresolvedType) RawType() Type {
	panic("RawType() invalid on UnresolvedType")
}

func (v *UnresolvedType) IsSigned() bool {
	panic("IsSigned() invalid on UnresolvedType")
}

func (v *UnresolvedType) LevelsOfIndirection() int {
	panic("LevelsOfIndirection() invalid on UnresolvedType")
}

func (v *UnresolvedType) IsIntegerType() bool {
	panic("IsIntegerType() invalid on UnresolvedType")
}

func (v *UnresolvedType) IsFloatingType() bool {
	panic("IsFloatingType() invalid on UnresolvedType")
}

func (v *UnresolvedType) CanCastTo(t Type) bool {
	panic("CanCastTo() invalid on UnresolvedType")
}

func (v *UnresolvedType) Attrs() AttrGroup {
	panic("Attrs() invalid on UnresolvedType")
}

func (v *UnresolvedType) Equals(t Type) bool {
	panic("Equals() invalid on UnresolvedType")
}
