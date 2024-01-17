package forms

import (
	"fmt"
)

var IsNotInForm = Form{
	Fields: []Field{
		{
			Name: "choices",
			Validators: []Validator{
				IsOptional{Default: []any{}},
				IsList{},
			},
		},
	},
}

func MakeIsNotInValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isNotIn := &IsNotIn{}
	if params, err := IsNotInForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsNotInForm.Coerce(isNotIn, params); err != nil {
		return nil, err
	}
	return isNotIn, nil
}

type IsNotIn struct {
	Values []interface{} `json:"values"`
}

func (f IsNotIn) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	for _, v := range f.Values {
		if v == input {
			return nil, fmt.Errorf("illegal value: %v", v)
		}
	}
	return input, nil
}
