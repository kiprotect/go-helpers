package forms

import (
	"testing"
)

func TestSwitchFromConfig(t *testing.T) {
	config := map[string]interface{}{
		"fields": []map[string]interface{}{
			{
				"name": "type",
				"validators": []map[string]interface{}{
					{
						"type": "CanBeAnything",
					},
				},
			},
			{
				"name": "example",
				"validators": []map[string]interface{}{
					{
						"type": "Switch",
						"params": map[string]interface{}{
							"key": "type",
							"cases": map[string]interface{}{
								"string": []map[string]interface{}{
									{
										"type": "IsString",
									},
								},
								"integer": []map[string]interface{}{
									{
										"type": "IsInteger",
									},
								},
							},
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
	if _, err := form.Validate(map[string]interface{}{"example": 4, "type": "integer"}); err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": "bar", "type": "string"}); err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": "bar", "type": "integer"}); err == nil {
		t.Fatalf("expected an error")
	}
}
