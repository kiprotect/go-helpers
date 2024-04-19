// KIProtect Go-Helpers - Golang Utility Functions
// Copyright (C) 2019-2024  KIProtect GmbH (HRB 208395B) - Germany
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

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
