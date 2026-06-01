package output

import (
	"fmt"
	"reflect"
	"strconv"
)

// TableFromTags derives a TableDef from the `col:"Header"` struct tags on T, in
// field order. Each column reads its field by reflection and formats it the
// same way the hand-written closures did (strings as-is, bools via FormatBool,
// integers and floats via strconv). Fields without a `col` tag are skipped, so
// computed columns can be appended to the returned TableDef explicitly.
func TableFromTags[T any]() TableDef {
	t := reflect.TypeFor[T]()
	var columns []Column
	for i := range t.NumField() {
		header, ok := t.Field(i).Tag.Lookup("col")
		if !ok {
			continue
		}
		field := i
		columns = append(columns, Column{
			Header: header,
			Value:  func(v any) string { return formatField(reflect.ValueOf(v).Field(field)) },
		})
	}
	return TableDef{Columns: columns}
}

// FieldsFromTags derives detail Fields from the `detail:"Label"` struct tags on
// v, in field order. Fields without a `detail` tag are skipped, so computed or
// conditional fields can be appended by the caller.
func FieldsFromTags(v any) []Field {
	rv := reflect.ValueOf(v)
	t := rv.Type()
	var fields []Field
	for i := range t.NumField() {
		label, ok := t.Field(i).Tag.Lookup("detail")
		if !ok {
			continue
		}
		fields = append(fields, Field{Label: label, Value: formatField(rv.Field(i))})
	}
	return fields
}

func formatField(rv reflect.Value) string {
	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	default:
		return fmt.Sprint(rv.Interface())
	}
}
