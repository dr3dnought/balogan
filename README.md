# balogan

The `balogan` package provides useful and simple logger which helps you to customize and structure your logs level easily.  
Moreover, the balogan is very flexible and can be extended by you to suit your needs.  

I am following a consept, that colorful logs and other visual extenshions are ***redundant*** and that's why balogan will ***never*** support such things from the box.  
By the way, I strive to make the logger as extensible as possible, so that anyone can create an addon with the functionality they need  

## Why `balogan`?
* ***Fast start***. You can use balogan without spending your time to configuration. It's to easy
* ***Flexibility***. You can easily extend logger funcionality by writing your own prefixes and writers.
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
  // Define logger with log level
  logger := balogan.New(balogan.DebugLevel, balogan.DefaultWriter)
  logger.Debug("Debug message") // Output: Debug message
  logger.Debugf("Hello, %s", "world!") // Output: Hello world!
}
```
> `balogan.DefaultWriter` will write logs to `stdout`  

Every log function has `format` alternative. It works like `Sprintf` function in `fmt` package.

### Using prefixes

So, now our log message is not informative. Let's improve it.

```go
package main

import "github.com/dr3dnought/balogan"

func main() {
  // Define logger with log level
  logger := balogan.New(
      balogan.DebugLevel,
      balogan.DefaultWriter,
      balogan.WithLogLevel(balogan.DebugLevel),
      balogan.WithTimeStamp(),
  )
  logger.Debug("Debug message") // Output: DEBUG 2024-12-13T03:10:56+03:00 Debug message
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
  // Define logger with log level
  logger := balogan.New(
      balogan.DebugLevel,
      balogan.DefaultWriter,
      WithGOOS(),
  )
  logger.Debug("Debug message") // Output: darwin Debug message
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
  // Define logger with log level
  logger := balogan.New(
      balogan.DebugLevel,
      balogan.DefaultWriter,
      balogan.WithLogLevel(balogan.DebugLevel),
  )
  logger.Debug("some message") // Output: DEBUG some message

  // Add temporary prefix
  logger.WithTemporaryPrefix(balogan.WithTag("database")).Debug("some message")

  // Output: DEBUG database some message 
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
	logger := balogan.New(balogan.DebugLevel, balogan.DefaultWriter, balogan.WithLogLevel(balogan.DebugLevel))
	logger.Debug("debug message") // Output: DEBUG debug message

	ctx := context.Background()
	ctx = logger.WithTemporaryPrefix(balogan.WithLogLevel(balogan.ErrorLevel)).WithContext(ctx)
	logWithCtx(ctx) // Output: ERROR error message
}

func logWithCtx(ctx context.Context) {
	logger, _ := balogan.FromContext(ctx)
	logger.Error("error message") 
}
```

That's way we put modificated logger into context and use it from context in `logWithCtx` func.

