package parser

func hexRuneToInt(r rune) int {
	if r >= '0' && r <= '9' {
		return int(r - '0')
	} else if r >= 'A' && r <= 'F' {
		return int(r - 'A') + 10
	} else if r >= 'a' && r <= 'f' {
		return int(r - 'a') + 10
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
