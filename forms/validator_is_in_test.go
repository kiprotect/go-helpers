package forms

import (
	"testing"
)

func TestIsInFromConfig(t *testing.T) {
	config := map[string]interface{}{
		"fields": []map[string]interface{}{
			{
				"name": "example",
				"validators": []map[string]interface{}{
					{
						"type": "IsIn",
						"config": map[string]interface{}{
							"choices": []interface{}{"a", "b", "c"},
						},
					},
				},
			},
		},
	}
	context := &FormDescriptionContext{
		Validators: Validators,
	}
	form, err := FromConfig(config, context)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": "a"}); err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": "b"}); err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": "d"}); err == nil {
		t.Fatalf("expected an error")
	}
}
