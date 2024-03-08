package comparator

import (
	"ccd-comparator-data-diff-rapid/helper"
	"ccd-comparator-data-diff-rapid/jsonx"
	"fmt"
	"time"
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

type StaticFieldChangeRule struct {
	concurrentEventThresholdMilliseconds int64
	isScanReportMask                     bool
}

type FieldChangeCountRule struct {
	fieldChangeThreshold int
}

type ArrayFieldChangeRule struct {
	concurrentEventThresholdMilliseconds int64
	isScanReportMask                     bool
}

func NewStaticFieldChangeRule(concurrentEventThresholdMilliseconds int64, isScanReportMask bool) *StaticFieldChangeRule {
	return &StaticFieldChangeRule{
		concurrentEventThresholdMilliseconds: concurrentEventThresholdMilliseconds,
		isScanReportMask:                     isScanReportMask,
	}
}

func NewArrayFieldChangeRule(concurrentEventThresholdMilliseconds int64, isScanReportMask bool) *ArrayFieldChangeRule {
	return &ArrayFieldChangeRule{
		concurrentEventThresholdMilliseconds: concurrentEventThresholdMilliseconds,
		isScanReportMask:                     isScanReportMask,
	}
}

func NewFieldChangeCountRule(fieldChangeThreshold int) *FieldChangeCountRule {
	return &FieldChangeCountRule{fieldChangeThreshold}
}

func (r StaticFieldChangeRule) CheckForViolation(fieldName string, fieldChanges []EventFieldChange) []Violation {
	var violations []Violation

	for currentIndex, currentChange := range fieldChanges {
		if currentChange.OperationType != NoChange {
			for previousIndex := 0; previousIndex < currentIndex; previousIndex++ {
				previousChange := fieldChanges[previousIndex]
				if previousChange.OperationType != NoChange {
					if currentChange.NewRecord == previousChange.OldRecord {
						timeDifference := currentChange.CreatedDate.Sub(previousChange.CreatedDate).Milliseconds()
						if checkThreshold(r.concurrentEventThresholdMilliseconds, timeDifference) {
							preCreatedDate := helper.FormatTimeStamp(previousChange.CreatedDate)
							message := generateViolationMessage("SV", fieldName, previousChange.NewRecord,
								previousChange.SourceEventId, r.isScanReportMask, preCreatedDate,
								currentChange.NewRecord, currentChange.SourceEventId, currentChange.CreatedDate)
							v := Violation{
								sourceEventId:            currentChange.SourceEventId,
								previousEventCreatedDate: preCreatedDate,
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

func (f FieldChangeCountRule) CheckForViolation(fieldName string, fieldChanges []EventFieldChange) []Violation {
	var violations []Violation

	count := 0
	for _, difference := range fieldChanges {
		count++
		if count > f.fieldChangeThreshold {
			message := fmt.Sprintf("JsonNode field change threshold %d exceeded for field %s.",
				f.fieldChangeThreshold, fieldName)

			v := Violation{
				sourceEventId: difference.SourceEventId,
				message:       message,
			}
			violations = append(violations, v)
		}
	}

	return violations
}

func (a ArrayFieldChangeRule) CheckForViolation(path string, fieldChanges []EventFieldChange) []Violation {
	var violations []Violation

	for currentIndex, currentChange := range fieldChanges {
		if currentChange.OperationType.IsArrayOperation() {
			var currentArray []jsonx.Change
			jsonx.MustUnmarshal([]byte(currentChange.NewRecord), &currentArray)

			for previousIndex := 0; previousIndex < currentIndex; previousIndex++ {
				previousChange := fieldChanges[previousIndex]
				if previousChange.OperationType.IsArrayOperation() {
					var previousArray []jsonx.Change
					jsonx.MustUnmarshal([]byte(previousChange.NewRecord), &previousArray)

					for _, currentItem := range currentArray {
						for _, previousItem := range previousArray {
							if isCrossCheckViolation(currentItem, previousItem) {
								timeDifference := currentChange.CreatedDate.Sub(previousChange.CreatedDate).Milliseconds()
								if checkThreshold(a.concurrentEventThresholdMilliseconds, timeDifference) {
									preCreatedDate := helper.FormatTimeStamp(previousChange.CreatedDate)
									message := fmt.Sprintf("%s:Field '%s':'%s' %s in event id %d on %s, "+
										"but '%s' %s in event id %d on %s",
										"AF", path, processInputValue(previousItem.Value,
											a.isScanReportMask), previousItem.ChangeType(),
										previousChange.SourceEventId, preCreatedDate,
										processInputValue(currentItem.Value, a.isScanReportMask),
										currentItem.ChangeType(),
										currentChange.SourceEventId, helper.FormatTimeStamp(currentChange.CreatedDate))

									v := Violation{
										sourceEventId:            currentChange.SourceEventId,
										previousEventCreatedDate: preCreatedDate,
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
		}
	}

	return violations
}

func isCrossCheckViolation(current jsonx.Change, previous jsonx.Change) bool {
	return current.Compare(previous) && current.HasCrossMatch(previous)
}

func generateViolationMessage(code, path, previousChangeRecord string, previousChangeEventId int64,
	isScanReportMask bool, preCreatedDate string, currentChangeRecord string, currentChangeEventId int64,
	currentChangeDate time.Time) string {
	return fmt.Sprintf("%s:Field '%s' changed to '%s' in event id %d on %s, "+
		"but reverted back to the previous value '%s' in event id %d on %s",
		code, path, processInputValue(previousChangeRecord, isScanReportMask),
		previousChangeEventId, preCreatedDate,
		processInputValue(currentChangeRecord, isScanReportMask), currentChangeEventId,
		helper.FormatTimeStamp(currentChangeDate))
}

func processInputValue(input string, isScanReportMask bool) string {
	if isScanReportMask {
		return "***"
	}

	maxLength := 250
	if len(input) > maxLength {
		return input[:maxLength]
	}
	return input
}

func checkThreshold(thresholdMilliseconds, timeDifference int64) bool {
	return thresholdMilliseconds == -1 || timeDifference <= thresholdMilliseconds
}
