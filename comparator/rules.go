package comparator

import (
	"ccd-comparator-data-diff-rapid/helper"
	"fmt"
	"time"
)

type Rule interface {
	CheckForViolation(fieldName string, fieldChanges []EventFieldChange) *Pair[int64, string]
}

type SameValueAfterChangeRule struct {
	concurrentEventThresholdMilliseconds int64
	isScanReportMask                     bool
}

type FieldChangeCountRule struct {
	fieldChangeThreshold int
}

func NewSameValueAfterChangeRule(concurrentEventThresholdMilliseconds int64, isScanReportMask bool) *SameValueAfterChangeRule {
	return &SameValueAfterChangeRule{
		concurrentEventThresholdMilliseconds: concurrentEventThresholdMilliseconds,
		isScanReportMask:                     isScanReportMask,
	}
}

func NewFieldChangeCountRule(fieldChangeThreshold int) *FieldChangeCountRule {
	return &FieldChangeCountRule{fieldChangeThreshold}
}

func (r SameValueAfterChangeRule) CheckForViolation(fieldName string,
	fieldChanges []EventFieldChange) *Pair[int64, string] {

	var previousDate time.Time
	var previousValue = "\n\r"
	var previousSourceEventId int64

	for _, difference := range fieldChanges {
		oldValue := difference.OldRecord
		newValue := difference.NewRecord
		currentDate := difference.CreatedDate

		if newValue == previousValue && !previousDate.IsZero() &&
			(r.concurrentEventThresholdMilliseconds == 0 ||
				currentDate.Sub(previousDate).Milliseconds() <= r.concurrentEventThresholdMilliseconds) {

			sourceEventId := difference.SourceEventId
			message := fmt.Sprintf("field %s changed in event id %d with the value %s on %s. "+
				"The field %s was updated back to old value %s in event id %d on %s",
				fieldName, previousSourceEventId, processInputValue(oldValue, r.isScanReportMask),
				helper.FormatTimeStamp(previousDate), fieldName, processInputValue(newValue, r.isScanReportMask),
				sourceEventId, helper.FormatTimeStamp(currentDate))

			return NewPair(sourceEventId, helper.FormatTimeStamp(previousDate)+"->"+message)
		}

		previousSourceEventId = difference.SourceEventId
		previousValue = oldValue
		previousDate = currentDate
	}

	return nil
}

func processInputValue(input string, isScanReportMask bool) string {
	if isScanReportMask {
		return "***"
	}

	maxLength := 25
	if len(input) > maxLength {
		return input[:maxLength]
	}
	return input
}

func (r FieldChangeCountRule) CheckForViolation(fieldName string,
	fieldDifferences []EventFieldChange) *Pair[int64, string] {

	count := 0
	for _, difference := range fieldDifferences {
		count++
		if count > r.fieldChangeThreshold {
			message := fmt.Sprintf("JsonNode field change threshold %d exceeded for field %s.",
				r.fieldChangeThreshold, fieldName)

			return NewPair(difference.SourceEventId, message)
		}
	}

	return nil
}
