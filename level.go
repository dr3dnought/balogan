package balogan

import "os"

type LogLevel int

const (
	TraceLevel LogLevel = iota - 1
	DebugLevel
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

func (level LogLevel) String() string {
	switch level {
	case TraceLevel:
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarningLevel:
		return "WARNING"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// IsEnabled checks if the given level should be logged based on the minimum level.
func (level LogLevel) IsEnabled(minLevel LogLevel) bool {
	return level >= minLevel
}

// ShouldExit returns true if the log level should cause the program to exit.
func (level LogLevel) ShouldExit() bool {
	return level >= FatalLevel
}

// ShouldPanic returns true if the log level should cause a panic.
func (level LogLevel) ShouldPanic() bool {
	return level == PanicLevel
}

// Exit terminates the program with exit code 1.
// Used internally by Fatal level logging.
func (level LogLevel) Exit() {
	if level.ShouldExit() {
		os.Exit(1)
	}
}

// Panic causes a panic with the given message.
// Used internally by Panic level logging.
func (level LogLevel) Panic(message string) {
	if level.ShouldPanic() {
		panic(message)
	}
}
