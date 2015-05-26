package util

import (
	"runtime"
)

var (
	TEXT_RESET   string = ""
	TEXT_BOLD    string = ""
	TEXT_RED     string = ""
	TEXT_GREEN   string = ""
	TEXT_YELLOW  string = ""
	TEXT_BLUE    string = ""
	TEXT_MAGENTA string = ""
	TEXT_CYAN    string = ""
	TEXT_WHITE   string = ""
)

func init() {
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd":
		TEXT_RESET = "\x1B[00m"
		TEXT_BOLD = "\x1B[01m"
		TEXT_RED = "\x1B[31m"
		TEXT_GREEN = "\x1B[32m"
		TEXT_YELLOW = "\x1B[33m"
		TEXT_BLUE = "\x1B[34m"
		TEXT_MAGENTA = "\x1B[35m"
		TEXT_CYAN = "\x1B[36m"
		TEXT_WHITE = "\x1B[37m"
	}
}

func Bold(s string) string {
	return TEXT_BOLD + s + TEXT_RESET
}

func Red(s string) string {
	return TEXT_RED + s + TEXT_RESET
}

func Green(s string) string {
	return TEXT_GREEN + s + TEXT_RESET
}

func Yellow(s string) string {
	return TEXT_YELLOW + s + TEXT_RESET
}

func Blue(s string) string {
	return TEXT_BLUE + s + TEXT_RESET
}

func Magenta(s string) string {
	return TEXT_MAGENTA + s + TEXT_RESET
}

func Cyan(s string) string {
	return TEXT_CYAN + s + TEXT_RESET
}

func White(s string) string {
	return TEXT_WHITE + s + TEXT_RESET
}
