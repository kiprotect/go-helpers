// KIProtect Go-Helpers - Golang Utility Functions
// Copyright (C) 2019-2021  KIProtect GmbH (HRB 208395B) - Germany
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

package forms

import (
	"encoding/json"
	"fmt"
	"github.com/kiprotect/go-helpers/errors"
	"reflect"
	"strings"
)

type Validator interface {
	Validate(input interface{}, values map[string]interface{}) (interface{}, error)
}

type ContextValidator interface {
	ValidateWithContext(input interface{}, values map[string]interface{}, context map[string]interface{}) (interface{}, error)
}

type TransformFunction func(interface{}, map[string]interface{}) (interface{}, error)

// https://stackoverflow.com/questions/35790935/using-reflection-in-go-to-get-the-name-of-a-struct
func GetType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func (f Field) MarshalJSON() ([]byte, error) {
	if serializedField, err := f.Serialize(); err != nil {
		return nil, err
	} else {
		return json.Marshal(serializedField)
	}
}

type Serializable interface {
	Serialize() (map[string]interface{}, error)
}

func SerializeValidators(validators []Validator) ([]*ValidatorDescription, error) {
	descriptions := []*ValidatorDescription{}
	for _, validator := range validators {
		var description *ValidatorDescription
		validatorType := GetType(validator)
		if serializableValidator, ok := validator.(Serializable); ok {
			if config, err := serializableValidator.Serialize(); err != nil {
				return nil, err
			} else {
				description = &ValidatorDescription{
					Type:   validatorType,
					Config: config,
				}
			}
		} else {
			config := map[string]interface{}{}
			if err := Coerce(config, validator); err != nil {
				return nil, fmt.Errorf("error serializing validator %v: %v", validator, err)
			}
			description = &ValidatorDescription{
				Type:   validatorType,
				Config: config,
			}
		}
		descriptions = append(descriptions, description)
	}
	return descriptions, nil

}

func (f *Field) Serialize() (map[string]interface{}, error) {
	if descriptions, err := SerializeValidators(f.Validators); err != nil {
		return nil, err
	} else {
		m := map[string]interface{}{
			"name":       f.Name,
			"validators": descriptions,
		}

		if f.Description != "" {
			m["description"] = f.Description
		}

		if f.Global {
			m["global"] = true
		}
		return m, nil
	}
}

type ValidatorDescriptions []*ValidatorDescription

type Field struct {
	ValidatorDescriptions []*ValidatorDescription `json:"validators"`
	Validators            []Validator             `json:"-"`
	Name                  string                  `json:"name"`
	Global                bool                    `json:"global,omitempty"`
	Description           string                  `json:"description,omitempty"`
	Examples              []FieldExample          `json:"examples,omitempty"`
}

type FieldExample struct {
	Value   any  `json:"value"`
	Invalid bool `json:"invalid"`
}

type Transform struct {
	Field     string              `json:"field"`
	Functions []TransformFunction `json:"-"`
}

type Preprocessor func(map[string]interface{}) map[string]interface{}

type ErrorAdder func(key string, err error)

type FormValidator func(map[string]interface{}, ErrorAdder) error

type FormError struct {
	errors.BaseChainableError
}

func MakeFormError(message, code string, data map[string]interface{}, base error) errors.ChainableError {
	return &FormError{
		BaseChainableError: *errors.MakeError(errors.ExternalError, makeErrorMessage(message, data), code, data, base),
	}
}

func makeErrorMessage(baseMessage string, data map[string]interface{}) string {
	messages := make([]string, 0)
	for key, value := range data {
		var strValue string
		if err, ok := value.(error); ok {
			strValue = err.Error()
		} else if str, ok := value.(string); ok {
			strValue = str
		} else {
			strValue = fmt.Sprint(value)
		}
		messages = append(messages, fmt.Sprintf("%s(%s)", key, strValue))
	}
	return baseMessage + ": " + strings.Join(messages, ", ")
}

type Form struct {
	Name                    string                   `json:"name,omitempty"`
	Strict                  bool                     `json:"strict,omitempty"`
	SanitizeKeys            bool                     `json:"sanitizeKeys,omitempty"`
	Validator               FormValidator            `json:"-"`
	Fields                  []Field                  `json:"fields"`
	Transforms              []Transform              `json:"-"`
	Preprocessor            Preprocessor             `json:"-"`
	PreprocessorDescription *PreprocessorDescription `json:"preprocessor,omitempty"`
	ErrorMsg                string                   `json:"errorMsg,omitempty"`
	Description             string                   `json:"description,omitempty"`
	Examples                []FormExample            `json:"examples,omitempty"`
}

type FormExample struct {
	Value   map[string]any `json:"value"`
	Invalid bool           `json:"invalid"`
}

// this is just a convenience function to avoid importing the "forms" module
func (f *Form) Coerce(target, source interface{}) error {
	return Coerce(target, source)
}

func sanitizeURLValues(input map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for key, value := range input {
		o[strings.ToLower(key)] = value
	}
	return o
}

func (f *Form) ValidateURL(inputs map[string][]string) (map[string]interface{}, error) {
	cInputs := make(map[string]interface{})
	for key, value := range inputs {
		cInputs[key] = value
	}
	return f.Validate(cInputs)
}

func (f *Form) makeError(message string, data map[string]interface{}) error {
	return MakeFormError(message, "FORM-ERROR", data, nil)
}

func (f *Form) ValidateGeneric(inputs interface{}) (map[string]interface{}, error) {
	v := reflect.ValueOf(inputs)
	if v.Kind() != reflect.Map {
		return nil, f.makeError("invalid input type: not a map", nil)
	}
	inputsStringMap := make(map[string]interface{})
	for _, key := range v.MapKeys() {
		value := v.MapIndex(key)
		strKey, ok := key.Interface().(string)
		if !ok {
			return nil, f.makeError("invalid input type", nil)
		}
		inputsStringMap[strKey] = value
	}
	return f.Validate(inputsStringMap)
}

func (f *Form) MakeValidationError(data map[string]interface{}) error {
	return MakeFormError(f.ErrorMessage(), "FORM-ERROR", data, nil)
}

func (f *Form) ErrorMessage() string {
	errorMessage := f.ErrorMsg
	if errorMessage == "" {
		errorMessage = "invalid input data"
	}
	return errorMessage
}

func (f *Form) ValidateWithContext(inputs map[string]interface{}, context map[string]interface{}) (values map[string]interface{}, validationError error) {
	return f.validate(inputs, false, context)
}

func (f *Form) Validate(inputs map[string]interface{}) (values map[string]interface{}, validationError error) {
	return f.validate(inputs, false, nil)
}

func (f *Form) ValidateUpdateWithContext(inputs map[string]interface{}, context map[string]interface{}) (values map[string]interface{}, validationError error) {
	return f.validate(inputs, true, context)
}

func (f *Form) ValidateUpdate(inputs map[string]interface{}) (values map[string]interface{}, validationError error) {
	return f.validate(inputs, true, nil)
}

func (f *Form) validate(inputs map[string]interface{}, update bool, context map[string]interface{}) (values map[string]interface{}, validationError error) {

	errors := make(map[string]interface{})
	values = make(map[string]interface{})
	var sanitizedInput map[string]interface{}
	if f.SanitizeKeys {
		sanitizedInput = sanitizeURLValues(inputs)
	} else {
		sanitizedInput = inputs
	}

	setError := func(key string, err error) {
		if _, ok := err.(*FormError); ok {
			// form errors we include in their structured form
			errors[key] = err
		} else {
			// for normal errors we just include the message
			errors[key] = err.Error()
		}
	}

	var err error
	var value interface{}
	for _, field := range f.Fields {
		keys := []string{field.Name}

		if field.Name == "*" {
			// this is a wildcard field that should be applied to all
			// input values (e.g. useful for global validators)
			keys = make([]string, 0)
			for k, _ := range sanitizedInput {
				keys = append(keys, k)
			}
		}

		for _, key := range keys {

			value = sanitizedInput[key]

			// if no validators are given, we simply copy the raw value
			if len(field.Validators) == 0 {
				values[key] = value
			}

			for _, validator := range field.Validators {

				// we skip empty fields if we're in "update mode"
				if update && value == nil {
					continue
				}

				if contextValidator, ok := validator.(ContextValidator); ok && context != nil {
					value, err = contextValidator.ValidateWithContext(value, values, context)
				} else {
					value, err = validator.Validate(value, values)
				}
				if err != nil {
					setError(key, err)
					break
				}
				if value == nil {
					break //if the value is nil we break out of the processing
				}
				values[key] = value
			}
		}
	}

	if len(errors) == 0 {
		for _, transform := range f.Transforms {
			for _, function := range transform.Functions {
				value, err := function(values[transform.Field], values)
				if err != nil {
					setError(transform.Field, err)
				} else {
					values[transform.Field] = value
				}
				break
			}
		}
	}

	errorMessage := f.ErrorMessage()

	// needed in case the validator raises a form error but does not add
	// any field errors.

	hasError := false
	if f.Validator != nil && len(errors) == 0 {
		// if there's a validator function defined we call it
		if err := f.Validator(values, setError); err != nil {
			errorMessage = err.Error()
			hasError = true
		}
	}

	if f.Strict {
		for k, _ := range sanitizedInput {
			found := false
			for _, field := range f.Fields {
				if field.Name == k {
					found = true
					break
				}
			}
			if !found {
				setError(k, fmt.Errorf("field is unexpected"))
			}
		}
	}

	if len(errors) > 0 || hasError {
		validationError = f.makeError(errorMessage, errors)
	}

	return
}
