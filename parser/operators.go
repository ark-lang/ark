package parser

type OpCategory int

const (
	OP_ARITHMETIC OpCategory = iota
	OP_COMPARISON
	OP_BITWISE
	OP_LOGICAL
	OP_ASSIGN
)

func (v OpCategory) PrettyString() string {
	switch v {
	case OP_ARITHMETIC:
		return "arithmetic"
	case OP_COMPARISON:
		return "comparison"
	case OP_BITWISE:
		return "bitwise"
	case OP_LOGICAL:
		return "logical"
	case OP_ASSIGN:
		return "assignment"
	default:
		panic("missing opcategory")
	}
}

//go:generate stringer -type=BinOpType
type BinOpType int

const (
	BINOP_ERR BinOpType = iota

	BINOP_ADD
	BINOP_SUB
	BINOP_MUL
	BINOP_DIV
	BINOP_MOD

	BINOP_GREATER
	BINOP_LESS
	BINOP_GREATER_EQ
	BINOP_LESS_EQ
	BINOP_EQ
	BINOP_NOT_EQ

	BINOP_BIT_AND
	BINOP_BIT_OR
	BINOP_BIT_XOR
	BINOP_BIT_LEFT
	BINOP_BIT_RIGHT

	BINOP_LOG_AND
	BINOP_LOG_OR
)

var binOpStrings = []string{"", "+", "-", "*", "/", "%", ">", "<", ">=", "<=",
	"==", "!=", "&", "|", "^", "<<", ">>", "&&", "||"}

func stringToBinOpType(s string) BinOpType {
	for i, str := range binOpStrings {
		if str == s {
			return BinOpType(i)
		}
	}
	return BINOP_ERR
}

func (v BinOpType) OpString() string {
	return binOpStrings[v]
}

func (v BinOpType) Category() OpCategory {
	switch v {
	case BINOP_ADD, BINOP_SUB, BINOP_MUL, BINOP_DIV, BINOP_MOD:
		return OP_ARITHMETIC
	case BINOP_GREATER, BINOP_LESS, BINOP_GREATER_EQ, BINOP_LESS_EQ, BINOP_EQ, BINOP_NOT_EQ:
		return OP_COMPARISON
	case BINOP_BIT_AND, BINOP_BIT_OR, BINOP_BIT_XOR, BINOP_BIT_LEFT, BINOP_BIT_RIGHT:
		return OP_BITWISE
	case BINOP_LOG_AND, BINOP_LOG_OR:
		return OP_LOGICAL
	default:
		panic("missing op category")
	}
}

func newBinOpPrecedenceMap() map[BinOpType]int {
	m := make(map[BinOpType]int)

	// lowest to highest
	precedences := [][]BinOpType{
		{BINOP_LOG_OR},
		{BINOP_LOG_AND},
		{BINOP_BIT_OR},
		{BINOP_BIT_AND},
		{BINOP_EQ, BINOP_NOT_EQ},
		{BINOP_GREATER, BINOP_LESS, BINOP_GREATER_EQ, BINOP_LESS_EQ},
		{BINOP_BIT_LEFT, BINOP_BIT_RIGHT},
		{BINOP_ADD, BINOP_SUB},
		{BINOP_MUL, BINOP_DIV, BINOP_MOD},
	}

	for i, list := range precedences {
		for _, op := range list {
			m[op] = i + 1
		}
	}

	return m
}

//go:generate stringer -type=UnOpType
type UnOpType int

const (
	UNOP_ERR UnOpType = iota

	UNOP_LOG_NOT

	UNOP_BIT_NOT

	UNOP_ADDRESS
	UNOP_DEREF
)

func stringToUnOpType(s string) UnOpType {
	switch s {
	case "!":
		return UNOP_LOG_NOT

	case "~":
		return UNOP_BIT_NOT

	case "&":
		return UNOP_ADDRESS
	case "^":
		return UNOP_DEREF

	default:
		return UNOP_ERR
	}
}
