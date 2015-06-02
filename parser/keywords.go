package parser

const (
	KEYWORD_ALLOC    string = "alloc"
	KEYWORD_AS       string = "as"
	KEYWORD_BREAK    string = "break"
	KEYWORD_CONTINUE string = "continue"
	KEYWORD_ELSE     string = "else"
	KEYWORD_ENUM     string = "enum"
	KEYWORD_EXT      string = "ext"
	KEYWORD_FOR      string = "for"
	KEYWORD_FREE     string = "free"
	KEYWORD_FUNC     string = "func"
	KEYWORD_IF       string = "if"
	KEYWORD_IMPL     string = "impl"
	KEYWORD_MATCH    string = "match"
	KEYWORD_MUT      string = "mut"
	KEYWORD_RETURN   string = "return"
	KEYWORD_SET      string = "set"
	KEYWORD_SIZEOF   string = "sizeof"
	KEYWORD_STRUCT   string = "struct"
	KEYWORD_USE      string = "use"
	KEYWORD_VOID     string = "void"
)

// Contains a map with all keywords as keys, and true as values
var keywordMap map[string]bool

func init() {
	keywordMap = make(map[string]bool)

	keywordMap[KEYWORD_ALLOC] = true
	keywordMap[KEYWORD_AS] = true
	keywordMap[KEYWORD_BREAK] = true
	keywordMap[KEYWORD_CONTINUE] = true
	keywordMap[KEYWORD_ELSE] = true
	keywordMap[KEYWORD_ENUM] = true
	keywordMap[KEYWORD_EXT] = true
	keywordMap[KEYWORD_FOR] = true
	keywordMap[KEYWORD_FREE] = true
	keywordMap[KEYWORD_FUNC] = true
	keywordMap[KEYWORD_IF] = true
	keywordMap[KEYWORD_IMPL] = true
	keywordMap[KEYWORD_MATCH] = true
	keywordMap[KEYWORD_MUT] = true
	keywordMap[KEYWORD_RETURN] = true
	keywordMap[KEYWORD_SET] = true
	keywordMap[KEYWORD_SIZEOF] = true
	keywordMap[KEYWORD_STRUCT] = true
	keywordMap[KEYWORD_USE] = true
	keywordMap[KEYWORD_VOID] = true
}
