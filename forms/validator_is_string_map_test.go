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

func TestIsStringMapFromConfig(t *testing.T) {
	config := map[string]interface{}{
		"fields": []map[string]interface{}{
			{
				"name": "example",
				"validators": []map[string]interface{}{
					{
						"type": "IsStringMap",
						"config": map[string]interface{}{
							"form": map[string]interface{}{
								"fields": []map[string]interface{}{
									{
										"name": "foo",
										"validators": []map[string]interface{}{
											{
												"type": "IsString",
											},
										},
									},
								},
							},
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
	if _, err := form.Validate(map[string]interface{}{"example": map[string]interface{}{"foo": "bar"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": map[string]interface{}{"foo": 12}}); err == nil {
		t.Fatalf("expected an error")
	}
}
