package forms

var IsOptionalForm = Form{
	Fields: []Field{
		{
			Name: "default",
			Validators: []Validator{
				IsOptional{},
			},
		},
	},
}

func MakeIsOptionalValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	isOptional := &IsOptional{}
	if params, err := IsOptionalForm.Validate(config); err != nil {
		return nil, err
	} else if err := IsOptionalForm.Coerce(isOptional, params); err != nil {
		return nil, err
	}
	return isOptional, nil
}

type IsOptional struct {
	Default          interface{}        `json:"default"`
	DefaultGenerator func() interface{} `json:"-"`
}

func (f IsOptional) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	if input == nil || input == "" {
		//if a default value is defined we return that instead
		if f.Default != nil {
			return f.Default, nil
		} else if f.DefaultGenerator != nil {
			return f.DefaultGenerator(), nil
		}
		return nil, nil
	}
	return input, nil
}
