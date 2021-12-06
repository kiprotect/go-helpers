package forms

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

var IsBytesForm = Form{
	Fields: []Field{
		{
			Name: "encoding",
			Validators: []Validator{
				IsOptional{Default: "base64"},
				IsIn{Choices: []interface{}{"base64", "base64-url", "hex"}},
			},
		},
		{
			Name: "minLength",
			Validators: []Validator{
				IsOptional{},
				IsInteger{HasMin: true, Min: 0},
			},
		},
		{
			Name: "maxLength",
			Validators: []Validator{
				IsOptional{},
				IsInteger{HasMin: true, Min: 0},
			},
		},
	},
}

func MakeIsBytesValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isBytes := &IsBytes{}
	if params, err := IsBytesForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsBytesForm.Coerce(isBytes, params); err != nil {
		return nil, err
	}
	return isBytes, nil
}

type IsBytes struct {
	Encoding  string `json:"encoding"`
	MinLength int    `json:"minLength" coerce:"convert"`
	MaxLength int    `json:"maxLength" coerce:"convert"`
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
