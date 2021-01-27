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

package strings

import (
	"fmt"
	"strings"
)

//Helper function to check if a list of strings contains a given string.
func Contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

//Helper function to check if at least one string in list has a given prefix.
func HasPrefix(a string, list []string) bool {
	for _, b := range list {
		if strings.HasPrefix(b, a) {
			return true
		}
	}
	return false
}

func ToListOfStr(value interface{}) ([]string, error) {

	valList, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Not a slice to begin with, cannot convert!")
	}

	valListOfStr := make([]string, len(valList))
	for idx, element := range valList {
		strElement, ok := element.(string)
		if !ok {
			return nil, fmt.Errorf("Non-string type found, aborting!")
		}
		valListOfStr[idx] = strElement
	}

	return valListOfStr, nil
}
