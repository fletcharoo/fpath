package internal

import (
	"fmt"
	"reflect"
	"strconv"
)

// LookupPath retrieves a value from a nested data structure at the location
// of the provided path.
func LookupPath(data any, path []string) (result any, err error) {
	pathLen := len(path)
	if pathLen == 0 {
		return data, nil
	}

	key := path[0]

	switch d := data.(type) {
	case map[string]any:
		result, err = lookupPathMap(d, key)
	case []any:
		result, err = lookupPathSlice(d, key)
	default:
		result, err = lookupPathStruct(d, key)
	}

	if err != nil {
		err = fmt.Errorf("failed to get value at %q", key)
		return
	}

	if pathLen == 1 {
		return result, nil
	}

	return LookupPath(result, path[1:])
}

func lookupPathMap(data map[string]any, key string) (result any, err error) {
	result, ok := data[key]
	if !ok {
		err = fmt.Errorf("key %q not found", key)
		return
	}

	return result, nil
}

func lookupPathSlice(data []any, key string) (result any, err error) {
	index, err := strconv.Atoi(key)
	if err != nil {
		err = fmt.Errorf("failed to parse %q as int", key)
		return
	}

	if len(data) < index+1 {
		err = fmt.Errorf("index %d out of bounds", index)
		return
	}

	return data[index], nil
}

func lookupPathStruct(data any, key string) (result any, err error) {
	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		err = fmt.Errorf("unsupported type %T", data)
		return
	}

	for i := range val.NumField() {
		fieldType := typ.Field(i)
		fieldName := fieldType.Name

		if fieldName != key {
			continue
		}

		field := val.Field(i)
		if !field.CanInterface() {
			err = fmt.Errorf("failed to assert data to interface")
			return
		}

		return field.Interface(), nil
	}

	err = fmt.Errorf("field %q not found", key)
	return
}
