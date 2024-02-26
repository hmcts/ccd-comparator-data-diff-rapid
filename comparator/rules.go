package comparator

import (
	"ccd-comparator-data-diff-rapid/helper"
	"fmt"
)

type Rule interface {
	CheckForViolation(fieldName string, fieldChanges []EventFieldChange) []Violation
}

type Violation struct {
	sourceEventId            int64
	previousEventCreatedDate string
	previousEventUserId      string
	message                  string
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

func (r SameValueAfterChangeRule) CheckForViolation(fieldName string, fieldChanges []EventFieldChange) []Violation {
	var violations []Violation

	for currentIndex, currentChange := range fieldChanges {
		if currentChange.OperationType != NoChange {
			for previousIndex := 0; previousIndex < currentIndex; previousIndex++ {
				previousChange := fieldChanges[previousIndex]
				if previousChange.OperationType != NoChange {
					if currentChange.NewRecord == previousChange.OldRecord {
						timeDifference := currentChange.CreatedDate.Sub(previousChange.CreatedDate).Milliseconds()
						if r.concurrentEventThresholdMilliseconds == -1 || timeDifference <= r.
							concurrentEventThresholdMilliseconds {
							sourceEventId := currentChange.SourceEventId
							message := fmt.Sprintf("Field '%s' changed to '%s' in event id %d on %s, "+
								"but reverted back to the previous value '%s' in event id %d on %s",
								fieldName, processInputValue(previousChange.NewRecord,
									r.isScanReportMask),
								previousChange.SourceEventId, helper.FormatTimeStamp(previousChange.CreatedDate),
								processInputValue(currentChange.NewRecord, r.isScanReportMask), currentChange.SourceEventId,
								helper.FormatTimeStamp(currentChange.CreatedDate))
							v := Violation{
								sourceEventId:            sourceEventId,
								previousEventCreatedDate: helper.FormatTimeStamp(previousChange.CreatedDate),
								previousEventUserId:      previousChange.UserId,
								message:                  message,
							}
							violations = append(violations, v)
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

func (r FieldChangeCountRule) CheckForViolation(fieldName string, fieldChanges []EventFieldChange) []Violation {
	var violations []Violation

	count := 0
	for _, difference := range fieldChanges {
		count++
		if count > r.fieldChangeThreshold {
			message := fmt.Sprintf("JsonNode field change threshold %d exceeded for field %s.",
				r.fieldChangeThreshold, fieldName)

			v := Violation{
				sourceEventId: difference.SourceEventId,
				message:       message,
			}
			violations = append(violations, v)
		}
	}

	return violations
}
