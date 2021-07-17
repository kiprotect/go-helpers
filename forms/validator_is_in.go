package forms

import (
	"fmt"
	"strings"
)

var IsInForm = Form{
	Fields: []Field{
		{
			Name: "choices",
			Validators: []Validator{
				IsList{},
			},
		},
	},
}

func MakeIsInValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isIn := &IsIn{}
	if params, err := IsInForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsInForm.Coerce(isIn, params); err != nil {
		return nil, err
	}
	return isIn, nil
}

type IsIn struct {
	Choices []interface{} `json:"choices"`
}

func (f IsIn) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	found := false
	for _, v := range f.Choices {
		if v == input {
			found = true
			break
		}
	}
	if !found {
		choices := make([]string, len(f.Choices))
		for i, choice := range f.Choices {
			choices[i] = fmt.Sprintf("%v", choice)
		}
		return nil, fmt.Errorf("invalid choice, must be one of: %s", strings.Join(choices, ", "))
	}
	return input, nil
}
