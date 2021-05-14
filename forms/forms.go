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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/kiprotect/go-helpers/errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Serializable interface {
	Serialize(context map[string]interface{}) interface{}
}

type Validator interface {
	Validate(input interface{}, values map[string]interface{}) (interface{}, error)
}

type ContextValidator interface {
	ValidateWithContext(input interface{}, values map[string]interface{}, context map[string]interface{}) (interface{}, error)
}

type TransformFunction func(interface{}, map[string]interface{}) (interface{}, error)

// https://stackoverflow.com/questions/35790935/using-reflection-in-go-to-get-the-name-of-a-struct
func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

type SerializedValidator struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (f Field) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Serialize(map[string]interface{}{}))
}

func (f Field) Serialize(context map[string]interface{}) interface{} {

	validators := make([]SerializedValidator, 0)

	for _, validator := range f.Validators {
		var data interface{}
		if serializableValidator, ok := validator.(Serializable); ok {
			data = serializableValidator.Serialize(context)
		} else {
			// we include the raw validator
			data = validator
		}
		validators = append(validators, SerializedValidator{
			Type: ToSnakeCase(getType(validator)),
			Data: data,
		})
	}

	return map[string]interface{}{
		"name":       f.Name,
		"validators": validators,
	}
}

type Field struct {
	Validators []Validator `json:"validators"`
	Name       string      `json:"name"`
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
	SanitizeKeys bool          `json:"sanitizeKeys"`
	Validator    FormValidator `json:"-"`
	Fields       []Field       `json:"fields"`
	Transforms   []Transform   `json:"-"`
	Preprocessor Preprocessor  `json:"-"`
	ErrorMsg     string        `json:"-"`
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

func (f *Form) Serialize(context map[string]interface{}) interface{} {
	fields := make([]interface{}, len(f.Fields))
	for i, field := range f.Fields {
		fields[i] = field.Serialize(context)
	}
	return map[string]interface{}{
		"fields": fields,
	}
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

	if len(errors) > 0 || hasError {
		validationError = f.makeError(errorMessage, errors)
	}

	return
}

type IsOptional struct {
	Default          interface{}        `json:"default"`
	DefaultGenerator func() interface{} `json:"-"`
}

type Or struct {
	Options [][]Validator `json:"validators"`
}

type Switch struct {
	Key   string                 `json:"key"`
	Cases map[string][]Validator `json:"cases"`
}

type IsRequired struct{}

type IsString struct {
	MinLength int `json:"min_length"`
	MaxLength int `json:"max_length"`
}

type IsBytes struct {
	Encoding  string `json:"encoding"`
	MinLength int    `json:"min_length"`
	MaxLength int    `json:"max_length"`
}

type IsBoolean struct {
	Convert bool `json:"convert"`
}

type MatchesRegex struct {
	Regex *regexp.Regexp `json:"regexp"`
}

type IsStringList struct {
	Validators []Validator `json:"validators"`
}

type IsTime struct {
	Format string `json:"format"`
	ToUTC  bool   `json:"to_utc"`
	Raw    bool
}

type IsUUID struct {
	ConvertToBinary bool `json:"convert_to_binary"`
}

type IsHex struct {
	ConvertToBinary bool `json:"convert_to_binary"`
	Strict          bool `json:"strict"`
	MinLength       int  `json:"min_length"`
	MaxLength       int  `json:"max_length"`
}

func (f Or) Validate(input interface{}, inputs map[string]interface{}) (interface{}, error) {
	return f.validate(input, inputs, nil)
}

func (f Or) ValidateWithContext(input interface{}, inputs map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	return f.validate(input, inputs, context)
}

func (f Or) validate(input interface{}, inputs map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	for _, option := range f.Options {
		value := input
		var err error
		for _, validator := range option {
			if contextValidator, ok := validator.(ContextValidator); ok && context != nil {
				if value, err = contextValidator.ValidateWithContext(value, inputs, context); err != nil {
					break
				}
			} else if value, err = validator.Validate(value, inputs); err != nil {
				break
			}
		}
		if err == nil {
			return value, nil
		}
	}
	return nil, fmt.Errorf("no possible option worked out")
}

func (f IsBytes) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {

	// we see if the input is already a []byte instance
	if bytes, ok := input.([]byte); ok {
		return bytes, nil
	}

	// if not and no encoding is defined we throw an error
	if f.Encoding == "" {
		return nil, fmt.Errorf("not a byte array and no encoding given")
	}

	// we try to convert the input to a string
	str, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("expected a string")
	}

	var b []byte
	var err error

	// we try to decode the string
	switch f.Encoding {
	case "base64":
		if b, err = base64.StdEncoding.DecodeString(str); err != nil {
			return nil, err
		}
	case "base64-url":
		if b, err = base64.URLEncoding.DecodeString(str); err != nil {
			return nil, err
		}
	case "hex":
		if b, err = hex.DecodeString(str); err != nil {
			return nil, err
		}
	default:
		// no encoding matched
		return nil, fmt.Errorf("invalid encoding: %s", f.Encoding)
	}
	if f.MinLength != 0 && len(b) < f.MinLength {
		return nil, fmt.Errorf("binary array must be at least %d bytes long", f.MinLength)
	}
	if f.MaxLength != 0 && len(b) > f.MaxLength {
		return nil, fmt.Errorf("binary array must be at most %d bytes long", f.MaxLength)
	}
	return b, nil
}

func (f IsHex) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	hexStr, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("not a valid hex string")
	}
	var rawHexStr string
	if !f.Strict {
		rawHexStr = strings.Replace(hexStr, "-", "", -1)
	} else {
		rawHexStr = hexStr
	}
	bStr, err := hex.DecodeString(rawHexStr)
	if err != nil {
		return nil, fmt.Errorf("not a valid hex string")
	}
	if f.MinLength != 0 && len(bStr) < f.MinLength {
		return nil, fmt.Errorf("binary string must be at least %d bytes long", f.MinLength)
	}
	if f.MaxLength != 0 && len(bStr) > f.MaxLength {
		return nil, fmt.Errorf("binary string must be at most %d bytes long", f.MaxLength)
	}
	if f.ConvertToBinary {
		return bStr, nil
	}
	return rawHexStr, nil
}

func (f IsUUID) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	uuidStr, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("not a valid UUID")
	}
	rawUUIDStr := strings.Replace(uuidStr, "-", "", -1)
	bStr, err := hex.DecodeString(rawUUIDStr)
	if err != nil {
		return nil, fmt.Errorf("not a valid UUID")
	}
	if len(bStr) != 16 {
		return nil, fmt.Errorf("not a valid UUID")
	}
	if f.ConvertToBinary {
		return bStr, nil
	}
	return uuidStr, nil
}

type IsStringMap struct {
	Form   *Form `json:"form"`
	Coerce interface{}
}

type IsList struct {
	Validators []Validator `json:"validators"`
}

func (f IsTime) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {

	toNumber := func() (int64, error) {
		var t int64
		if inputFloat, ok := input.(float64); ok {
			t = int64(inputFloat)
		} else if inputInt, ok := input.(int); ok {
			t = int64(inputInt)
		} else if inputInt64, ok := input.(int64); ok {
			t = inputInt64
		} else {
			return 0, fmt.Errorf("not a number")
		}
		return t, nil
	}

	var t time.Time
	var err error
	switch f.Format {
	case "":
		fallthrough
	case "rfc3339":
		inputStr, ok := input.(string)
		if !ok {
			return nil, fmt.Errorf("not a string")
		}
		t, err = time.Parse(time.RFC3339, inputStr)
	case "rfc3339-date":
		inputStr, ok := input.(string)
		if !ok {
			return nil, fmt.Errorf("not a string")
		}
		t, err = time.Parse("2006-01-02", inputStr)
	case "unix":
		if n, err := toNumber(); err != nil {
			return nil, err
		} else {
			if f.Raw {
				return input, nil
			}
			return time.Unix(n, 0), nil
		}
	case "unix-nano":
		var n int64
		if n, err = toNumber(); err == nil {
			if f.Raw {
				return n, nil
			}
			t = time.Unix(n/1e9, n%1e9)
		}
	case "unix-milli":
		var n int64
		if n, err = toNumber(); err == nil {
			if f.Raw {
				return n, nil
			}
			t = time.Unix(n/1e3, (n%1e3)*1e6)
		}
	default:
		return nil, fmt.Errorf("invalid time format: %s", f.Format)
	}
	if err != nil {
		return nil, err
	}
	if f.ToUTC {
		t = t.UTC()
	}
	if f.Raw {
		return input, nil
	}
	return t, nil

}

func (f IsList) ValidateWithContext(input interface{}, values map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	return f.validate(input, values, context)
}

func (f IsList) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	return f.validate(input, values, nil)
}

func (f IsList) validate(input interface{}, values map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	it := reflect.TypeOf(input)
	if it == nil || it.Kind() != reflect.Slice {
		return nil, fmt.Errorf("not a list")
	}
	vt := reflect.ValueOf(input)
	if f.Validators != nil {
		validatedList := make([]interface{}, vt.Len())
		for i := 0; i < vt.Len(); i++ {
			entry := vt.Index(i).Interface()
			for _, validator := range f.Validators {
				var err error

				makeError := func(err error) error {
					return MakeFormError("validation error in list value", "FORM-ERROR", map[string]interface{}{fmt.Sprintf("%d", i): err}, nil)
				}

				if contextValidator, ok := validator.(ContextValidator); ok && context != nil {
					if entry, err = contextValidator.ValidateWithContext(entry, values, context); err != nil {
						return nil, makeError(err)
					}
				} else {
					if entry, err = validator.Validate(entry, values); err != nil {
						return nil, makeError(err)
					}
				}
			}
			validatedList[i] = entry
		}
		return validatedList, nil
	}
	return input, nil
}

func (f IsStringMap) ValidateWithContext(input interface{}, values map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	return f.validate(input, values, context)
}

func (f IsStringMap) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	return f.validate(input, values, nil)
}

func (f IsStringMap) validate(input interface{}, values map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	sm, ok := input.(map[string]interface{})
	if !ok {
		m, ok := input.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("not a map")
		}
		sm = make(map[string]interface{})
		for k, v := range m {
			sk, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("not a string map")
			}
			sm[sk] = v
		}
	}
	// if a forms is defined for the string map we execute it
	if f.Form != nil {
		if context == nil {
			context = map[string]interface{}{"_parent": values}
		} else {
			context["_parent"] = values
		}
		if params, err := f.Form.ValidateWithContext(sm, context); err != nil {
			return nil, err
		} else {
			if f.Coerce != nil {
				target := New(f.Coerce)
				if err := Coerce(target, params); err != nil {
					return nil, err
				} else {
					return target, err
				}
			}
			return params, nil
		}
	}

	return sm, nil
}

type IsFloat struct {
	Convert bool    `json:"convert"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	HasMin  bool    `json:"has_min"`
	HasMax  bool    `json:"has_max"`
}

type IsInteger struct {
	Convert bool  `json:"convert"`
	Min     int64 `json:"min"`
	Max     int64 `json:"max"`
	HasMin  bool  `json:"has_min"`
	HasMax  bool  `json:"has_max"`
}

type IsIn struct {
	Choices []interface{} `json:"choices"`
}

type IsNotIn struct {
	Values []interface{} `json:"values"`
}

type OnlyIf struct {
	Function func(interface{}, map[string]interface{}) bool `json:"-"`
}

func (f IsFloat) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	var iv float64
	switch v := input.(type) {
	case float64:
		iv = v
	case float32:
		iv = float64(v)
	case int:
		iv = float64(v)
	case int64:
		iv = float64(v)
	case string:
		if !f.Convert {
			return nil, fmt.Errorf("not an integer")
		}
		i, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("not an integer")
		}
		iv = i
	default:
		return nil, fmt.Errorf("not an float")
	}
	if f.HasMin && iv < f.Min {
		return nil, fmt.Errorf("value must be larger than or equal %g", f.Min)
	}
	if f.HasMax && iv > f.Max {
		return nil, fmt.Errorf("value must be smaller than or equal %g", f.Max)
	}
	return iv, nil
}

func (f IsInteger) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	var iv int64
	switch v := input.(type) {
	case int64:
		iv = v
	case int:
		iv = int64(v)
	case uint:
		iv = int64(v)
	case float64:
		if float64(int64(v)) != v {
			return nil, fmt.Errorf("not an integer")
		}
		iv = int64(v)
	case string:
		if !f.Convert {
			return nil, fmt.Errorf("not an integer")
		}
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("not an integer")
		}
		iv = i
	default:
		return nil, fmt.Errorf("not an integer")
	}
	if f.HasMin && iv < f.Min {
		return nil, fmt.Errorf("value must be larger than or equal %d", f.Min)
	}
	if f.HasMax && iv > f.Max {
		return nil, fmt.Errorf("value must be smaller than or equal %d", f.Max)
	}
	return iv, nil
}

func (f IsStringList) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	strList := make([]string, 0)
	switch l := input.(type) {
	case []string:
		strList = l
		break
	case []interface{}:
		for _, v := range l {
			strV, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("not a string")
			}
			strList = append(strList, strV)
		}
	}
	for _, validator := range f.Validators {
		for i, v := range strList {
			res, err := validator.Validate(v, values)
			if err != nil {
				return nil, err
			}
			strRes, ok := res.(string)
			if !ok {
				return nil, fmt.Errorf("validator result is not a string")
			}
			strList[i] = strRes
		}
	}
	return strList, nil
}

func (f Switch) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	strValue, ok := values[f.Key].(string)
	if !ok {
		return nil, fmt.Errorf("switch key is not a string")
	}
	caseValue, ok := f.Cases[strValue]
	if !ok {
		// we check if a default value is defined
		caseValue, ok = f.Cases["default!"]
		if !ok {
			// no default defined either
			return input, nil
		}
	}
	var err error
	for _, validator := range caseValue {
		input, err = validator.Validate(input, values)
		if err != nil {
			return nil, err
		}
	}
	return input, nil
}

func (f OnlyIf) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	if f.Function(input, values) == true {
		return input, nil
	}
	return nil, nil
}

func (f IsOptional) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	if input == nil || input == "" {
		//if a default value is defined we return that instead
		if f.Default != nil {
			return f.Default, nil
		} else if f.DefaultGenerator != nil {
			return f.DefaultGenerator(), nil
		}
		return nil, nil
	}
	return input, nil
}

func (f IsRequired) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	if input == nil {
		return nil, fmt.Errorf("is required")
	}
	return input, nil
}

func (f IsString) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	str, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("expected a string")
	}
	if f.MinLength > 0 && len(str) < f.MinLength {
		return nil, fmt.Errorf("must be at least %d characters long", f.MinLength)
	}
	if f.MaxLength > 0 && len(str) > f.MaxLength {
		return nil, fmt.Errorf("must be at most %d characters long", f.MaxLength)
	}
	return str, nil
}

func (f IsBoolean) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	b, ok := input.(bool)
	if !ok {
		if f.Convert {
			s, ok := input.(string)
			if ok {
				if s == "true" {
					return true, nil
				} else if s == "false" {
					return false, nil
				}
			}
		}
		return nil, fmt.Errorf("expected a boolean")
	}
	return b, nil
}

func (f MatchesRegex) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	value, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("expected a string")
	}
	if matched := f.Regex.Match([]byte(value)); !matched {
		return nil, fmt.Errorf("regex '%s' did not match", f.Regex.String())
	}
	return value, nil
}

func (f IsIn) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	found := false
	for _, v := range f.Choices {
		if v == input {
			found = true
			break
		}
	}
	if !found {
		choices := make([]string, len(f.Choices))
		for i, choice := range f.Choices {
			choices[i] = fmt.Sprintf("%v", choice)
		}
		return nil, fmt.Errorf("invalid choice, must be one of: %s", strings.Join(choices, ", "))
	}
	return input, nil
}

func (f IsNotIn) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	for _, v := range f.Values {
		if v == input {
			return nil, fmt.Errorf("illegal value: %v", v)
		}
	}
	return input, nil
}
