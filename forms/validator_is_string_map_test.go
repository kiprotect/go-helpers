package forms

import (
	"testing"
)

func TestIsStringMapFromConfig(t *testing.T) {
	config := map[string]interface{}{
		"fields": []map[string]interface{}{
			{
				"name": "example",
				"validators": []map[string]interface{}{
					{
						"type": "IsStringMap",
						"params": map[string]interface{}{
							"form": map[string]interface{}{
								"fields": []map[string]interface{}{
									{
										"name": "foo",
										"validators": []map[string]interface{}{
											{
												"type": "IsString",
											},
										},
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
	if _, err := form.Validate(map[string]interface{}{"example": map[string]interface{}{"foo": "bar"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := form.Validate(map[string]interface{}{"example": map[string]interface{}{"foo": 12}}); err == nil {
		t.Fatalf("expected an error")
	}
}
