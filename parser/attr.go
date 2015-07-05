package parser

import (
	"github.com/ark-lang/ark/lexer"
	"github.com/ark-lang/ark/util"
)

/*
TODO replace with map

type AttrGroup struct {
	members map[string]*Attr
}

func Contains(s string) bool {}
func Get(s string) *Attr {}
*/

type AttrGroup map[string]*Attr

func (v AttrGroup) Contains(key string) bool {
	_, ok := v[key]
	return ok
}

func (v AttrGroup) Get(key string) *Attr {
	return v[key]
}

func (v AttrGroup) Set(key string, value *Attr) bool {
	_, ok := v[key]
	v[key] = value
	return ok
}

func (v AttrGroup) Extend(other AttrGroup) {
	for _, attr := range other {
		v[attr.Key] = attr
	}
}

func (v AttrGroup) Equals(other AttrGroup) bool {
	if len(v) != len(other) {
		return false
	}

	for key, value := range v {
		otherValue, ok := other[key]
		if !ok {
			return false
		}

		if value.Value != otherValue.Value {
			return false
		}
	}

	return true
}

type Attr struct {
	Key       string
	Value     string
	FromBlock bool
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

func (v *semanticAnalyzer) checkAttrsDistanceFromLine(attrs AttrGroup, line int, declType, declName string) {
	// Turn map into a list sorted by line number
	var sorted []*Attr
	for _, attr := range attrs {
		index := 0
		for idx, innerAttr := range sorted {
			if attr.lineNumber >= innerAttr.lineNumber {
				index = idx
			}
		}

		sorted = append(sorted, nil)
		copy(sorted[index+1:], sorted[index:])
		sorted[index] = attr
	}

	for i := len(sorted) - 1; i >= 0; i-- {
		if sorted[i].lineNumber < line-1 {
			// mute warnings from attribute blocks
			if !sorted[i].FromBlock {
				v.warn(sorted[i], "Gap of %d lines between declaration of %s `%s` and `%s` attribute", line-sorted[i].lineNumber, declType, declName, sorted[i].Key)
			}
		}
		line = sorted[i].lineNumber
	}
}

func (v *parser) parseAttrs() AttrGroup {
	ret := make(AttrGroup)
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
			if ret.Set(attr.Key, attr) {
				v.err("Duplicate attribute `%s`", attr.Key)
			}
			goto thing // hey, it works
		} else if !v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "]") {
			v.err("Expected `]` at the end of attribute, found `%s`", v.peek(0).Contents)
		}

		// eat the closing bracket
		v.consumeToken()

		if ret.Set(attr.Key, attr) {
			v.err("Duplicate attribute `%s`", attr.Key)
		}
	}
	return ret
}
