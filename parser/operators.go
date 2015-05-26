package parser

//go:generate stringer -type=BinOpType
type BinOpType int

const (
	BINOP_ERR BinOpType = iota
	
	BINOP_ADD
	BINOP_SUB
	BINOP_MUL
	BINOP_DIV
	BINOP_MOD
	
	BINOP_DOT
	
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
	
	BINOP_ASSIGN
)

func newBinOpPrecedenceMap() map[BinOpType]int {
	m := make(map[BinOpType]int)
	
	// lowest to highest
	precedences := [][]BinOpType {
		{ BINOP_ASSIGN },
		{ BINOP_LOG_OR },
		{ BINOP_LOG_AND },
		{ BINOP_BIT_OR },
		{ BINOP_BIT_AND },
		{ BINOP_EQ, BINOP_NOT_EQ },
		{ BINOP_GREATER, BINOP_LESS, BINOP_GREATER_EQ, BINOP_LESS_EQ },
		{ BINOP_ADD, BINOP_SUB },
		{ BINOP_MUL, BINOP_DIV, BINOP_MOD },
		{ BINOP_DOT },
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

func stringToBinOpType(s string) BinOpType {
	switch s {
	case "+": return BINOP_ADD
	case "-": return BINOP_SUB
	case "*": return BINOP_MUL
	case "/": return BINOP_DIV
	case "%": return BINOP_MOD
	
	case ".": return BINOP_DOT
	
	case ">": return BINOP_GREATER
	case "<": return BINOP_LESS
	case ">=": return BINOP_GREATER_EQ
	case "<=": return BINOP_LESS_EQ
	case "==": return BINOP_EQ
	case "!=": return BINOP_NOT_EQ
	
	case "&": return BINOP_BIT_AND
	case "|": return BINOP_BIT_OR
	case "^": return BINOP_BIT_XOR
	case "<<": return BINOP_BIT_LEFT
	case ">>": return BINOP_BIT_RIGHT
	
	case "&&": return BINOP_LOG_AND
	case "||": return BINOP_LOG_OR
	
	case "=": return BINOP_ASSIGN
	
	default: return BINOP_ERR
	}
}

func stringToUnOpType(s string) UnOpType {
	switch s {
	case "!": return UNOP_LOG_NOT
	
	case "~": return UNOP_BIT_NOT
	
	case "&": return UNOP_ADDRESS
	case "^": return UNOP_DEREF
	
	default: return UNOP_ERR
	}
}
