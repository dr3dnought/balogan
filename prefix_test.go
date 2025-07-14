package balogan

import (
	"strings"
	"testing"
	"time"
)

func TestWithLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"Debug level", DebugLevel, "DEBUG"},
		{"Info level", InfoLevel, "INFO"},
		{"Warning level", WarningLevel, "WARNING"},
		{"Error level", ErrorLevel, "ERROR"},
		{"Trace level", TraceLevel, "TRACE"},
		{"Fatal level", FatalLevel, "FATAL"},
		{"Panic level", PanicLevel, "PANIC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := WithLogLevel(tt.level)
			result := builder()

			if result != tt.expected {
				t.Errorf("WithLogLevel(%v) = %q, want %q", tt.level, result, tt.expected)
			}
		})
	}
}

func TestWithLogLevel_IgnoresArgs(t *testing.T) {
	builder := WithLogLevel(InfoLevel)

	result1 := builder("ignored", "args", 123)
	result2 := builder()

	if result1 != result2 {
		t.Errorf("WithLogLevel should ignore arguments: got %q and %q", result1, result2)
	}

	if result1 != "INFO" {
		t.Errorf("Expected 'INFO', got %q", result1)
	}
}

func TestWithTimeStamp(t *testing.T) {
	builder := WithTimeStamp()

	result := builder()

	if result == "" {
		t.Error("WithTimeStamp() returned empty string")
	}

	_, err := time.Parse(time.RFC3339, result)
	if err != nil {
		t.Errorf("WithTimeStamp() returned invalid RFC3339 format: %q, error: %v", result, err)
	}
}

func TestWithTimeStamp_IgnoresArgs(t *testing.T) {
	builder := WithTimeStamp()

	result1 := builder("ignored", "args", 123)
	result2 := builder()

	_, err1 := time.Parse(time.RFC3339, result1)
	_, err2 := time.Parse(time.RFC3339, result2)

	if err1 != nil || err2 != nil {
		t.Errorf("Both results should be valid timestamps: %v, %v", err1, err2)
	}
}

func TestWithTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected string
	}{
		{"Empty tag", "", ""},
		{"Simple tag", "test", "test"},
		{"Tag with spaces", "my tag", "my tag"},
		{"Tag with special chars", "tag-123_@#$%", "tag-123_@#$%"},
		{"Unicode tag", "тег-тест", "тег-тест"},
		{"Long tag", strings.Repeat("a", 100), strings.Repeat("a", 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := WithTag(tt.tag)
			result := builder()

			if result != tt.expected {
				t.Errorf("WithTag(%q) = %q, want %q", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestWithTag_IgnoresArgs(t *testing.T) {
	builder := WithTag("test-tag")

	result1 := builder("ignored", "args", 123)
	result2 := builder()

	if result1 != result2 {
		t.Errorf("WithTag should ignore arguments: got %q and %q", result1, result2)
	}

	if result1 != "test-tag" {
		t.Errorf("Expected 'test-tag', got %q", result1)
	}
}

func TestPrefixBuilderFunc_Consistency(t *testing.T) {
	levelBuilder := WithLogLevel(ErrorLevel)
	tagBuilder := WithTag("consistency-test")

	level1 := levelBuilder()
	level2 := levelBuilder()
	tag1 := tagBuilder()
	tag2 := tagBuilder()

	if level1 != level2 {
		t.Errorf("LogLevel builder should be consistent: %q != %q", level1, level2)
	}

	if tag1 != tag2 {
		t.Errorf("Tag builder should be consistent: %q != %q", tag1, tag2)
	}
}

func TestPrefixBuilderFunc_ThreadSafety(t *testing.T) {
	levelBuilder := WithLogLevel(WarningLevel)
	tagBuilder := WithTag("thread-test")

	results := make([]string, 10)
	for i := range results {
		if i%2 == 0 {
			results[i] = levelBuilder()
		} else {
			results[i] = tagBuilder()
		}
	}

	for i := 0; i < len(results); i += 2 {
		if results[i] != "WARNING" {
			t.Errorf("Expected 'WARNING', got %q", results[i])
		}
	}

	for i := 1; i < len(results); i += 2 {
		if results[i] != "thread-test" {
			t.Errorf("Expected 'thread-test', got %q", results[i])
		}
	}
}

func TestPrefixBuilderFunc_Integration(t *testing.T) {
	builders := []struct {
		name    string
		builder PrefixBuilderFunc
		check   func(string) bool
	}{
		{
			name:    "LogLevel",
			builder: WithLogLevel(DebugLevel),
			check: func(s string) bool {
				return s == "DEBUG"
			},
		},
		{
			name:    "TimeStamp",
			builder: WithTimeStamp(),
			check: func(s string) bool {
				_, err := time.Parse(time.RFC3339, s)
				return err == nil
			},
		},
		{
			name:    "Tag",
			builder: WithTag("integration-test"),
			check: func(s string) bool {
				return s == "integration-test"
			},
		},
	}

	for _, tt := range builders {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder()

			if !tt.check(result) {
				t.Errorf("Builder %s failed validation: %q", tt.name, result)
			}
		})
	}
}
