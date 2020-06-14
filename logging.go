package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gookit/color"
)

const (
	verbosePrefix = "VERBOSE"
	debugPrefix   = "DEBUG"
	errorPrefix   = "ERROR"
	warnPrefix    = "WARN"
	infoPrefix    = "INFO"
	initPrefix    = "INIT"
	chatPrefix    = "CHAT"

	debugLevel   = 2
	verboseLevel = 2
	infoLevel    = 1
	errorLevel   = 0
	warnLevel    = 0
)

var sprintf = fmt.Sprintf

// Logger - Interface that details an object that can log
type Logger interface {
	Loglevel() int
	SetLoglevel(int)
	UUID() string
}

// LogOutput - Log the given string with a timestamp and no prefix. Logging does
// not depend on the current loglevel of an object
func LogOutput(l Logger, m string) {
	log.Output(1, m)
}

// LogError - Log an error.
func LogError(l Logger, m string) {
	log.Output(1, sprintf("[%s] %s", errorPrefix, m))
}

// LogWarning - Log a warning
func LogWarning(l Logger, m string) {
	log.Output(1, sprintf("[%s] %s", warnPrefix, m))
}

// LogDebug - Log a debug message if the loglevel of the given object is three
// or greater
func LogDebug(l Logger, m string) {
	if l.Loglevel() >= debugLevel {
		log.Output(1, sprintf("[%s] %s", debugPrefix, m))
	}
}

// LogInfo - Log an informational notification if the loglevel is one or greater
func LogInfo(l Logger, m string) {
	if l.Loglevel() >= infoLevel {
		log.Output(1, sprintf("[%s] %s", infoPrefix, m))
	}
}

// LogInit - Log an initialization message
func LogInit(l Logger, m string) {
	log.Output(1, sprintf("[%s] %s", initPrefix, m))
}

// LogVerbose - Log a message only when the loglevel of an object is 2 or greater
func LogVerbose(l Logger, m string) {
	if l.Loglevel() >= verboseLevel {
		log.Output(1, sprintf("[%s] %s", verbosePrefix, m))
	}
}

// LogChat - Log server chat
func logChat(l Logger, m string) {
	if l.Loglevel() >= infoLevel {
		log.Output(1, sprintf("[%s] %s", chatPrefix, m))
	}
}

// LogHTTP - Log an HTTP response code and string. Provides formatting for the
// response, and will output if the loglevel of the object is 1 or greater
func LogHTTP(l Logger, rc int, r *http.Request) {
	if l.Loglevel() >= infoLevel {
		rcs := formatResponseHeader(rc, r.Method)
		rinfo := sprintf("%s - %s %s",
			r.RemoteAddr,
			r.Host,
			r.RequestURI)
		log.Output(1, sprintf("%s %s", rcs, rinfo))
	}
}

// formatResponseCode - Provide string formatting for the given response code
func formatResponseHeader(r int, m string) string {
	white := color.FgWhite.Render
	black := color.FgBlack.Render
	greenbg := color.BgGreen.Render
	redbg := color.BgRed.Render
	bluebg := color.BgBlue.Render
	yellowbg := color.BgYellow.Render

	out := "[" + m + " " + strconv.Itoa(r) + "]"
	switch true {
	case 200 <= r && r <= 299:
		return greenbg(black(out))
	case 300 <= r && r <= 399:
		return bluebg(black(out))
	case 400 <= r && r <= 499:
		return yellowbg(black(out))
	default:
		return white(redbg(out))
	}
}
