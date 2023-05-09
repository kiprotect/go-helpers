package forms

import (
	"fmt"
)

var IsStringMapForm = Form{
	Fields: []Field{
		{
			Name: "form",
			Validators: []Validator{
				IsOptional{},
				IsStringMap{
					Form: &FormForm,
				},
			},
		},
	},
}

func MakeIsStringMapValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isStringMap := &IsStringMap{}
	if params, err := IsStringMapForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsStringMapForm.Coerce(isStringMap, params); err != nil {
		return nil, err
	} else {
		if isStringMap.Form != nil {
			if err := isStringMap.Form.Initialize(context); err != nil {
				return nil, err
			}
		}
	}
	return isStringMap, nil
}

type IsStringMap struct {
	Form   *Form       `json:"form,omitempty"`
	Coerce interface{} `json:"-"`
}

func (f IsStringMap) ValidateWithContext(input interface{}, values map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	return f.validate(input, values, context)
}

func (f IsStringMap) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	return f.validate(input, values, nil)
}

func (f IsStringMap) validate(input interface{}, values map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	sm, ok := input.(map[string]interface{})
	if !ok {
		m, ok := input.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("not a map")
		}
		sm = make(map[string]interface{})
		for k, v := range m {
			sk, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("not a string map")
			}
			sm[sk] = v
		}
	}
	// if a forms is defined for the string map we execute it
	if f.Form != nil {
		if context == nil {
			context = map[string]interface{}{"_parent": values}
		} else {
			context["_parent"] = values
		}
		if params, err := f.Form.ValidateWithContext(sm, context); err != nil {
			return nil, err
		} else {
			if f.Coerce != nil {
				target := New(f.Coerce)
				if err := Coerce(target, params); err != nil {
					return nil, err
				} else {
					return target, err
				}
			}
			return params, nil
		}
	}

	return sm, nil
}
