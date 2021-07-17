package forms

import (
	"fmt"
)

var OrForm = Form{
	Fields: []Field{
		{
			Name: "options",
			Validators: []Validator{
				IsList{
					Validators: []Validator{
						IsList{
							Validators: []Validator{
								IsStringMap{
									Form: &ValidatorDescriptionForm,
								},
							},
						},
					},
				},
			},
		},
	},
}

func MakeOrValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	or := &Or{}
	if params, err := OrForm.Validate(config); err != nil {
		return nil, err
	} else if err := OrForm.Coerce(or, params); err != nil {
		return nil, err
	} else {
		options := [][]Validator{}
		for _, optionDescription := range or.OptionDescriptions {
			validators := []Validator{}
			for _, validatorDescription := range optionDescription {
				if validator, err := ValidatorFromDescription(validatorDescription, context); err != nil {
					return nil, err
				} else {
					validators = append(validators, validator)
				}
			}
			options = append(options, validators)
		}
		or.Options = options
	}
	return or, nil
}

type Or struct {
	Options            [][]Validator             `json:"-"`
	OptionDescriptions [][]*ValidatorDescription `json:"options"`
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
