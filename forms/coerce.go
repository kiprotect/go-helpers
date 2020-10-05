package forms

import (
	"fmt"
	"reflect"
	"strings"
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

func Coerce(target interface{}, source interface{}) error {
	return coerce(target, source, make([]interface{}, 0))
}

func coerce(target interface{}, source interface{}, path []interface{}) error {
	targetType := typeOf(target)
	sourceType := typeOf(source)
	targetValue := valueOf(target)
	sourceValue := valueOf(source)
	switch sourceType.Kind() {
	case reflect.Slice:
		// this is a slice, so we expect the target to be a slice too
		if targetType.Kind() != reflect.Slice {
			return MakeCoerceError(fmt.Sprintf("expected an array to coerce an array into"), path)
		}
		elemType := targetType.Elem()
		targetSliceValue := reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0)
		for i := 0; i < sourceValue.Len(); i++ {
			path = append(path, i)
			sourceElemValue := sourceValue.Index(i)
			var targetValue reflect.Value
			if elemType.Kind() == reflect.Ptr {
				// the slice expects a pointer type
				targetValue = reflect.New(unpointType(elemType))
			} else {
				// the slice expects a literal type
				targetValue = reflect.New(elemType)
			}
			if err := coerce(targetValue.Interface(), sourceElemValue.Interface(), path); err != nil {
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
	case reflect.Map:
		// this is a map, so we expect the source to be a map too. Since we assign
		// map values to struct fields we further assume that the map has only string
		// keys, and we don't use reflection to iterate over map keys like we do
		// for the slices above.
		sourceMap, ok := source.(map[string]interface{})
		if !ok {
			return MakeCoerceError(fmt.Sprintf("expected a string map"), path)
		}
		if targetType.Kind() != reflect.Struct {
			return MakeCoerceError(fmt.Sprintf("expected a struct to coerce a map into"), path)
		}
		for i := 0; i < targetType.NumField(); i++ {
			targetFieldType := targetType.Field(i)
			targetFieldValue := targetValue.Field(i)
			sourceName := ToSnakeCase(targetFieldType.Name)
			sourceData, ok := sourceMap[sourceName]
			path = append(path, sourceName)
			if !ok {
				tags := ExtractTags(targetFieldType, "coerce")
				required := false
				for _, tag := range tags {
					if tag.Flag && tag.Name == "required" {
						required = true
					}
				}
				if required {
					return MakeCoerceError(fmt.Sprintf("missing value for required key '%s'", sourceName), path)
				}
				continue
			}
			sourceValue = valueOf(sourceData)
			sourceValueType := typeOf(sourceData)
			if !targetFieldValue.CanSet() {
				return MakeCoerceError(fmt.Sprintf("struct value '%s' of type '%s' is not assignable", targetFieldType.Name, targetFieldType.Type), path)
			}
			// if the target value is not assignable to the source value, it is probably a
			// struct itself that the source value should be coerced into
			if !targetFieldType.Type.AssignableTo(sourceValueType) {
				var targetFieldValuePtr reflect.Value
				if targetFieldValue.Type().Kind() == reflect.Ptr {
					if targetFieldValue.IsZero() {
						// this pointer is uninitialized, we have to initialize it first
						newFieldValue := reflect.New(targetFieldValue.Type().Elem())
						targetFieldValue.Set(newFieldValue)
					}
					targetFieldValuePtr = targetFieldValue
				} else {
					targetFieldValuePtr = targetFieldValue.Addr()
				}
				// we first check if we can generate interface values for both source and target
				if targetFieldValuePtr.CanInterface() && sourceValue.CanInterface() {
					// we then try to coerce the source interface value into the target interface value
					path = append(path, sourceName)
					if err := coerce(targetFieldValuePtr.Interface(), sourceValue.Interface(), path); err != nil {
						return err
					}
				} else {
					return MakeCoerceError(fmt.Sprintf("cannot assign map value '%s' to struct field '%s'", sourceName, targetFieldType.Name), path)
				}
			} else {
				targetFieldValue.Set(sourceValue)
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
