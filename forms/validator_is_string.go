package forms

import (
	"fmt"
)

var IsStringForm = Form{
	Fields: []Field{
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

func MakeIsStringValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isString := &IsString{}
	if params, err := IsStringForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsStringForm.Coerce(isString, params); err != nil {
		return nil, err
	}
	return isString, nil
}

type IsString struct {
	MinLength int `json:"minLength,omitempty" coerce:"convert"`
	MaxLength int `json:"maxLength,omitempty" coerce:"convert"`
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
