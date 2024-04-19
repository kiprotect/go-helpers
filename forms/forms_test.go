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
	"fmt"
	"testing"
)

func testCases(t *testing.T, form Form, testCases []map[string]interface{}, valid bool) {
	for i, testCase := range testCases {
		_, err := form.Validate(testCase)
		if valid && err != nil {
			t.Fatalf("case %d should be valid but raised: %s", i, err)
		} else if !valid && err == nil {
			t.Fatalf("case %d should raise an error but didn't", i)
		}
	}
}

func TestMarshalField(t *testing.T) {
	f := Field{
		Name: "test",
		Validators: []Validator{
			IsOptional{},
		},
	}

	if _, err := f.Serialize(); err != nil {
		t.Fatal(err)
	}

	m, err := json.Marshal(f)

	if err != nil {
		t.Fatal(err)
	}

	var d map[string]interface{}

	if err := json.Unmarshal(m, &d); err != nil {
		t.Fatal(err)
	}

	fmt.Println(d)

	validators, ok := d["validators"].([]interface{})

	if !ok {
		t.Fatalf("validators missing")
	}

	validator, ok := validators[0].(map[string]interface{})

	if !ok {
		t.Fatalf("validator missing")
	}

	if v, ok := validator["type"]; !ok {
		t.Fatalf("type is missing")
	} else if vStr, ok := v.(string); !ok {
		t.Fatalf("type is not a string")
	} else if vStr != "IsOptional" {
		t.Fatalf("expected IsOptional")
	}

}

func TestWildcards(t *testing.T) {
	form := Form{
		Fields: []Field{
			Field{
				Name: "*",
				Validators: []Validator{
					IsString{},
				},
			},
		},
	}
	validTestCases := []map[string]interface{}{
		map[string]interface{}{
			"a": "b",
			"c": "d",
			"e": "f",
		},
	}
	invalidTestCases := []map[string]interface{}{
		map[string]interface{}{
			"a": 1,
			"b": "c",
		},
		map[string]interface{}{
			"a": "b",
			"c": 1,
		},
	}

	testCases(t, form, validTestCases, true)
	testCases(t, form, invalidTestCases, false)

}

type AB struct {
	A string
	B int `coerce:"convert"`
}

func TestStringMapCoerce(t *testing.T) {

	form := Form{
		Fields: []Field{
			Field{
				Name: "map",
				Validators: []Validator{
					IsStringMap{
						Coerce: AB{},
						Form: &Form{
							Fields: []Field{
								Field{
									Name: "a",
									Validators: []Validator{
										IsString{},
									},
								},
								Field{
									Name: "b",
									Validators: []Validator{
										IsInteger{},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	validData := map[string]interface{}{
		"map": map[string]interface{}{"a": "foo", "b": 10},
	}

	if params, err := form.Validate(validData); err != nil {
		t.Fatal(err)
	} else if _, ok := params["map"].(*AB); !ok {
		t.Fatalf("expected an AB struct")
	}

}

func TestIsStringMapWithSubform(t *testing.T) {
	form := Form{
		Fields: []Field{
			Field{
				Name: "map",
				Validators: []Validator{
					IsStringMap{
						Coerce: &AB{},
						Form: &Form{
							Fields: []Field{
								Field{
									Name: "a",
									Validators: []Validator{
										IsString{},
									},
								},
								Field{
									Name: "b",
									Validators: []Validator{
										IsInteger{},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	validTestCases := []map[string]interface{}{
		map[string]interface{}{
			"map": map[string]interface{}{"a": "foo", "b": 10},
		},
	}
	invalidTestCases := []map[string]interface{}{
		map[string]interface{}{
			"map": nil,
		},
		map[string]interface{}{
			"map": map[string]interface{}{"a": 1, "b": "ar"},
		},
	}
	testCases(t, form, validTestCases, true)
	testCases(t, form, invalidTestCases, false)
}

func TestIsList(t *testing.T) {
	form := Form{
		Fields: []Field{
			Field{
				Name: "list",
				Validators: []Validator{
					IsList{},
				},
			},
		},
	}
	validTestCases := []map[string]interface{}{
		map[string]interface{}{
			"list": []string{"a", "b", "c"},
		},
		map[string]interface{}{
			"list": []interface{}{"a", 1, nil},
		},
		map[string]interface{}{
			"list": []int{-4, 1, 344},
		},
		map[string]interface{}{
			"list": []int{},
		},
	}
	invalidTestCases := []map[string]interface{}{
		map[string]interface{}{
			"list": map[string]interface{}{"a": 1, "b": nil},
		},
		map[string]interface{}{
			"list": nil,
		},
		map[string]interface{}{
			"list": "foo",
		},
		map[string]interface{}{
			"list": 3444,
		},
	}
	testCases(t, form, validTestCases, true)
	testCases(t, form, invalidTestCases, false)
}
