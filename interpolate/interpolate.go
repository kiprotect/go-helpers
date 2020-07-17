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

package interpolate

import (
	"bytes"
	"text/template"
)

type InterpolationError struct {
	msg string
}

func (self *InterpolationError) Error() string {
	return self.msg
}

func Interpolate(s string, args interface{}) (so string, er error) {
	defer func() {
		if r := recover(); r != nil {
			so = s
			er = &InterpolationError{"Interpolate crashed..."}
		}
	}()
	// we change the default delimiters because they clash with Jinja...
	t, err := template.New("foo").Delims("{", "}").Parse(s)
	buf := new(bytes.Buffer)
	err = t.Execute(buf, args)
	return buf.String(), err
}
