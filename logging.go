package main

import (
	"fmt"
	"log"
)

const (
	verbosePrefix = "VERBOSE"
	debugPrefix   = "DEBUG"
	errorPrefix   = "ERROR"
	warnPrefix    = "WARN"
	infoPrefix    = "INFO"
	initPrefix    = "INIT"

	debugLevel   = 2
	verboseLevel = 2
	infoLevel    = 1
	errorLevel   = 0
	warnLevel    = 0
)

var sprintf = fmt.Sprintf

// Logger -
type Logger interface {
	Loglevel() int
	SetLoglevel(int)
	UUID() string
}

// LogOutput -
func LogOutput(l Logger, m string) {
	log.Output(1, m)
}

// LogError -
func LogError(l Logger, m string) {
	log.Output(1, sprintf("[%s] %s", errorPrefix, m))
}

// LogWarning -
func LogWarning(l Logger, m string) {
	log.Output(1, sprintf("[%s] %s", warnPrefix, m))
}

// LogDebug -
func LogDebug(l Logger, m string) {
	if l.Loglevel() >= debugLevel {
		log.Output(1, sprintf("[%s] %s", debugPrefix, m))
	}
}

// LogInfo -
func LogInfo(l Logger, m string) {
	if l.Loglevel() >= infoLevel {
		log.Output(1, sprintf("[%s] %s", infoPrefix, m))
	}
}

// LogInit -
func LogInit(l Logger, m string) {
	log.Output(1, sprintf("[%s] %s", initPrefix, m))
}

// LogVerbose -
func LogVerbose(l Logger, m string) {
	if l.Loglevel() >= verboseLevel {
		log.Output(1, sprintf("[%s] %s", verbosePrefix, m))
	}
}
