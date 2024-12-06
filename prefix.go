package balogan

import (
	"time"
)

type PrefixBuilderFunc func(args ...interface{}) string

func WithLogLevel(level LogLevel) PrefixBuilderFunc {
	return func(args ...interface{}) string {
		return level.String()
	}
}

func WithTimeStamp() PrefixBuilderFunc {
	return func(args ...interface{}) string {
		return time.Now().Format(time.RFC3339)
	}
}

func WithTag(tag string) PrefixBuilderFunc {
	return func(args ...interface{}) string {
		return tag
	}
}
