package parser

import (
	"github.com/ark-lang/ark-go/lexer"
	"github.com/ark-lang/ark-go/util"
)

type Attr struct {
	Key string
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

func (v *parser) parseAttr() *Attr {
	if v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		attr := &Attr{}
		
		// eat the opening bracket
		v.consumeToken()

		attr.Key = v.consumeToken().Contents
		
		if v.tokenMatches(0, lexer.TOKEN_OPERATOR, "=") {
			v.consumeToken() // eat =
			
			if v.tokenMatches(0, lexer.TOKEN_IDENTIFIER, "") {
				attr.Value = v.consumeToken().Contents
			} else {
				v.err("Expected attribute value, found `%s`", v.peek(0).Contents)
			}
		}
		
		if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
			v.err("Expected `]` at the end of attribute, found `%s`", v.peek(0).Contents)
		}

		// eat the closing bracket
		v.consumeToken()
		return attr
	}
	return nil
}
