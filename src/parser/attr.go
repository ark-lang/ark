package parser

import (
	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util"
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

func (v *parser) parseAttrs() AttrGroup {
	ret := make(AttrGroup)
	for v.tokenMatches(0, lexer.TOKEN_SEPARATOR, "[") {
		// eat the opening bracket
		v.consumeToken()
	thing:
		attr := &Attr{
			Key: v.consumeToken().Contents,
		}
		attr.setPos(v.peek(0).Where.Start())

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
