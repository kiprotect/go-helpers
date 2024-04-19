// KIProtect Go-Helpers - Golang Utility Functions
// Copyright (C) 2019-2024  KIProtect GmbH (HRB 208395B) - Germany
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
	"fmt"
)

var OrForm = Form{
	Fields: []Field{
		{
			Name: "options",
			Validators: []Validator{
				IsOptional{Default: []map[string]any{}},
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

func (f Or) Serialize() (map[string]interface{}, error) {
	optionDescriptions := make([][]*ValidatorDescription, len(f.Options))
	for i, option := range f.Options {
		if descriptions, err := SerializeValidators(option); err != nil {
			return nil, err
		} else {
			optionDescriptions[i] = descriptions
		}
	}
	return map[string]interface{}{
		"options": optionDescriptions,
	}, nil
}

func MakeOrValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	or := &Or{}
	if params, err := OrForm.Validate(config); err != nil {
		return nil, err
	} else if err := OrForm.Coerce(or, params); err != nil {
		return nil, err
	} else {
		options := [][]Validator{}
		for _, optionDescription := range or.OptionsDescriptions {
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
	OptionsDescriptions [][]*ValidatorDescription `json:"options"`
	Options             [][]Validator             `json:"-"`
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
