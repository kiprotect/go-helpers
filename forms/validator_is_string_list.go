package forms

import (
	"fmt"
)

var IsStringListForm = Form{
	Fields: []Field{
		{
			Name: "validators",
			Validators: []Validator{
				IsOptional{},
				IsList{
					Validators: []Validator{
						IsStringMap{
							Form: &ValidatorDescriptionForm,
						},
					},
				},
			},
		},
	},
}

func (f IsStringList) Serialize() (map[string]interface{}, error) {
	if validators, err := SerializeValidators(f.Validators); err != nil {
		return nil, err
	} else {
		return map[string]interface{}{
			"validators": validators,
		}, nil
	}
}

func MakeIsStringListValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isStringList := &IsStringList{}
	if params, err := IsStringListForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsStringListForm.Coerce(isStringList, params); err != nil {
		return nil, err
	} else {
		validators := []Validator{}
		for _, validatorDescription := range isStringList.ValidatorDescriptions {
			if validator, err := ValidatorFromDescription(validatorDescription, context); err != nil {
				return nil, err
			} else {
				validators = append(validators, validator)
			}
		}
		isStringList.Validators = validators
	}
	return isStringList, nil
}

type IsStringList struct {
	Validators            []Validator             `json:"-"`
	ValidatorDescriptions []*ValidatorDescription `json:"validators"`
}

func (f IsStringList) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	strList := make([]string, 0)
	switch l := input.(type) {
	case []string:
		strList = l
		break
	case []interface{}:
		for _, v := range l {
			strV, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("not a string")
			}
			strList = append(strList, strV)
		}
	}
	for _, validator := range f.Validators {
		for i, v := range strList {
			res, err := validator.Validate(v, values)
			if err != nil {
				return nil, err
			}
			strRes, ok := res.(string)
			if !ok {
				return nil, fmt.Errorf("validator result is not a string")
			}
			strList[i] = strRes
		}
	}
	return strList, nil
}
