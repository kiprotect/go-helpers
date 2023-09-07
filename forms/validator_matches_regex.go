package forms

import (
	"fmt"
	"regexp"
)

var MatchesRegexForm = Form{
	Fields: []Field{
		{
			Name: "regexp",
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
	if regexp, err := regexp.Compile(matchesRegex.Source); err != nil {
		return nil, err
	} else {
		matchesRegex.Regexp = regexp
	}
	return matchesRegex, nil
}

func (f MatchesRegex) Serialize() (map[string]interface{}, error) {
	return map[string]interface{}{
		"regexp": f.Regexp.String(),
	}, nil
}

type MatchesRegex struct {
	Source string         `json:"regexp"`
	Regexp *regexp.Regexp `json:"-"`
}

func (f MatchesRegex) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	value, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("MatchesRegex: expected a string")
	}
	if matched := f.Regexp.Match([]byte(value)); !matched {
		return nil, fmt.Errorf("regex '%s' did not match", f.Regexp.String())
	}
	return value, nil
}
