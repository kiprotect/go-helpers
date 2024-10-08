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
	"strconv"
)

var IsFloatForm = Form{
	Fields: []Field{
		{
			Name: "convert",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "hasMin",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "hasMax",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "min",
			Validators: []Validator{
				IsOptional{},
				IsFloat{HasMin: true, Min: 0},
			},
		},
		{
			Name: "max",
			Validators: []Validator{
				IsOptional{},
				IsFloat{HasMin: true, Min: 0},
			},
		},
	},
}

func MakeIsFloatValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isFloat := &IsFloat{}
	if params, err := IsFloatForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsFloatForm.Coerce(isFloat, params); err != nil {
		return nil, err
	}
	return isFloat, nil
}

type IsFloat struct {
	Convert bool    `json:"convert,omitempty"`
	Min     float64 `json:"min,omitempty" coerce:"convert"`
	Max     float64 `json:"max,omitempty" coerce:"convert"`
	HasMin  bool    `json:"hasMin,omitempty"`
	HasMax  bool    `json:"hasMax,omitempty"`
}

func (f IsFloat) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	var iv float64
	switch v := input.(type) {
	case float64:
		iv = v
	case float32:
		iv = float64(v)
	case int:
		iv = float64(v)
	case int64:
		iv = float64(v)
	case string:
		if !f.Convert {
			return nil, fmt.Errorf("not an integer")
		}
		i, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("not an integer")
		}
		iv = i
	default:
		return nil, fmt.Errorf("not an float")
	}
	if f.HasMin && iv < f.Min {
		return nil, fmt.Errorf("value must be larger than or equal %g", f.Min)
	}
	if f.HasMax && iv > f.Max {
		return nil, fmt.Errorf("value must be smaller than or equal %g", f.Max)
	}
	return iv, nil
}
