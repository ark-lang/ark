package parser

import (
	"github.com/ark-lang/ark/src/util"
)

type Type interface {
	TypeName() string
	LevelsOfIndirection() int // number of pointers you have to go through to get to the actual type
	IsIntegerType() bool      // true for all int types
	IsFloatingType() bool     // true for all floating-point types
	IsSigned() bool           // true for all signed integer types
	CanCastTo(Type) bool      // true if the receiver can be typecast to the parameter
	Attrs() AttrGroup         // fetches the attributes associated with the type
	Equals(Type) bool         // compares whether two types are equal
	ActualType() Type         // returns the actual type disregarding named types

	resolveType(Locatable, *Resolver, *Scope) Type
}

//go:generate stringer -type=PrimitiveType
type PrimitiveType int

const (
	PRIMITIVE_s8 PrimitiveType = iota
	PRIMITIVE_s16
	PRIMITIVE_s32
	PRIMITIVE_s64
	PRIMITIVE_s128

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
	case PRIMITIVE_s8, PRIMITIVE_s16, PRIMITIVE_s32, PRIMITIVE_s64, PRIMITIVE_s128,
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
	case PRIMITIVE_s8, PRIMITIVE_s16, PRIMITIVE_s32, PRIMITIVE_s64, PRIMITIVE_s128, PRIMITIVE_int:
		return true
	default:
		return false
	}
}

func (v PrimitiveType) TypeName() string {
	return v.String()[10:]
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

func (v PrimitiveType) ActualType() Type {
	return v
}

// StructType

type StructType struct {
	Variables    []*VariableDecl
	attrs        AttrGroup
	ParentEnum   *EnumType
	ParentModule *Module
}

func (v *StructType) String() string {
	result := "(" + util.Blue("StructType") + ": "
	for _, attr := range v.attrs {
		result += attr.String() + " "
	}
	result += "\n"
	for _, decl := range v.Variables {
		result += "\t" + decl.String() + "\n"
	}
	return result + ")"
}

func (v *StructType) TypeName() string {
	res := "struct { "

	for i, variable := range v.Variables {
		res += variable.Variable.Name + ": " + variable.Variable.Type.TypeName()

		if i < len(v.Variables)-1 {
			res += ", "
		} else {
			res += " "
		}
	}

	return res + "}"
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

func (v *StructType) GetVariableDecl(s string) *VariableDecl {
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
	other, ok := t.(*StructType)
	if !ok {
		return false
	}

	if !v.Attrs().Equals(other.Attrs()) {
		return false
	}

	if len(v.Variables) != len(other.Variables) {
		return false
	}

	for idx, _ := range v.Variables {
		variable, otherVariable := v.Variables[idx].Variable, other.Variables[idx].Variable
		if variable.Name != otherVariable.Name {
			return false
		}
		if !variable.Type.Equals(otherVariable.Type) {
			return false
		}
	}

	return true
}

func (v *StructType) ActualType() Type {
	return v
}

// NamedType

type NamedType struct {
	Name string
	Type Type
}

func (v *NamedType) ActualType() Type {
	return v.Type.ActualType()
}

func (v *NamedType) String() string {
	return "(" + util.Blue("NamedType") + ": " + v.Type.TypeName() + ")"
}

func (v *NamedType) TypeName() string {
	return v.Name
}
func (v *NamedType) IsSigned() bool {
	return v.Type.IsSigned()
}

func (v *NamedType) LevelsOfIndirection() int {
	return v.Type.LevelsOfIndirection()
}

func (v *NamedType) IsIntegerType() bool {
	return v.Type.IsIntegerType()
}

func (v *NamedType) IsFloatingType() bool {
	return v.Type.IsFloatingType()
}

func (v *NamedType) CanCastTo(t Type) bool {
	return v.ActualType().CanCastTo(t)
}

func (v *NamedType) Attrs() AttrGroup {
	return v.Type.Attrs()
}

func (v *NamedType) Equals(t Type) bool {
	if other, ok := t.(*NamedType); ok {
		return v == other
	}

	return false
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

func (v ArrayType) ActualType() Type {
	return v
}

// TraitType

type TraitType struct {
	Name         string
	Functions    []*FunctionDecl
	attrs        AttrGroup
	ParentModule *Module
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

	panic("please implement the rest of this, if we ever need it")

	// TODO: Check trait function types
	//return true
}

func (v *TraitType) ActualType() Type {
	return v
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

func (v PointerType) LevelsOfIndirection() int {
	return v.Addressee.LevelsOfIndirection() + 1
}

func (v PointerType) IsIntegerType() bool {
	return true
}

func (v PointerType) IsFloatingType() bool {
	return false
}

func (v PointerType) CanCastTo(t Type) bool {
	if t.IsIntegerType() {
		return true
	}
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

	return v.Addressee.Equals(other.Addressee)
}

func (v PointerType) ActualType() Type {
	return v
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
	result := "("
	for idx, mem := range v.Members {
		result += mem.TypeName()

		// if we are not at the last component
		if idx < len(v.Members)-1 {
			result += ", "
		}
	}
	result += ")"
	return result
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
	return v != t && v.Equals(t.ActualType())
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

func (v *TupleType) ActualType() Type {
	return v
}

// EnumType
type EnumType struct {
	Name         string
	Simple       bool
	Members      []EnumTypeMember
	attrs        AttrGroup
	ParentModule *Module
}

type EnumTypeMember struct {
	Name string
	Type Type
	Tag  int
}

func (v *EnumType) String() string {
	result := "(" + util.Blue("EnumType") + ": "
	for _, attr := range v.attrs {
		result += attr.String() + " "
	}
	result += v.Name + "\n"
	for _, mem := range v.Members {
		result += "\t" + mem.Name + ": " + mem.Type.TypeName() + "\n"
	}
	return result + util.Magenta(" <"+v.MangledName(MANGLE_ARK_UNSTABLE)+"> ") + ")"
}

func (v *EnumType) TypeName() string {
	return v.Name
}

func (v *EnumType) IsSigned() bool {
	return false
}

func (v *EnumType) LevelsOfIndirection() int {
	return 0
}

func (v *EnumType) IsIntegerType() bool {
	return v.Simple
}

func (v *EnumType) IsFloatingType() bool {
	return false
}

func (v *EnumType) CanCastTo(t Type) bool {
	return v.Simple && t.IsIntegerType()
}

func (v *EnumType) MemberIndex(name string) int {
	for idx, member := range v.Members {
		if member.Name == name {
			return idx
		}
	}
	return -1
}

func (v *EnumType) Attrs() AttrGroup {
	return v.attrs
}

func (v *EnumType) Equals(t Type) bool {
	other, ok := t.(*EnumType)
	if !ok {
		return false
	}

	if v.Name != other.Name {
		return false
	}

	if !v.Attrs().Equals(other.Attrs()) {
		return false
	}

	if len(v.Members) != len(other.Members) {
		return false
	}

	for idx, member := range v.Members {
		otherMember := other.Members[idx]

		if member.Name != otherMember.Name {
			return false
		}
		if !member.Type.Equals(otherMember.Type) {
			return false
		}
		if member.Tag != otherMember.Tag {
			return false
		}
	}

	return true
}

func (v *EnumType) ActualType() Type {
	return v
}

// UnresolvedType
type UnresolvedType struct {
	Name unresolvedName
}

func (v *UnresolvedType) String() string {
	return "(" + util.Blue("UnresolvedType") + ": " + v.Name.String() + ")"
}

func (v *UnresolvedType) TypeName() string {
	return v.Name.String()
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

func (v *UnresolvedType) ActualType() Type {
	return v
}
