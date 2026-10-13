package submission

import (
	"encoding/json"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ValidationError is returned when submission data
// does not match the published form schema.
//
// Fields maps a field name to its validation message.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

// formSchema is the part of the form version schema
// used for validation.
type formSchema struct {
	Fields []schemaField `json:"fields"`
}

type schemaField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// ValidateData validates submission data against a form version schema.
//
// It returns a *ValidationError when one or more fields are invalid.
// A schema without fields is not validated.
func ValidateData(schema any, data any) error {

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil
	}

	var parsed formSchema

	if err := json.Unmarshal(schemaJSON, &parsed); err != nil {
		return nil
	}

	if len(parsed.Fields) == 0 {
		return nil
	}

	values, ok := data.(map[string]any)
	if !ok {
		return &ValidationError{
			Fields: map[string]string{
				"data": "must be an object",
			},
		}
	}

	fieldErrors := make(map[string]string)
	known := make(map[string]bool)

	for _, field := range parsed.Fields {

		if field.Name == "" {
			continue
		}

		known[field.Name] = true

		value := values[field.Name]

		if isEmpty(value) {
			if field.Required {
				fieldErrors[field.Name] = "is required"
			}

			continue
		}

		if message := validateType(field.Type, value); message != "" {
			fieldErrors[field.Name] = message
		}
	}

	// Reject fields that are not part of the form.
	for name := range values {
		if !known[name] {
			fieldErrors[name] = "is not a field of this form"
		}
	}

	if len(fieldErrors) > 0 {
		return &ValidationError{
			Fields: fieldErrors,
		}
	}

	return nil
}

func isEmpty(value any) bool {

	switch v := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	}

	return false
}

// validateType returns a validation message,
// or an empty string when the value is valid.
//
// Unknown field types are not type-checked.
func validateType(fieldType string, value any) string {

	switch strings.ToLower(fieldType) {

	case "text", "textarea", "tel", "phone", "password":
		if _, ok := value.(string); !ok {
			return "must be a string"
		}

	case "email":
		text, ok := value.(string)
		if !ok || !isEmail(text) {
			return "must be a valid email address"
		}

	case "number":
		switch v := value.(type) {
		case float64:
		case string:
			if _, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err != nil {
				return "must be a number"
			}
		default:
			return "must be a number"
		}

	case "url":
		text, ok := value.(string)
		if !ok {
			return "must be a valid URL"
		}

		parsed, err := url.Parse(text)
		if err != nil || parsed.Host == "" ||
			(parsed.Scheme != "http" && parsed.Scheme != "https") {
			return "must be a valid URL"
		}

	case "date":
		text, ok := value.(string)
		if !ok {
			return "must be a date in YYYY-MM-DD format"
		}

		if _, err := time.Parse("2006-01-02", text); err != nil {
			return "must be a date in YYYY-MM-DD format"
		}

	case "checkbox", "boolean":
		if _, ok := value.(bool); !ok {
			return "must be true or false"
		}
	}

	return ""
}

func isEmail(value string) bool {

	address, err := mail.ParseAddress(value)

	// ParseAddress also accepts "Name <user@example.com>".
	if err != nil || address.Address != value {
		return false
	}

	domain := value[strings.LastIndex(value, "@")+1:]

	return strings.Contains(domain, ".") &&
		!strings.HasPrefix(domain, ".") &&
		!strings.HasSuffix(domain, ".")
}
