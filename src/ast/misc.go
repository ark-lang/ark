package ast

import (
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
)

// escape for debug output
// only things that can't be displayed need to be escaped
func EscapeString(s string) string {
	out := make([]rune, 0)
	sr := []rune(s)

main_loop:
	for _, r := range sr {
		for i, escapeVal := range []rune(parser.SIMPLE_ESCAPE_VALUES) {
			if r == escapeVal {
				out = append(out, '\\', []rune(parser.SIMPLE_ESCAPE_NAMES)[i])
				continue main_loop
			}
		}
		out = append(out, r)
	}

	return string(out)
}

func colorizeEscapedString(input string) string {
	inputRunes := []rune(input)
	outputRunes := make([]rune,
		len(inputRunes))

	outputRunes = append(outputRunes, []rune(util.TEXT_YELLOW)...)
	for i := 0; i < len(inputRunes); i++ {
		if inputRunes[i] == '\\' {
			outputRunes = append(outputRunes, []rune(util.TEXT_RESET+util.TEXT_CYAN)...)
			i++
			outputRunes = append(outputRunes, '\\', inputRunes[i])
			outputRunes = append(outputRunes, []rune(util.TEXT_YELLOW)...)
			continue
		}
		outputRunes = append(outputRunes, inputRunes[i])
	}

	outputRunes = append(outputRunes, []rune(util.TEXT_RESET)...)
	return string(outputRunes)
}
