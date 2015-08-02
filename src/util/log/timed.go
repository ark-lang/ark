package log

import (
	"strings"
	"time"

	"github.com/ark-lang/ark/src/util"
)

// magic
var indent int = 0

func Timed(title string, fn func()) {
	Verbose("main", strings.Repeat(" ", indent))
	Verboseln("main", util.TEXT_BOLD+util.TEXT_GREEN+"Started "+title+util.TEXT_RESET)
	start := time.Now()

	indent++
	fn()
	indent--

	duration := time.Since(start)
	Verbose("main", strings.Repeat(" ", indent))
	Verboseln("main", util.TEXT_BOLD+util.TEXT_GREEN+"Ended "+title+util.TEXT_RESET+" (%.2fms)", float32(duration)/1000000)
}
