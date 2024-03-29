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
	previousEventId          int64
	previousEventCreatedDate string
	previousEventUserId      string
	ruleType                 RuleType
	message                  string
	previousEventName        string
}

type StaticFieldChangeRule struct {
	concurrentEventTimeLimit int64
	isScanReportMask         bool
	ruleType                 RuleType
}

type FieldChangeCountRule struct {
	fieldChangeThreshold int
	ruleType             RuleType
}

type ArrayFieldChangeRule struct {
	concurrentEventTimeLimit int64
	isScanReportMask         bool
	searchStartTime          time.Time
	ruleType                 RuleType
}

func NewStaticFieldChangeRule(concurrentEventTimeLimit int64, isScanReportMask bool) *StaticFieldChangeRule {
	return &StaticFieldChangeRule{
		concurrentEventTimeLimit: concurrentEventTimeLimit,
		isScanReportMask:         isScanReportMask,
		ruleType:                 RuleTypeStaticFieldChange,
	}
}

func NewArrayFieldChangeRule(concurrentEventTimeLimit int64, isScanReportMask bool,
	searchStartTime time.Time) *ArrayFieldChangeRule {
	return &ArrayFieldChangeRule{
		concurrentEventTimeLimit: concurrentEventTimeLimit,
		isScanReportMask:         isScanReportMask,
		searchStartTime:          searchStartTime,
		ruleType:                 RuleTypeArrayFieldChange,
	}
}

func NewFieldChangeCountRule(fieldChangeThreshold int) *FieldChangeCountRule {
	return &FieldChangeCountRule{fieldChangeThreshold, RuleTypeFieldChangeCount}
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
						if checkThreshold(r.concurrentEventTimeLimit, timeDifference) {
							preCreatedDate := helper.FormatTimeStamp(previousChange.CreatedDate)
							message := fmt.Sprintf("Field '%s' changed to '%s' in event id %d on %s, "+
								"but reverted back to the previous value '%s' in event id %d on %s",
								fieldName, processInputValue(previousChange.NewRecord, r.isScanReportMask),
								previousChange.SourceEventId, preCreatedDate,
								processInputValue(currentChange.NewRecord, r.isScanReportMask), currentChange.SourceEventId,
								helper.FormatTimeStamp(currentChange.CreatedDate))
							v := Violation{
								sourceEventId:            currentChange.SourceEventId,
								previousEventId:          previousChange.SourceEventId,
								previousEventCreatedDate: preCreatedDate,
								previousEventUserId:      previousChange.UserId,
								previousEventName:        previousChange.SourceEventName,
								ruleType:                 r.ruleType,
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
				ruleType:      f.ruleType,
			}
			violations = append(violations, v)
		}
	}

	return violations
}

func (a ArrayFieldChangeRule) CheckForViolation(path string, fieldChanges []EventFieldChange) []Violation {
	var violations []Violation

	for currentIndex := len(fieldChanges) - 1; currentIndex > 0; currentIndex-- {
		currentChange := fieldChanges[currentIndex]

		// stop comparing if the change date is older than the search start date
		if currentChange.CreatedDate.Before(a.searchStartTime) {
			break
		}

		if currentChange.OperationType.IsArrayOperation() {
			for previousIndex := currentIndex - 1; previousIndex >= 0; previousIndex-- {
				previousChange := fieldChanges[previousIndex]

				timeDifference := currentChange.CreatedDate.Sub(previousChange.CreatedDate).Milliseconds()
				if !previousChange.OperationType.IsArrayOperation() || !checkThreshold(a.concurrentEventTimeLimit,
					timeDifference) {
					continue
				}

				var currentArray, previousArray []jsonx.Change
				jsonx.MustUnmarshal([]byte(currentChange.NewRecord), &currentArray)
				jsonx.MustUnmarshal([]byte(previousChange.NewRecord), &previousArray)

				for _, currentItem := range currentArray {
					for _, previousItem := range previousArray {
						if isCrossCheckViolation(currentItem, previousItem) {
							preCreatedDate := helper.FormatTimeStamp(previousChange.CreatedDate)
							message := fmt.Sprintf("Field '%s':'%s' %s in event id %d on %s, "+
								"but '%s' %s in event id %d on %s",
								path, processInputValue(previousItem.Value,
									a.isScanReportMask), previousItem.ChangeType(),
								previousChange.SourceEventId, preCreatedDate,
								processInputValue(currentItem.Value, a.isScanReportMask),
								currentItem.ChangeType(),
								currentChange.SourceEventId, helper.FormatTimeStamp(currentChange.CreatedDate))

							v := Violation{
								sourceEventId:            currentChange.SourceEventId,
								previousEventId:          previousChange.SourceEventId,
								previousEventCreatedDate: preCreatedDate,
								previousEventUserId:      previousChange.UserId,
								previousEventName:        previousChange.SourceEventName,
								ruleType:                 a.ruleType,
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

func isCrossCheckViolation(current jsonx.Change, previous jsonx.Change) bool {
	return current.Compare(previous) && current.HasCrossMatch(previous)
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
