package forms

import (
	"fmt"
	"sort"
)

var CasesForm = Form{
	Fields: []Field{
		{
			Name: "*",
			Validators: []Validator{
				IsOptional{Default: []map[string]any{}},
				IsList{
					Validators: []Validator{
						IsOptional{Default: map[string]any{}},
						IsStringMap{
							Form: &ValidatorDescriptionForm,
						},
					},
				},
			},
		},
	},
}

var SwitchForm = Form{
	Fields: []Field{
		{
			Name: "key",
			Validators: []Validator{
				IsOptional{Default: "_"},
				IsString{},
			},
		},
		{
			Name: "cases",
			Validators: []Validator{
				IsOptional{Default: map[string]any{}},
				IsStringMap{
					Form: &CasesForm,
				},
			},
		},
	},
}

func (f Switch) Serialize() (map[string]interface{}, error) {
	casesDescriptions := make(map[string][]*ValidatorDescription)
	for key, validators := range f.Cases {
		if descriptions, err := SerializeValidators(validators); err != nil {
			return nil, err
		} else {
			casesDescriptions[key] = descriptions
		}
	}

	defaultDescriptions, err := SerializeValidators(f.Default)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"key":        f.Key,
		"exhaustive": f.Exhaustive,
		"default":    defaultDescriptions,
		"cases":      casesDescriptions,
	}, nil
}

func MakeSwitchValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	switchValidator := &Switch{}
	if params, err := SwitchForm.Validate(config); err != nil {
		return nil, err
	} else if err := SwitchForm.Coerce(switchValidator, params); err != nil {
		return nil, err
	} else {

		values := []string{}

		for value, _ := range switchValidator.CasesDescriptions {
			values = append(values, value)
		}

		// we sort the case values to make sure they will be evaluated in a deterministic order
		sort.Strings(values)

		cases := map[string][]Validator{}

		for _, key := range values {
			caseDescription := switchValidator.CasesDescriptions[key]
			validators := []Validator{}
			for _, validatorDescription := range caseDescription {
				if validator, err := ValidatorFromDescription(validatorDescription, context); err != nil {
					return nil, err
				} else {
					validators = append(validators, validator)
				}
			}
			cases[key] = validators
		}
		switchValidator.Cases = cases
	}
	return switchValidator, nil
}

type Switch struct {
	Key                 string                             `json:"key"`
	Exhaustive          bool                               `json:"exhaustive"`
	Default             []Validator                        `json:"-"`
	Cases               map[string][]Validator             `json:"-"`
	CasesDescriptions   map[string][]*ValidatorDescription `json:"cases"`
	DefaultDescriptions []*ValidatorDescription            `json:"default"`
}

func (f Switch) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	strValue, ok := values[f.Key].(string)

	if !ok {
		return nil, fmt.Errorf("switch key is not a string")
	}

	caseValue, ok := f.Cases[strValue]

	if !ok {

		if f.Default != nil {
			caseValue = f.Default
			ok = true
		}

		if !ok {

			// the switch case is not covered, we return nil
			if input == nil && !f.Exhaustive {
				return nil, nil
			}

			// no default defined either
			return nil, fmt.Errorf("unknown switch case value: '%s'", strValue)
		}

	}

	var err error

	for _, validator := range caseValue {
		input, err = validator.Validate(input, values)
		if err != nil {
			return nil, err
		}
	}

	return input, nil
}
