package forms

import (
	"encoding/hex"
	"fmt"
	"strings"
)

var IsUUIDForm = Form{
	Fields: []Field{
		{
			Name: "convertToBinary",
			Validators: []Validator{
				IsOptional{Default: false},
				IsBoolean{},
			},
		},
	},
}

func MakeIsUUIDValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isUUID := &IsUUID{}
	if params, err := IsUUIDForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsUUIDForm.Coerce(isUUID, params); err != nil {
		return nil, err
	}
	return isUUID, nil
}

type IsUUID struct {
	ConvertToBinary bool `json:"convertToBinary"`
}

func (f IsUUID) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	uuidStr, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("not a valid UUID")
	}
	rawUUIDStr := strings.Replace(uuidStr, "-", "", -1)
	bStr, err := hex.DecodeString(rawUUIDStr)
	if err != nil {
		return nil, fmt.Errorf("not a valid UUID")
	}
	if len(bStr) != 16 {
		return nil, fmt.Errorf("not a valid UUID")
	}
	if f.ConvertToBinary {
		return bStr, nil
	}
	return uuidStr, nil
}
