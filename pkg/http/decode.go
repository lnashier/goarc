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

// RequestValidate validates a model for required fields and valid values
func RequestValidate(req *http.Request, m RequestValidation) error {
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		if err := getRequiredValidationError(m); err != nil {
			return err
		}
	}
	return m.Validate(req)
}

// RequestParse is combination func of RequestDecode and RequestValidate.
func RequestParse(req *http.Request, m RequestValidation) error {
	if err := RequestDecode(req, m); err != nil {
		return err
	}
	return RequestValidate(req, m)
}

func getRequiredValidationError(v any) error {
	getFieldsWithValidationType := func(validationType string, thing any) ([]*reflect.Value, []string) {
		tagName := "validate"

		valueOfT := reflect.ValueOf(thing).Elem()
		valuesOfT := []reflect.Value{valueOfT}
		typeOfT := valueOfT.Type()

		if valueOfT.Kind() == reflect.Slice || valueOfT.Kind() == reflect.Array {
			valuesOfT = make([]reflect.Value, valueOfT.Len())
			for i := 0; i < valueOfT.Len(); i++ {
				valuesOfT[i] = valueOfT.Index(i).Elem()
				typeOfT = valuesOfT[i].Type()
			}
		}

		if len(valuesOfT) == 0 {
			return []*reflect.Value{}, []string{}
		}

		maxFieldsCount := valuesOfT[0].NumField() * len(valuesOfT)
		list := make([]*reflect.Value, maxFieldsCount)
		nameList := make([]string, maxFieldsCount)
		listIndex := 0

		for _, valueOfT := range valuesOfT {
			for i := 0; i < valueOfT.NumField(); i++ {
				tag, ok := typeOfT.Field(i).Tag.Lookup(tagName)
				if ok {
					tagParts := strings.Split(strings.ToLower(tag), ";")

					for _, tagPart := range tagParts {
						if strings.TrimSpace(tagPart) == validationType {
							name := typeOfT.Field(i).Name
							jsonTag := typeOfT.Field(i).Tag.Get("json")
							if jsonTag != "" {
								name = strings.TrimSpace(strings.Split(jsonTag, ",")[0])
							}

							field := valueOfT.Field(i)
							list[listIndex] = &field
							nameList[listIndex] = name
							listIndex++
						}
					}
				}
			}
		}

		return list[0:listIndex], nameList
	}

	fields, fieldNames := getFieldsWithValidationType("required", v)
	for idx, field := range fields {
		if field.Interface() == reflect.Zero(field.Type()).Interface() {
			return fmt.Errorf("'%v' is required", fieldNames[idx])
		}
	}
	return nil
}
