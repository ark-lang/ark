package parser

type Type interface {
	GetTypeName() string
	GetRawType() Type            // type disregarding pointers
	GetLevelsOfIndirection() int // number of pointers you have to go through to get to the actual type
}

//go:generate stringer -type=PrimitiveType
type PrimitiveType int

const (
	PRIMITIVE_i8 PrimitiveType = iota
	PRIMITIVE_i16
	PRIMITIVE_i32
	PRIMITIVE_i64
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
)

func (v PrimitiveType) GetTypeName() string {
	return v.String()[10:]
}

func (v PrimitiveType) GetRawType() Type {
	return v
}

func (v PrimitiveType) GetLevelsOfIndirection() int {
	return 0
}

// StructType

type StructType struct {
	Name string
}

func (v *StructType) GetTypeName() string {
	return v.Name
}

func (v *StructType) GetRawType() Type {
	return v
}

func (v *StructType) GetLevelsOfIndirection() int {
	return 0
}

// PointerTyper

type PointerType struct {
	Addressee Type
}

func (v *PointerType) GetTypeName() string {
	return "^" + v.Addressee.GetTypeName()
}

func (v *PointerType) GetRawType() Type {
	return v.Addressee.GetRawType()
}

func (v *PointerType) GetLevelsOfIndirection() int {
	return v.Addressee.GetLevelsOfIndirection() + 1
}
