package balogan

import (
	"context"
	"os"
	"sync"
	"time"
)

// Condition represents a function that determines whether logging should occur.
// It takes no parameters and returns a boolean indicating if the log message should be processed.
// This is the most basic type of condition used for simple true/false logic.
//
// Example usage:
//
//	condition := func() bool { return os.Getenv("DEBUG") == "true" }
//	logger.When(condition).Debug("Conditional debug message")
type Condition func() bool

// ContextCondition represents a function that determines whether logging should occur based on context.
// It receives a context.Context parameter and returns a boolean indicating if logging should proceed.
// This is useful for conditions that depend on request-specific or goroutine-local data.
//
// Example usage:
//
//	condition := func(ctx context.Context) bool {
//		return ctx.Value("user_role") == "admin"
//	}
//	logger.WithContextCondition(condition).Info("Admin-only message")
type ContextCondition func(context.Context) bool

// LevelCondition represents a function that determines whether logging should occur based on log level and fields.
// It receives the log level and the current logger fields, returning a boolean for conditional logic.
// This is useful for complex conditions that depend on both the severity and the structured data.
//
// Example usage:
//
//	condition := func(level LogLevel, fields Fields) bool {
//		return level >= ErrorLevel && fields["service"] == "payment"
//	}
//	logger.WithLevelCondition(condition).Error("Payment service error")
type LevelCondition func(LogLevel, Fields) bool

// Predefined conditions for common logging scenarios.
// These conditions can be used directly without additional configuration.
var (
	// Environment conditions check the ENV environment variable

	// InProduction evaluates to true when ENV environment variable equals "production"
	InProduction = func() bool { return os.Getenv("ENV") == "production" }

	// InDevelopment evaluates to true when ENV environment variable equals "development"
	InDevelopment = func() bool { return os.Getenv("ENV") == "development" }

	// InTesting evaluates to true when ENV environment variable equals "test"
	InTesting = func() bool { return os.Getenv("ENV") == "test" }

	// InStaging evaluates to true when ENV environment variable equals "staging"
	InStaging = func() bool { return os.Getenv("ENV") == "staging" }

	// Debug conditions check debug-related environment variables

	// DebugEnabled evaluates to true when DEBUG environment variable equals "true"
	DebugEnabled = func() bool { return os.Getenv("DEBUG") == "true" }

	// VerboseMode evaluates to true when VERBOSE environment variable equals "true"
	VerboseMode = func() bool { return os.Getenv("VERBOSE") == "true" }

	// Time-based conditions evaluate based on current time

	// WorkingHours evaluates to true during business hours (9 AM to 5 PM)
	WorkingHours = func() bool {
		hour := time.Now().Hour()
		return hour >= 9 && hour <= 17
	}

	// Weekend evaluates to true on Saturday and Sunday
	Weekend = func() bool {
		day := time.Now().Weekday()
		return day == time.Saturday || day == time.Sunday
	}

	// Weekday evaluates to true on Monday through Friday (opposite of Weekend)
	Weekday = func() bool { return !Weekend() }
)

// Always returns a condition that always evaluates to true.
// This condition will allow all log messages to pass through.
//
// Example:
//
//	logger.When(Always()).Info("This will always log")
func Always() Condition {
	return func() bool { return true }
}

// Never returns a condition that always evaluates to false.
// This condition will block all log messages from being written.
//
// Example:
//
//	logger.When(Never()).Info("This will never log")
func Never() Condition {
	return func() bool { return false }
}

// EnvEquals returns a condition that checks if an environment variable equals a specific value.
// The condition evaluates to true only when the environment variable matches the expected value.
//
// Parameters:
//
//	key: The name of the environment variable to check.
//	value: The expected value of the environment variable.
//
// Example:
//
//	logger.When(EnvEquals("LOG_LEVEL", "debug")).Debug("Debug message")
func EnvEquals(key, value string) Condition {
	return func() bool {
		return os.Getenv(key) == value
	}
}

// EnvExists returns a condition that checks if an environment variable is set.
// The condition evaluates to true when the environment variable exists, regardless of its value.
//
// Parameters:
//
//	key: The name of the environment variable to check for existence.
//
// Example:
//
//	logger.When(EnvExists("DEBUG")).Info("Debug mode is enabled")
func EnvExists(key string) Condition {
	return func() bool {
		_, exists := os.LookupEnv(key)
		return exists
	}
}

// RandomSample returns a condition that randomly samples log messages based on percentage.
// The condition uses pseudo-random sampling based on current time to determine if logging should occur.
//
// Parameters:
//
//	percentage: The percentage (0-100) of messages that should be logged.
//	            Values <= 0 will never log, values >= 100 will always log.
//
// Example:
//
//	logger.When(RandomSample(10)).Debug("Only 10% of these will be logged")
func RandomSample(percentage int) Condition {
	if percentage <= 0 {
		return Never()
	}
	if percentage >= 100 {
		return Always()
	}

	return func() bool {
		// Use a more random approach with better distribution
		return (time.Now().UnixNano()/1000)%100 < int64(percentage)
	}
}

// RateLimit creates a condition that limits the number of log messages per second.
// The condition maintains internal state to track message count and resets every second.
// This is useful for preventing log spam in high-frequency scenarios.
//
// Parameters:
//
//	maxPerSecond: Maximum number of log messages allowed per second.
//	              Values <= 0 will never allow logging.
//
// Example:
//
//	logger.When(RateLimit(10)).Error("Rate limited error message")
//
// Note: This condition is thread-safe and can be used across multiple goroutines.
func RateLimit(maxPerSecond int) Condition {
	if maxPerSecond <= 0 {
		return Never()
	}

	var (
		mu        sync.Mutex
		lastReset time.Time
		count     int
	)

	return func() bool {
		mu.Lock()
		defer mu.Unlock()

		now := time.Now()
		if now.Sub(lastReset) >= time.Second {
			lastReset = now
			count = 0
		}

		if count < maxPerSecond {
			count++
			return true
		}

		return false
	}
}

// TimeRange creates a condition that allows logging only during specified time range.
// The condition evaluates the current hour and checks if it falls within the specified range.
// Supports both same-day ranges (9-17) and overnight ranges (22-6).
//
// Parameters:
//
//	startHour: The starting hour (0-23) of the allowed time range.
//	endHour: The ending hour (0-23) of the allowed time range.
//
// Example:
//
//	logger.When(TimeRange(9, 17)).Info("Business hours only")
//	logger.When(TimeRange(22, 6)).Error("Night shift logging")
func TimeRange(startHour, endHour int) Condition {
	return func() bool {
		hour := time.Now().Hour()
		if startHour <= endHour {
			return hour >= startHour && hour <= endHour
		}
		// Handle overnight ranges (e.g., 22:00 to 06:00)
		return hour >= startHour || hour <= endHour
	}
}

// HasContextValue returns a context condition that checks if a specific key exists in the context.
// The condition evaluates to true when the context contains the specified key with any non-nil value.
//
// Parameters:
//
//	key: The context key to check for existence.
//
// Example:
//
//	logger.WithContextCondition(HasContextValue("request_id")).Info("Request logged")
func HasContextValue(key interface{}) ContextCondition {
	return func(ctx context.Context) bool {
		return ctx != nil && ctx.Value(key) != nil
	}
}

// ContextValueEquals returns a context condition that checks if a context value equals an expected value.
// The condition evaluates to true when the context contains the key and its value matches the expected value.
//
// Parameters:
//
//	key: The context key to check.
//	expectedValue: The expected value for comparison.
//
// Example:
//
//	logger.WithContextCondition(ContextValueEquals("user_role", "admin")).Info("Admin action")
func ContextValueEquals(key interface{}, expectedValue interface{}) ContextCondition {
	return func(ctx context.Context) bool {
		if ctx == nil {
			return false
		}
		return ctx.Value(key) == expectedValue
	}
}

// OnlyLevel returns a level condition that allows logging only for a specific log level.
// The condition evaluates to true only when the log level exactly matches the target level.
//
// Parameters:
//
//	targetLevel: The specific log level to allow.
//
// Example:
//
//	logger.WithLevelCondition(OnlyLevel(ErrorLevel)).Log(ErrorLevel, "Only errors")
func OnlyLevel(targetLevel LogLevel) LevelCondition {
	return func(level LogLevel, fields Fields) bool {
		return level == targetLevel
	}
}

// MinLevel returns a level condition that allows logging for levels at or above the minimum level.
// The condition evaluates to true when the log level is greater than or equal to the minimum level.
//
// Parameters:
//
//	minLevel: The minimum log level to allow.
//
// Example:
//
//	logger.WithLevelCondition(MinLevel(WarningLevel)).Info("This won't log")
//	logger.WithLevelCondition(MinLevel(WarningLevel)).Error("This will log")
func MinLevel(minLevel LogLevel) LevelCondition {
	return func(level LogLevel, fields Fields) bool {
		return level >= minLevel
	}
}

// HasField returns a level condition that checks if a specific field exists in the log fields.
// The condition evaluates to true when the specified field is present, regardless of its value.
//
// Parameters:
//
//	fieldName: The name of the field to check for existence.
//
// Example:
//
//	logger.WithLevelCondition(HasField("user_id")).WithField("user_id", 123).Info("User action")
func HasField(fieldName string) LevelCondition {
	return func(level LogLevel, fields Fields) bool {
		_, exists := fields[fieldName]
		return exists
	}
}

// FieldEquals returns a level condition that checks if a field equals a specific value.
// The condition evaluates to true when the field exists and its value matches the expected value.
//
// Parameters:
//
//	fieldName: The name of the field to check.
//	expectedValue: The expected value for comparison.
//
// Example:
//
//	logger.WithLevelCondition(FieldEquals("environment", "production")).
//		WithField("environment", "production").Info("Production log")
func FieldEquals(fieldName string, expectedValue interface{}) LevelCondition {
	return func(level LogLevel, fields Fields) bool {
		value, exists := fields[fieldName]
		return exists && value == expectedValue
	}
}

// And returns a condition that evaluates to true only when all provided conditions are true.
// The condition uses short-circuit evaluation, stopping at the first false condition.
//
// Parameters:
//
//	conditions: Variable number of conditions that must all be true.
//
// Example:
//
//	logger.When(And(InDevelopment, DebugEnabled, WorkingHours)).Debug("Complex condition")
func And(conditions ...Condition) Condition {
	return func() bool {
		for _, condition := range conditions {
			if !condition() {
				return false
			}
		}
		return true
	}
}

// Or returns a condition that evaluates to true when at least one of the provided conditions is true.
// The condition uses short-circuit evaluation, stopping at the first true condition.
//
// Parameters:
//
//	conditions: Variable number of conditions where at least one must be true.
//
// Example:
//
//	logger.When(Or(InDevelopment, InTesting, DebugEnabled)).Debug("Debug in dev, test, or debug mode")
func Or(conditions ...Condition) Condition {
	return func() bool {
		for _, condition := range conditions {
			if condition() {
				return true
			}
		}
		return false
	}
}

// Not returns a condition that inverts the result of the provided condition.
// The condition evaluates to true when the provided condition is false, and vice versa.
//
// Parameters:
//
//	condition: The condition to invert.
//
// Example:
//
//	logger.When(Not(InProduction)).Debug("Debug in non-production environments")
func Not(condition Condition) Condition {
	return func() bool {
		return !condition()
	}
}

// Any returns a condition that evaluates to true when at least one of the provided conditions is true.
// This is an alias for Or() provided for better readability in some contexts.
//
// Parameters:
//
//	conditions: Variable number of conditions where at least one must be true.
//
// Example:
//
//	logger.When(Any(Weekend, After(18))).Info("Leisure time logging")
func Any(conditions ...Condition) Condition {
	return Or(conditions...)
}

// All returns a condition that evaluates to true only when all provided conditions are true.
// This is an alias for And() provided for better readability in some contexts.
//
// Parameters:
//
//	conditions: Variable number of conditions that must all be true.
//
// Example:
//
//	logger.When(All(InProduction, HasField("user_id"))).Info("Production user action")
func All(conditions ...Condition) Condition {
	return And(conditions...)
}

// CountBased returns a condition that allows only a specific number of log messages total.
// Once the maximum count is reached, the condition will always return false.
// This is useful for limiting debug output to a specific number of occurrences.
//
// Parameters:
//
//	maxCount: Maximum total number of log messages to allow.
//
// Example:
//
//	logger.When(CountBased(5)).Debug("This will only log 5 times total")
//
// Note: This condition is thread-safe and maintains state across multiple calls.
func CountBased(maxCount int) Condition {
	var (
		mu    sync.Mutex
		count int
	)

	return func() bool {
		mu.Lock()
		defer mu.Unlock()

		if count < maxCount {
			count++
			return true
		}
		return false
	}
}

// SampleEveryN returns a condition that allows every Nth log message to pass through.
// This provides deterministic sampling with better distribution than random sampling.
// The first message is always logged, then every Nth message after that.
//
// Parameters:
//
//	n: The sampling interval. Every Nth message will be logged.
//	   Values <= 1 will allow all messages (same as Always()).
//
// Example:
//
//	logger.When(SampleEveryN(10)).Debug("Every 10th debug message will be logged")
//
// Note: This condition is thread-safe and maintains internal counter state.
func SampleEveryN(n int) Condition {
	if n <= 1 {
		return Always()
	}

	var (
		mu    sync.Mutex
		count int
	)

	return func() bool {
		mu.Lock()
		defer mu.Unlock()

		count++
		return count%n == 1
	}
}
