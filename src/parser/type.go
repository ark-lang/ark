package parser

// IDEA for rewrite: have separate:
// --> Type - raw type info
// --> TypeInstance - Type with generic arguments (ie. a complete type)
// --> TypeReference - an AST element for a literal reference to a type (used for warning/error messages etc)
// Currently, TypeReference takes the roles of both TypeInstance and TypeReference.

import (
	"fmt"
	"reflect"

	"github.com/ark-lang/ark/src/util"
)

func IsPointerOrReferenceType(t Type) bool {
	switch t.ActualType().(type) {
	case PointerType, ReferenceType:
		return true
	}
	return false
}

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
	IsVoidType() bool
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

	PRIMITIVE_int
	PRIMITIVE_uint
	PRIMITIVE_uintptr

	PRIMITIVE_bool
	PRIMITIVE_void
)

func (v PrimitiveType) IsVoidType() bool {
	return v == PRIMITIVE_void
}

func (v PrimitiveType) IsIntegerType() bool {
	switch v {
	case PRIMITIVE_s8, PRIMITIVE_s16, PRIMITIVE_s32, PRIMITIVE_s64, PRIMITIVE_s128,
		PRIMITIVE_u8, PRIMITIVE_u16, PRIMITIVE_u32, PRIMITIVE_u64, PRIMITIVE_u128,
		PRIMITIVE_int, PRIMITIVE_uint, PRIMITIVE_uintptr:
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
	if v == PRIMITIVE_uintptr && IsPointerOrReferenceType(t) {
		return true
	}

	return (v.IsIntegerType() || v.IsFloatingType()) &&
		(t.IsFloatingType() || t.IsIntegerType())
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
	Members           []*StructMember
	attrs             AttrGroup
	GenericParameters GenericSigil
}

type StructMember struct {
	Name string
	Type *TypeReference
}

func (v StructType) String() string {
	result := "(" + util.Blue("StructType") + ": "
	result += v.attrs.String()
	result += "\n"
	for _, mem := range v.Members {
		result += "\t" + mem.Name + ": " + mem.Type.String() + "\n"
	}
	return result + ")"
}

func (v StructType) TypeName() string {
	res := "struct" + v.GenericParameters.String() + " {"

	for i, mem := range v.Members {
		res += mem.Name + ": " + mem.Type.String()

		if i < len(v.Members)-1 {
			res += ", "
		}
	}

	return res + "}"
}

func (v StructType) IsSigned() bool {
	return false
}

func (v StructType) LevelsOfIndirection() int {
	return 0
}

func (v StructType) IsIntegerType() bool {
	return false
}

func (v StructType) IsVoidType() bool {
	return false
}

func (v StructType) IsFloatingType() bool {
	return false
}

func (v StructType) CanCastTo(t Type) bool {
	return false
}

func (v StructType) GetMember(name string) *StructMember {
	idx := v.MemberIndex(name)
	if idx != -1 {
		return v.Members[idx]
	}
	return nil
}

func (v StructType) addMember(name string, typ *TypeReference) StructType {
	v.Members = append(v.Members, &StructMember{Name: name, Type: typ})
	return v
}

func (v StructType) MemberIndex(name string) int {
	for idx, mem := range v.Members {
		if mem.Name == name {
			return idx
		}
	}
	return -1
}

func (v StructType) Attrs() AttrGroup {
	return v.attrs
}

func (v StructType) Equals(t Type) bool {
	other, ok := t.(StructType)
	if !ok {
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
	}

	return true
}

func (v StructType) ActualType() Type {
	return v
}

// NamedType

type NamedType struct {
	Name          string
	Type          Type
	ParentModule  *Module
	Methods       []*Function
	StaticMethods []*Function
}

func (v *NamedType) addMethod(fn *Function) {
	v.Methods = append(v.Methods, fn)
}

func (v *NamedType) addStaticMethod(fn *Function) {
	v.StaticMethods = append(v.StaticMethods, fn)
}

func (v *NamedType) GetMethod(name string) *Function {
	for _, fn := range v.Methods {
		if fn.Name == name {
			return fn
		}
	}

	return nil
}

func (v *NamedType) GetStaticMethod(name string) *Function {
	for _, fn := range v.StaticMethods {
		if fn.Name == name {
			return fn
		}
	}
	return nil
}

func (v *NamedType) ActualType() Type {
	return v.Type.ActualType()
}

func (v *NamedType) String() string {
	res := "(" + util.Blue("NamedType") + ": " + v.Name
	return res + " = " + v.Type.TypeName() + ")"
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

func (v *NamedType) IsVoidType() bool {
	return v.Type.IsVoidType()
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
	other, ok := t.(*NamedType)
	if !ok {
		return false
	}

	if v.ParentModule != other.ParentModule {
		return false
	}

	if v.Name != other.Name {
		return false
	}

	return true
}

// ArrayType

type ArrayType struct {
	MemberType *TypeReference

	IsFixedLength bool
	Length        int // TODO change to uint64

	attrs AttrGroup
}

// IMPORTANT:
// Using this function is no longer important, just make sure to use
// .Equals() to compare two types.
func ArrayOf(t *TypeReference, isFixedLength bool, length int) ArrayType {
	return ArrayType{MemberType: t, IsFixedLength: isFixedLength, Length: length}
}

func (v ArrayType) String() string {
	result := "(" + util.Blue("ArrayType") + ": "
	for _, attr := range v.attrs {
		result += attr.String() + " "
	}
	return result + v.TypeName() + ")" //+ util.Magenta(" <"+v.MangledName(MANGLE_ARK_UNSTABLE)+"> ") + ")"
}

func (v ArrayType) TypeName() string {
	var l string
	if v.IsFixedLength {
		l = fmt.Sprintf("%d", v.Length)
	}
	return "[" + l + "]" + v.MemberType.String()
}

func (v ArrayType) IsSigned() bool {
	return false
}

func (v ArrayType) LevelsOfIndirection() int {
	return 0
}

func (v ArrayType) IsVoidType() bool {
	return false
}

func (v ArrayType) IsIntegerType() bool {
	return false
}

func (v ArrayType) IsFloatingType() bool {
	return false
}

func (v ArrayType) CanCastTo(t Type) bool {
	return t.ActualType().Equals(v)
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

	if v.IsFixedLength != other.IsFixedLength {
		return false
	}

	if v.IsFixedLength && v.Length != other.Length {
		return false
	}

	return true
}

func (v ArrayType) ActualType() Type {
	return v
}

// Reference

type ReferenceType struct {
	Referrer  *TypeReference // TODO rename to Element
	IsMutable bool
}

func ReferenceTo(t *TypeReference, mutable bool) ReferenceType {
	return ReferenceType{Referrer: t, IsMutable: mutable}
}

func (v ReferenceType) TypeName() string {
	if v.IsMutable {
		return "&mut " + v.Referrer.String()
	}
	return "&" + v.Referrer.String()
}

func (v ReferenceType) LevelsOfIndirection() int {
	return v.Referrer.BaseType.LevelsOfIndirection() + 1
}

func (v ReferenceType) IsIntegerType() bool {
	return false
}

func (v ReferenceType) IsFloatingType() bool {
	return false
}

func (v ReferenceType) IsVoidType() bool {
	return false
}

func (v ReferenceType) CanCastTo(t Type) bool {
	return IsPointerOrReferenceType(t) || t.ActualType() == PRIMITIVE_uintptr
}

func (v ReferenceType) Attrs() AttrGroup {
	return nil
}

func (v ReferenceType) IsSigned() bool {
	return false
}

func (v ReferenceType) Equals(t Type) bool {
	ref, ok := t.(ReferenceType)
	if ok {
		return v.IsMutable == ref.IsMutable && v.Referrer.Equals(ref.Referrer)
	}

	/*ptr, ok := t.(PointerType)
	if ok {
		return v.Referrer.Equals(ptr.Addressee)
	}*/

	return false
}

func (v ReferenceType) ActualType() Type {
	return v
}

// PointerType

type PointerType struct {
	Addressee *TypeReference
	IsMutable bool
}

// IMPORTANT:
// Using this function is no longer important, just make sure to use
// .Equals() to compare two types.
func PointerTo(t *TypeReference, mutable bool) PointerType {
	return PointerType{Addressee: t, IsMutable: mutable}
}

func (v PointerType) TypeName() string {
	if v.IsMutable {
		return "^mut " + v.Addressee.String()
	}
	return "^" + v.Addressee.String()
}

func (v PointerType) LevelsOfIndirection() int {
	return v.Addressee.BaseType.LevelsOfIndirection() + 1
}

func (v PointerType) IsIntegerType() bool {
	return false
}

func (v PointerType) IsFloatingType() bool {
	return false
}

func (v PointerType) IsVoidType() bool {
	return false
}

func (v PointerType) CanCastTo(t Type) bool {
	return IsPointerOrReferenceType(t) || t.ActualType() == PRIMITIVE_uintptr
}

func (v PointerType) Attrs() AttrGroup {
	return nil
}

func (v PointerType) IsSigned() bool {
	return false
}

func (v PointerType) Equals(t Type) bool {
	other, ok := t.(PointerType)
	if ok {
		return v.IsMutable == other.IsMutable && v.Addressee.Equals(other.Addressee)
	}

	return false
}

func (v PointerType) ActualType() Type {
	return v
}

// TupleType

func tupleOf(types ...*TypeReference) Type {
	return TupleType{Members: types}
}

type TupleType struct {
	Members []*TypeReference
}

func (v TupleType) String() string {
	result := "(" + util.Blue("TupleType") + ": "
	for _, mem := range v.Members {
		result += "\t" + mem.String() + "\n"
	}
	return result + ")"
}

func (v TupleType) TypeName() string {
	result := "("
	for idx, mem := range v.Members {
		result += mem.String()

		// if we are not at the last component
		if idx < len(v.Members)-1 {
			result += ", "
		}
	}
	result += ")"
	return result
}

func (v TupleType) IsSigned() bool {
	return false
}

func (v TupleType) LevelsOfIndirection() int {
	return 0
}

func (v TupleType) IsIntegerType() bool {
	return false
}

func (v TupleType) IsFloatingType() bool {
	return false
}

func (v TupleType) IsVoidType() bool {
	return false
}

func (v TupleType) CanCastTo(t Type) bool {
	return v.Equals(t.ActualType())
}

func (v TupleType) addMember(decl *TypeReference) {
	v.Members = append(v.Members, decl)
}

func (v TupleType) Attrs() AttrGroup {
	return nil
}

func (v TupleType) Equals(t Type) bool {
	other, ok := t.(TupleType)
	if !ok {
		return false
	}

	if len(v.Members) != len(other.Members) {
		return false
	}

	for idx, mem := range v.Members {
		if !mem.Equals(other.Members[idx]) {
			return false
		}
	}

	return true
}

func (v TupleType) ActualType() Type {
	return v
}

// InterfaceType

type InterfaceType struct {
	Functions []*Function
	attrs     AttrGroup
}

func (v InterfaceType) String() string {
	result := "(" + util.Blue("InterfaceType") + ": "
	for _, function := range v.Functions {
		result += "\t" + function.String() + "\n"
	}
	result += "}"
	return result + ")"
}

func (v InterfaceType) TypeName() string {
	result := "interface {\n"
	for _, function := range v.Functions {
		result += "\t" + function.String() + "\n"
	}
	result += "}"
	return result
}

func (v InterfaceType) IsSigned() bool {
	return false
}

func (v InterfaceType) LevelsOfIndirection() int {
	return 0
}

func (v InterfaceType) IsIntegerType() bool {
	return false
}

func (v InterfaceType) IsFloatingType() bool {
	return false
}

func (v InterfaceType) IsVoidType() bool {
	return false
}

func (v InterfaceType) CanCastTo(t Type) bool {
	return v.Equals(t.ActualType())
}

func (v InterfaceType) addFunction(fn *Function) InterfaceType {
	v.Functions = append(v.Functions, fn)
	return v
}

func (v InterfaceType) GetFunction(name string) *Function {
	for _, fn := range v.Functions {
		if fn.Name == name {
			return fn
		}
	}

	return nil
}

func (v InterfaceType) MatchesType(t Type) {

}

func (v InterfaceType) Attrs() AttrGroup {
	return nil
}

func (v InterfaceType) Equals(t Type) bool {
	other, ok := t.(InterfaceType)
	if !ok {
		return false
	}

	if len(v.Functions) != len(other.Functions) {
		return false
	}

	for idx, mem := range v.Functions {
		if mem != other.Functions[idx] {
			return false
		}
	}

	return true
}

func (v InterfaceType) ActualType() Type {
	return v
}

// EnumType
type EnumType struct {
	Simple            bool
	GenericParameters GenericSigil
	Members           []EnumTypeMember
	attrs             AttrGroup
}

type EnumTypeMember struct {
	Name string
	Type Type
	Tag  int
}

func (v EnumType) GetMember(name string) (EnumTypeMember, bool) {
	for _, member := range v.Members {
		if member.Name == name {
			return member, true
		}
	}
	return EnumTypeMember{}, false
}

func (v EnumType) String() string {
	result := "(" + util.Blue("EnumType") + ": "
	result += v.attrs.String()

	result += "\n"

	for _, mem := range v.Members {
		result += "\t" + mem.Name + ": " + mem.Type.TypeName() + "\n"
	}
	return result + ")"
}

func (v EnumType) TypeName() string {
	res := "enum" + v.GenericParameters.String() + " {"

	for idx, mem := range v.Members {
		res += mem.Name + ": " + mem.Type.TypeName()
		if idx < len(v.Members)-1 {
			res += ", "
		}
	}

	return res + "}"
}

func (v EnumType) IsSigned() bool {
	return false
}

func (v EnumType) LevelsOfIndirection() int {
	return 0
}

func (v EnumType) IsIntegerType() bool {
	return v.Simple
}

func (v EnumType) IsFloatingType() bool {
	return false
}

func (v EnumType) IsVoidType() bool {
	return false
}

func (v EnumType) CanCastTo(t Type) bool {
	return v.Simple && t.IsIntegerType()
}

func (v EnumType) MemberIndex(name string) int {
	for idx, member := range v.Members {
		if member.Name == name {
			return idx
		}
	}
	return -1
}

func (v EnumType) Attrs() AttrGroup {
	return v.attrs
}

func (v EnumType) Equals(t Type) bool {
	other, ok := t.(EnumType)
	if !ok {
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

func (v EnumType) ActualType() Type {
	return v
}

// FunctionType
type FunctionType struct {
	attrs AttrGroup

	GenericParameters GenericSigil
	Parameters        []*TypeReference
	Return            *TypeReference
	IsVariadic        bool

	Receiver Type // non-nil if non-static method
}

func (v FunctionType) String() string {
	result := "(" + util.Blue("FunctionType") + ": "
	return result + ")"
}

func (v FunctionType) TypeName() string {
	res := ""

	res += v.attrs.String()
	if res != "" {
		res += " "
	}

	res += "func"

	if v.Receiver != nil {
		res += "(v: " + v.Receiver.TypeName() + ")"
	}

	res += v.GenericParameters.String()

	res += "("

	for idx, para := range v.Parameters {
		res += para.String()
		if idx < len(v.Parameters)-1 {
			res += ", "
		}
	}

	res += ")"

	if v.Return != nil {
		res += " -> " + v.Return.String()
	}

	return res
}

func (v FunctionType) IsSigned() bool {
	return false
}

func (v FunctionType) LevelsOfIndirection() int {
	return 0
}

func (v FunctionType) IsIntegerType() bool {
	return false
}

func (v FunctionType) IsFloatingType() bool {
	return false
}

func (v FunctionType) IsVoidType() bool {
	return false
}

func (v FunctionType) CanCastTo(t Type) bool {
	return false
}

func (v FunctionType) Attrs() AttrGroup {
	return v.attrs
}

func (v FunctionType) Equals(t Type) bool {
	other, ok := t.(FunctionType)
	if !ok {
		return false
	}

	if !v.Attrs().Equals(other.Attrs()) {
		return false
	}

	if v.IsVariadic != other.IsVariadic {
		return false
	}

	if !v.Return.Equals(other.Return) {
		return false
	}

	if len(v.Parameters) != len(other.Parameters) {
		return false
	}

	for i, par := range v.Parameters {
		if !par.Equals(other.Parameters[i]) {
			return false
		}
	}

	if (v.Receiver == nil) != (other.Receiver == nil) {
		return false
	}

	if v.Receiver != nil && !v.Receiver.Equals(other.Receiver) {
		return false
	}

	return true
}

func (v FunctionType) ActualType() Type {
	return v
}

// MetaType
type metaType struct {
}

func (v metaType) IsSigned() bool {
	panic("IsSigned() invalid on metaType")
}

func (v metaType) LevelsOfIndirection() int {
	panic("LevelsOfIndirection() invalid on metaType")
}

func (v metaType) IsIntegerType() bool {
	panic("IsIntegerType() invalid on metaType")
}

func (v metaType) IsFloatingType() bool {
	panic("IsFloatingType() invalid on metaType")
}

func (v metaType) IsVoidType() bool {
	panic("IsVoidType() invalid on metaType")
}

func (v metaType) CanCastTo(t Type) bool {
	panic("CanCastTo() invalid on metaType")
}

func (v metaType) Attrs() AttrGroup {
	panic("Attrs() invalid on metaType")
}

func (v metaType) Equals(t Type) bool {
	panic("Equals() invalid on metaType")
}

// Generics helper objects

type GenericSigil []*SubstitutionType

func (v GenericSigil) String() string {
	if len(v) <= 0 {
		return ""
	}

	str := "<"
	for i, par := range v {
		str += par.TypeName()
		if i < len(v)-1 {
			str += ", "
		}
	}
	return str + ">"
}

// SubstitutionType
type SubstitutionType struct {
	attrs       AttrGroup
	Name        string
	Constraints []Type // should be all interface types
}

func NewSubstitutionType(name string, constraints []Type) *SubstitutionType {
	return &SubstitutionType{Name: name, Constraints: constraints}
}

func (v *SubstitutionType) String() string {
	return "(" + util.Blue("SubstitutionType") + ": " + v.TypeName() + " " + fmt.Sprintf("%p", v) + ")"
}

func (v *SubstitutionType) TypeName() string {
	str := v.Name

	if len(v.Constraints) > 0 {
		str += ":"
		for _, c := range v.Constraints {
			str += " " + c.TypeName()
		}
	}

	return str
}

func (v *SubstitutionType) ActualType() Type {
	return v
}

func (v *SubstitutionType) Equals(t Type) bool {
	return v == t
}

func (v *SubstitutionType) Attrs() AttrGroup {
	return v.attrs
}

func (v *SubstitutionType) CanCastTo(t Type) bool {
	return false
}

func (v *SubstitutionType) IsFloatingType() bool {
	return false
}

func (v *SubstitutionType) IsIntegerType() bool {
	return false
}

func (v *SubstitutionType) IsSigned() bool {
	return false
}

func (v *SubstitutionType) IsVoidType() bool {
	return false
}

func (v *SubstitutionType) LevelsOfIndirection() int {
	return 0
}

// GenericInstance
// Substition GenericContext to real type mappings override parameters to self mappings.
type GenericContext struct {
	submap map[*SubstitutionType]*TypeReference
	Outer  *GenericContext
}

func NewGenericContext(parameters GenericSigil, arguments []*TypeReference) *GenericContext {
	v := &GenericContext{
		submap: make(map[*SubstitutionType]*TypeReference),
	}

	if len(parameters) != len(arguments) {
		panic("len(parameters) != len(arguments): " + fmt.Sprintf("%d != %d", len(parameters), len(arguments)))
	}

	for i, par := range parameters {
		v.submap[par] = arguments[i]
	}

	return v
}

func getTypeGenericParameters(typ Type) GenericSigil {
	switch typ := typ.ActualType().(type) {
	case EnumType:
		return typ.GenericParameters
	case FunctionType:
		return typ.GenericParameters
	case StructType:
		return typ.GenericParameters
	case PointerType:
		return getTypeGenericParameters(typ.Addressee.BaseType)
	case ReferenceType:
		return getTypeGenericParameters(typ.Referrer.BaseType)

	case PrimitiveType, *SubstitutionType, ArrayType, TupleType:
		return nil

	case *NamedType:
		panic("oh no")

	default:
		panic("unim base type: " + reflect.TypeOf(typ).String())
	}
}

func NewGenericContextFromTypeReference(typref *TypeReference) *GenericContext {
	return NewGenericContext(getTypeGenericParameters(typref.BaseType), typref.GenericArguments)
}

// Like Get, but only gets value where key is substitution type. Returns nil if no value for key.
func (v *GenericContext) GetSubstitutionType(t *SubstitutionType) *TypeReference {
	if x, ok := v.submap[t]; ok {
		return x
	} else if v.Outer != nil {
		return v.Outer.GetSubstitutionType(t)
	}
	return nil
}

// If the key is a substitution type and is not found in this generic instance, Get() checks GenericContext.Outer if not nil.
func (v *GenericContext) Get(t *TypeReference) *TypeReference {
	if v == nil {
		panic("called Get() on nil GenericContext")
	}

	if sub, ok := t.BaseType.(*SubstitutionType); ok {
		if x, ok := v.submap[sub]; ok {
			return x
		} else if v.Outer != nil {
			return v.Outer.Get(t)
		}
	}
	return t
}

// Replace goes through the *TypeReference and returns a new one with all the SubstitutionTypes that are mapped in this GenericContext.
func (v *GenericContext) Replace(t *TypeReference) *TypeReference {
	if sub, ok := t.BaseType.(*SubstitutionType); ok {
		if subt := v.GetSubstitutionType(sub); subt != nil {
			return subt
		}
		return t
	}

	return &TypeReference{
		BaseType:         v.replaceType(t.BaseType),
		GenericArguments: v.replaceTypeReferences(t.GenericArguments),
	}
}

func (v *GenericContext) replaceTypeReferences(ts []*TypeReference) []*TypeReference {
	ret := make([]*TypeReference, 0, len(ts))
	for _, t := range ts {
		ret = append(ret, v.Replace(t))
	}
	return ret
}

// Function internal to GenericContext.
func (v *GenericContext) replaceType(ty Type) Type {
	switch t := ty.(type) {
	case EnumType:
		for i, mem := range t.Members {
			t.Members[i].Type = v.replaceType(mem.Type) // TODO replace if Members becomes ptr
		}
		return t

	case FunctionType:
		t.Parameters = v.replaceTypeReferences(t.Parameters)
		if t.Receiver != nil {
			t.Receiver = v.replaceType(t.Receiver)
		}
		t.Return = v.Replace(t.Return)
		return t

	case StructType:
		for i, mem := range t.Members {
			t.Members[i] = &StructMember{
				Name: mem.Name,
				Type: v.Replace(mem.Type),
			}
		}
		return t

	case PrimitiveType:
		return t

	case *NamedType:
		return &NamedType{
			Name:         t.Name,
			Methods:      t.Methods,
			ParentModule: t.ParentModule,
			Type:         t.Type,
		}

	case PointerType:
		t.Addressee = v.Replace(t.Addressee)
		return t

	case ReferenceType:
		t.Referrer = v.Replace(t.Referrer)
		return t

	case ArrayType:
		t.MemberType = v.Replace(t.MemberType)
		return t

	case TupleType:
		for i, mem := range t.Members {
			t.Members[i] = v.Replace(mem)
		}
		return t

	default:
		if t == nil {
			panic("t is nil")
		}
		panic("unim: " + reflect.TypeOf(t).String())
	}
}

// UnresolvedType
type UnresolvedType struct {
	metaType
	Name UnresolvedName
}

func (v UnresolvedType) String() string {
	return "(" + util.Blue("UnresolvedType") + ": " + v.Name.String() + ")"
}

func (v UnresolvedType) TypeName() string {
	return "unresolved(" + v.Name.String() + ")"
}

func (v UnresolvedType) Equals(t Type) bool {
	panic("Equals() invalid on unresolved type: name == " + v.Name.String())
}

func (v UnresolvedType) ActualType() Type {
	return v
}

func TypeWithoutPointers(t Type) Type {
	if ptr, ok := t.(PointerType); ok {
		return TypeWithoutPointers(ptr.Addressee.BaseType)
	}

	return t
}
