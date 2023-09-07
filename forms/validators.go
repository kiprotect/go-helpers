package forms

type ValidatorDefinition struct {
	Maker ValidatorMaker
	Form  Form
}

var Validators = map[string]ValidatorDefinition{
	"IsNil":         ValidatorDefinition{MakeIsNilValidator, IsNilForm},
	"IsString":      ValidatorDefinition{MakeIsStringValidator, IsStringForm},
	"IsStringList":  ValidatorDefinition{MakeIsStringListValidator, IsStringListForm},
	"CanBeAnything": ValidatorDefinition{MakeCanBeAnythingValidator, CanBeAnythingForm},
	"IsBytes":       ValidatorDefinition{MakeIsBytesValidator, IsBytesForm},
	"IsBoolean":     ValidatorDefinition{MakeIsBooleanValidator, IsBooleanForm},
	"IsFloat":       ValidatorDefinition{MakeIsFloatValidator, IsFloatForm},
	"IsHex":         ValidatorDefinition{MakeIsHexValidator, IsHexForm},
	"IsIn":          ValidatorDefinition{MakeIsInValidator, IsInForm},
	"IsInteger":     ValidatorDefinition{MakeIsIntegerValidator, IsIntegerForm},
	"IsList":        ValidatorDefinition{MakeIsListValidator, IsListForm},
	"IsNotIn":       ValidatorDefinition{MakeIsNotInValidator, IsNotInForm},
	"IsOptional":    ValidatorDefinition{MakeIsOptionalValidator, IsOptionalForm},
	"IsRequired":    ValidatorDefinition{MakeIsRequiredValidator, IsRequiredForm},
	"IsStringMap":   ValidatorDefinition{MakeIsStringMapValidator, IsStringMapForm},
	"IsTime":        ValidatorDefinition{MakeIsTimeValidator, IsTimeForm},
	"IsUUID":        ValidatorDefinition{MakeIsUUIDValidator, IsUUIDForm},
	"MatchesRegex":  ValidatorDefinition{MakeMatchesRegexValidator, MatchesRegexForm},
	"Or":            ValidatorDefinition{MakeOrValidator, OrForm},
	"Switch":        ValidatorDefinition{MakeSwitchValidator, SwitchForm},
}
