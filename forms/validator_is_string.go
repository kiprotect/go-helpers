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
		return nil, fmt.Errorf("IsString: expected a string")
	}
	if f.MinLength > 0 && len(str) < f.MinLength {
		return nil, fmt.Errorf("must be at least %d characters long", f.MinLength)
	}
	if f.MaxLength > 0 && len(str) > f.MaxLength {
		return nil, fmt.Errorf("must be at most %d characters long", f.MaxLength)
	}
	return str, nil
}
