package log

import (
	"strings"
	"time"

	"github.com/ark-lang/ark/src/util"
)

// magic
var indent int = 0

func Timed(titleColored, titleUncolored string, fn func()) {
	var bold string
	if indent == 0 {
		bold = util.TEXT_BOLD
	}

	if titleUncolored != "" {
		titleUncolored = " " + titleUncolored
	}

	Verbose("main", strings.Repeat(" ", indent))
	Verboseln("main", bold+util.TEXT_GREEN+"Started "+titleColored+util.TEXT_RESET+titleUncolored)
	start := time.Now()

	indent++
	fn()
	indent--

	duration := time.Since(start)
	Verbose("main", strings.Repeat(" ", indent))
	Verboseln("main", bold+util.TEXT_GREEN+"Ended "+titleColored+util.TEXT_RESET+titleUncolored+" (%.2fms)", float32(duration)/1000000)
}
