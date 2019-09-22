package console

import (
	"log"
)

type LogLevel int

const (
	LevelUnknown = LogLevel(0)
	LevelFatal   = LogLevel(1)
	LevelError   = LogLevel(2)
	LevelWarning = LogLevel(3)
	LevelInfo    = LogLevel(4)
	LevelSuccess = LogLevel(5)
)

func PrintString(level LogLevel, text string) {
	PrintInterface(level, text)
}

func PrintInterface(level LogLevel, i interface{}) {
	log.Println(i)
}
