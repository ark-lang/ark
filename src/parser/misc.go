package parser

import (
	"strings"

	"github.com/ark-lang/ark/src/util"
)

func hexRuneToInt(r rune) int {
	if r >= '0' && r <= '9' {
		return int(r - '0')
	} else if r >= 'A' && r <= 'F' {
		return int(r-'A') + 10
	} else if r >= 'a' && r <= 'f' {
		return int(r-'a') + 10
	} else {
		return -1
	}
}

func octRuneToInt(r rune) int {
	if r >= '0' && r <= '7' {
		return int(r - '0')
	} else {
		return -1
	}
}

func binRuneToInt(r rune) int {
	if r == '0' {
		return 0
	} else if r == '1' {
		return 1
	} else {
		return -1
	}
}

const (
	SIMPLE_ESCAPE_VALUES string = "\a\b\f\n\r\t\v\\'\""
	SIMPLE_ESCAPE_NAMES  string = "abfnrtv\\'\""
)

func UnescapeString(s string) string {
	out := make([]rune, 0)
	sr := []rune(s)

	for i := 0; i < len(sr); i++ {
		if sr[i] == '\\' {
			i++
			out = append(out, []rune(SIMPLE_ESCAPE_VALUES)[strings.IndexRune(SIMPLE_ESCAPE_NAMES, sr[i])])
		} else {
			out = append(out, sr[i])
		}
	}

	return string(out)
}

// escape for debug output
// only things that can't be displayed need to be escaped
func EscapeString(s string) string {
	out := make([]rune, 0)
	sr := []rune(s)

main_loop:
	for _, r := range sr {
		for i, escapeVal := range []rune(SIMPLE_ESCAPE_VALUES) {
			if r == escapeVal {
				out = append(out, '\\', []rune(SIMPLE_ESCAPE_NAMES)[i])
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
