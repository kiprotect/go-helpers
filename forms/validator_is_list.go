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
	"reflect"
)

var IsListForm = Form{
	Fields: []Field{
		{
			Name: "validators",
			Validators: []Validator{
				IsOptional{},
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
}

func (f IsList) Serialize() (map[string]interface{}, error) {
	if validators, err := SerializeValidators(f.Validators); err != nil {
		return nil, err
	} else {
		return map[string]interface{}{
			"validators": validators,
		}, nil
	}
}

func MakeIsListValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isList := &IsList{}
	if params, err := IsListForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsListForm.Coerce(isList, params); err != nil {
		return nil, err
	} else {
		validators := []Validator{}
		for _, validatorDescription := range isList.ValidatorDescriptions {
			if validator, err := ValidatorFromDescription(validatorDescription, context); err != nil {
				return nil, err
			} else {
				validators = append(validators, validator)
			}
		}
		isList.Validators = validators
	}
	return isList, nil
}

type IsList struct {
	Validators            []Validator             `json:"-"`
	ValidatorDescriptions []*ValidatorDescription `json:"validators"`
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
