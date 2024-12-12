package main

import (
	"github.com/dr3dnought/balogan"
)

func main() {
	logger := balogan.New(balogan.DebugLevel, balogan.NewStdOutLogWriter())
	logger.Debug("bob")

}
