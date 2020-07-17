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

func RecursiveToStringMap(value interface{}) (map[string]interface{}, bool) {
	stringMap, ok := recursiveToStringMap(value, false)
	if ok {
		return stringMap.(map[string]interface{}), true
	}
	return nil, false
}

func recursiveToStringMap(value interface{}, innerCall bool) (interface{}, bool) {
	valueStrMap, ok := value.(map[string]interface{})
	if ok {
		newValueStrMap := make(map[string]interface{})
		for key, value := range valueStrMap {
			newValue, ok := recursiveToStringMap(value, true)
			if !ok {
				return nil, false
			}
			newValueStrMap[key] = newValue
		}
		return newValueStrMap, true
	}
	valueIfMap, ok := value.(map[interface{}]interface{})
	// if this is a generic map we try to convert it to a string map and
	// reeturn the result
	if ok {
		valueStrMap = make(map[string]interface{})
		for key, value := range valueIfMap {
			strKey, ok := key.(string)
			if !ok {
				return nil, false
			}
			newValue, ok := recursiveToStringMap(value, true)
			if !ok {
				return nil, false
			}
			valueStrMap[strKey] = newValue
		}
		return valueStrMap, true
	}
	// if this isn't an inner call this is not a string map
	if !innerCall {
		return nil, false
	}
	valueList, ok := value.([]interface{})
	if ok {
		// this is a list, we convert each of its elements
		newValueList := make([]interface{}, len(valueList))
		for i, value := range valueList {
			if newValue, ok := recursiveToStringMap(value, true); !ok {
				return nil, false
			} else {
				newValueList[i] = newValue
			}
		}
		return newValueList, true
	}
	// we do not modify this value
	return value, true
}

func ToStringMap(value interface{}) (map[string]interface{}, bool) {
	valueStrMap, ok := value.(map[string]interface{})
	if ok {
		return valueStrMap, true
	}
	valueIfMap, ok := value.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}
	valueStrMap = make(map[string]interface{})
	for key, value := range valueIfMap {
		strKey, ok := key.(string)
		if !ok {
			return nil, false
		}
		valueStrMap[strKey] = value
	}
	return valueStrMap, true
}

func ToStringMapList(value interface{}) ([]map[string]interface{}, bool) {
	lv, ok := value.([]interface{})
	if !ok {
		ll, ok := value.([]map[string]interface{})
		if !ok {
			return nil, false
		}
		return ll, true
	}
	ll := make([]map[string]interface{}, len(lv))
	for i, e := range lv {
		m, ok := ToStringMap(e)
		if !ok {
			return nil, false
		}
		ll[i] = m
	}
	return ll, true
}
