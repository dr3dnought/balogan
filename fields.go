package balogan

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"strings"
)

// Fields represents a map of key-value pairs for structured logging.
type Fields map[string]interface{}

// FieldsFormatter defines how fields should be formatted in log output.
type FieldsFormatter interface {
	Format(fields Fields) string
}

// JSONFormatter formats fields as JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) Format(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}

	data, err := json.Marshal(fields)
	if err != nil {
		return fmt.Sprintf(`{"error":"failed to marshal fields: %v"}`, err)
	}

	return string(data)
}

// KeyValueFormatter formats fields as key=value pairs.
type KeyValueFormatter struct {
	Separator string
}

func (f *KeyValueFormatter) Format(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}

	separator := f.Separator
	if separator == "" {
		separator = " "
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(fields))
	for _, k := range keys {
		v := fields[k]
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}

	return strings.Join(pairs, separator)
}

// LogfmtFormatter formats fields in logfmt style (key=value with proper escaping).
type LogfmtFormatter struct{}

func (f *LogfmtFormatter) Format(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(fields))
	for _, k := range keys {
		v := fields[k]
		valueStr := fmt.Sprintf("%v", v)
		if strings.Contains(valueStr, " ") || strings.Contains(valueStr, "=") {
			valueStr = fmt.Sprintf(`"%s"`, strings.ReplaceAll(valueStr, `"`, `\"`))
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, valueStr))
	}

	return strings.Join(pairs, " ")
}

// DefaultFieldsFormatter is the default formatter for fields.
var DefaultFieldsFormatter FieldsFormatter = &KeyValueFormatter{}

// WithFields returns a PrefixBuilderFunc that formats the given fields.
func WithFields(fields Fields) PrefixBuilderFunc {
	return WithFieldsFormatter(fields, DefaultFieldsFormatter)
}

// WithFieldsFormatter returns a PrefixBuilderFunc that formats fields using the specified formatter.
func WithFieldsFormatter(fields Fields, formatter FieldsFormatter) PrefixBuilderFunc {
	return func(args ...any) string {
		return formatter.Format(fields)
	}
}

// WithJSONFields returns a PrefixBuilderFunc that formats fields as JSON.
func WithJSONFields(fields Fields) PrefixBuilderFunc {
	return WithFieldsFormatter(fields, &JSONFormatter{})
}

// WithLogfmtFields returns a PrefixBuilderFunc that formats fields in logfmt style.
func WithLogfmtFields(fields Fields) PrefixBuilderFunc {
	return WithFieldsFormatter(fields, &LogfmtFormatter{})
}

// Copy creates a deep copy of the fields.
func (f Fields) Copy() Fields {
	copy := make(Fields, len(f))
	maps.Copy(copy, f)
	return copy
}

// With returns a new Fields with the given key-value pair added.
func (f Fields) With(key string, value interface{}) Fields {
	copy := f.Copy()
	copy[key] = value
	return copy
}

// WithFields returns a new Fields with all given fields merged.
func (f Fields) WithFields(fields Fields) Fields {
	copy := f.Copy()
	for k, v := range fields {
		copy[k] = v
	}
	return copy
}
