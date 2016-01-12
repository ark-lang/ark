package parser

const (
	KEYWORD_ALLOC     string = "alloc"
	KEYWORD_AS        string = "as"
	KEYWORD_BREAK     string = "break"
	KEYWORD_C         string = "C"
	KEYWORD_CAST      string = "cast"
	KEYWORD_DEFAULT   string = "default"
	KEYWORD_DEFER     string = "defer"
	KEYWORD_DO        string = "do"
	KEYWORD_ELSE      string = "else"
	KEYWORD_ENUM      string = "enum"
	KEYWORD_EXT       string = "ext"
	KEYWORD_FALSE     string = "false"
	KEYWORD_FOR       string = "for"
	KEYWORD_FREE      string = "free"
	KEYWORD_FUNC      string = "func"
	KEYWORD_LEN       string = "len"
	KEYWORD_IF        string = "if"
	KEYWORD_IMPL      string = "impl"
	KEYWORD_MATCH     string = "match"
	KEYWORD_MODULE    string = "module"
	KEYWORD_MUT       string = "mut"
	KEYWORD_NEXT      string = "next"
	KEYWORD_RETURN    string = "return"
	KEYWORD_SET       string = "set"
	KEYWORD_SIZEOF    string = "sizeof"
	KEYWORD_STRUCT    string = "struct"
	KEYWORD_INTERFACE string = "interface"
	KEYWORD_TRAIT     string = "trait"
	KEYWORD_TRUE      string = "true"
	KEYWORD_USE       string = "use"
	KEYWORD_VOID      string = "void"
)

var keywordList = []string{
	KEYWORD_ALLOC,
	KEYWORD_AS,
	KEYWORD_BREAK,
	KEYWORD_C,
	KEYWORD_CAST,
	KEYWORD_DEFAULT,
	KEYWORD_DEFER,
	KEYWORD_DO,
	KEYWORD_ELSE,
	KEYWORD_ENUM,
	KEYWORD_EXT,
	KEYWORD_FALSE,
	KEYWORD_FOR,
	KEYWORD_FREE,
	KEYWORD_FUNC,
	KEYWORD_LEN,
	KEYWORD_IF,
	KEYWORD_IMPL,
	KEYWORD_MATCH,
	KEYWORD_MODULE,
	KEYWORD_MUT,
	KEYWORD_NEXT,
	KEYWORD_RETURN,
	KEYWORD_SET,
	KEYWORD_SIZEOF,
	KEYWORD_STRUCT,
	KEYWORD_INTERFACE,
	KEYWORD_TRAIT,
	KEYWORD_TRUE,
	KEYWORD_USE,
	KEYWORD_VOID,
}

// Contains a map with all keywords as keys, and true as values
// Uses a map for quick lookup time when checking for reserved vars
var keywordMap map[string]bool

func init() {
	keywordMap = make(map[string]bool)

	for _, key := range keywordList {
		keywordMap[key] = true
	}
}

func isReservedKeyword(s string) bool {
	if m := keywordMap[s]; m {
		return true
	}

	// names starting with a _ followed by an uppercase letter
	// are reserved as they can interfere with name mangling
	if len(s) >= 2 && s[0] == '_' && (s[1] >= 'A' && s[1] <= 'Z') {
		return true
	}

	return false
}
