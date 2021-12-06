package forms

import (
	"fmt"
	"strconv"
)

var IsIntegerForm = Form{
	Fields: []Field{
		{
			Name: "convert",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "hasMin",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "hasMax",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
		{
			Name: "min",
			Validators: []Validator{
				IsOptional{},
				IsInteger{HasMin: true, Min: 0},
			},
		},
		{
			Name: "max",
			Validators: []Validator{
				IsOptional{},
				IsInteger{HasMin: true, Min: 0},
			},
		},
	},
}

func MakeIsIntegerValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isInteger := &IsInteger{}
	if params, err := IsIntegerForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsIntegerForm.Coerce(isInteger, params); err != nil {
		return nil, err
	}
	return isInteger, nil
}

type IsInteger struct {
	Convert bool  `json:"convert"`
	Min     int64 `json:"min" coerce:"convert"`
	Max     int64 `json:"max" coerce:"convert"`
	HasMin  bool  `json:"hasMin"`
	HasMax  bool  `json:"hasMax"`
}

func (f IsInteger) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	var iv int64
	switch v := input.(type) {
	case int64:
		iv = v
	case int:
		iv = int64(v)
	case uint:
		iv = int64(v)
	case float64:
		if float64(int64(v)) != v {
			return nil, fmt.Errorf("not an integer")
		}
		iv = int64(v)
	case string:
		if !f.Convert {
			return nil, fmt.Errorf("not an integer")
		}
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("not an integer")
		}
		iv = i
	default:
		return nil, fmt.Errorf("not an integer")
	}
	if f.HasMin && iv < f.Min {
		return nil, fmt.Errorf("value must be larger than or equal %d", f.Min)
	}
	if f.HasMax && iv > f.Max {
		return nil, fmt.Errorf("value must be smaller than or equal %d", f.Max)
	}
	return iv, nil
}
