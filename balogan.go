package balogan

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/dr3dnought/gospadi"
)

var DefaultWriter = NewStdOutLogWriter()

type ErrorHandler interface {
	Handle(err error)
}

type DefaultErrorHandler struct{}

func (h *DefaultErrorHandler) Handle(error) {}

type Logger struct {
	mutex sync.Mutex

	level    LogLevel
	writers  []LogWriter
	prefixes []PrefixBuilderFunc

	errorHandler ErrorHandler

	concurrency bool
}

// The simpliest way to create new Balogan Logger instance.
//
// If writers accept nil, the StdOutLogWriter will
// be used as a default value.
//
// Balogan does not provide prefixes which will be used
// as a default value.
func New(level LogLevel, writer LogWriter, prefixes ...PrefixBuilderFunc) *Logger {
	return &Logger{
		level: level,
		writers: (func() []LogWriter {
			if writer == nil {
				return []LogWriter{NewStdOutLogWriter()}
			}

			return []LogWriter{writer}
		}()),
		prefixes:     prefixes,
		errorHandler: &DefaultErrorHandler{},
	}
}

type BaloganConfig struct {
	Level    LogLevel
	Writers  []LogWriter
	Prefixes []PrefixBuilderFunc

	Concurrency bool
}

func NewFromConfig(cfg *BaloganConfig) *Logger {
	return &Logger{
		level:    cfg.Level,
		writers:  cfg.Writers,
		prefixes: cfg.Prefixes,

		concurrency: cfg.Concurrency,
	}
}

// Creates new Balogan Logger instance ith previous conifg except prefixes.
// Accept additional prefixes which will be added to the end of prefix part of log message.
//
// Provided prefixes DO NOT APPLY for Balogan Logger instance from which the method was called.
func (l *Logger) WithTemporaryPrefix(builder ...PrefixBuilderFunc) *Logger {
	prefixes := append(l.prefixes, builder...)

	return &Logger{
		level:    l.level,
		writers:  l.writers,
		prefixes: prefixes,
	}
}

// Logf logs a formatted message at the specified level.
// It checks if the log level is enabled, then writes the message with provided prefixes.
//
// Parameters:
//
//	level: The log level of the message.
//	format: The format string for the message.
//	args: The arguments for the format string.
func (l *Logger) Logf(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)

	l.write(fmt.Sprintf("%s %s %s", level, l.buildPrefixStr(), message))
}

// Log logs a message at the specified level.
// It checks if the log level is enabled, then writes the message with provided prefixes.
//
// Parameters:
//
//	level: The log level of the message.
//	args: The arguments to be converted to a string message.
func (l *Logger) Log(level LogLevel, args ...interface{}) {
	if level < l.level {
		return
	}

	l.write(fmt.Sprintf("%s %s %s", level, l.buildPrefixStr(), fmt.Sprint(args...)))
}

// Debug logs a message at the DEBUG level.
// It calls the general Log method with the DEBUG level.
//
// Parameters:
//
//	args: The arguments to be logged.
func (l *Logger) Debug(args ...interface{}) {
	l.Log(DebugLevel, args...)
}

// Debugf logs a formatted message at the DEBUG level.
// It calls the general Logf method with the DEBUG level.
//
// Parameters:
//
//	format: The format string for the message.
//	args: The arguments for the format string.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logf(DebugLevel, format, args...)
}

// Info logs a message at the INFO level.
// It calls the general Log method with the INFO level.
// Parameters:
//
//	args: The arguments to be logged.
func (l *Logger) Info(args ...interface{}) {
	l.Log(InfoLevel, args...)
}

// Infof logs a formatted message at the INFO level.
// It calls the general Logf method with the INFO level.
//
// Parameters:
//
//	format: The format string for the message.
//	args: The arguments for the format string.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logf(InfoLevel, format, args...)
}

// Warning logs a message at the WARNING level.
// It calls the general Log method with the WARNING level.
//
// Parameters:
//
//	args: The arguments to be logged.
func (l *Logger) Warning(args ...interface{}) {
	l.Log(WarningLevel, args...)
}

// Warningf logs a formatted message at the WARNING level.
// It calls the general Logf method with the WARNING level.
//
// Parameters:
//
//	format: The format string for the message.
//	args: The arguments for the format string.
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.Logf(WarningLevel, format, args...)
}

// Error logs a message at the ERROR level.
// It calls the general Log method with the ERROR level.
//
// Parameters:
//
//	args: The arguments to be logged.
func (l *Logger) Error(args ...interface{}) {
	l.Log(ErrorLevel, args...)
}

// Errorf logs a formatted message at the ERROR level.
// It calls the general Logf method with the ERROR level.
//
// Parameters:
//
//	format: The format string for the message.
//	args: The arguments for the format string.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logf(ErrorLevel, format, args...)
}

// Close closes all writers associated with the logger.
// It ensures that all log messages are flushed and resources are released.
//
// If you use some Writer except StdOutLogWriter we highly recommend to call this method.
//
// Returns:
//
//	An error if any of the writers fail to close.
func (l *Logger) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var errs []error
	if l.concurrency {
		var wg sync.WaitGroup
		for _, writer := range l.writers {
			wg.Add(1)
			go func(w LogWriter) {
				defer wg.Done()
				if err := w.Close(); err != nil {
					errs = append(errs, err)
				}
			}(writer)
		}

		wg.Wait()
	} else {
		for _, writer := range l.writers {
			if closeErr := writer.Close(); closeErr != nil {
				errs = append(errs, closeErr)
			}
		}
	}

	return errors.Join(errs...)
}

func (l *Logger) buildPrefixStr(args ...any) string {
	return strings.Join(gospadi.Map(l.prefixes, func(f PrefixBuilderFunc) string {
		return f(args)
	}), " ")
}

func (l *Logger) write(message string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.concurrency {
		var wg sync.WaitGroup
		var errsMu sync.Mutex
		var errs []error
		for _, writer := range l.writers {
			wg.Add(1)
			go func(w LogWriter) {
				defer wg.Done()
				if _, err := w.Write([]byte(message)); err != nil {
					errsMu.Unlock()
					errs = append(errs, err)
					errsMu.Lock()
				}
			}(writer)
		}

		wg.Wait()
		for _, err := range errs {
			l.errorHandler.Handle(err)
		}
	} else {
		var errs []error
		for _, writer := range l.writers {
			_, err := writer.Write([]byte(message))
			if err != nil {
				errs = append(errs, err)
			}
		}
		for _, err := range errs {
			l.errorHandler.Handle(err)
		}
	}
}
