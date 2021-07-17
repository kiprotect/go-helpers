package forms

import (
	"fmt"
	"time"
)

var IsTimeForm = Form{
	Fields: []Field{
		{
			Name: "toUTC",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "raw",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "form",
			Validators: []Validator{
				IsOptional{
					Default: "rfc3339",
				},
				IsIn{
					Choices: []interface{}{"rfc3339", "rfc3339-date", "unix", "unix-nano", "unix-milli"},
				},
			},
		},
	},
}

func MakeIsTimeValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isTime := &IsTime{}
	if params, err := IsTimeForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsTimeForm.Coerce(isTime, params); err != nil {
		return nil, err
	}
	return isTime, nil
}

type IsTime struct {
	Format string `json:"format"`
	ToUTC  bool   `json:"toUTC"`
	Raw    bool   `json:"raw"`
}

func (f IsTime) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {

	toNumber := func() (int64, error) {
		var t int64
		if inputFloat, ok := input.(float64); ok {
			t = int64(inputFloat)
		} else if inputInt, ok := input.(int); ok {
			t = int64(inputInt)
		} else if inputInt64, ok := input.(int64); ok {
			t = inputInt64
		} else {
			return 0, fmt.Errorf("not a number")
		}
		return t, nil
	}

	var t time.Time
	var err error
	switch f.Format {
	case "":
		fallthrough
	case "rfc3339":
		inputStr, ok := input.(string)
		if !ok {
			return nil, fmt.Errorf("not a string")
		}
		t, err = time.Parse(time.RFC3339, inputStr)
	case "rfc3339-date":
		inputStr, ok := input.(string)
		if !ok {
			return nil, fmt.Errorf("not a string")
		}
		t, err = time.Parse("2006-01-02", inputStr)
	case "unix":
		if n, err := toNumber(); err != nil {
			return nil, err
		} else {
			if f.Raw {
				return input, nil
			}
			return time.Unix(n, 0), nil
		}
	case "unix-nano":
		var n int64
		if n, err = toNumber(); err == nil {
			if f.Raw {
				return n, nil
			}
			t = time.Unix(n/1e9, n%1e9)
		}
	case "unix-milli":
		var n int64
		if n, err = toNumber(); err == nil {
			if f.Raw {
				return n, nil
			}
			t = time.Unix(n/1e3, (n%1e3)*1e6)
		}
	default:
		return nil, fmt.Errorf("invalid time format: %s", f.Format)
	}
	if err != nil {
		return nil, err
	}
	if f.ToUTC {
		t = t.UTC()
	}
	if f.Raw {
		return input, nil
	}
	return t, nil

}
