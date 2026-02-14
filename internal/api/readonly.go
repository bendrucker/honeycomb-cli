package api

import "encoding/json"

// StripReadOnly removes readOnly fields from JSON data for the given schema.
// Returns data unchanged if the schema has no readOnly fields.
func StripReadOnly(data []byte, schema string) ([]byte, error) {
	fields, ok := readOnlyFields[schema]
	if !ok {
		return data, nil
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	for key := range fields {
		delete(m, key)
	}

	return json.Marshal(m)
}

// MarshalStrippingReadOnly marshals v to JSON, then strips readOnly fields for the given schema.
func MarshalStrippingReadOnly(v any, schema string) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return StripReadOnly(data, schema)
}
