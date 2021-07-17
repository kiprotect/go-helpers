package forms

type CanBeAnything struct {
}

func MakeCanBeAnythingValidator(config map[string]interface{}, context *FormDescriptionContext) (Validator, error) {
	return &CanBeAnything{}, nil
}

func (f CanBeAnything) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	return input, nil
}
