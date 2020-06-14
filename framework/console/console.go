package console

import (
	"sync"
)

// LogLevel represents a the importance of a particular console message
type LogLevel int

type logger struct {
	mut   sync.Mutex
	pipes []func(LogLevel, interface{})
}

var singleton logger

const (
	// LevelUnknown is the default log level
	LevelUnknown = LogLevel(0)
	// LevelFatal is used for reporting unrecoverable errors
	LevelFatal = LogLevel(1)
	// LevelError is used for reporting recoverable errors
	LevelError = LogLevel(2)
	// LevelWarning is used for reporting undesirable situations, but are not unexpected (e.g. missing texture)
	LevelWarning = LogLevel(3)
	// LevelInfo is used for generic messages
	LevelInfo = LogLevel(4)
	// LevelSuccess is used for reporting success messages
	LevelSuccess = LogLevel(5)
)

func AddOutputPipe(cb func(LogLevel, interface{})) {
	singleton.mut.Lock()
	singleton.pipes = append(singleton.pipes, cb)
	singleton.mut.Unlock()
}

// PrintString prints pass string to output stream
func PrintString(level LogLevel, text string) {
	PrintInterface(level, text)
}

// PrintInterface will print anything to output stream
func PrintInterface(level LogLevel, i interface{}) {
	singleton.mut.Lock()
	for _, cb := range singleton.pipes {
		cb(level, i)
	}
	singleton.mut.Unlock()
}
