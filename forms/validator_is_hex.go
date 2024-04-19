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
	"encoding/hex"
	"fmt"
	"strings"
)

var IsHexForm = Form{
	Fields: []Field{
		{
			Name: "convertToBinary",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "strict",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
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

func MakeIsHexValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isHex := &IsHex{}
	if params, err := IsHexForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsHexForm.Coerce(isHex, params); err != nil {
		return nil, err
	}
	return isHex, nil
}

type IsHex struct {
	ConvertToBinary bool `json:"convertToBinary"`
	Strict          bool `json:"strict"`
	MinLength       int  `json:"minLength" coerce:"convert"`
	MaxLength       int  `json:"maxLength" coerce:"convert"`
}

func (f IsHex) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	hexStr, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("not a valid hex string")
	}
	var rawHexStr string
	if !f.Strict {
		rawHexStr = strings.Replace(hexStr, "-", "", -1)
	} else {
		rawHexStr = hexStr
	}
	bStr, err := hex.DecodeString(rawHexStr)
	if err != nil {
		return nil, fmt.Errorf("not a valid hex string")
	}
	if f.MinLength != 0 && len(bStr) < f.MinLength {
		return nil, fmt.Errorf("binary string must be at least %d bytes long", f.MinLength)
	}
	if f.MaxLength != 0 && len(bStr) > f.MaxLength {
		return nil, fmt.Errorf("binary string must be at most %d bytes long", f.MaxLength)
	}
	if f.ConvertToBinary {
		return bStr, nil
	}
	return rawHexStr, nil
}
