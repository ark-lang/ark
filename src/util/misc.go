package util

import "unicode"

func CapitalizeFirst(s string) string {
	runes := []rune(s)

	runes[0] = unicode.ToTitle(runes[0])

	return string(runes)
}
