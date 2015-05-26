package parser

import (
	"strings"
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

func parseEscapeSequence(s []rune) (rune, int) {
	if s[0] != '\\' {
		return -1, 0
	}
	return 0, 0
}

func unescapeString(s string) string {
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
