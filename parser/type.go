package parser

type Type interface {
	GetTypeName() string
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

type StructType struct {
	Name string
}

func (v *StructType) GetTypeName() string {
	return v.Name
}
