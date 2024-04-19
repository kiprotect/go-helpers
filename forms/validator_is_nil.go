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
