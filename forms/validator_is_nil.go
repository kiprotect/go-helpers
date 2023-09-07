package forms

import (
	"fmt"
)

var IsNilForm = Form{
	Fields: []Field{
		{
			Name: "allowNull",
			Validators: []Validator{
				IsOptional{Default: true},
				IsBoolean{},
			},
		},
	},
}

func MakeIsNilValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {

	isNil := &IsNil{}

	if params, err := IsNilForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsNilForm.Coerce(isNil, params); err != nil {
		return nil, err
	}

	return isNil, nil
}

type IsNil struct {
	AllowNull bool `json:"allowNull"`
}

func (f IsNil) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {

	if input != nil {

		if f.AllowNull && (input == "" || input == 0) {
			return nil, nil
		}

		return nil, fmt.Errorf("IsNil: expected a nil value, got '%v'", input)
	}

	return nil, nil
}
