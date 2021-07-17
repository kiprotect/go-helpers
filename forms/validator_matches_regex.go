package forms

import (
	"fmt"
	"regexp"
)

var MatchesRegexForm = Form{
	Fields: []Field{
		{
			Name: "regex",
			Validators: []Validator{
				IsString{},
				// to do: add regex validation
			},
		},
	},
}

func MakeMatchesRegexValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	matchesRegex := &MatchesRegex{}
	if params, err := MatchesRegexForm.Validate(config); err != nil {
		return nil, err
	} else if err := MatchesRegexForm.Coerce(matchesRegex, params); err != nil {
		return nil, err
	}
	return matchesRegex, nil
}

type MatchesRegex struct {
	Regex *regexp.Regexp `json:"regexp"`
}

func (f MatchesRegex) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	value, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("expected a string")
	}
	if matched := f.Regex.Match([]byte(value)); !matched {
		return nil, fmt.Errorf("regex '%s' did not match", f.Regex.String())
	}
	return value, nil
}
