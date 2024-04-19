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
	"strings"
)

var IsInForm = Form{
	Fields: []Field{
		{
			Name: "choices",
			Validators: []Validator{
				IsOptional{Default: []any{}},
				IsList{},
			},
		},
	},
}

func MakeIsInValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isIn := &IsIn{}
	if params, err := IsInForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsInForm.Coerce(isIn, params); err != nil {
		return nil, err
	}
	return isIn, nil
}

type IsIn struct {
	Choices []interface{} `json:"choices"`
}

func (f IsIn) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	found := false
	for _, v := range f.Choices {
		if v == input {
			found = true
			break
		}
	}
	if !found {
		choices := make([]string, len(f.Choices))
		for i, choice := range f.Choices {
			choices[i] = fmt.Sprintf("%v", choice)
		}
		return nil, fmt.Errorf("invalid choice, must be one of: %s", strings.Join(choices, ", "))
	}
	return input, nil
}
