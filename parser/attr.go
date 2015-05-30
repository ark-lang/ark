package parser

import (
	"github.com/ark-lang/ark-go/lexer"
	"github.com/ark-lang/ark-go/util"
)

type Attr struct {
	Key   string
	Value string
}

func (v *Attr) String() string {
	result := "[" + v.Key
	if v.Value == "" {
		result += "]"
	} else {
		result += "=" + v.Value + "]"
	}
	return util.Green(result)
}

func (v *parser) parseAttrs() []*Attr {
	ret := make([]*Attr, 0)
	for v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		// eat the opening bracket
		v.consumeToken()
	thing:
		attr := &Attr{}

		attr.Key = v.consumeToken().Contents

		if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
			v.consumeToken() // eat =

			if v.tokenMatches(0, lexer.TOKEN_STRING, "") {
				attr.Value = v.consumeToken().Contents
				attr.Value = attr.Value[1 : len(attr.Value)-1]
			} else {
				v.err("Expected attribute value, found `%s`", v.peek(0).Contents)
			}
		}

		if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, ",") {
			v.consumeToken()
			ret = append(ret, attr)
			goto thing // hey, it works
		} else if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
			v.err("Expected `]` at the end of attribute, found `%s`", v.peek(0).Contents)
		}

		// eat the closing bracket
		v.consumeToken()

		ret = append(ret, attr)
	}
	return ret
}
