// KIProtect Go-Helpers - Golang Utility Functions
// Copyright (C) 2019-2021  KIProtect GmbH (HRB 208395B) - Germany
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a opy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

package forms

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

type Tag struct {
	Name  string
	Value string
	Flag  bool
}

func ExtractTags(field reflect.StructField, tag string) []Tag {
	tags := make([]Tag, 0)
	if value, ok := field.Tag.Lookup(tag); ok {
		strTags := strings.Split(value, ",")
		for _, tag := range strTags {
			kv := strings.Split(value, ":")
			if len(kv) == 1 {
				tags = append(tags, Tag{
					Name:  tag,
					Value: "",
					Flag:  true,
				})
			} else {
				tags = append(tags, Tag{
					Name:  kv[0],
					Value: kv[1],
					Flag:  false,
				})
			}
		}
	}
	return tags
}

type CoerceError struct {
	Path    []interface{}
	Message string
}

func MakeCoerceError(message string, path []interface{}) *CoerceError {
	return &CoerceError{
		Message: message,
		Path:    path,
	}
}

func (c CoerceError) Error() string {
	pathComponents := make([]string, len(c.Path))
	for i, key := range c.Path {
		pathComponents[i] = fmt.Sprintf("%v", key)
	}
	return fmt.Sprintf("%s (%s)", c.Message, strings.Join(pathComponents, "."))
}

func New(target interface{}) interface{} {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Struct {
		targetType = targetType.Elem()
	}
	return reflect.New(targetType).Interface()
}

func Coerce(target interface{}, source interface{}) error {
	return coerce(target, source, make([]interface{}, 0), nil)
}

func coerce(target interface{}, source interface{}, path []interface{}, tags []Tag) error {
	targetType := typeOf(target)
	sourceType := typeOf(source)
	targetValue := valueOf(target)
	sourceValue := valueOf(source)

	tryAssign := func(st, tt reflect.Type, sv, tv reflect.Value) bool {
		if tt.AssignableTo(st) {
			// the source can be directly assigned to the target
			tv.Set(sv)
			return true
		}

		if st.ConvertibleTo(tt) && tags != nil {

			convert := false
			for _, tag := range tags {
				if tag.Flag && tag.Name == "convert" {
					convert = true
					break
				}
			}

			// conversion needs to be specified explicitly, as it can lead to weird errors...
			if convert {
				tv.Set(sv.Convert(tt))
				return true
			}
		}
		return false
	}

	if tryAssign(sourceType, targetType, sourceValue, targetValue) {
		return nil
	}

	// if this is an embedded struct we try to assign to it
	if targetType.Kind() == reflect.Struct && targetType.NumField() == 1 {
		field := targetType.Field(0)

		assignableTo := field.Type.AssignableTo(sourceType)
		convertibleTo := sourceType.ConvertibleTo(field.Type)

		if assignableTo || convertibleTo {

			if assignableTo {
				newStructValue := reflect.NewAt(field.Type, unsafe.Pointer(targetValue.Field(0).UnsafeAddr()))
				newStructValue.Elem().Set(sourceValue)
				return nil
			}

			convert := false
			for _, tag := range tags {
				if tag.Flag && tag.Name == "convert" {
					convert = true
					break
				}
			}

			// conversion needs to be specified explicitly, as it can lead to weird errors...
			if convert {
				newStructValue := reflect.NewAt(field.Type, unsafe.Pointer(targetValue.Field(0).UnsafeAddr()))
				newStructValue.Elem().Set(sourceValue.Convert(field.Type))
				return nil
			}
		}
	}

	if targetType.Kind() == reflect.Interface {
		// this is an interface
		targetValue.Set(sourceValue)
		return nil
	}

	switch sourceType.Kind() {
	case reflect.Slice:
		// this is a slice, so we expect the target to be a slice too
		if targetType.Kind() != reflect.Slice {
			return MakeCoerceError(fmt.Sprintf("expected an array to coerce an array into"), path)
		}
		elemType := targetType.Elem()
		targetSliceValue := reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0)
		for i := 0; i < sourceValue.Len(); i++ {
			slicePath := append(path, i)
			sourceElemValue := sourceValue.Index(i)
			var targetValue reflect.Value
			if elemType.Kind() == reflect.Ptr {
				// the slice expects a pointer type
				targetValue = reflect.New(unpointType(elemType))
			} else {
				// the slice expects a literal type
				targetValue = reflect.New(elemType)
			}
			if err := coerce(targetValue.Interface(), sourceElemValue.Interface(), slicePath, nil); err != nil {
				return err
			}
			if elemType.Kind() == reflect.Ptr {
				// the slice expects a pointer type
				targetSliceValue = reflect.Append(targetSliceValue, targetValue)
			} else {
				// the slice expects a literal type
				targetSliceValue = reflect.Append(targetSliceValue, unpointValue(targetValue))
			}
		}
		targetValue.Set(targetSliceValue)
	case reflect.Struct:
		if targetType.Kind() != reflect.Map {
			return MakeCoerceError(fmt.Sprintf("expected a map to coerce a struct into, got '%s'", targetType.Kind()), path)
		}
		for i := 0; i < sourceType.NumField(); i++ {
			sourceFieldType := sourceType.Field(i)
			sourceFieldValue := sourceValue.Field(i)

			coerceTags := ExtractTags(sourceFieldType, "coerce")
			jsonTags := ExtractTags(sourceFieldType, "json")

			targetMap, ok := target.(map[string]interface{})
			if !ok {
				return MakeCoerceError(fmt.Sprintf("expected a string map"), path)
			}

			var targetName string

			for _, tag := range coerceTags {
				if !tag.Flag && tag.Name == "name" {
					targetName = tag.Value
				}
			}
			if targetName == "" {
				if len(jsonTags) > 0 && jsonTags[0].Flag {
					if jsonTags[0].Name == "-" {
						continue // this field should not be serialized
					}
					targetName = jsonTags[0].Name
				} else {
					targetName = ToSnakeCase(sourceFieldType.Name)
				}
			}

			sourceMapValue := sourceFieldValue.Interface()

			switch sourceFieldValue.Type().Kind() {
			case reflect.Map:
				newMap := map[string]interface{}{}
				if err := coerce(newMap, sourceMapValue, append(path, sourceFieldType.Name), coerceTags); err != nil {
					return err
				}
				sourceMapValue = newMap
			case reflect.Slice:
				newSlice := []interface{}{}
				if err := coerce(newSlice, sourceMapValue, append(path, sourceFieldType.Name), coerceTags); err != nil {
					return err
				}
				sourceMapValue = newSlice
			}

			targetMap[targetName] = sourceMapValue

		}
	case reflect.Map:
		// this is a map, so we expect the source to be a map too. Since we assign
		// map values to struct fields we further assume that the map has only string
		// keys, and we don't use reflection to iterate over map keys like we do
		// for the slices above.
		if targetType.Kind() != reflect.Struct && targetType.Kind() != reflect.Map {
			return MakeCoerceError(fmt.Sprintf("expected a struct or a map to coerce a map into, got '%s'", targetType.Kind()), path)
		}
		sourceMap, ok := source.(map[string]interface{})
		if !ok {
			return MakeCoerceError(fmt.Sprintf("expected a string map"), path)
		}

		if targetType.Kind() == reflect.Struct {
			for i := 0; i < targetType.NumField(); i++ {
				targetFieldType := targetType.Field(i)
				targetFieldValue := targetValue.Field(i)

				coerceTags := ExtractTags(targetFieldType, "coerce")
				jsonTags := ExtractTags(targetFieldType, "json")

				var sourceName string

				for _, tag := range coerceTags {
					if !tag.Flag && tag.Name == "name" {
						sourceName = tag.Value
					}
				}
				if sourceName == "" {
					if len(jsonTags) > 0 && jsonTags[0].Flag {
						sourceName = jsonTags[0].Name
					} else {
						sourceName = ToSnakeCase(targetFieldType.Name)
					}
				}

				sourceData, ok := sourceMap[sourceName]
				mapPath := append(path, sourceName)
				if targetFieldType.Anonymous {
					// this is an anonymous field
					mapPath = append(path, fmt.Sprintf("%s(anonymous)", sourceName))
					sourceData = sourceMap
					ok = true
				}
				if !ok {
					required := false
					for _, tag := range coerceTags {
						if tag.Flag && tag.Name == "required" {
							required = true
						}
					}
					if required {
						return MakeCoerceError(fmt.Sprintf("missing value for required key '%s'", sourceName), mapPath)
					}
					continue
				}
				sourceValue = valueOf(sourceData)
				sourceValueType := typeOf(sourceData)
				if !targetFieldValue.CanSet() {
					return MakeCoerceError(fmt.Sprintf("struct value '%s' of type '%s' is not assignable", targetFieldType.Name, targetFieldType.Type), mapPath)
				}
				// if the target value is not assignable to the source value, it is probably a
				// struct itself that the source value should be coerced into
				if !targetFieldType.Type.AssignableTo(sourceValueType) {
					var targetFieldValuePtr reflect.Value

					if targetFieldValue.IsZero() {
						switch targetFieldValue.Type().Kind() {
						case reflect.Ptr:
							// this pointer is uninitialized, we have to initialize it first
							targetFieldValue.Set(reflect.New(targetFieldValue.Type().Elem()))
						case reflect.Map:
							targetFieldValue.Set(reflect.MakeMap(targetFieldValue.Type()))
						case reflect.Slice:
							targetFieldValue.Set(reflect.MakeSlice(targetFieldValue.Type(), 0, 0))
						}
					}

					if targetFieldValue.Type().Kind() == reflect.Ptr {
						targetFieldValuePtr = targetFieldValue
					} else {
						targetFieldValuePtr = targetFieldValue.Addr()
					}
					// we first check if we can generate interface values for both source and target
					if targetFieldValuePtr.CanInterface() && sourceValue.CanInterface() {
						// we then try to coerce the source interface value into the target interface value
						if err := coerce(targetFieldValuePtr.Interface(), sourceValue.Interface(), mapPath, coerceTags); err != nil {
							return err
						}
					} else {
						return MakeCoerceError(fmt.Sprintf("cannot assign map value '%s' to struct field '%s'", sourceName, targetFieldType.Name), mapPath)
					}
				} else {
					targetFieldValue.Set(sourceValue)
				}
			}
		} else {
			// this is a map
			if targetType.Key().Kind() != reflect.String {
				return MakeCoerceError(fmt.Sprintf("expected a string map, got '%s'", targetType.Key().Kind()), path)
			}
			for k, v := range sourceMap {
				kv := valueOf(k)
				// we create a new target type
				tv := reflect.New(targetType.Elem())
				// we try to assign the source value to the target
				if err := coerce(tv.Interface(), v, append(path, k), nil); err != nil {
					return err
				}
				// we assign the map value to the unpointed value
				targetValue.SetMapIndex(kv, unpointValue(tv))
			}
		}
		break
	default:
		return MakeCoerceError(fmt.Sprintf("cannot coerce source of type '%s' into target of type '%s'", sourceType.Kind(), targetType.Kind()), path)
	}
	return nil
}

func unpointValue(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		return reflect.Indirect(value)
	}
	return value
}

func unpointType(typ reflect.Type) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		return typ.Elem()
	}
	return typ
}

func valueOf(value interface{}) reflect.Value {
	return unpointValue(reflect.ValueOf(value))
}

func typeOf(value interface{}) reflect.Type {
	return unpointType(reflect.TypeOf(value))
}
