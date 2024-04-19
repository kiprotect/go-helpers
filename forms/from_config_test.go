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
	"encoding/json"
	"testing"
)

func TestFromConfig(t *testing.T) {
	config := map[string]interface{}{
		"fields": []map[string]interface{}{
			{
				"name": "example",
				"validators": []map[string]interface{}{
					{
						"type": "IsString",
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
	if _, err := form.Validate(map[string]interface{}{"example": 4}); err == nil {
		t.Fatalf("expected an error")
	}
	if params, err := form.Validate(map[string]interface{}{"example": "bar"}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	} else if params["example"] != "bar" {
		t.Fatalf("expected value 'bar'")
	}
}

func TestRoundTrip(t *testing.T) {

	form := &Form{
		Fields: []Field{
			{
				Name: "example",
				Validators: []Validator{
					IsString{},
				},
			},
		},
	}

	context := &FormDescriptionContext{
		Validators: Validators,
	}

	bytes, err := json.Marshal(form)

	if err != nil {
		t.Fatal(err)
	}

	config := map[string]interface{}{}
	if err := json.Unmarshal(bytes, &config); err != nil {
		t.Fatal(err)
	}

	recoveredForm, err := FromConfig(config, context)

	if err != nil {
		t.Fatal(err)
	}
	if _, err := recoveredForm.Validate(map[string]interface{}{"example": 4}); err == nil {
		t.Fatalf("expected an error")
	}
	if params, err := recoveredForm.Validate(map[string]interface{}{"example": "bar"}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	} else if params["example"] != "bar" {
		t.Fatalf("expected value 'bar'")
	}
}
