package forms

import (
	"fmt"
)

var IsBooleanForm = Form{
	Fields: []Field{
		{
			Name: "convert",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
	},
}

func MakeIsBooleanValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isBoolean := &IsBoolean{}
	if params, err := IsBooleanForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsBooleanForm.Coerce(isBoolean, params); err != nil {
		return nil, err
	}
	return isBoolean, nil
}

type IsBoolean struct {
	Convert bool `json:"convert"`
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
