package balogan

import (
	"strings"
	"testing"
)

func TestFields_Copy(t *testing.T) {
	original := Fields{
		"key1": "value1",
		"key2": 42,
	}

	copy := original.Copy()

	copy["key3"] = "value3"

	if _, exists := original["key3"]; exists {
		t.Error("Original fields were modified when copy was changed")
	}

	if len(original) != 2 {
		t.Errorf("Expected original to have 2 fields, got %d", len(original))
	}

	if len(copy) != 3 {
		t.Errorf("Expected copy to have 3 fields, got %d", len(copy))
	}
}

func TestFields_With(t *testing.T) {
	original := Fields{
		"key1": "value1",
	}

	new := original.With("key2", "value2")

	if len(original) != 1 {
		t.Error("Original fields were modified")
	}

	if len(new) != 2 {
		t.Error("New fields don't have correct length")
	}

	if new["key2"] != "value2" {
		t.Error("New field was not added correctly")
	}
}

func TestFields_WithFields(t *testing.T) {
	original := Fields{
		"key1": "value1",
	}

	additional := Fields{
		"key2": "value2",
		"key3": "value3",
	}

	merged := original.WithFields(additional)

	if len(original) != 1 {
		t.Error("Original fields were modified")
	}

	if len(merged) != 3 {
		t.Errorf("Expected merged to have 3 fields, got %d", len(merged))
	}

	for k, v := range additional {
		if merged[k] != v {
			t.Errorf("Field %s was not merged correctly", k)
		}
	}
}

func TestJSONFormatter_Format(t *testing.T) {
	formatter := &JSONFormatter{}

	result := formatter.Format(Fields{})
	if result != "" {
		t.Errorf("Expected empty string for empty fields, got %s", result)
	}

	fields := Fields{
		"string": "value",
		"number": 42,
		"bool":   true,
	}

	result = formatter.Format(fields)

	if !strings.Contains(result, `"string":"value"`) {
		t.Error("JSON should contain string field")
	}

	if !strings.Contains(result, `"number":42`) {
		t.Error("JSON should contain number field")
	}

	if !strings.Contains(result, `"bool":true`) {
		t.Error("JSON should contain bool field")
	}
}

func TestKeyValueFormatter_Format(t *testing.T) {
	formatter := &KeyValueFormatter{}

	result := formatter.Format(Fields{})
	if result != "" {
		t.Errorf("Expected empty string for empty fields, got %s", result)
	}

	fields := Fields{
		"key1": "value1",
		"key2": 42,
	}

	result = formatter.Format(fields)

	if !strings.Contains(result, "key1=value1") {
		t.Error("Result should contain key1=value1")
	}

	if !strings.Contains(result, "key2=42") {
		t.Error("Result should contain key2=42")
	}
}

func TestKeyValueFormatter_FormatWithCustomSeparator(t *testing.T) {
	formatter := &KeyValueFormatter{Separator: " | "}

	fields := Fields{
		"key1": "value1",
		"key2": "value2",
	}

	result := formatter.Format(fields)

	if !strings.Contains(result, " | ") {
		t.Error("Result should contain custom separator")
	}
}

func TestLogfmtFormatter_Format(t *testing.T) {
	formatter := &LogfmtFormatter{}

	fields := Fields{
		"simple":      "value",
		"with_space":  "value with space",
		"with_equals": "value=with=equals",
	}

	result := formatter.Format(fields)

	if !strings.Contains(result, "simple=value") {
		t.Error("Simple value should not be quoted")
	}

	if !strings.Contains(result, `with_space="value with space"`) {
		t.Error("Value with space should be quoted")
	}

	if !strings.Contains(result, `with_equals="value=with=equals"`) {
		t.Error("Value with equals should be quoted")
	}
}

func TestLogger_WithJSON(t *testing.T) {
	logger := New(InfoLevel, NewStdOutLogWriter())

	jsonLogger := logger.WithJSON()

	// Check that formatter was changed
	if _, ok := jsonLogger.fieldsFormatter.(*JSONFormatter); !ok {
		t.Error("WithJSON should set JSONFormatter")
	}

	// Check that original logger is unchanged
	if _, ok := logger.fieldsFormatter.(*JSONFormatter); ok {
		t.Error("Original logger should not be modified")
	}
}

func TestLogger_WithLogfmt(t *testing.T) {
	logger := New(InfoLevel, NewStdOutLogWriter())

	logfmtLogger := logger.WithLogfmt()

	// Check that formatter was changed
	if _, ok := logfmtLogger.fieldsFormatter.(*LogfmtFormatter); !ok {
		t.Error("WithLogfmt should set LogfmtFormatter")
	}

	// Check that original logger is unchanged
	if _, ok := logger.fieldsFormatter.(*LogfmtFormatter); ok {
		t.Error("Original logger should not be modified")
	}
}

func TestLogger_WithKeyValue(t *testing.T) {
	logger := New(InfoLevel, NewStdOutLogWriter())

	kvLogger := logger.WithKeyValue()

	// Check that formatter was changed
	if _, ok := kvLogger.fieldsFormatter.(*KeyValueFormatter); !ok {
		t.Error("WithKeyValue should set KeyValueFormatter")
	}
}

func TestLogger_WithKeyValueSeparator(t *testing.T) {
	logger := New(InfoLevel, NewStdOutLogWriter())

	customLogger := logger.WithKeyValueSeparator(" | ")

	// Check that formatter was changed
	formatter, ok := customLogger.fieldsFormatter.(*KeyValueFormatter)
	if !ok {
		t.Error("WithKeyValueSeparator should set KeyValueFormatter")
	}

	if formatter.Separator != " | " {
		t.Errorf("Expected separator ' | ', got '%s'", formatter.Separator)
	}
}

func TestLogger_FormatSwitching(t *testing.T) {
	logger := New(InfoLevel, NewStdOutLogWriter())

	// Start with JSON
	jsonLogger := logger.WithJSON()
	if _, ok := jsonLogger.fieldsFormatter.(*JSONFormatter); !ok {
		t.Error("Should be JSON formatter")
	}

	// Switch to logfmt
	logfmtLogger := jsonLogger.WithLogfmt()
	if _, ok := logfmtLogger.fieldsFormatter.(*LogfmtFormatter); !ok {
		t.Error("Should be logfmt formatter")
	}

	// Switch to key-value
	kvLogger := logfmtLogger.WithKeyValue()
	if _, ok := kvLogger.fieldsFormatter.(*KeyValueFormatter); !ok {
		t.Error("Should be key-value formatter")
	}

	// Original loggers should not be affected
	if _, ok := jsonLogger.fieldsFormatter.(*JSONFormatter); !ok {
		t.Error("Original JSON logger should not be modified")
	}

	if _, ok := logfmtLogger.fieldsFormatter.(*LogfmtFormatter); !ok {
		t.Error("Original logfmt logger should not be modified")
	}
}
