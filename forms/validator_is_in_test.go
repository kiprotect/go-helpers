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
	"testing"
)

func TestIsInFromConfig(t *testing.T) {
	config := map[string]interface{}{
		"fields": []map[string]interface{}{
			{
				"name": "example",
				"validators": []map[string]interface{}{
					{
						"type": "IsIn",
						"config": map[string]interface{}{
							"choices": []interface{}{"a", "b", "c"},
						},
					},
				},
			},
		},
	}
	context := &FormDescriptionContext{
		Validators: Validators,
	}
	form, err := FromConfig(config, context)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": "a"}); err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": "b"}); err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": "d"}); err == nil {
		t.Fatalf("expected an error")
	}
}
