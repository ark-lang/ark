package log

import (
	"fmt"
	"github.com/ark-lang/ark/util"
	"os"
	"strings"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelVerbose
	LevelInfo
	LevelWarning
	LevelError
)

var LevelMap = map[string]LogLevel{
	"debug":   LevelDebug,
	"verbose": LevelVerbose,
	"info":    LevelInfo,
	"warning": LevelWarning,
	"error":   LevelError,
}

var currentLevel LogLevel
var enabledTags map[string]bool
var enableAll bool

func init() {
	currentLevel = LevelInfo
	enabledTags = make(map[string]bool)
	enableAll = false
}

func SetLevel(level string) {
	lvl, ok := LevelMap[level]
	if !ok {
		fmt.Printf("Invalid log level")
		os.Exit(util.EXIT_FAILURE_SETUP)
	}

	currentLevel = lvl
}

func SetTags(tags string) {
	enabledTags = make(map[string]bool)
	enableAll = false

	for _, tag := range strings.Split(tags, ",") {
		if tag == "all" {
			enableAll = true
		} else {
			enabledTags[tag] = true
		}
	}
}

func AtLevel(level LogLevel) bool {
	return level >= currentLevel
}

func Log(level LogLevel, tag string, msg string, args ...interface{}) {
	if !enableAll {
		if !enabledTags[tag] {
			return
		}
	}

	if AtLevel(level) {
		fmt.Printf("["+tag+"] "+msg, args...)
	}
}

func Logln(level LogLevel, tag string, msg string, args ...interface{}) {
	if !enableAll {
		if !enabledTags[tag] {
			return
		}
	}

	if AtLevel(level) {
		fmt.Printf("["+tag+"] "+msg+"\n", args...)
	}
}

func Debug(tag string, msg string, args ...interface{}) {
	Log(LevelDebug, tag, msg, args...)
}

func Debugln(tag string, msg string, args ...interface{}) {
	Logln(LevelDebug, tag, msg, args...)
}

func Verbose(tag string, msg string, args ...interface{}) {
	Log(LevelVerbose, tag, msg, args...)
}

func Verboseln(tag string, msg string, args ...interface{}) {
	Logln(LevelVerbose, tag, msg, args...)
}

func Info(tag string, msg string, args ...interface{}) {
	Log(LevelInfo, tag, msg, args...)
}

func Infoln(tag string, msg string, args ...interface{}) {
	Logln(LevelInfo, tag, msg, args...)
}

func Warning(tag string, msg string, args ...interface{}) {
	Log(LevelWarning, tag, msg, args...)
}

func Warningln(tag string, msg string, args ...interface{}) {
	Logln(LevelWarning, tag, msg, args...)
}

func Error(tag string, msg string, args ...interface{}) {
	Log(LevelError, tag, msg, args...)
}

func Errorln(tag string, msg string, args ...interface{}) {
	Logln(LevelError, tag, msg, args...)
}
