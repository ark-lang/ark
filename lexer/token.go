package lexer

//go:generate stringer -type=TokenType
type TokenType int

const (
	TOKEN_CHARACTER TokenType = iota
	TOKEN_IDENTIFIER
	TOKEN_SEPARATOR
	TOKEN_OPERATOR
	TOKEN_NUMBER
	TOKEN_ERRONEOUS
	TOKEN_STRING
	TOKEN_UNKNOWN
	TOKEN_END_OF_FILE
)

type Token struct {
	Type                   TokenType
	Contents               string
	CharNumber, LineNumber int
	Filename               string
}
