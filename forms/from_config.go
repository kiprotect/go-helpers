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

type IsValidConfig struct {
}

func (i IsValidConfig) ValidateWithContext(input interface{}, values map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	return input, nil
}

func (i IsValidConfig) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	return input, nil
}

var ValidatorDescriptionForm = Form{
	Fields: []Field{
		{
			Name: "type",
			Validators: []Validator{
				IsString{},
			},
		},
		{
			Name: "config",
			Validators: []Validator{
				IsOptional{},
				IsStringMap{},
				IsValidConfig{},
			},
		},
	},
}

var FieldExampleForm = Form{
	Fields: []Field{
		{
			Name: "value",
			Validators: []Validator{
				CanBeAnything{},
			},
		},
		{
			Name: "invalid",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
	},
}

var FormExampleForm = Form{
	Fields: []Field{
		{
			Name: "value",
			Validators: []Validator{
				IsStringMap{},
			},
		},
		{
			Name: "invalid",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
	},
}

var FieldForm = Form{
	Fields: []Field{
		{
			Name: "name",
			Validators: []Validator{
				IsString{},
			},
		},
		{
			Name: "examples",
			Validators: []Validator{
				IsOptional{},
				IsList{
					Validators: []Validator{
						IsStringMap{
							Form: &FieldExampleForm,
						},
					},
				},
			},
		},
		{
			Name: "validators",
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
}

var FormValidatorDescriptionForm = Form{}

var PreprocessorDescriptionForm = Form{}

var FormForm = Form{
	Fields: []Field{
		{
			Name: "strict",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "sanitizeKeys",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "name",
			Validators: []Validator{
				IsOptional{},
				IsString{},
			},
		},
		{
			Name: "examples",
			Validators: []Validator{
				IsOptional{},
				IsList{
					Validators: []Validator{
						IsStringMap{
							Form: &FormExampleForm,
						},
					},
				},
			},
		},
		{
			Name: "fields",
			Validators: []Validator{
				IsList{
					Validators: []Validator{
						IsStringMap{
							Form: &FieldForm,
						},
					},
				},
			},
		},
		{
			Name: "validator",
			Validators: []Validator{
				IsOptional{},
				IsStringMap{
					Form: &FormValidatorDescriptionForm,
				},
			},
		},
		{
			Name: "preprocessor",
			Validators: []Validator{
				IsOptional{},
				IsStringMap{
					Form: &PreprocessorDescriptionForm,
				},
			},
		},
		{
			Name: "errorMsg",
			Validators: []Validator{
				IsOptional{},
				IsString{},
			},
		},
	},
}

type ValidatorMaker func(map[string]interface{}, *FormDescriptionContext) (Validator, error)

type ValidatorDescription struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type FormValidatorDescription struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type PreprocessorDescription struct {
	Type string `json:"type"`
}

type FormDescriptionContext struct {
	Validators map[string]ValidatorDefinition
}

func ValidatorFromDescription(config *ValidatorDescription, context *FormDescriptionContext) (Validator, error) {
	if definition, ok := context.Validators[config.Type]; !ok {
		return nil, fmt.Errorf("unknown validator type: '%s'", config.Type)
	} else {
		return definition.Maker(config.Config, context)
	}
}

func (f *Form) Initialize(context *FormDescriptionContext) error {

	fields := []Field{}

	for _, field := range f.Fields {

		validators := []Validator{}

		for _, validatorDescription := range field.ValidatorDescriptions {
			if validator, err := ValidatorFromDescription(validatorDescription, context); err != nil {
				return err
			} else {
				validators = append(validators, validator)
			}
		}

		field.Validators = validators

		fields = append(fields, field)

	}

	f.Fields = fields

	return nil
}

func FromConfig(config map[string]interface{}, context *FormDescriptionContext) (*Form, error) {
	form := &Form{}

	if params, err := FormForm.Validate(config); err != nil {
		return nil, err
	} else if err := FormForm.Coerce(form, params); err != nil {
		return nil, err
	}
	return form, form.Initialize(context)
}
