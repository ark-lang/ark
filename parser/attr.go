package parser

import (
	"github.com/ark-lang/ark/lexer"
	"github.com/ark-lang/ark/util"
)

type Attr struct {
	Key   string
	Value string
	nodePos
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

func getAttr(attrs []*Attr, s string) *Attr {
	for _, a := range attrs {
		if s == a.Key {
			return a
		}
	}
	return nil
}

func (v *semanticAnalyzer) checkAttrsDistanceFromLine(attrs []*Attr, line int, declType, declName string) {
	for i := len(attrs) - 1; i >= 0; i-- {
		if attrs[i].lineNumber < line-1 {
			v.warn(attrs[i], "Gap of %d lines between declaration of %s `%s` and `%s` attribute", line-attrs[i].lineNumber, declType, declName, attrs[i].Key)
		}
		line = attrs[i].lineNumber
	}
}

func (v *parser) parseAttrs() []*Attr {
	ret := make([]*Attr, 0)
	for v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		// eat the opening bracket
		v.consumeToken()
	thing:
		attr := &Attr{
			Key: v.consumeToken().Contents,
		}
		attr.setPos(v.peek(0).Filename, v.peek(0).LineNumber, v.peek(0).CharNumber)

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
