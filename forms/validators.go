package forms

var Validators = map[string]ValidatorMaker{
	"IsString":      MakeIsStringValidator,
	"IsStringList":  MakeIsStringListValidator,
	"CanBeAnything": MakeCanBeAnythingValidator,
	"IsBytes":       MakeIsBytesValidator,
	"IsBoolean":     MakeIsBooleanValidator,
	"IsFloat":       MakeIsFloatValidator,
	"IsHex":         MakeIsHexValidator,
	"IsIn":          MakeIsInValidator,
	"IsInteger":     MakeIsIntegerValidator,
	"IsList":        MakeIsListValidator,
	"IsNotIn":       MakeIsNotInValidator,
	"IsOptional":    MakeIsOptionalValidator,
	"IsRequired":    MakeIsRequiredValidator,
	"IsStringMap":   MakeIsStringMapValidator,
	"IsTime":        MakeIsTimeValidator,
	"IsUUID":        MakeIsUUIDValidator,
	"MatchesRegex":  MakeMatchesRegexValidator,
	"Or":            MakeOrValidator,
	"Switch":        MakeSwitchValidator,
}
