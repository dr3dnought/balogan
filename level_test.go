package balogan

import (
	"testing"
)

func TestLogLevel_Constants(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected int
		name     string
	}{
		{TraceLevel, -1, "TraceLevel"},
		{DebugLevel, 0, "DebugLevel"},
		{InfoLevel, 1, "InfoLevel"},
		{WarningLevel, 2, "WarningLevel"},
		{ErrorLevel, 3, "ErrorLevel"},
		{FatalLevel, 4, "FatalLevel"},
		{PanicLevel, 5, "PanicLevel"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if int(test.level) != test.expected {
				t.Errorf("%s expected to be %d, got %d", test.name, test.expected, int(test.level))
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{TraceLevel, "TRACE"},
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarningLevel, "WARNING"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{PanicLevel, "PANIC"},
		{LogLevel(999), "UNKNOWN"},
		{LogLevel(-99), "UNKNOWN"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.level.String()
			if result != test.expected {
				t.Errorf("Level %d.String() expected %q, got %q", int(test.level), test.expected, result)
			}
		})
	}
}

func TestLogLevel_IsEnabled(t *testing.T) {
	tests := []struct {
		level    LogLevel
		minLevel LogLevel
		expected bool
		name     string
	}{
		{InfoLevel, InfoLevel, true, "same_level"},
		{WarningLevel, InfoLevel, true, "higher_level"},
		{DebugLevel, InfoLevel, false, "lower_level"},

		{TraceLevel, TraceLevel, true, "trace_same"},
		{DebugLevel, TraceLevel, true, "debug_vs_trace"},
		{TraceLevel, DebugLevel, false, "trace_vs_debug"},

		{FatalLevel, ErrorLevel, true, "fatal_vs_error"},
		{PanicLevel, ErrorLevel, true, "panic_vs_error"},
		{ErrorLevel, FatalLevel, false, "error_vs_fatal"},

		{PanicLevel, TraceLevel, true, "panic_vs_trace"},
		{TraceLevel, PanicLevel, false, "trace_vs_panic"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.level.IsEnabled(test.minLevel)
			if result != test.expected {
				t.Errorf("Level %s.IsEnabled(%s) expected %v, got %v",
					test.level.String(), test.minLevel.String(), test.expected, result)
			}
		})
	}
}

func TestLogLevel_ShouldExit(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected bool
	}{
		{TraceLevel, false},
		{DebugLevel, false},
		{InfoLevel, false},
		{WarningLevel, false},
		{ErrorLevel, false},
		{FatalLevel, true},
		{PanicLevel, true},
	}

	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			result := test.level.ShouldExit()
			if result != test.expected {
				t.Errorf("Level %s.ShouldExit() expected %v, got %v",
					test.level.String(), test.expected, result)
			}
		})
	}
}

func TestLogLevel_ShouldPanic(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected bool
	}{
		{TraceLevel, false},
		{DebugLevel, false},
		{InfoLevel, false},
		{WarningLevel, false},
		{ErrorLevel, false},
		{FatalLevel, false},
		{PanicLevel, true},
	}

	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			result := test.level.ShouldPanic()
			if result != test.expected {
				t.Errorf("Level %s.ShouldPanic() expected %v, got %v",
					test.level.String(), test.expected, result)
			}
		})
	}
}

func TestLogLevel_Exit(t *testing.T) {
	tests := []struct {
		level      LogLevel
		shouldCall bool
	}{
		{TraceLevel, false},
		{DebugLevel, false},
		{InfoLevel, false},
		{WarningLevel, false},
		{ErrorLevel, false},
		{FatalLevel, true},
		{PanicLevel, true},
	}

	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			if test.level.ShouldExit() != test.shouldCall {
				t.Errorf("Level %s.ShouldExit() expected %v for Exit() logic",
					test.level.String(), test.shouldCall)
			}
		})
	}
}

func TestLogLevel_Panic(t *testing.T) {
	nonPanicLevels := []LogLevel{
		TraceLevel, DebugLevel, InfoLevel, WarningLevel, ErrorLevel, FatalLevel,
	}

	for _, level := range nonPanicLevels {
		t.Run("no_panic_"+level.String(), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Level %s.Panic() should not panic, but got: %v", level.String(), r)
				}
			}()

			level.Panic("test message")
		})
	}

	t.Run("should_panic_"+PanicLevel.String(), func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("PanicLevel.Panic() should panic, but did not")
			} else {
				if r != "test panic message" {
					t.Errorf("PanicLevel.Panic() expected panic message 'test panic message', got %v", r)
				}
			}
		}()

		PanicLevel.Panic("test panic message")
	})
}

func TestLogLevel_Ordering(t *testing.T) {
	levels := []LogLevel{TraceLevel, DebugLevel, InfoLevel, WarningLevel, ErrorLevel, FatalLevel, PanicLevel}

	for i := 0; i < len(levels)-1; i++ {
		if levels[i] >= levels[i+1] {
			t.Errorf("Log levels should be in ascending order: %s (%d) should be less than %s (%d)",
				levels[i].String(), int(levels[i]), levels[i+1].String(), int(levels[i+1]))
		}
	}
}

func TestLogLevel_Comprehensive(t *testing.T) {
	allLevels := []LogLevel{TraceLevel, DebugLevel, InfoLevel, WarningLevel, ErrorLevel, FatalLevel, PanicLevel}

	for _, currentLevel := range allLevels {
		for _, minLevel := range allLevels {
			expected := currentLevel >= minLevel
			result := currentLevel.IsEnabled(minLevel)

			if result != expected {
				t.Errorf("Level %s.IsEnabled(%s) expected %v, got %v",
					currentLevel.String(), minLevel.String(), expected, result)
			}
		}
	}
}
