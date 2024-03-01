package helper

import (
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
