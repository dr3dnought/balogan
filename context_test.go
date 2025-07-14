package balogan

import (
	"context"
	"fmt"
	"testing"
)

func TestLogger_WithContext(t *testing.T) {
	logger := New(InfoLevel, NewStdOutLogWriter())

	ctx := context.Background()
	ctxWithLogger := logger.WithContext(ctx)

	retrievedLogger, ok := FromContext(ctxWithLogger)
	if !ok {
		t.Error("Expected to find logger in context")
	}

	if retrievedLogger != logger {
		t.Error("Retrieved logger should be the same instance")
	}
}

func TestFromContext_Success(t *testing.T) {
	logger := New(DebugLevel, NewStdOutLogWriter())
	logger = logger.WithField("test", "value")

	ctx := context.Background()
	ctxWithLogger := logger.WithContext(ctx)

	retrievedLogger, ok := FromContext(ctxWithLogger)

	if !ok {
		t.Error("Expected to successfully retrieve logger from context")
	}

	if retrievedLogger == nil {
		t.Error("Retrieved logger should not be nil")
	}

	if len(retrievedLogger.GetFields()) != len(logger.GetFields()) {
		t.Error("Retrieved logger should have same fields as original")
	}

	if retrievedLogger != logger {
		t.Error("Retrieved logger should be the exact same instance")
	}
}

func TestFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	logger, ok := FromContext(ctx)

	if ok {
		t.Error("Expected NOT to find logger in empty context")
	}

	if logger != nil {
		t.Error("Logger should be nil when not found in context")
	}
}

func TestFromContext_WrongType(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextKey, "not a logger")

	logger, ok := FromContext(ctx)

	if ok {
		t.Error("Expected NOT to find logger when wrong type in context")
	}

	if logger != nil {
		t.Error("Logger should be nil when wrong type in context")
	}
}

func TestFromContext_DifferentKey(t *testing.T) {
	ctx := context.Background()
	logger := New(InfoLevel, NewStdOutLogWriter())
	ctx = context.WithValue(ctx, "different:key", logger)

	retrievedLogger, ok := FromContext(ctx)

	if ok {
		t.Error("Expected NOT to find logger with different key")
	}

	if retrievedLogger != nil {
		t.Error("Logger should be nil when using different key")
	}
}

func TestLogger_WithContext_PreservesExistingValues(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "existing", "value")

	logger := New(InfoLevel, NewStdOutLogWriter())
	ctxWithLogger := logger.WithContext(ctx)

	existingValue := ctxWithLogger.Value("existing")
	if existingValue != "value" {
		t.Error("Existing context values should be preserved")
	}

	retrievedLogger, ok := FromContext(ctxWithLogger)
	if !ok {
		t.Error("Logger should be present in context")
	}

	if retrievedLogger != logger {
		t.Error("Retrieved logger should match original")
	}
}

func TestLogger_WithContext_Chaining(t *testing.T) {
	logger1 := New(InfoLevel, NewStdOutLogWriter()).WithField("logger", "first")
	ctx := context.Background()
	ctx1 := logger1.WithContext(ctx)

	logger2 := New(DebugLevel, NewStdOutLogWriter()).WithField("logger", "second")
	ctx2 := logger2.WithContext(ctx1)

	retrievedLogger, ok := FromContext(ctx2)
	if !ok {
		t.Error("Expected to find logger in context")
	}

	if retrievedLogger != logger2 {
		t.Error("Should retrieve the most recently added logger")
	}

	fields := retrievedLogger.GetFields()
	if fields["logger"] != "second" {
		t.Error("Should have fields from second logger")
	}
}

func TestLogger_WithContext_MultipleDifferentLoggers(t *testing.T) {
	logger1 := New(InfoLevel, NewStdOutLogWriter()).WithField("instance", "1")
	logger2 := New(ErrorLevel, NewStdOutLogWriter()).WithField("instance", "2")

	ctx := context.Background()

	ctx1 := logger1.WithContext(ctx)
	retrieved1, ok1 := FromContext(ctx1)

	ctx2 := logger2.WithContext(ctx)
	retrieved2, ok2 := FromContext(ctx2)

	if !ok1 || !ok2 {
		t.Error("Both loggers should be retrievable from their respective contexts")
	}

	if retrieved1 == retrieved2 {
		t.Error("Retrieved loggers should be different instances")
	}

	fields1 := retrieved1.GetFields()
	fields2 := retrieved2.GetFields()

	if fields1["instance"] == fields2["instance"] {
		t.Error("Loggers should have different field values")
	}
}

func TestFromContext_NilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("FromContext should not panic with nil context")
		}
	}()

	logger, ok := FromContext(context.TODO())

	if ok {
		t.Error("Expected NOT to find logger in nil context")
	}

	if logger != nil {
		t.Error("Logger should be nil with nil context")
	}
}

func TestContext_ThreadSafety(t *testing.T) {
	logger := New(InfoLevel, NewStdOutLogWriter()).WithField("thread_test", "value")
	ctx := context.Background()

	const numGoroutines = 100
	const numOperations = 100

	results := make(chan bool, numGoroutines*numOperations)

	for i := range numGoroutines {
		go func(goroutineID int) {
			localLogger := logger.WithField("goroutine_id", goroutineID)

			for j := 0; j < numOperations; j++ {
				ctxWithLogger := localLogger.WithContext(ctx)

				retrievedLogger, ok := FromContext(ctxWithLogger)

				if !ok || retrievedLogger == nil {
					results <- false
					continue
				}

				fields := retrievedLogger.GetFields()
				if fields["thread_test"] != "value" || fields["goroutine_id"] != goroutineID {
					results <- false
					continue
				}

				results <- true
			}
		}(i)
	}

	successCount := 0
	totalExpected := numGoroutines * numOperations

	for range totalExpected {
		if <-results {
			successCount++
		}
	}

	if successCount != totalExpected {
		t.Errorf("Expected %d successful operations, got %d", totalExpected, successCount)
	}
}

func TestContext_RealWorldUsage(t *testing.T) {
	baseLogger := New(InfoLevel, NewStdOutLogWriter()).WithFields(Fields{
		"service": "test_service",
		"version": "1.0.0",
	})

	processStep := func(ctx context.Context, step string) error {
		logger, ok := FromContext(ctx)
		if !ok {
			return fmt.Errorf("no logger in context")
		}

		stepLogger := logger.WithField("step", step)

		fields := stepLogger.GetFields()
		expectedFields := []string{"service", "version", "request_id", "user_id", "step"}
		for _, field := range expectedFields {
			if _, exists := fields[field]; !exists {
				return fmt.Errorf("missing field: %s", field)
			}
		}

		return nil
	}

	processRequest := func(requestID string, userID int) error {
		reqLogger := baseLogger.WithFields(Fields{
			"request_id": requestID,
			"user_id":    userID,
		})

		ctx := context.Background()
		ctx = reqLogger.WithContext(ctx)

		return processStep(ctx, "validation")
	}

	err := processRequest("req-123", 456)
	if err != nil {
		t.Errorf("Request processing failed: %v", err)
	}
}
