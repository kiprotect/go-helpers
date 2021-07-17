package forms

import (
	"fmt"
)

var IsRequiredForm = Form{
	Fields: []Field{},
}

func MakeIsRequiredValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isRequired := &IsRequired{}
	if params, err := IsRequiredForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsRequiredForm.Coerce(isRequired, params); err != nil {
		return nil, err
	}
	return isRequired, nil
}

type IsRequired struct{}

func (f IsRequired) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	if input == nil {
		return nil, fmt.Errorf("is required")
	}
	return input, nil
}
