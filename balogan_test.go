package balogan

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
)

type MockWriter struct {
	bytes.Buffer
	closed bool
}

func (w *MockWriter) Close() error {
	w.closed = true
	return nil
}

func (w *MockWriter) IsClosed() bool {
	return w.closed
}

func TestNew(t *testing.T) {
	logger := New(InfoLevel, nil)
	if logger == nil {
		t.Fatal("New() returned nil")
	}
	if len(logger.writers) != 1 {
		t.Errorf("Expected 1 writer, got %d", len(logger.writers))
	}

	mockWriter := &MockWriter{}
	logger = New(DebugLevel, mockWriter)
	if logger == nil {
		t.Fatal("New() returned nil")
	}
	if len(logger.writers) != 1 {
		t.Errorf("Expected 1 writer, got %d", len(logger.writers))
	}
	if logger.writers[0] != mockWriter {
		t.Error("Writer not set correctly")
	}

	logger = New(ErrorLevel, mockWriter, WithLogLevel(InfoLevel), WithTag("test"))
	if len(logger.prefixes) != 2 {
		t.Errorf("Expected 2 prefixes, got %d", len(logger.prefixes))
	}
}

func TestNewFromConfig(t *testing.T) {
	mockWriter := &MockWriter{}

	config := &BaloganConfig{
		Level:           DebugLevel,
		Writers:         []LogWriter{mockWriter},
		Prefixes:        []PrefixBuilderFunc{WithTag("config-test")},
		Concurrency:     true,
		Fields:          Fields{"test": "value"},
		FieldsFormatter: &JSONFormatter{},
	}

	logger := NewFromConfig(config)
	if logger == nil {
		t.Fatal("NewFromConfig() returned nil")
	}
	if logger.level != DebugLevel {
		t.Errorf("Expected level %v, got %v", DebugLevel, logger.level)
	}
	if len(logger.writers) != 1 {
		t.Errorf("Expected 1 writer, got %d", len(logger.writers))
	}
	if !logger.concurrency {
		t.Error("Concurrency not set")
	}
	if len(logger.fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(logger.fields))
	}

	logger = NewFromConfig(nil)
	if logger == nil {
		t.Fatal("NewFromConfig(nil) returned nil")
	}
}

func TestLogger_LogLevels(t *testing.T) {
	mockWriter := &MockWriter{}
	logger := New(TraceLevel, mockWriter)

	tests := []struct {
		name     string
		level    LogLevel
		method   func()
		expected string
	}{
		{
			name:  "Trace",
			level: TraceLevel,
			method: func() {
				logger.Trace("trace message")
			},
			expected: "TRACE trace message",
		},
		{
			name:  "Debug",
			level: DebugLevel,
			method: func() {
				logger.Debug("debug message")
			},
			expected: "DEBUG debug message",
		},
		{
			name:  "Info",
			level: InfoLevel,
			method: func() {
				logger.Info("info message")
			},
			expected: "INFO info message",
		},
		{
			name:  "Warning",
			level: WarningLevel,
			method: func() {
				logger.Warning("warning message")
			},
			expected: "WARNING warning message",
		},
		{
			name:  "Error",
			level: ErrorLevel,
			method: func() {
				logger.Error("error message")
			},
			expected: "ERROR error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter.Reset()
			tt.method()

			output := strings.TrimSpace(mockWriter.String())
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestLogger_LogLevelsFormatted(t *testing.T) {
	mockWriter := &MockWriter{}
	logger := New(TraceLevel, mockWriter)

	tests := []struct {
		name     string
		method   func()
		expected string
	}{
		{
			name: "Tracef",
			method: func() {
				logger.Tracef("trace %s", "formatted")
			},
			expected: "TRACE trace formatted",
		},
		{
			name: "Debugf",
			method: func() {
				logger.Debugf("debug %d", 42)
			},
			expected: "DEBUG debug 42",
		},
		{
			name: "Infof",
			method: func() {
				logger.Infof("info %s %d", "test", 123)
			},
			expected: "INFO info test 123",
		},
		{
			name: "Warningf",
			method: func() {
				logger.Warningf("warning %v", true)
			},
			expected: "WARNING warning true",
		},
		{
			name: "Errorf",
			method: func() {
				logger.Errorf("error %f", 3.14)
			},
			expected: "ERROR error 3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter.Reset()
			tt.method()

			output := strings.TrimSpace(mockWriter.String())
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	mockWriter := &MockWriter{}
	logger := New(WarningLevel, mockWriter)

	logger.Trace("trace message")
	logger.Debug("debug message")
	logger.Info("info message")

	if mockWriter.Len() > 0 {
		t.Errorf("Expected no output for disabled levels, got: %q", mockWriter.String())
	}

	logger.Warning("warning message")
	logger.Error("error message")

	output := mockWriter.String()
	if !strings.Contains(output, "warning message") {
		t.Error("Warning message not logged")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message not logged")
	}
}

func TestLogger_WithTemporaryPrefix(t *testing.T) {
	mockWriter := &MockWriter{}
	logger := New(InfoLevel, mockWriter, WithTag("base"))

	tempLogger := logger.WithTemporaryPrefix(WithTimeStamp(), WithLogLevel(ErrorLevel))

	tempLogger.Info("temp message")

	output := strings.TrimSpace(mockWriter.String())

	if !strings.Contains(output, "base") {
		t.Error("Base prefix not found")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Temporary level prefix not found")
	}
	if !strings.Contains(output, "temp message") {
		t.Error("Message not found")
	}

	mockWriter.Reset()
	logger.Info("original message")
	output = strings.TrimSpace(mockWriter.String())
	if !strings.Contains(output, "base") {
		t.Error("Original logger lost base prefix")
	}
	if strings.Contains(output, "ERROR") {
		t.Error("Original logger should not have temporary prefix")
	}
}

func TestLogger_StructuredLogging(t *testing.T) {
	mockWriter := &MockWriter{}
	logger := New(InfoLevel, mockWriter)

	logger.WithField("user", "john").Info("user logged in")
	output := strings.TrimSpace(mockWriter.String())
	if !strings.Contains(output, "user=john") {
		t.Errorf("Field not found in output: %q", output)
	}

	mockWriter.Reset()
	logger.WithFields(Fields{"id": 123, "role": "admin"}).Info("user action")
	output = strings.TrimSpace(mockWriter.String())
	if !strings.Contains(output, "id=123") || !strings.Contains(output, "role=admin") {
		t.Errorf("Fields not found in output: %q", output)
	}

	mockWriter.Reset()
	logger.WithJSON().WithField("data", "value").Info("json test")
	output = strings.TrimSpace(mockWriter.String())
	if !strings.Contains(output, `"data":"value"`) {
		t.Errorf("JSON format not found in output: %q", output)
	}

	mockWriter.Reset()
	logger.WithLogfmt().WithField("key", "value with spaces").Info("logfmt test")
	output = strings.TrimSpace(mockWriter.String())
	if !strings.Contains(output, `key="value with spaces"`) {
		t.Errorf("Logfmt format not found in output: %q", output)
	}
}

func TestLogger_Close(t *testing.T) {
	mockWriter1 := &MockWriter{}
	mockWriter2 := &MockWriter{}

	logger := New(InfoLevel, mockWriter1)
	logger.writers = append(logger.writers, mockWriter2)

	err := logger.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	if !mockWriter1.IsClosed() {
		t.Error("First writer not closed")
	}
	if !mockWriter2.IsClosed() {
		t.Error("Second writer not closed")
	}
}

func TestLogger_CloseConcurrent(t *testing.T) {
	mockWriter1 := &MockWriter{}
	mockWriter2 := &MockWriter{}

	config := &BaloganConfig{
		Level:       InfoLevel,
		Writers:     []LogWriter{mockWriter1, mockWriter2},
		Concurrency: true,
	}
	logger := NewFromConfig(config)

	err := logger.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	if !mockWriter1.IsClosed() {
		t.Error("First writer not closed")
	}
	if !mockWriter2.IsClosed() {
		t.Error("Second writer not closed")
	}
}

func TestLogger_ConcurrentWrites(t *testing.T) {
	mockWriter := &MockWriter{}
	config := &BaloganConfig{
		Level:       InfoLevel,
		Writers:     []LogWriter{mockWriter},
		Concurrency: true,
	}
	logger := NewFromConfig(config)

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Info("concurrent message", id)
		}(i)
	}
	wg.Wait()

	output := mockWriter.String()
	expectedCount := 10
	actualCount := strings.Count(output, "concurrent message")
	if actualCount != expectedCount {
		t.Errorf("Expected %d messages, got %d", expectedCount, actualCount)
	}
}

func TestLogger_GetFields(t *testing.T) {
	logger := New(InfoLevel, nil)
	logger.fields = Fields{"existing": "value"}

	fields := logger.GetFields()

	fields["new"] = "should not affect original"

	if _, exists := logger.fields["new"]; exists {
		t.Error("Modifying returned fields should not affect original")
	}
}

func TestLogger_ErrorHandler(t *testing.T) {
	handledErrors := make([]error, 0)
	errorHandler := &MockErrorHandler{
		HandleFunc: func(err error) {
			handledErrors = append(handledErrors, err)
		},
	}

	mockWriter := &MockWriter{}
	logger := New(InfoLevel, mockWriter)
	logger.errorHandler = errorHandler

	logger.Info("test message")

	if len(handledErrors) > 0 {
		t.Error("No errors should be handled for successful write")
	}
}

type MockErrorHandler struct {
	HandleFunc func(error)
}

func (h *MockErrorHandler) Handle(err error) {
	if h.HandleFunc != nil {
		h.HandleFunc(err)
	}
}

func TestLogger_LogAndLogf(t *testing.T) {
	mockWriter := &MockWriter{}
	logger := New(InfoLevel, mockWriter)

	logger.Log(InfoLevel, "direct log", "message")
	output := strings.TrimSpace(mockWriter.String())
	fmt.Println(output)
	if !strings.Contains(output, "direct log message") {
		t.Errorf("Log output incorrect: %q", output)
	}

	mockWriter.Reset()
	logger.Logf(InfoLevel, "formatted %s %d", "log", 42)
	output = strings.TrimSpace(mockWriter.String())
	if !strings.Contains(output, "formatted log 42") {
		t.Errorf("Logf output incorrect: %q", output)
	}
}

func TestLogger_WithFieldsFormatter(t *testing.T) {
	mockWriter := &MockWriter{}
	logger := New(InfoLevel, mockWriter)

	customFormatter := &CustomFormatter{}

	logger.WithFieldsFormatter(customFormatter).WithField("test", "value").Info("message")

	output := strings.TrimSpace(mockWriter.String())
	if !strings.Contains(output, "CUSTOM:test=value") {
		t.Errorf("Custom formatter not used: %q", output)
	}
}

type CustomFormatter struct{}

func (f *CustomFormatter) Format(fields Fields) string {
	return "CUSTOM:" + DefaultFieldsFormatter.Format(fields)
}
