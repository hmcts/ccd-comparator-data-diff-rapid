package helper

import (
	"encoding/json"
	"fmt"
	"time"
)

const defaultTimeStampLayout = "2006-01-02T15:04:05.999999"

func MustParseTime(layout, value string) time.Time {
	if layout == "" {
		layout = defaultTimeStampLayout
	}

	parsedTime, err := time.Parse(layout, value)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse time: %s", err))
	}
	return parsedTime
}

func FormatTimeStamp(sourceTimeStamp time.Time) string {
	return sourceTimeStamp.Format(defaultTimeStampLayout)

}

func removeIDField(data interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if key == "id" {
				delete(v, key)
			} else {
				removeIDField(val)
			}
		}
	case []interface{}:
		for _, val := range v {
			removeIDField(val)
		}
	}
}

func MustUnmarshal(data []byte, v interface{}) {
	var jsonData interface{}

	// Unmarshal the JSON data
	if err := json.Unmarshal(data, &jsonData); err != nil {
		panic(fmt.Sprintf("error occurred while processing the JSON: %s", err))
	}

	removeIDField(jsonData)

	modifiedData, err := json.Marshal(jsonData)
	if err != nil {
		panic(fmt.Sprintf("error occurred while marshaling JSON: %s", err))
	}

	if err := json.Unmarshal(modifiedData, v); err != nil {
		panic(fmt.Sprintf("error occurred while processing the modified JSON: %s", err))
	}
}
