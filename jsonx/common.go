package jsonx

import (
	"encoding/json"
	"fmt"
)

func removeIDNullAndEmpty(data any) any {
	switch v := data.(type) {
	case map[string]any:
		for key, val := range v {
			if key == "id" {
				delete(v, key)
			} else if val == nil || isEmptyArray(val) || isEmptyMap(val) || val == "" {
				delete(v, key)
			} else {
				v[key] = removeIDNullAndEmpty(val)
			}
		}
	case []any:
		if len(v) == 0 {
			return nil
		}
		for i, val := range v {
			v[i] = removeIDNullAndEmpty(val)
		}
	}
	return data
}

func isEmptyArray(val any) bool {
	arr, ok := val.([]any)
	return ok && len(arr) == 0
}

func isEmptyMap(val any) bool {
	m, ok := val.(map[string]any)
	return ok && len(m) == 0
}

func MustUnmarshal(data []byte, v any) {
	var jsonData any

	// Unmarshal the JSON data
	if err := json.Unmarshal(data, &jsonData); err != nil {
		panic(fmt.Sprintf("error occurred while processing the JSON: %s", err))
	}

	removeIDNullAndEmpty(jsonData)

	modifiedData := MustMarshal(jsonData)

	if err := json.Unmarshal(modifiedData, v); err != nil {
		panic(fmt.Sprintf("error occurred while processing the modified JSON: %s", err))
	}
}

func MustMarshal(data any) []byte {
	result, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("error occurred while marshaling JSON: %s", err))
	}

	return result
}
