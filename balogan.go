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

	// Structured logging fields
	fields          Fields
	fieldsFormatter FieldsFormatter
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
		prefixes:        prefixes,
		errorHandler:    &DefaultErrorHandler{},
		fields:          make(Fields),
		fieldsFormatter: DefaultFieldsFormatter,
	}
}

type BaloganConfig struct {
	Level    LogLevel
	Writers  []LogWriter
	Prefixes []PrefixBuilderFunc

	Concurrency bool

	// Structured logging configuration
	Fields          Fields
	FieldsFormatter FieldsFormatter
}

func NewFromConfig(cfg *BaloganConfig) *Logger {
	fields := cfg.Fields
	if fields == nil {
		fields = make(Fields)
	}

	fieldsFormatter := cfg.FieldsFormatter
	if fieldsFormatter == nil {
		fieldsFormatter = DefaultFieldsFormatter
	}

	return &Logger{
		level:           cfg.Level,
		writers:         cfg.Writers,
		prefixes:        cfg.Prefixes,
		errorHandler:    &DefaultErrorHandler{},
		concurrency:     cfg.Concurrency,
		fields:          fields,
		fieldsFormatter: fieldsFormatter,
	}
}

// Creates new Balogan Logger instance ith previous conifg except prefixes.
// Accept additional prefixes which will be added to the end of prefix part of log message.
//
// Provided prefixes DO NOT APPLY for Balogan Logger instance from which the method was called.
func (l *Logger) WithTemporaryPrefix(builder ...PrefixBuilderFunc) *Logger {
	prefixes := append(l.prefixes, builder...)

	return &Logger{
		level:           l.level,
		writers:         l.writers,
		prefixes:        prefixes,
		errorHandler:    l.errorHandler,
		concurrency:     l.concurrency,
		fields:          l.fields.Copy(),
		fieldsFormatter: l.fieldsFormatter,
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
	if !level.IsEnabled(l.level) {
		return
	}

	message := fmt.Sprintf(format, args...)
	fullMessage := l.buildMessage(level, message)
	l.write(fullMessage)
}

// Log logs a message at the specified level.
// It checks if the log level is enabled, then writes the message with provided prefixes.
//
// Parameters:
//
//	level: The log level of the message.
//	args: The arguments to be converted to a string message.
func (l *Logger) Log(level LogLevel, args ...interface{}) {
	if !level.IsEnabled(l.level) {
		return
	}

	message := fmt.Sprint(args...)
	fullMessage := l.buildMessage(level, message)
	l.write(fullMessage)
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

// Trace logs a message at the TRACE level.
// It calls the general Log method with the TRACE level.
//
// Parameters:
//
//	args: The arguments to be logged.
func (l *Logger) Trace(args ...interface{}) {
	l.Log(TraceLevel, args...)
}

// Tracef logs a formatted message at the TRACE level.
// It calls the general Logf method with the TRACE level.
//
// Parameters:
//
//	format: The format string for the message.
//	args: The arguments for the format string.
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.Logf(TraceLevel, format, args...)
}

// Fatal logs a message at the FATAL level and then exits the program.
// It calls the general Log method with the FATAL level, then calls os.Exit(1).
//
// Parameters:
//
//	args: The arguments to be logged.
func (l *Logger) Fatal(args ...interface{}) {
	l.Log(FatalLevel, args...)
	FatalLevel.Exit()
}

// Fatalf logs a formatted message at the FATAL level and then exits the program.
// It calls the general Logf method with the FATAL level, then calls os.Exit(1).
//
// Parameters:
//
//	format: The format string for the message.
//	args: The arguments for the format string.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logf(FatalLevel, format, args...)
	FatalLevel.Exit()
}

// Panic logs a message at the PANIC level and then panics.
// It calls the general Log method with the PANIC level, then calls panic().
//
// Parameters:
//
//	args: The arguments to be logged.
func (l *Logger) Panic(args ...interface{}) {
	message := fmt.Sprint(args...)
	l.Log(PanicLevel, args...)
	PanicLevel.Panic(message)
}

// Panicf logs a formatted message at the PANIC level and then panics.
// It calls the general Logf method with the PANIC level, then calls panic().
//
// Parameters:
//
//	format: The format string for the message.
//	args: The arguments for the format string.
func (l *Logger) Panicf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.Logf(PanicLevel, format, args...)
	PanicLevel.Panic(message)
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

// WithField returns a new Logger instance with the specified field added.
// The new logger inherits all configuration from the current logger.
//
// Parameters:
//
//	key: The field key.
//	value: The field value.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		level:           l.level,
		writers:         l.writers,
		prefixes:        l.prefixes,
		errorHandler:    l.errorHandler,
		concurrency:     l.concurrency,
		fields:          l.fields.With(key, value),
		fieldsFormatter: l.fieldsFormatter,
	}
}

// WithFields returns a new Logger instance with the specified fields added.
// The new logger inherits all configuration from the current logger.
//
// Parameters:
//
//	fields: A map of field key-value pairs to add.
func (l *Logger) WithFields(fields Fields) *Logger {
	return &Logger{
		level:           l.level,
		writers:         l.writers,
		prefixes:        l.prefixes,
		errorHandler:    l.errorHandler,
		concurrency:     l.concurrency,
		fields:          l.fields.WithFields(fields),
		fieldsFormatter: l.fieldsFormatter,
	}
}

// WithFieldsFormatter returns a new Logger instance with the specified fields formatter.
// This allows changing how fields are formatted in log output.
//
// Parameters:
//
//	formatter: The FieldsFormatter to use for formatting fields.
func (l *Logger) WithFieldsFormatter(formatter FieldsFormatter) *Logger {
	return &Logger{
		level:           l.level,
		writers:         l.writers,
		prefixes:        l.prefixes,
		errorHandler:    l.errorHandler,
		concurrency:     l.concurrency,
		fields:          l.fields.Copy(),
		fieldsFormatter: formatter,
	}
}

// WithJSON returns a new Logger instance configured to format fields as JSON.
// This is a convenient shortcut for WithFieldsFormatter(&JSONFormatter{}).
//
// Example:
//
//	logger.WithJSON().WithField("user", "john").Info("User logged in")
//	// Output: INFO {"user":"john"} User logged in
func (l *Logger) WithJSON() *Logger {
	return l.WithFieldsFormatter(&JSONFormatter{})
}

// WithLogfmt returns a new Logger instance configured to format fields in logfmt style.
// This is a convenient shortcut for WithFieldsFormatter(&LogfmtFormatter{}).
//
// Example:
//
//	logger.WithLogfmt().WithField("user", "john doe").Info("User logged in")
//	// Output: INFO user="john doe" User logged in
func (l *Logger) WithLogfmt() *Logger {
	return l.WithFieldsFormatter(&LogfmtFormatter{})
}

// WithKeyValue returns a new Logger instance configured to format fields as key=value pairs.
// This is a convenient shortcut for WithFieldsFormatter(&KeyValueFormatter{}).
// This is the default format, so this method is mainly useful for explicitly switching back
// from another format.
//
// Example:
//
//	logger.WithKeyValue().WithField("user", "john").Info("User logged in")
//	// Output: INFO user=john User logged in
func (l *Logger) WithKeyValue() *Logger {
	return l.WithFieldsFormatter(&KeyValueFormatter{})
}

// WithKeyValueSeparator returns a new Logger instance configured to format fields
// as key=value pairs with a custom separator between pairs.
//
// Parameters:
//
//	separator: The separator to use between key=value pairs.
//
// Example:
//
//	logger.WithKeyValueSeparator(" | ").WithFields(Fields{"a": 1, "b": 2}).Info("Test")
//	// Output: INFO a=1 | b=2 Test
func (l *Logger) WithKeyValueSeparator(separator string) *Logger {
	return l.WithFieldsFormatter(&KeyValueFormatter{Separator: separator})
}

// GetFields returns a copy of the current fields.
func (l *Logger) GetFields() Fields {
	return l.fields.Copy()
}

func (l *Logger) buildPrefixStr(args ...any) string {
	return strings.Join(gospadi.Map(l.prefixes, func(f PrefixBuilderFunc) string {
		return f(args)
	}), " ")
}

func (l *Logger) buildMessage(level LogLevel, message string) string {
	parts := []string{level.String()}

	prefixStr := l.buildPrefixStr()
	if prefixStr != "" {
		parts = append(parts, prefixStr)
	}

	if len(l.fields) > 0 {
		fieldsStr := l.fieldsFormatter.Format(l.fields)
		if fieldsStr != "" {
			parts = append(parts, fieldsStr)
		}
	}

	parts = append(parts, message)

	return strings.Join(parts, " ")
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
					errsMu.Lock()
					errs = append(errs, err)
					errsMu.Unlock()
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
