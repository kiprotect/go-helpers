// KIProtect Go-Helpers - Golang Utility Functions
// Copyright (C) 2019-2021  KIProtect GmbH (HRB 208395B) - Germany
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

package settings

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestContext(t *testing.T) {
	includes := map[string]string{
		"/bar/include.yml": "foo: $PATH\nbar: $DIR",
		"/main.yml": `
$include:
  - bar/include.yml
`,
	}
	reader := func(path string) ([]byte, error) {
		include, ok := includes[path]
		t.Log(path)
		if !ok {
			return nil, fmt.Errorf("not found")
		}
		return []byte(include), nil
	}

	settings, err := loadYaml("/main.yml", reader)
	if err != nil {
		t.Fatal(err)
	}
	settingsMap, ok := settings.(map[string]interface{})
	if !ok {
		t.Fatal("expected a map")
	}
	if foo, ok := settingsMap["foo"].(string); !ok || foo != "/bar/include.yml" {
		t.Fatal("PATH context variable not present")
	}
	if bar, ok := settingsMap["bar"].(string); !ok || bar != "/bar" {
		t.Fatal("DIR context variable not present")
	}
}

func TestRefs(t *testing.T) {
	settings := map[string]interface{}{
		"test": map[string]interface{}{
			"$ref": "foo.bar",
		},
		"list": []interface{}{
			map[string]interface{}{
				"$ref": "foo.bar",
			},
		},
		"foo": map[string]interface{}{
			"bar": 10,
		},
	}
	refSettings, err := parseRefs(settings, settings)
	if err != nil {
		t.Fatal(err)
	}
	refSettingsMap, ok := refSettings.(map[string]interface{})
	if !ok {
		t.Fatal("expected a map")
	}
	if i, ok := refSettingsMap["test"].(int); !ok || i != 10 {
		t.Error("ref not parsed")
	}
	if i, ok := refSettingsMap["list"].([]interface{})[0].(int); !ok || i != 10 {
		t.Error("ref not parsed")
	}
}

func TestIncludes(t *testing.T) {
	includes := map[string]string{
		"/test/another-include.yml": "foo: bar",
		"/test/and-another-one.yml": "bar: baz",
		"/test/list-include.yml":    "zoop: zap",
		"/test/map-include.yml":     "bam: bom",
		"/test/bar.yml": `
$include:
  - another-include.yml
  - and-another-one.yml
deep:
  - list
  - $include: list-include.yml
  - map:
      $include: map-include.yml`,
	}
	reader := func(path string) ([]byte, error) {
		include, ok := includes[path]
		t.Log(path)
		if !ok {
			return nil, fmt.Errorf("not found")
		}
		return []byte(include), nil
	}

	settings, err := loadYaml("/test/bar.yml", reader)
	if err != nil {
		t.Fatal(err)
	}
	mapSettings, ok := settings.(map[string]interface{})
	if !ok {
		t.Fatal("expected a settings map")
	}
	if v, ok := mapSettings["foo"].(string); !ok || v != "bar" {
		t.Error("'foo' value missing")
	}
	if v, ok := mapSettings["bar"].(string); !ok || v != "baz" {
		t.Error("'foo' value missing")
	}
	if v, ok := mapSettings["deep"].([]interface{})[1].(map[string]interface{})["zoop"]; !ok || v != "zap" {
		t.Error("'zoop' value missing")
	}
	if v, ok := mapSettings["deep"].([]interface{})[2].(map[string]interface{})["map"].(map[string]interface{})["bam"].(string); !ok || v != "bom" {
		t.Error("'bom' value missing")
	}
}

func TestEnvVariable(t *testing.T) {
	varsStruct := map[string]interface{}{
		"vars": map[string]interface{}{
			"test": map[string]interface{}{
				"type":     "string",
				"source":   "env",
				"variable": "GO_HELPERS_TEST",
			},
		},
	}
	os.Setenv("GO_HELPERS_TEST", "test")
	vars, err := ParseVars(varsStruct)
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := vars["test"].(string); !ok || value != "test" {
		t.Fatalf("env input didn't work")
	}
}

func TestPromptVariable(t *testing.T) {
	varsStruct := map[string]interface{}{
		"vars": map[string]interface{}{
			"test": map[string]interface{}{
				"type":   "string",
				"source": "prompt",
			},
		},
	}
	var stdin bytes.Buffer
	stdin.Write([]byte("test"))
	vars, err := parseVars(varsStruct, &stdin)
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := vars["test"].(string); !ok || value != "test" {
		t.Fatalf("prompt input didn't work")
	}
}

func TestVariableTypes(t *testing.T) {
	varsStruct := map[string]interface{}{
		"vars": map[string]interface{}{
			"intString": map[string]interface{}{
				"type":   "int",
				"source": "literal",
				"value":  "345345",
			},
			"int": map[string]interface{}{
				"type":   "int",
				"source": "literal",
				"value":  345345,
			},
			"float": map[string]interface{}{
				"type":   "float",
				"source": "literal",
				"value":  345345.34435,
			},
			"floatString": map[string]interface{}{
				"type":   "float",
				"source": "literal",
				"value":  "345345.34435",
			},
		},
	}
	vars, err := ParseVars(varsStruct)
	if err != nil {
		t.Fatal(err)
	}
	if intValue, ok := vars["int"].(int); !ok {
		t.Fatal("int not parsed")
	} else if intValue != 345345 {
		t.Fatal("value does not match")
	}
	if intValue, ok := vars["intString"].(int); !ok {
		t.Fatal("int not parsed")
	} else if intValue != 345345 {
		t.Fatal("value does not match")
	}
	if floatValue, ok := vars["float"].(float64); !ok {
		t.Fatal("float not parsed")
	} else if floatValue != 345345.34435 {
		t.Fatal("value does not match")
	}
	if floatValue, ok := vars["floatString"].(float64); !ok {
		t.Fatal("float not parsed")
	} else if floatValue != 345345.34435 {
		t.Fatal("value does not match")
	}
}

func TestVariableParsing(t *testing.T) {
	varsStruct := map[string]interface{}{
		"vars": map[string]interface{}{
			"test": map[string]interface{}{
				"type":   "string",
				"source": "literal",
				"value":  "testing",
			},
			"foo": map[string]interface{}{
				"type":   "string",
				"source": "literal",
				"value":  "hey",
			},
		},
		"foo": map[string]interface{}{
			"bar": "$test",
			"bam": map[string]interface{}{
				"boom": "$test",
			},
		},
		"fooz": []interface{}{
			"$test",
			map[string]interface{}{
				"bar": "$test",
			},
		},
		"bom":  "another $test example $foo",
		"bom2": "$test another example $foo",
		"bum":  "escaped $$test",
		"bam":  "double escaped $$$$test",
	}
	vars, err := ParseVars(varsStruct)
	if err != nil {
		t.Error(err)
	}
	if testValue, ok := vars["test"]; !ok {
		t.Errorf("test variable not found")
	} else if testValue != "testing" {
		t.Errorf("expected 'testing' as value")
	}
	if err := InsertVars(varsStruct, vars); err != nil {
		t.Error(err)
	}
	barValue := varsStruct["foo"].(map[string]interface{})["bar"]
	if barValue != "testing" {
		t.Error("variable not inserted into map")
	}
	barBamBoomValue := varsStruct["foo"].(map[string]interface{})["bam"].(map[string]interface{})["boom"]
	if barBamBoomValue != "testing" {
		t.Error("variable not inserted into nested map")
	}
	foozValue := varsStruct["fooz"].([]interface{})[0]
	if foozValue != "testing" {
		t.Error("variable not inserted into array")
	}
	foozBarValue := varsStruct["fooz"].([]interface{})[1].(map[string]interface{})["bar"]
	if foozBarValue != "testing" {
		t.Error("variable not inserted into map nested in array")
	}
	bomValue := varsStruct["bom"].(string)
	if bomValue != "another testing example hey" {
		t.Log(bomValue)
		t.Error("variable not inserted into string")
	}
	bom2Value := varsStruct["bom2"].(string)
	if bom2Value != "testing another example hey" {
		t.Log(bom2Value)
		t.Error("variable not inserted into string")
	}
	bumValue := varsStruct["bum"].(string)
	if bumValue != "escaped $test" {
		t.Error("escaping didn't work")
	}
	bamValue := varsStruct["bam"].(string)
	if bamValue != "double escaped $$test" {
		t.Error("double-escaping didn't work")
	}
}

func TestListMerging(t *testing.T) {
	a := map[string]interface{}{
		"a": []string{"a", "b", "c"},
	}
	b := map[string]interface{}{
		"a": []string{"d", "e", "f"},
	}

	Merge(a, b)

	resultList, ok := a["a"].([]interface{})

	if !ok {
		t.Fatal("not a list")
	}

	if len(resultList) != 6 {
		t.Fatal("invalid length")
	}
}
