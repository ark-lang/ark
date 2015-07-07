package lexer

//go:generate stringer -type=TokenType
type TokenType int

const (
	TOKEN_RUNE TokenType = iota
	TOKEN_IDENTIFIER
	TOKEN_SEPARATOR
	TOKEN_OPERATOR
	TOKEN_NUMBER
	TOKEN_ERRONEOUS
	TOKEN_STRING
	TOKEN_UNKNOWNS
	TOKEN_DOCCOMMENT
	TOKEN_END_OF_FILE
)

type Token struct {
	Type                         TokenType
	Contents                     string
	CharNumber, LineNumber       int
	EndCharNumber, EndLineNumber int
	Filename                     string
}
