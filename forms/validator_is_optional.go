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

var IsOptionalForm = Form{
	Fields: []Field{
		{
			Name: "default",
			Validators: []Validator{
				IsOptional{},
				CanBeAnything{},
			},
		},
	},
}

func MakeIsOptionalValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isOptional := &IsOptional{}
	if params, err := IsOptionalForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsOptionalForm.Coerce(isOptional, params); err != nil {
		return nil, err
	}
	return isOptional, nil
}

type IsOptional struct {
	Default          interface{}        `json:"default,omitempty"`
	DefaultGenerator func() interface{} `json:"-"`
}

func (f IsOptional) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	if input == nil || input == "" {
		//if a default value is defined we return that instead
		if f.Default != nil {
			return f.Default, nil
		} else if f.DefaultGenerator != nil {
			return f.DefaultGenerator(), nil
		}
		return nil, nil
	}
	return input, nil
}
