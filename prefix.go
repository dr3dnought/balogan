package balogan

import (
	"time"
)

type PrefixBuilderFunc func(args ...any) string

func WithLogLevel(level LogLevel) PrefixBuilderFunc {
	return func(args ...any) string {
		return level.String()
	}
}

func WithTimeStamp() PrefixBuilderFunc {
	return func(args ...any) string {
		return time.Now().Format(time.RFC3339)
	}
}

func WithTag(tag string) PrefixBuilderFunc {
	return func(args ...any) string {
		return tag
	}
}
