package helper

import (
	"encoding/json"
	"fmt"
	"time"
)

type OperationType string

const (
	Added         OperationType = "ADDED"
	Deleted       OperationType = "DELETED"
	Modified      OperationType = "MODIFIED"
	ArrayModified OperationType = "ARRAY_MODIFIED"
	ArrayExtended OperationType = "ARRAY_EXTENDED"
	ArrayShrunk   OperationType = "ARRAY_SHRUNK"
	NoChange      OperationType = "NO_CHANGE"
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

func MustUnmarshal(data []byte, v interface{}) {
	if err := json.Unmarshal(data, v); err != nil {
		panic(fmt.Sprintf("err occurred while processing the JSON: %s", err))
	}
}
