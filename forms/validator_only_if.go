package forms

type OnlyIf struct {
	Function func(interface{}, map[string]interface{}) bool `json:"-"`
}

func (f OnlyIf) Validate(input interface{}, values map[string]interface{}) (interface{}, error) {
	if f.Function(input, values) == true {
		return input, nil
	}
	return nil, nil
}
