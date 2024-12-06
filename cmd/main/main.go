package main

import "github.com/dr3dnought/balogan"

func main() {
	fileWriter, err := balogan.NewFileLogWriter("test.log")
	if err != nil {
		panic(err)
	}
	logger := balogan.New(
		balogan.InfoLevel,
		[]balogan.LogWriter{balogan.NewStdOutLogWriter(), fileWriter},
		balogan.WithLogLevel(balogan.DebugLevel),
		balogan.WithTimeStamp(),
	)

	logger.WithTemporaryPrefix(balogan.WithTag("TAG")).Close(4, "Debug message: %s", "test format")
}
