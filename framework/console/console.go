package console

import (
	"sync"
)

// LogLevel represents a the importance of a particular console message
type LogLevel int

type loggerBufferedEntry struct {
	level LogLevel
	value interface{}
}

type logger struct {
	mut   sync.Mutex
	pipes []func(LogLevel, interface{})

	storeEntries bool
	entries      []loggerBufferedEntry
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

// AddOutputPipe allows for adding additional callbacks for any engine logs
// This will add the provided callback, and will not replace any existing ones.
func AddOutputPipe(cb func(LogLevel, interface{})) {
	singleton.mut.Lock()
	singleton.pipes = append(singleton.pipes, cb)

	// Any buffered entries are immediately output to the new pipe
	if singleton.storeEntries == true {
		for _, entry := range singleton.entries {
			cb(entry.level, entry.value)
		}
	}
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
	singleton.entries = append(singleton.entries, loggerBufferedEntry{
		level: level,
		value: i,
	})
	singleton.mut.Unlock()
}

// DisableBufferedLogs stop storing new log entries. It can be called repeatedly, but cannot be undone once called
// The purpose is such that different output pipes might be attached at runtime, and historical entries might be desirable
func DisableBufferedLogs() {
	singleton.storeEntries = false
	singleton.entries = nil
}

func init() {
	singleton.storeEntries = true
	singleton.entries = make([]loggerBufferedEntry, 0)
}
