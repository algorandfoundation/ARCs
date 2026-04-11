package arc

import (
	"fmt"
	"strings"
	"time"
)

func IsScalarDateField(key string) bool {
	switch key {
	case "created", "last-call-deadline", "idle-since":
		return true
	default:
		return false
	}
}

func IsStringSequenceField(key string) bool {
	_, ok := canonicalStringSequenceFields[key]
	return ok
}

func IsIntSequenceField(key string) bool {
	_, ok := canonicalIntSequenceFields[key]
	return ok
}

func (document *Document) StringField(key string) string {
	return stringField(document, key)
}

func (document *Document) IntField(key string) (int, bool) {
	return intField(document, key)
}

func (document *Document) BoolField(key string) (bool, bool) {
	return boolField(document, key)
}

func (document *Document) IntSequenceField(key string) []int {
	return intSequenceField(document, key)
}

func (document *Document) StringSequenceField(key string, allowDates bool) []string {
	return stringSequenceField(document, key, allowDates)
}

func NormalizeScalarDateValue(value any) (string, bool) {
	switch typed := value.(type) {
	case time.Time:
		return typed.Format("2006-01-02"), true
	case string:
		trimmed := strings.TrimSpace(typed)
		if datePattern.MatchString(trimmed) {
			return trimmed, true
		}
		if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
			return parsed.Format("2006-01-02"), true
		}
	}
	return "", false
}

func hasField(values map[string]any, key string) bool {
	_, ok := values[key]
	return ok
}

func hasValue(values map[string]any, key string) bool {
	value, ok := values[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	default:
		return true
	}
}

func stringField(document *Document, key string) string {
	value, ok := document.Fields[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case time.Time:
		return typed.Format("2006-01-02")
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func intField(document *Document, key string) (int, bool) {
	value, ok := document.Fields[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	}
	return 0, false
}

func boolField(document *Document, key string) (bool, bool) {
	value, ok := document.Fields[key]
	if !ok {
		return false, false
	}
	typed, ok := value.(bool)
	return typed, ok
}

func intSequenceField(document *Document, key string) []int {
	value, ok := document.Fields[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []any:
		out := make([]int, 0, len(typed))
		for _, item := range typed {
			switch value := item.(type) {
			case int:
				out = append(out, value)
			case int64:
				out = append(out, int(value))
			case float64:
				out = append(out, int(value))
			default:
				return nil
			}
		}
		return out
	default:
		return nil
	}
}

func stringSequenceField(document *Document, key string, allowDates bool) []string {
	value, ok := document.Fields[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			switch value := item.(type) {
			case string:
				trimmed := strings.TrimSpace(value)
				if trimmed == "" {
					return nil
				}
				out = append(out, trimmed)
			case time.Time:
				if !allowDates {
					return nil
				}
				out = append(out, value.Format("2006-01-02"))
			default:
				return nil
			}
		}
		return out
	default:
		return nil
	}
}
