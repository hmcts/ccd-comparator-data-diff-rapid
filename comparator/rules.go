package comparator

import (
	"ccd-comparator-data-diff-rapid/helper"
	"fmt"
)

type Rule interface {
	CheckForViolation(fieldName string, fieldChanges []EventFieldChange) []*Pair[int64, string]
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

func (r SameValueAfterChangeRule) CheckForViolation(fieldName string, fieldChanges []EventFieldChange) []*Pair[int64,
	string] {
	var violations []*Pair[int64, string]

	for currentIndex, currentChange := range fieldChanges {
		if currentChange.OperationType != helper.NoChange {
			for previousIndex := 0; previousIndex < currentIndex; previousIndex++ {
				previousChange := fieldChanges[previousIndex]
				if previousChange.OperationType != helper.NoChange {
					if currentChange.NewRecord == previousChange.OldRecord {
						timeDifference := currentChange.CreatedDate.Sub(previousChange.CreatedDate).Milliseconds()
						if r.concurrentEventThresholdMilliseconds == -1 || timeDifference <= r.
							concurrentEventThresholdMilliseconds {
							sourceEventId := currentChange.SourceEventId
							message := fmt.Sprintf("Field '%s' changed to '%s' in event id %d on %s, "+
								"but reverted back to the previous value '%s' in event id %d on %s",
								trimInputValue(fieldName), processInputValue(previousChange.NewRecord,
									r.isScanReportMask),
								previousChange.SourceEventId, helper.FormatTimeStamp(previousChange.CreatedDate),
								processInputValue(currentChange.NewRecord, r.isScanReportMask), currentChange.SourceEventId,
								helper.FormatTimeStamp(currentChange.CreatedDate))
							violations = append(violations, NewPair(sourceEventId, message))
						}
					}
				}
			}
		}
	}

	return violations
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

func trimInputValue(input string) string {
	maxLength := 50
	if len(input) > maxLength {
		return input[:maxLength]
	}
	return input
}

func (r FieldChangeCountRule) CheckForViolation(fieldName string,
	fieldDifferences []EventFieldChange) []*Pair[int64, string] {
	var violations []*Pair[int64, string]

	count := 0
	for _, difference := range fieldDifferences {
		count++
		if count > r.fieldChangeThreshold {
			message := fmt.Sprintf("JsonNode field change threshold %d exceeded for field %s.",
				r.fieldChangeThreshold, fieldName)

			violations = append(violations, NewPair(difference.SourceEventId, message))
		}
	}

	return violations
}
