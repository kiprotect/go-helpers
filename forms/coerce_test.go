package forms

import (
	"testing"
)

type BasicTestStruct struct {
	Foo        string
	Bar        int
	StringList []string
	IntList    []int
	Map        map[string]interface{}
	Interface  interface{}
}

type EmbeddedStruct struct {
	Embedded map[string]string
	Name string
}

type ComplexTestStruct struct {
	EmbeddedStruct
	Foo        string
	Bar        int
	Baz        Baz
	BazPtr     *Baz
	ZapList    []Zap
	ZapListPtr []*Zap
}

type Baz struct {
	Baz string
}

type Zap struct {
	Zap string
}

func TestBasicCoerce(t *testing.T) {
	testMap := map[string]interface{}{
		"foo":         "test",
		"bar":         4,
		"string_list": []string{"a", "b", "c"},
		"map": map[string]interface{}{"test": "test"},
		"interface": "foo",
	}
	bt := &BasicTestStruct{}
	if err := Coerce(bt, testMap); err != nil {
		t.Fatal(err)
	}
	if bt.Interface != "foo" {
		t.Fatalf("expected 'foo' as value of Interface")
	}
	if bt.Foo != "test" {
		t.Fatalf("expected 'test' as value of Foo")
	}
	if bt.Bar != 4 {
		t.Fatalf("expected 4 as value of Bar")
	}
	if len(bt.StringList) != 3 {
		t.Fatalf("expected a list of length 3")
	}
	if bt.StringList[0] != "a" {
		t.Fatalf("expected value 'a' for string list")
	}
}

func TestComplexCoerce(t *testing.T) {
	testMap := map[string]interface{}{
		"embedded": map[string]string{"foo" : "foo"},
		"name": "slim shady",
		"foo": "test",
		"bar": 4,
		"baz": map[string]interface{}{
			"baz": "baz",
		},
		"baz_ptr": map[string]interface{}{
			"baz": "baz",
		},
		"zap_list": []map[string]interface{}{
			{
				"zap": "baz",
			},
			{
				"zap": "barz",
			},
		},
		"zap_list_ptr": []map[string]interface{}{
			{
				"zap": "baz",
			},
			{
				"zap": "barz",
			},
		},
	}
	bt := &ComplexTestStruct{}
	if err := Coerce(bt, testMap); err != nil {
		t.Fatal(err)
	}
	if bt.Name != "slim shady" {
		t.Fatalf("name doesn't match")
	}
	if bt.Embedded["foo"] != "foo" {
		t.Fatalf("expected embedded value to be set")
	}
	if bt.Foo != "test" {
		t.Fatalf("expected 'test' as value of Foo")
	}
	if bt.Bar != 4 {
		t.Fatalf("expected 4 as value of Bar")
	}
	if len(bt.ZapList) != 2 {
		t.Fatalf("expected a list of length 1 for ZapList, got %d", len(bt.ZapList))
		if bt.ZapList[0].Zap != "baz" {
			t.Fatalf("Expected a value of 'baz' for the first element of ZapList")
		}
		if bt.ZapList[1].Zap != "barz" {
			t.Fatalf("Expected a value of 'baz' for the first element of ZapList")
		}
	}
	if len(bt.ZapListPtr) != 2 {
		t.Fatalf("expected a list of length 1 for ZapListPtr, got %d", len(bt.ZapListPtr))
		if bt.ZapListPtr[0].Zap != "baz" {
			t.Fatalf("Expected a value of 'baz' for the first element of ZapListPtr")
		}
		if bt.ZapListPtr[1].Zap != "barz" {
			t.Fatalf("Expected a value of 'baz' for the first element of ZapListPtr")
		}
	}
}

func TestBasicTypeMismatch(t *testing.T) {
	testMap := map[string]interface{}{
		"foo": 5,
	}
	bt := &BasicTestStruct{}
	if err := Coerce(bt, testMap); err == nil {
		t.Fatalf("should throw an error")
	}
}

func TestComplexTypeMismatch1(t *testing.T) {
	testMap := map[string]interface{}{
		"baz": 5,
	}
	bt := &ComplexTestStruct{}
	if err := Coerce(bt, testMap); err == nil {
		t.Fatalf("should throw an error")
	}
}

func TestComplexTypeMismatch2(t *testing.T) {
	testMap := map[string]interface{}{
		"baz_ptr": 5,
	}
	bt := &ComplexTestStruct{}
	if err := Coerce(bt, testMap); err == nil {
		t.Fatalf("should throw an error")
	}
}

func TestComplexTypeMismatch3(t *testing.T) {
	testMap := map[string]interface{}{
		"zap_list": 5,
	}
	bt := &ComplexTestStruct{}
	if err := Coerce(bt, testMap); err == nil {
		t.Fatalf("should throw an error")
	}
}
