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

package settings

import (
	"bufio"
	"fmt"
	"github.com/kiprotect/go-helpers/forms"
	"github.com/kiprotect/go-helpers/maps"
	"github.com/kiprotect/go-helpers/yaml"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Settings struct {
	Values map[string]interface{}
}

type SettingsError struct {
	msg string
}

func (self *SettingsError) Error() string {
	return self.msg
}

func (self *Settings) Get(key string) (interface{}, error) {
	elements := strings.Split(key, ".")
	currentMap := self.Values
	for i := range elements {
		currentValue, ok := currentMap[elements[i]]
		if !ok {
			return nil, &SettingsError{fmt.Sprintf("Key '%v' not found!", key)}
		}
		if i == len(elements)-1 {
			return currentValue, nil
		}
		mapValue, ok := currentValue.(map[string]interface{})
		if !ok {
			return nil, &SettingsError{fmt.Sprintf("Key '%v' not found!", key)}
		}
		currentMap = mapValue
	}
	return currentMap, nil
}

func (self *Settings) String(key string) (string, bool) {
	value, err := self.Get(key)
	if err != nil {
		return "", false
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", false
	}
	return stringValue, true

}

func (self *Settings) Update(values map[string]interface{}) {
	if self.Values == nil {
		self.Values = make(map[string]interface{})
	}
	Merge(self.Values, values)
}

func (self *Settings) Int(key string) (int, bool) {
	value, err := self.Get(key)
	if err != nil {
		return 0, false
	}
	intValue, ok := value.(int)
	if !ok {
		return 0, false
	}
	return intValue, true

}

func (self *Settings) Bool(key string) (bool, bool) {
	value, err := self.Get(key)
	if err != nil {
		return false, false
	}
	boolValue, ok := value.(bool)
	if !ok {
		return false, false
	}
	return boolValue, true

}

func (self *Settings) Set(key string, value interface{}) {
	self.Values[key] = value
}

func mergeMaps(a, b interface{}) (interface{}, bool) {
	aMapValue, aMapOk := a.(map[string]interface{})
	bMapValue, bMapOk := b.(map[string]interface{})

	if aMapOk && bMapOk {
		Merge(aMapValue, bMapValue)
		return aMapValue, true
	}

	return nil, false
}

func mergeSlices(a, b interface{}) (interface{}, bool) {

	rtA := reflect.TypeOf(a)
	rtB := reflect.TypeOf(b)

	if rtA.Kind() != reflect.Slice || rtB.Kind() != reflect.Slice {
		return nil, false
	}

	return mergeLists(a, b), true

}

func mergeArrays(a, b interface{}) (interface{}, bool) {

	rtA := reflect.TypeOf(a)
	rtB := reflect.TypeOf(b)

	if rtA.Kind() != reflect.Array || rtB.Kind() != reflect.Array {
		return nil, false
	}

	return mergeLists(a, b), true

}

func mergeLists(a, b interface{}) interface{} {
	vA := reflect.ValueOf(a)
	vB := reflect.ValueOf(b)

	c := make([]interface{}, vA.Len()+vB.Len())

	for i := 0; i < vA.Len(); i++ {
		c[i] = vA.Index(i).Interface()
	}

	for i := 0; i < vB.Len(); i++ {
		c[i+vA.Len()] = vB.Index(i).Interface()
	}

	return c
}

//Merges two maps in place, such that entries in a will be recursively updated/created using
//entries from b.
func Merge(a map[string]interface{}, b map[string]interface{}) {

	for key, value := range b {
		aValue, aOk := a[key]
		if !aOk {
			// this key does not yet exist in a, so we just add it
			a[key] = value
			continue
		}

		if v, ok := mergeMaps(aValue, value); ok {
			a[key] = v
			continue
		}

		if v, ok := mergeSlices(aValue, value); ok {
			a[key] = v
			continue
		}

		if v, ok := mergeArrays(aValue, value); ok {
			a[key] = v
			continue
		}

		// otherwise we just assign the value to a and we're done
		a[key] = value
	}
}

func (self *Settings) getSettingsFiles(settingsPath string) []string {
	paths := make([]string, 0)
	files, err := ioutil.ReadDir(settingsPath)
	if err != nil {
		return paths
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		r, err := regexp.MatchString(".yml", file.Name())
		if err == nil && r {
			paths = append(paths, path.Join(settingsPath, file.Name()))
		}
	}
	return paths
}

func (self *Settings) Load(settingsPath string) error {
	fi, err := os.Stat(settingsPath)
	if err != nil {
		return err
	}
	var settingsFiles []string
	if fi.Mode().IsDir() {
		settingsFiles = self.getSettingsFiles(settingsPath)
	} else {
		settingsFiles = []string{settingsPath}
	}
	if self.Values == nil {
		self.Values = make(map[string]interface{})
	}
	for _, settingsFile := range settingsFiles {
		log.Debugf("Adding settings from %v...", settingsFile)
		settings, err := LoadYaml(settingsFile)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("load yaml %v: %v", settingsFile, err)
		}
		Merge(self.Values, settings)
	}
	if values, err := ParseVars(self.Values); err != nil {
		return err
	} else {
		return InsertVars(self.Values, values)
	}
}

var VarsForm = forms.Form{
	Fields: []forms.Field{
		forms.Field{
			Name: "type",
			Validators: []forms.Validator{
				forms.IsRequired{},
				forms.IsIn{Choices: []interface{}{"string", "int", "float", "any"}},
			},
		},
		forms.Field{
			Name: "source",
			Validators: []forms.Validator{
				forms.IsRequired{},
				forms.IsIn{Choices: []interface{}{"prompt", "env", "literal"}},
			},
		},
		forms.Field{
			Name: "config",
			Validators: []forms.Validator{
				forms.IsOptional{Default: map[string]interface{}{}},
				forms.IsStringMap{},
			},
		},
	},
}

var PromptForm = forms.Form{
	Fields: []forms.Field{
		forms.Field{
			Name: "sensitive",
			Validators: []forms.Validator{
				forms.IsOptional{Default: false},
				forms.IsBoolean{},
			},
		},
	},
}

var EnvForm = forms.Form{
	Fields: []forms.Field{
		forms.Field{
			Name: "variable",
			Validators: []forms.Validator{
				forms.IsRequired{},
				forms.IsString{},
			},
		},
	},
}

var LiteralForm = forms.Form{
	Fields: []forms.Field{
		forms.Field{
			Name: "value",
			Validators: []forms.Validator{
				forms.IsRequired{},
			},
		},
	},
}

func ParseVars(settings map[string]interface{}) (map[string]interface{}, error) {
	return parseVars(settings, os.Stdin)
}

func parseVars(settings map[string]interface{}, reader io.Reader) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	varsObj, ok := settings["vars"]
	if !ok {
		// no variables defined
		return values, nil
	}
	vars, ok := maps.ToStringMap(varsObj)
	if !ok {
		return nil, fmt.Errorf("invalid variables format")
	}
	for key, configObj := range vars {
		config, ok := maps.ToStringMap(configObj)
		if !ok {
			return nil, fmt.Errorf("not a map")
		}
		params, err := VarsForm.Validate(config)
		if err != nil {
			return nil, err
		}
		var value interface{}
		switch params["source"].(string) {
		case "prompt":
			promptParams, err := PromptForm.Validate(config)
			if err != nil {
				return nil, err
			}
			fmt.Printf("Please provide a value for variable '%s': ", key)
			if promptParams["sensitive"].(bool) {
				bytesValue, err := terminal.ReadPassword(0)
				if err != nil {
					return nil, err
				}
				value = string(bytesValue)
				// since nothing is echoed we input a newline ourselves
				fmt.Printf("\n")
			} else {
				scanner := bufio.NewScanner(reader)
				if ok := scanner.Scan(); ok {
					value = scanner.Text()
				} else {
					return nil, fmt.Errorf("cannot read from stdin")
				}
				if scanner.Err() != nil {
					return nil, err
				}
			}
		case "env":
			envParams, err := EnvForm.Validate(config)
			if err != nil {
				return nil, err
			}
			variable := envParams["variable"].(string)
			if envValue, ok := os.LookupEnv(variable); !ok {
				return nil, fmt.Errorf("environment variable '%s' is undefind", variable)
			} else {
				value = envValue
			}
		case "literal":
			literalParams, err := LiteralForm.Validate(config)
			if err != nil {
				return nil, err
			}
			value = literalParams["value"]
		}
		switch params["type"].(string) {
		case "string":
			_, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("variable '%s' is not a string", key)
			}
		case "int":
			if strValue, ok := value.(string); ok {
				if intValue, err := strconv.ParseInt(strValue, 10, 0); err != nil {
					return nil, fmt.Errorf("variable '%s' is not an integer", key)
				} else {
					value = int(intValue)
				}
			} else if _, ok := value.(int); !ok {
				return nil, fmt.Errorf("variable '%s' is not an integer", key)
			}
		case "float":
			if strValue, ok := value.(string); ok {
				if floatValue, err := strconv.ParseFloat(strValue, 64); err != nil {
					return nil, fmt.Errorf("variable '%s' is not a float", key)
				} else {
					value = float64(floatValue)
				}
			} else if _, ok := value.(float64); !ok {
				return nil, fmt.Errorf("variable '%s' is not a float", key)
			}
		case "any":
			break
		}
		values[key] = value
	}

	return values, nil
}

var fullVarRegex = regexp.MustCompile(`^\$([a-z][a-z0-9_]*)$`)
var innerVarRegex = regexp.MustCompile(`^(|.*?[^\$])?\$([a-z][a-z0-9_]*)(.*)$`)
var escapeRegex = regexp.MustCompile(`\$\$`)

func replaceStringVar(value string, values map[string]interface{}) (interface{}, error) {
	if match := fullVarRegex.FindStringSubmatch(value); match != nil {
		varName := match[1]
		if newValue, ok := values[varName]; ok {
			return newValue, nil
		} else {
			return nil, fmt.Errorf("undefined variable: '%s'", varName)
		}
	} else {
		remainingValue := value
		fullPrefix := ""
		for {
			match := innerVarRegex.FindStringSubmatch(remainingValue)
			if match == nil {
				break
			}
			prefix := match[1]
			varName := match[2]
			suffix := match[3]
			if newValue, ok := values[varName]; ok {
				var newStrValue string
				switch v := newValue.(type) {
				case string:
					newStrValue = v
				case int:
					newStrValue = strconv.FormatInt(int64(v), 10)
				case float64:
					newStrValue = strconv.FormatFloat(v, 'f', -1, 64)
				default:
					return nil, fmt.Errorf("invalid interpolation value")
				}
				fullPrefix += prefix + newStrValue
				remainingValue = suffix
			} else {
				return nil, fmt.Errorf("undefined variable: '%s' %v", varName, values)
			}
		}
		return fullPrefix + remainingValue, nil
	}
	return value, nil
}

func unescape(value interface{}) interface{} {
	if strValue, ok := value.(string); ok {
		return escapeRegex.ReplaceAllString(strValue, "$")
	}
	return value
}

func InsertVars(data interface{}, values map[string]interface{}) error {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if strValue, ok := value.(string); ok {
				if newValue, err := replaceStringVar(strValue, values); err != nil {
					return err
				} else {
					v[key] = unescape(newValue)
				}
			} else if err := InsertVars(value, values); err != nil {
				return err
			}
		}
	case []interface{}:
		for i, value := range v {
			if strValue, ok := value.(string); ok {
				if newValue, err := replaceStringVar(strValue, values); err != nil {
					return err
				} else {
					v[i] = unescape(newValue)
				}
			} else if err := InsertVars(value, values); err != nil {
				return err
			}
		}
	}
	return nil
}

func getPath(basePath, filePath string) (string, error) {
	if strings.HasPrefix(filePath, "/") {
		return "", fmt.Errorf("absolute paths are not allowed for security reasons")
	}
	dir := filepath.Dir(basePath)
	return filepath.Join(dir, filePath), nil
}

type Reader func(string) ([]byte, error)

func loadIncludes(data interface{}, filePath string, reader Reader) (interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		newValues := make(map[string]interface{})
		for key, value := range v {
			if key == "$include" {
				includes := make([]string, 0, 1)
				if strValue, ok := value.(string); ok {
					includes = append(includes, strValue)
				} else if listValue, ok := value.([]interface{}); ok {
					for _, value := range listValue {
						if strValue, ok := value.(string); ok {
							includes = append(includes, strValue)
						} else {
							return nil, fmt.Errorf("expected a string")
						}
					}
				} else {
					return nil, fmt.Errorf("invalid $include format (neither a string nor a list of strings)")
				}
				for _, include := range includes {
					if newPath, err := getPath(filePath, include); err != nil {
						return nil, err
					} else {
						if newValue, err := loadYaml(newPath, reader); err != nil {
							return nil, err
						} else {
							// we merge the values
							Merge(newValues, newValue)
						}
					}
				}
			} else {
				if result, err := loadIncludes(value, filePath, reader); err != nil {
					return nil, err
				} else {
					newValues[key] = result
				}
			}
		}
		return newValues, nil
	case []interface{}:
		newValues := make([]interface{}, 0, len(v))
		for _, value := range v {
			if result, err := loadIncludes(value, filePath, reader); err != nil {
				return nil, err
			} else {
				newValues = append(newValues, result)
			}
		}
		return newValues, nil
	}
	return data, nil

}

func LoadYaml(filePath string) (map[string]interface{}, error) {
	return loadYaml(filePath, ioutil.ReadFile)
}

func loadYaml(filePath string, reader Reader) (map[string]interface{}, error) {

	fileContent, err := reader(filePath)
	if err != nil {
		return nil, err
	}

	var settings map[interface{}]interface{}
	yamlerror := yaml.Unmarshal(fileContent, &settings)
	if yamlerror != nil {
		return nil, yamlerror
	}
	deepStringMap, ok := maps.RecursiveToStringMap(settings)
	if !ok {
		return nil, fmt.Errorf("Non-string keys encountered in file '%s'", filePath)
	}
	if mapWithIncludes, err := loadIncludes(deepStringMap, filePath, reader); err != nil {
		return nil, err
	} else {
		return mapWithIncludes.(map[string]interface{}), nil
	}
}

func MakeSettings(settingsPaths []string) (*Settings, error) {
	settings := new(Settings)
	for _, path := range settingsPaths {
		err := settings.Load(path)
		if err != nil {
			return nil, fmt.Errorf("Error loading settings from path '%s': %s", path, err.Error())
		}
	}
	return settings, nil
}
