package parser

import (
	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util"
)

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

func (v AttrGroup) String() string {
	var strs []string
	for _, attr := range v {
		strs = append(strs, attr.String())
	}

	var res string
	for i, str := range strs {
		res += str
		if i < len(strs)-1 {
			res += " "
		}
	}
	return res
}

type Attr struct {
	Key       string
	Value     string
	FromBlock bool
	pos       lexer.Position
}

func (v *Attr) String() string {
	result := "[" + v.Key
	if v.Value == "" {
		result += "]"
	} else {
		result += "=\"" + v.Value + "\"]"
	}
	return util.Green(result)
}

func (v *Attr) Pos() lexer.Position {
	return v.pos
}

func (v *Attr) SetPos(pos lexer.Position) {
	v.pos = pos
}
