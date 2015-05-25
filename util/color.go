package util

import (
	"runtime"
)

var (
	TEXT_BOLD  string = ""
	TEXT_RESET string = ""
	TEXT_RED   string = ""
	TEXT_GREEN string = ""
)

func init() {
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd":
		TEXT_BOLD = "\x1B[01m"
		TEXT_RESET = "\x1B[00m"
		TEXT_RED = "\x1B[31m"
		TEXT_GREEN = "\x1B[32m"
	}
}
