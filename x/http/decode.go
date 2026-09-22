package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// RequestValidation allows request objects to be validated
type RequestValidation interface {
	Validate(*http.Request) error
}

// RequestDecode decodes a model from the http.Request body.
// It closes http.Request body after reading.
func RequestDecode(req *http.Request, v any) error {
	defer req.Body.Close()
	return json.NewDecoder(req.Body).Decode(v)
}

// RequestValidate validates a model for required fields (see the `validate`
// struct tag) on POST/PUT requests, then calls v.Validate.
func RequestValidate(req *http.Request, v RequestValidation) error {
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		if err := requiredFieldsError(v); err != nil {
			return err
		}
	}
	return v.Validate(req)
}

// RequestParse is combination func of RequestDecode and RequestValidate.
func RequestParse(req *http.Request, v RequestValidation) error {
	if err := RequestDecode(req, v); err != nil {
		return err
	}
	return RequestValidate(req, v)
}

// requiredFieldsError reports the first field tagged `validate:"required"`
// (optionally combined with other tag values, separated by ";") whose value
// is its type's zero value.
//
// v must be a non-nil pointer to a struct, or a slice/array of structs or
// struct pointers (nil elements are skipped); any other shape is reported
// as an error rather than causing a panic.
func requiredFieldsError(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("validate: expected a non-nil pointer, got %T", v)
	}
	elem := rv.Elem()

	switch elem.Kind() {
	case reflect.Struct:
		return firstMissingRequiredField(elem)
	case reflect.Slice, reflect.Array:
		for i := 0; i < elem.Len(); i++ {
			item := elem.Index(i)
			if item.Kind() == reflect.Pointer {
				if item.IsNil() {
					continue
				}
				item = item.Elem()
			}
			if item.Kind() != reflect.Struct {
				return fmt.Errorf("validate: expected a struct or *struct element, got %s", item.Kind())
			}
			if err := firstMissingRequiredField(item); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("validate: expected a struct, got %s", elem.Kind())
	}
}

func firstMissingRequiredField(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			// unexported: skip rather than risk a panic on Value.Interface
			continue
		}
		tag, ok := sf.Tag.Lookup("validate")
		if !ok || !hasTagValue(tag, "required") {
			continue
		}
		if v.Field(i).IsZero() {
			return fmt.Errorf("'%s' is required", fieldName(sf))
		}
	}
	return nil
}

func hasTagValue(tag, want string) bool {
	for _, part := range strings.Split(strings.ToLower(tag), ";") {
		if strings.TrimSpace(part) == want {
			return true
		}
	}
	return false
}

func fieldName(sf reflect.StructField) string {
	if jsonTag := sf.Tag.Get("json"); jsonTag != "" {
		if name := strings.TrimSpace(strings.Split(jsonTag, ",")[0]); name != "" && name != "-" {
			return name
		}
	}
	return sf.Name
}
