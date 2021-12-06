package forms

import (
	"encoding/json"
	"testing"
)

func TestFromConfig(t *testing.T) {
	config := map[string]interface{}{
		"fields": []map[string]interface{}{
			{
				"name": "example",
				"validators": []map[string]interface{}{
					{
						"type": "IsString",
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
	if _, err := form.Validate(map[string]interface{}{"example": 4}); err == nil {
		t.Fatalf("expected an error")
	}
	if params, err := form.Validate(map[string]interface{}{"example": "bar"}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	} else if params["example"] != "bar" {
		t.Fatalf("expected value 'bar'")
	}
}

func TestRoundTrip(t *testing.T) {

	form := &Form{
		Fields: []Field{
			{
				Name: "example",
				Validators: []Validator{
					IsString{},
				},
			},
		},
	}

	context := &FormDescriptionContext{
		Validators: Validators,
	}

	bytes, err := json.Marshal(form)

	if err != nil {
		t.Fatal(err)
	}

	config := map[string]interface{}{}
	if err := json.Unmarshal(bytes, &config); err != nil {
		t.Fatal(err)
	}

	recoveredForm, err := FromConfig(config, context)

	if err != nil {
		t.Fatal(err)
	}
	if _, err := recoveredForm.Validate(map[string]interface{}{"example": 4}); err == nil {
		t.Fatalf("expected an error")
	}
	if params, err := recoveredForm.Validate(map[string]interface{}{"example": "bar"}); err != nil {
		t.Fatalf("expected no error but got %v", err)
	} else if params["example"] != "bar" {
		t.Fatalf("expected value 'bar'")
	}
}
