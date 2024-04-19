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
