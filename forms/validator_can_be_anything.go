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

type CanBeAnything struct {
}

var CanBeAnythingForm = Form{
	Fields: []Field{},
}

func MakeCanBeAnythingValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	return &CanBeAnything{}, nil
}

func (f CanBeAnything) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	return input, nil
}
