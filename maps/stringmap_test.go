// KIProtect Go-Helpers - Golang Utility Functions
// Copyright (C) 2020  KIProtect GmbH (HRB 208395B) - Germany
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

package maps

import (
	"testing"
)

func TestRecursiveStringMap(t *testing.T) {

	invalidStruct := map[interface{}]interface{}{
		"test": map[interface{}]interface{}{
			1: "test",
		},
	}

	validStruct := map[interface{}]interface{}{
		"test": map[interface{}]interface{}{
			"foo": "test",
		},
		"test2": []interface{}{
			map[interface{}]interface{}{
				"foo": "bar",
			},
		},
		"test3": map[string]interface{}{
			"foo": "bar",
			"deep": []interface{}{
				map[string]interface{}{
					"foo": "bar",
				},
			},
		},
	}

	if _, ok := RecursiveToStringMap(invalidStruct); ok {
		t.Error("should not work")
	}

	if stringMap, ok := RecursiveToStringMap(validStruct); !ok {
		t.Error("should work")
	} else {
		if _, ok := stringMap["test"].(map[string]interface{}); !ok {
			t.Error("should be a string map")
		}
		if deepValue, ok := stringMap["test2"].([]interface{}); !ok {
			t.Error("should be a list")
		} else if _, ok := deepValue[0].(map[string]interface{}); !ok {
			t.Error("should be a string map")
		}
		if deepValue, ok := stringMap["test3"].(map[string]interface{}); !ok {
			t.Error("should be a string map")
		} else if veryDeepValue, ok := deepValue["deep"].([]interface{}); !ok {
			t.Error("should be a list")
		} else if _, ok := veryDeepValue[0].(map[string]interface{}); !ok {
			t.Error("should be a string map")
		}
	}

}
