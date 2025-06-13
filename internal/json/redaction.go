// Package json - Redaction Integration for News Content
// This module provides content redaction capabilities that integrate seamlessly
// with the existing Sonic JSON adapter for high-performance JSON processing.
package json

import (
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// RedactionConfig holds the configuration for content redaction
type RedactionConfig struct {
	Enabled           bool                      `json:"enabled"`
	Patterns          map[string]*regexp.Regexp `json:"-"` // Compiled patterns
	SensitiveFields   map[string]bool           `json:"sensitive_fields"`
	FieldRules        map[string]RedactionRule  `json:"field_rules"`
	DefaultMask       string                    `json:"default_mask"`
	PreservePrefixLen int                       `json:"preserve_prefix_len"`
	PreserveSuffixLen int                       `json:"preserve_suffix_len"`
}

// RedactionRule defines how a specific field should be redacted
type RedactionRule struct {
	Pattern           string `json:"pattern"`
	Replacement       string `json:"replacement"`
	PreservePrefixLen int    `json:"preserve_prefix_len"`
	PreserveSuffixLen int    `json:"preserve_suffix_len"`
	Enabled           bool   `json:"enabled"`
}

// Global redaction configuration
var (
	redactionConfig = getDefaultRedactionConfig()
)

// getRedactionEnabledFromEnv checks if redaction is enabled via environment variable
func getRedactionEnabledFromEnv() bool {
	value := os.Getenv("NEWS_REDACTION_ENABLED")
	if value == "" {
		return false // Disabled by default for performance
	}
	enabled, _ := strconv.ParseBool(value)
	return enabled
}

// getDefaultRedactionConfig creates the default redaction configuration
func getDefaultRedactionConfig() *RedactionConfig {
	config := &RedactionConfig{
		Enabled:           true, // Always enable in config, runtime check via environment
		Patterns:          make(map[string]*regexp.Regexp),
		SensitiveFields:   make(map[string]bool),
		FieldRules:        make(map[string]RedactionRule),
		DefaultMask:       "***",
		PreservePrefixLen: 2,
		PreserveSuffixLen: 2,
	}

	// Define sensitive field patterns for news articles
	patterns := map[string]string{
		"email":       `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
		"phone":       `(\+?1[-.\s]?)?\(?[2-9][0-8][0-9]\)?[-.\s]?[2-9][0-9]{2}[-.\s]?[0-9]{4}`,
		"ssn":         `\b\d{3}-?\d{2}-?\d{4}\b`,
		"credit_card": `\b(?:\d{4}[-\s]?){3}\d{4}\b`,
		"ip_address":  `\b(?:\d{1,3}\.){3}\d{1,3}\b`,
		"url":         `https?://[^\s]+`,
	}

	// Compile patterns
	for name, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			config.Patterns[name] = compiled
		}
	}

	// Define sensitive fields commonly found in news content
	config.SensitiveFields = map[string]bool{
		"content":          true,
		"summary":          true,
		"title":            false, // Usually safe
		"meta_description": true,
		"email":            true,  // User email fields
		"author_email":     true,  // Legacy field name
		"source_url":       false, // URLs might be intentionally public
	}

	// Define field-specific rules
	config.FieldRules = map[string]RedactionRule{
		"content": {
			Pattern:           "email|phone|ssn|credit_card",
			Replacement:       "[REDACTED]",
			PreservePrefixLen: 0,
			PreserveSuffixLen: 0,
			Enabled:           true,
		},
		"summary": {
			Pattern:           "email|phone|ssn",
			Replacement:       "[SENSITIVE INFO REMOVED]",
			PreservePrefixLen: 1,
			PreserveSuffixLen: 1,
			Enabled:           true,
		},
		"email": {
			Pattern:           "email",
			Replacement:       "[EMAIL PROTECTED]",
			PreservePrefixLen: 2,
			PreserveSuffixLen: 0,
			Enabled:           true,
		},
		"author_email": {
			Pattern:           "email",
			Replacement:       "[EMAIL PROTECTED]",
			PreservePrefixLen: 2,
			PreserveSuffixLen: 0,
			Enabled:           true,
		},
	}

	return config
}

// SetRedactionEnabled enables or disables redaction globally
func SetRedactionEnabled(enabled bool) {
	redactionConfig.Enabled = enabled
}

// IsRedactionEnabled returns whether redaction is currently enabled
func IsRedactionEnabled() bool {
	// Check environment variable at runtime for flexibility
	envEnabled := getRedactionEnabledFromEnv()
	return envEnabled && redactionConfig.Enabled
}

// UpdateRedactionConfig allows updating the redaction configuration
func UpdateRedactionConfig(config *RedactionConfig) {
	redactionConfig = config
}

// GetRedactionConfig returns the current redaction configuration
func GetRedactionConfig() *RedactionConfig {
	return redactionConfig
}

// redactValue performs redaction on a single value based on field name and type
func redactValue(fieldName string, value interface{}) interface{} {
	if !IsRedactionEnabled() {
		return value
	}

	// Only process string values
	strValue, ok := value.(string)
	if !ok {
		return value
	}

	// Check if this field should be redacted
	if shouldRedact, exists := redactionConfig.SensitiveFields[strings.ToLower(fieldName)]; !exists || !shouldRedact {
		return value
	}

	// Apply field-specific rules if they exist
	if rule, exists := redactionConfig.FieldRules[strings.ToLower(fieldName)]; exists && rule.Enabled {
		return applyRedactionRule(strValue, rule)
	}

	// Apply default redaction patterns
	return applyDefaultRedaction(strValue)
}

// applyRedactionRule applies a specific redaction rule to a string
func applyRedactionRule(text string, rule RedactionRule) string {
	result := text

	// Apply patterns based on rule
	for patternName, pattern := range redactionConfig.Patterns {
		if strings.Contains(rule.Pattern, patternName) {
			result = pattern.ReplaceAllStringFunc(result, func(match string) string {
				return createMaskedReplacement(match, rule.Replacement, rule.PreservePrefixLen, rule.PreserveSuffixLen)
			})
		}
	}

	return result
}

// applyDefaultRedaction applies default redaction patterns to a string
func applyDefaultRedaction(text string) string {
	result := text

	// Apply all available patterns with default settings
	for _, pattern := range redactionConfig.Patterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return createMaskedReplacement(match, redactionConfig.DefaultMask, redactionConfig.PreservePrefixLen, redactionConfig.PreserveSuffixLen)
		})
	}

	return result
}

// createMaskedReplacement creates a masked version of sensitive data
func createMaskedReplacement(original, replacement string, prefixLen, suffixLen int) string {
	if len(original) <= prefixLen+suffixLen {
		return replacement
	}

	if prefixLen == 0 && suffixLen == 0 {
		return replacement
	}

	prefix := ""
	suffix := ""

	if prefixLen > 0 && prefixLen < len(original) {
		prefix = original[:prefixLen]
	}

	if suffixLen > 0 && suffixLen < len(original) {
		suffix = original[len(original)-suffixLen:]
	}

	return prefix + replacement + suffix
}

// redactStruct performs deep redaction on struct fields
func redactStruct(v reflect.Value) reflect.Value {
	if !IsRedactionEnabled() {
		return v
	}

	if v.Kind() != reflect.Struct {
		return v
	}

	// Create a copy of the struct to avoid modifying the original
	newValue := reflect.New(v.Type()).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)
		fieldName := getFieldName(fieldType)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		var newFieldValue reflect.Value

		switch field.Kind() {
		case reflect.String:
			// Redact string fields
			redacted := redactValue(fieldName, field.String())
			newFieldValue = reflect.ValueOf(redacted)

		case reflect.Slice:
			// Handle slices (like tags, categories, articles)
			newSlice := reflect.MakeSlice(field.Type(), field.Len(), field.Cap())
			for j := 0; j < field.Len(); j++ {
				item := field.Index(j)
				if item.Kind() == reflect.Struct {
					newSlice.Index(j).Set(redactStruct(item))
				} else if item.Kind() == reflect.Ptr && !item.IsNil() && item.Elem().Kind() == reflect.Struct {
					// Handle pointers to structs
					redactedStruct := redactStruct(item.Elem())
					newPtr := reflect.New(item.Elem().Type())
					newPtr.Elem().Set(redactedStruct)
					newSlice.Index(j).Set(newPtr)
				} else {
					newSlice.Index(j).Set(item)
				}
			}
			newFieldValue = newSlice

		case reflect.Struct:
			// Recursively handle nested structs
			newFieldValue = redactStruct(field)

		case reflect.Ptr:
			// Handle pointers
			if !field.IsNil() {
				pointedValue := field.Elem()
				if pointedValue.Kind() == reflect.Struct {
					redactedStruct := redactStruct(pointedValue)
					newPtr := reflect.New(pointedValue.Type())
					newPtr.Elem().Set(redactedStruct)
					newFieldValue = newPtr
				} else {
					newFieldValue = field
				}
			} else {
				newFieldValue = field
			}

		case reflect.Interface:
			// Handle interface{} fields (like PaginatedResponse.Data)
			if !field.IsNil() {
				interfaceValue := field.Elem()
				redactedInterface := applyRedaction(interfaceValue.Interface())
				newFieldValue = reflect.ValueOf(redactedInterface)
			} else {
				newFieldValue = field
			}

		default:
			// For other types, just copy the value
			newFieldValue = field
		}

		if newFieldValue.IsValid() && newValue.Field(i).CanSet() {
			newValue.Field(i).Set(newFieldValue)
		}
	}

	return newValue
}

// getFieldName extracts the field name from struct field, checking JSON tags
func getFieldName(field reflect.StructField) string {
	// Check for JSON tag first
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		name := strings.Split(jsonTag, ",")[0]
		if name != "-" && name != "" {
			return name
		}
	}
	// Fall back to field name
	return strings.ToLower(field.Name)
}

// MarshalWithRedaction performs JSON marshaling with redaction applied
func MarshalWithRedaction(v interface{}) ([]byte, error) {
	if !IsRedactionEnabled() {
		return Marshal(v)
	}

	// Apply redaction to the data structure
	redactedValue := applyRedaction(v)

	// Use the normal marshal function on the redacted data
	return Marshal(redactedValue)
}

// MarshalForCacheWithRedaction performs cache-optimized marshaling with redaction
func MarshalForCacheWithRedaction(v interface{}) ([]byte, error) {
	if !IsRedactionEnabled() {
		return MarshalForCache(v)
	}

	// Apply redaction to the data structure
	redactedValue := applyRedaction(v)

	// Use the cache marshal function on the redacted data
	return MarshalForCache(redactedValue)
}

// applyRedaction applies redaction to any data structure
func applyRedaction(v interface{}) interface{} {
	if !IsRedactionEnabled() {
		return v
	}

	value := reflect.ValueOf(v)
	if !value.IsValid() {
		return v
	}

	switch value.Kind() {
	case reflect.Struct:
		redactedStruct := redactStruct(value)
		return redactedStruct.Interface()

	case reflect.Ptr:
		if value.IsNil() {
			return v
		}
		pointedValue := value.Elem()
		if pointedValue.Kind() == reflect.Struct {
			redactedStruct := redactStruct(pointedValue)
			// Create new pointer to the redacted struct
			newPtr := reflect.New(pointedValue.Type())
			newPtr.Elem().Set(redactedStruct)
			return newPtr.Interface()
		}
		return v

	case reflect.Slice:
		// Handle slices of structs
		newSlice := reflect.MakeSlice(value.Type(), value.Len(), value.Cap())
		for i := 0; i < value.Len(); i++ {
			item := value.Index(i)
			redactedItem := applyRedaction(item.Interface())
			newSlice.Index(i).Set(reflect.ValueOf(redactedItem))
		}
		return newSlice.Interface()

	default:
		return v
	}
}

// RedactionStats provides statistics about redaction operations
type RedactionStats struct {
	Enabled       bool              `json:"enabled"`
	PatternsCount int               `json:"patterns_count"`
	FieldsCount   int               `json:"sensitive_fields_count"`
	RulesCount    int               `json:"field_rules_count"`
	Patterns      map[string]string `json:"available_patterns"`
}

// GetRedactionStats returns current redaction statistics and configuration
func GetRedactionStats() RedactionStats {
	patterns := make(map[string]string)
	for name := range redactionConfig.Patterns {
		patterns[name] = "configured"
	}

	return RedactionStats{
		Enabled:       IsRedactionEnabled(),
		PatternsCount: len(redactionConfig.Patterns),
		FieldsCount:   len(redactionConfig.SensitiveFields),
		RulesCount:    len(redactionConfig.FieldRules),
		Patterns:      patterns,
	}
}
