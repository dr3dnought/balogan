# balogan
[![Release](https://img.shields.io/github/release/dr3dnought/balogan?style=flat-square)](https://github.com/dr3dnought/balogan/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dr3dnought/balogan)](https://golang.org)
[![Go Reference](https://pkg.go.dev/badge/github.com/dr3dnought/balogan.svg)](https://pkg.go.dev/github.com/dr3dnought/balogan)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/dr3dnought/balogan/branch/main/graph/badge.svg)](https://codecov.io/gh/dr3dnought/balogan)


The `balogan` package provides a **powerful yet simple** logger with extensive customization capabilities and modern structured logging support. Designed for developers who need **flexibility without complexity**.

**Key Features:**
- 🚀 **Fast start** - works out of the box with zero configuration
- 📊 **7 log levels** from TRACE to PANIC with proper filtering
- 🏗️ **Structured logging** with fields (JSON, logfmt, key=value formats)
- 🔧 **Highly extensible** - custom prefixes, writers, and formatters
- ⚡ **Concurrent logging** to multiple destinations
- 🛡️ **Error handling** for robust logging operations
- 🧵 **Thread-safe** with immutable logger instances

I am following a concept that colorful logs and other visual extensions are ***redundant*** and that's why balogan will ***never*** support such things out of the box. However, I strive to make the logger as extensible as possible, so that anyone can create an addon with the functionality they need.  

## Why Choose `balogan`?

### 🚀 **Zero Configuration Start**
Works immediately without any setup - just import and use!

### 📊 **Complete Logging Solution**  
- **7 log levels**: TRACE (-1) → DEBUG (0) → INFO (1) → WARNING (2) → ERROR (3) → FATAL (4) → PANIC (5)
- **Smart filtering**: Only logs at or above your configured level
- **Critical handling**: FATAL exits program, PANIC causes panic

### 🏗️ **Modern Structured Logging**
- **Multiple formats**: JSON, logfmt, key=value (default)
- **Convenient API**: `.WithJSON()`, `.WithLogfmt()`, `.WithField()`
- **Chainable methods**: `logger.WithJSON().WithField("user", "john").Info("message")`

### 🔧 **Maximum Flexibility**
- **Custom prefixes**: timestamps, tags, or your own functions
- **Multiple writers**: stdout, files, or custom destinations  
- **Extensible formatters**: implement your own field formatting

### ⚡ **Performance & Safety**
- **Concurrent logging**: parallel writes to multiple destinations
- **Thread-safe**: use from multiple goroutines safely
- **Immutable loggers**: each method returns new instance
- **Error handling**: robust error handling for all operations 

## Installation

```bash
go get -u github.com/dr3dnought/balogan
```

## Quick Start

### 1. Basic Logging

```go
package main

import "github.com/dr3dnought/balogan"

func main() {
    // Create logger - works immediately!
    logger := balogan.New(balogan.InfoLevel, balogan.DefaultWriter)
    
    // Basic logging
    logger.Info("Application started")
    logger.Warning("This is a warning")
    logger.Error("Something went wrong")
    
    // Format strings
    logger.Infof("User %s logged in", "john")
}
```

### 2. Structured Logging (Modern Approach)

```go
// Add context with fields
logger.WithField("user_id", 123).Info("User action")
// Output: INFO user_id=123 User action

// Multiple fields
logger.WithFields(balogan.Fields{
    "request_id": "req-456",
    "method":     "POST",
    "duration":   "150ms",
}).Info("Request processed")
// Output: INFO duration=150ms method=POST request_id=req-456 Request processed

// Different formats
logger.WithJSON().WithField("event", "login").Info("User event")
// Output: INFO {"event":"login"} User event

logger.WithLogfmt().WithField("service", "api server").Info("Service started")  
// Output: INFO service="api server" Service started
```

### 3. Enhanced Logging with Prefixes

```go
// Add timestamps and tags
logger := balogan.New(
    balogan.InfoLevel,
    balogan.DefaultWriter,
    balogan.WithTimeStamp(),
    balogan.WithTag("[API]"),
)

logger.Info("Server started")
// Output: INFO 2024-12-13T15:30:45Z [API] Server started

// Combine with structured fields
logger.WithField("port", 8080).Info("Listening on port")
// Output: INFO 2024-12-13T15:30:45Z [API] port=8080 Listening on port
```

> 💡 **Default Behavior**: 
> - Uses `stdout` for output
> - `KeyValueFormatter` for fields (key=value format)
> - Shows logs at configured level and above  

Every log function has `format` alternative. It works like `Sprintf` function in `fmt` package.

## Log Levels

balogan supports 7 different log levels with increasing severity:

| Level | Value | Description | Behavior |
|-------|-------|-------------|----------|
| `TraceLevel` | -1 | Most detailed information, typically only for diagnosing problems | Standard logging |
| `DebugLevel` | 0 | Information useful for debugging | Standard logging |
| `InfoLevel` | 1 | General informational messages | Standard logging |
| `WarningLevel` | 2 | Warning messages for potentially harmful situations | Standard logging |
| `ErrorLevel` | 3 | Error messages for error conditions | Standard logging |
| `FatalLevel` | 4 | Critical errors that cause program termination | **Calls `os.Exit(1)` after logging** |
| `PanicLevel` | 5 | Most severe errors that cause program panic | **Calls `panic()` after logging** |

### Level Filtering

When you set a log level, only messages at that level or higher will be output:

```go
package main

import "github.com/dr3dnought/balogan"

func main() {
    // Logger with InfoLevel will show INFO, WARNING, ERROR, FATAL, PANIC
    // but will skip TRACE and DEBUG
    logger := balogan.New(balogan.InfoLevel, balogan.DefaultWriter)
    
    logger.Trace("This won't be shown")   // Skipped
    logger.Debug("This won't be shown")   // Skipped
    logger.Info("This will be shown")     // Output: INFO This will be shown
    logger.Warning("This will be shown")  // Output: WARNING This will be shown
    logger.Error("This will be shown")    // Output: ERROR This will be shown
}
```

### All Available Methods

```go
// Standard logging methods
logger.Trace("trace message")
logger.Tracef("trace with %s", "parameters")

logger.Debug("debug message")
logger.Debugf("debug with %s", "parameters")

logger.Info("info message")
logger.Infof("info with %s", "parameters")

logger.Warning("warning message")
logger.Warningf("warning with %s", "parameters")

logger.Error("error message")
logger.Errorf("error with %s", "parameters")

// Critical methods (terminate program)
logger.Fatal("fatal message")        // Logs and calls os.Exit(1)
logger.Fatalf("fatal with %s", "parameters")

logger.Panic("panic message")        // Logs and calls panic()
logger.Panicf("panic with %s", "parameters")
```

> ⚠️ **Important**: `Fatal` and `Panic` methods will terminate your program after logging! Use them only for truly critical errors.

## Structured Logging

balogan supports structured logging with fields (key-value pairs) that can be formatted in different ways. This allows you to add context to your logs in a machine-readable format.

### Basic Structured Logging

```go
package main

import "github.com/dr3dnought/balogan"

func main() {
    logger := balogan.New(balogan.InfoLevel, balogan.DefaultWriter)
    
    // Add single field
    logger.WithField("user_id", 123).Info("User logged in")
    // Output: INFO user_id=123 User logged in
    
    // Add multiple fields
    logger.WithFields(balogan.Fields{
        "request_id": "req-456",
        "method":     "POST", 
        "duration":   "42ms",
    }).Info("Request processed")
    // Output: INFO duration=42ms method=POST request_id=req-456 Request processed
}
```

### Field Formatters

balogan provides three built-in field formatters:

#### 1. Key-Value Formatter (Default)

**This is the default formatter** - used when no specific formatter is set.

```go
// Default behavior - no need to specify formatter
logger.WithField("key", "value").Info("message")
// Output: INFO key=value message

// Explicit usage (same result)
logger.WithKeyValue().WithField("key", "value").Info("message")
// Output: INFO key=value message

// Custom separator
logger.WithKeyValueSeparator(" | ").WithFields(balogan.Fields{
    "user": "john",
    "role": "admin",
}).Info("User info")
// Output: INFO role=admin | user=john User info
```

#### 2. JSON Formatter
```go
// Long way (still supported)
jsonLogger := logger.WithFieldsFormatter(&balogan.JSONFormatter{})

// Short way (recommended)
logger.WithJSON().WithFields(balogan.Fields{
    "user_id": 123,
    "action":  "login",
}).Info("User action")
// Output: INFO {"action":"login","user_id":123} User action
```

#### 3. Logfmt Formatter
```go
// Long way (still supported)
logfmtLogger := logger.WithFieldsFormatter(&balogan.LogfmtFormatter{})

// Short way (recommended)
logger.WithLogfmt().WithFields(balogan.Fields{
    "service": "api",
    "message": "request processed",
}).Info("Service event")
// Output: INFO service=api message="request processed" Service event
```

### Quick Format Switching

balogan provides convenient methods for quick format switching:

```go
// Switch to JSON format
jsonLogger := logger.WithJSON()

// Switch to logfmt format  
logfmtLogger := logger.WithLogfmt()

// Switch to key-value format (default)
kvLogger := logger.WithKeyValue()

// Key-value with custom separator
customLogger := logger.WithKeyValueSeparator(" | ")

// You can chain format switches
logger.WithJSON().
    WithField("step", 1).
    Info("Step 1") // JSON format

logger.WithLogfmt().
    WithField("step", 2).
    Info("Step 2") // Logfmt format
```

### Changing Default Formatter Globally

You can change the default formatter for all new loggers:

```go
package main

import "github.com/dr3dnought/balogan"

func init() {
    // Set JSON as default formatter for all loggers
    balogan.DefaultFieldsFormatter = &balogan.JSONFormatter{}
    
    // Or logfmt
    // balogan.DefaultFieldsFormatter = &balogan.LogfmtFormatter{}
    
    // Or key-value with custom separator
    // balogan.DefaultFieldsFormatter = &balogan.KeyValueFormatter{Separator: " | "}
}

func main() {
    // This logger will now use JSON by default
    logger := balogan.New(balogan.InfoLevel, balogan.DefaultWriter)
    logger.WithField("user", "john").Info("User logged in")
    // Output: INFO {"user":"john"} User logged in
}
```

### Persistent Fields

Create loggers with persistent fields for components:

```go
// Component logger with persistent fields
dbLogger := logger.WithFields(balogan.Fields{
    "component": "database",
    "host":      "localhost",
})

dbLogger.Info("Connection established")
// Output: INFO component=database host=localhost Connection established

dbLogger.WithField("query_time", "15ms").Debug("Query executed") 
// Output: DEBUG component=database host=localhost query_time=15ms Query executed
```

### Combining with Prefixes

Structured fields work seamlessly with traditional prefixes:

```go
logger := balogan.New(
    balogan.InfoLevel,
    balogan.DefaultWriter,
    balogan.WithTimeStamp(),
    balogan.WithTag("[API]"),
)

logger.WithField("endpoint", "/users").Info("Request received")
// Output: INFO 2024-12-13T03:10:56+03:00 [API] endpoint=/users Request received
```

### Custom Field Formatters

You can create custom field formatters by implementing the `FieldsFormatter` interface:

```go
package main

import (
    "fmt"
    "strings"
    "github.com/dr3dnought/balogan"
)

// Custom formatter that formats fields as XML
type XMLFormatter struct{}

func (f *XMLFormatter) Format(fields balogan.Fields) string {
    if len(fields) == 0 {
        return ""
    }
    
    var parts []string
    for key, value := range fields {
        parts = append(parts, fmt.Sprintf("<%s>%v</%s>", key, value, key))
    }
    
    return "<fields>" + strings.Join(parts, "") + "</fields>"
}

func main() {
    logger := balogan.New(balogan.InfoLevel, balogan.DefaultWriter)
    xmlLogger := logger.WithFieldsFormatter(&XMLFormatter{})
    
    xmlLogger.WithFields(balogan.Fields{
        "user": "john",
        "age":  30,
    }).Info("User data")
    // Output: INFO <fields><age>30</age><user>john</user></fields> User data
}
```

### Using prefixes

So, now our log message is not informative. Let's improve it.

```go
package main

import "github.com/dr3dnought/balogan"

func main() {
  // Define logger with log level and prefixes
  logger := balogan.New(
      balogan.InfoLevel,
      balogan.DefaultWriter,
      balogan.WithTimeStamp(),
      balogan.WithTag("[APP]"),
  )
  logger.Info("Application started") // Output: INFO 2024-12-13T03:10:56+03:00 [APP] Application started
}
```

Now, it looks like good log string.

You can use other standart prefixes, here's the whole list:

* `WithLogLevel(level balogan.LogLevel)` accepts the log level and prints it's string version.
* `WithTimeStamp()` prints `time.Now()` and prints it in `RFC3339` format
* `WithTag(tag string)` accepts string and prints it.

> The sequence of prefixes depends on the order in which their functions are called

## Custom prefixes

You can easily write your own prefix, just create a function which type is `balogan.PrefixBuilderFunc`

Let's create a prefix, which will print `GOOS` as an example:

```go
package main

import (
  "github.com/dr3dnought/balogan"
  "runtime"
)

func WithGOOS() balogan.PrefixBuilderFunc {
  return func(args ...any) string {
    return runtime.GOOS
  }
}

func main() {
  // Define logger with custom prefix
  logger := balogan.New(
      balogan.InfoLevel,
      balogan.DefaultWriter,
      WithGOOS(),
      balogan.WithTag("[SYSTEM]"),
  )
  logger.Info("System information") // Output: INFO darwin [SYSTEM] System information
}
```

## Temporary prefixes

Let's imagine case, when you need to print prefix without configuring new logger.

So, every balogan instance has method `WithTemporaryPrefix` that accepts addition prefixes and produces new instance of your logger with new prefixes.

Here's sample:

```go
package main

import "github.com/dr3dnought/balogan"

func main() {
  // Define base logger
  logger := balogan.New(
      balogan.InfoLevel,
      balogan.DefaultWriter,
      balogan.WithTimeStamp(),
  )
  logger.Info("Application message") // Output: INFO 2024-12-13T03:10:56+03:00 Application message

  // Add temporary prefix for database operations
  dbLogger := logger.WithTemporaryPrefix(balogan.WithTag("[DB]"))
  dbLogger.Info("Database connected") // Output: INFO 2024-12-13T03:10:56+03:00 [DB] Database connected
  
  // Add temporary prefix for API operations
  apiLogger := logger.WithTemporaryPrefix(balogan.WithTag("[API]"))
  apiLogger.Warning("Rate limit exceeded") // Output: WARNING 2024-12-13T03:10:56+03:00 [API] Rate limit exceeded
}
```

Like you can see, it's very extandable!

## Integration `context.Context`

balogan provides functions for putting logger and sub-logger in `context.Context`

```go
package main

import (
	"context"

	"github.com/dr3dnought/balogan"
)

func main() {
	logger := balogan.New(
		balogan.InfoLevel, 
		balogan.DefaultWriter, 
		balogan.WithTimeStamp(),
		balogan.WithTag("[MAIN]"),
	)
	logger.Info("Application started") // Output: INFO 2024-12-13T03:10:56+03:00 [MAIN] Application started

	ctx := context.Background()
	ctx = logger.WithTemporaryPrefix(balogan.WithTag("[WORKER]")).WithContext(ctx)
	processWithCtx(ctx) // Output: INFO 2024-12-13T03:10:56+03:00 [MAIN] [WORKER] Processing task
}

func processWithCtx(ctx context.Context) {
	logger, _ := balogan.FromContext(ctx)
	logger.Info("Processing task") 
}
```

That's way we put modificated logger into context and use it from context in `logWithCtx` func.

## Real-World Examples

### Web Application Logging

```go
package main

import (
    "net/http"
    "time"
    "github.com/dr3dnought/balogan"
)

func main() {
    // Create base logger with timestamps
    logger := balogan.New(
        balogan.InfoLevel,
        balogan.DefaultWriter,
        balogan.WithTimeStamp(),
        balogan.WithTag("[WEB]"),
    )
    
    // HTTP request handler
    http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Create request-specific logger with structured fields
        reqLogger := logger.WithJSON().WithFields(balogan.Fields{
            "request_id": generateRequestID(),
            "method":     r.Method,
            "path":       r.URL.Path,
            "ip":         r.RemoteAddr,
            "user_agent": r.UserAgent(),
        })
        
        reqLogger.Info("Request received")
        
        // Process request...
        time.Sleep(50 * time.Millisecond) // Simulate processing
        
        // Log response with additional fields
        reqLogger.WithFields(balogan.Fields{
            "status":   200,
            "duration": time.Since(start).String(),
            "size":     "1.2KB",
        }).Info("Request completed")
    })
    
    logger.Info("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}

func generateRequestID() string {
    return "req-" + time.Now().Format("20060102150405")
}
```

### Microservice with Different Components

```go
package main

import "github.com/dr3dnought/balogan"

func main() {
    // Base application logger
    logger := balogan.New(
        balogan.InfoLevel,
        balogan.DefaultWriter,
        balogan.WithTimeStamp(),
    )
    
    // Database component logger (JSON for machine processing)
    dbLogger := logger.WithJSON().WithFields(balogan.Fields{
        "component": "database",
        "version":   "1.0.0",
    })
    
    // API component logger (logfmt for human readable)
    apiLogger := logger.WithLogfmt().WithFields(balogan.Fields{
        "component": "api",
        "version":   "2.1.0",
    })
    
    // Cache component logger (key-value with custom separator)
    cacheLogger := logger.WithKeyValueSeparator(" | ").WithFields(balogan.Fields{
        "component": "cache",
        "type":      "redis",
    })
    
    // Usage examples
    dbLogger.WithField("query_time", "15ms").Info("Query executed")
    // Output: INFO 2024-12-13T15:30:45Z {"component":"database","query_time":"15ms","version":"1.0.0"} Query executed
    
    apiLogger.WithField("endpoint", "/users").Info("API call")  
    // Output: INFO 2024-12-13T15:30:45Z component=api endpoint=/users version=2.1.0 API call
    
    cacheLogger.WithField("hit_rate", "95%").Info("Cache stats")
    // Output: INFO 2024-12-13T15:30:45Z component=cache | hit_rate=95% | type=redis Cache stats
}
```

### Error Handling and Monitoring

```go
package main

import (
    "fmt"
    "github.com/dr3dnought/balogan"
)

func processPayment(logger *balogan.Logger, paymentID string) {
    // Create payment-specific logger
    paymentLogger := logger.WithFields(balogan.Fields{
        "payment_id": paymentID,
        "service":    "payment_processor",
    })
    
    paymentLogger.Info("Processing payment")
    
    // Simulate payment processing
    if err := validatePayment(paymentID); err != nil {
        // Log error with detailed context
        paymentLogger.WithJSON().WithFields(balogan.Fields{
            "error_code":    "VALIDATION_FAILED",
            "error_message": err.Error(),
            "retry_count":   0,
            "amount":        "99.99",
            "currency":      "USD",
        }).Error("Payment validation failed")
        return
    }
    
    // Success logging
    paymentLogger.WithField("status", "completed").Info("Payment processed successfully")
}

func validatePayment(paymentID string) error {
    // Simulation
    return fmt.Errorf("insufficient funds")
}

func main() {
    logger := balogan.New(balogan.InfoLevel, balogan.DefaultWriter)
    processPayment(logger, "pay-12345")
}
```

## API Reference Summary

### Logger Creation
```go
// Basic logger
logger := balogan.New(level, writer, prefixes...)

// From configuration
logger := balogan.NewFromConfig(&balogan.BaloganConfig{...})
```

### Log Levels (in order)
```go
balogan.TraceLevel   // -1 (most verbose)
balogan.DebugLevel   //  0
balogan.InfoLevel    //  1  
balogan.WarningLevel //  2
balogan.ErrorLevel   //  3
balogan.FatalLevel   //  4 (calls os.Exit(1))
balogan.PanicLevel   //  5 (calls panic())
```

### Logging Methods
```go
logger.Trace("message")   logger.Tracef("format", args...)
logger.Debug("message")   logger.Debugf("format", args...)
logger.Info("message")    logger.Infof("format", args...)
logger.Warning("message") logger.Warningf("format", args...)
logger.Error("message")   logger.Errorf("format", args...)
logger.Fatal("message")   logger.Fatalf("format", args...)
logger.Panic("message")   logger.Panicf("format", args...)
```

### Structured Logging
```go
// Single field
logger.WithField("key", "value")

// Multiple fields
logger.WithFields(balogan.Fields{"key1": "value1", "key2": "value2"})

// Format switching
logger.WithJSON()              // {"key":"value"}
logger.WithLogfmt()            // key=value or key="value with spaces"
logger.WithKeyValue()          // key=value (default)
logger.WithKeyValueSeparator(" | ") // key=value | key2=value2

// Formatter methods (long form)
logger.WithFieldsFormatter(&balogan.JSONFormatter{})
logger.WithFieldsFormatter(&balogan.LogfmtFormatter{})
logger.WithFieldsFormatter(&balogan.KeyValueFormatter{Separator: " | "})
```

### Prefixes
```go
balogan.WithTimeStamp()           // RFC3339 timestamp
balogan.WithTag("tag")            // Custom tag
balogan.WithLogLevel(level)       // Log level in output

// Custom prefix
func WithCustom() balogan.PrefixBuilderFunc {
    return func(args ...any) string { return "custom" }
}
```

### Temporary Extensions
```go
logger.WithTemporaryPrefix(prefix...)  // Add prefixes
logger.WithContext(ctx)                // Put in context
balogan.FromContext(ctx)               // Get from context
```

---

**balogan** - Simple, powerful, and extensible logging for Go! 🚀

