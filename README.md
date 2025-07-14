# balogan

The `balogan` package provides useful and simple logger which helps you to customize and structure your logs level easily.  
Moreover, the balogan is very flexible and can be extended by you to suit your needs.  

I am following a consept, that colorful logs and other visual extenshions are ***redundant*** and that's why balogan will ***never*** support such things from the box.  
By the way, I strive to make the logger as extensible as possible, so that anyone can create an addon with the functionality they need  

## Why `balogan`?
* ***Fast start***. You can use balogan without spending your time to configuration. It's to easy
* ***Flexibility***. You can easily extend logger funcionality by writing your own prefixes and writers.
* ***Complete log levels***. Seven log levels from TRACE to PANIC with proper filtering and critical behavior.
* ***Concurrency***. Balogan can write in different log sources concurrently.
* ***Error Handling***. Balogan provides functional to handle error for writing logs processes 

## Installation

```bash
go get -u github.com/dr3dnought/balogan
```

## Getting started

### Creating balogan instance

For using logger, import package `github.com/dr3dnought/balogan`

```go
package main

import "github.com/dr3dnought/balogan"

func main() {
  // Define logger with TraceLevel to show all messages
  logger := balogan.New(balogan.TraceLevel, balogan.DefaultWriter)
  
  // Different log levels
  logger.Trace("This is trace message")     // Output: TRACE This is trace message
  logger.Debug("Debug message")             // Output: DEBUG Debug message
  logger.Info("Information message")        // Output: INFO Information message
  logger.Warning("Warning message")         // Output: WARNING Warning message
  logger.Error("Error message")             // Output: ERROR Error message
  
  // Format alternatives
  logger.Debugf("Hello, %s", "world!")     // Output: DEBUG Hello, world!
  logger.Infof("User %s logged in", "john") // Output: INFO User john logged in
}
```
> `balogan.DefaultWriter` will write logs to `stdout`  

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

