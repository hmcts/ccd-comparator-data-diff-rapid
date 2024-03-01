package comparator

import (
	"ccd-comparator-data-diff-rapid/config"
	"ccd-comparator-data-diff-rapid/helper"
	"fmt"
	"sort"
	"strings"
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

type DynamicFieldChangeRule struct {
	concurrentEventThresholdMilliseconds int64
	isScanReportMask                     bool
	filter                               []config.Paths
}

func NewStaticFieldChangeRule(concurrentEventThresholdMilliseconds int64, isScanReportMask bool) *StaticFieldChangeRule {
	return &StaticFieldChangeRule{
		concurrentEventThresholdMilliseconds: concurrentEventThresholdMilliseconds,
		isScanReportMask:                     isScanReportMask,
	}
}

func NewDynamicFieldChangeRule(concurrentEventThresholdMilliseconds int64, isScanReportMask bool,
	filter []config.Paths) *DynamicFieldChangeRule {
	return &DynamicFieldChangeRule{
		concurrentEventThresholdMilliseconds: concurrentEventThresholdMilliseconds,
		isScanReportMask:                     isScanReportMask,
		filter:                               filter,
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
							message := generateViolationMessage("SV", fieldName, "", previousChange, r.isScanReportMask,
								preCreatedDate,
								currentChange)
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

var pathsMap = make(map[string][]string)

func (r DynamicFieldChangeRule) CheckForViolation(path string, fieldChanges []EventFieldChange) []Violation {
	var violations []Violation

	if len(pathsMap) == 0 {
		initializePathsMap(r.filter)
	}

	genericFields := pathsMap["*"]

	fields, ok := pathsMap[path]
	if !ok && len(genericFields) == 0 {
		return violations
	}

	fields = append(fields, genericFields...)

	for currentIndex, currentChange := range fieldChanges {
		if currentChange.OperationType != NoChange {
			for previousIndex := 0; previousIndex < currentIndex; previousIndex++ {
				previousChange := fieldChanges[previousIndex]
				if previousChange.OperationType != NoChange {
					for _, field := range fields {
						if compareTargetFieldStrings(currentChange.NewRecord, previousChange.OldRecord, field) {
							timeDifference := currentChange.CreatedDate.Sub(previousChange.CreatedDate).Milliseconds()
							if checkThreshold(r.concurrentEventThresholdMilliseconds, timeDifference) {
								preCreatedDate := helper.FormatTimeStamp(previousChange.CreatedDate)
								message := generateViolationMessage("DF", path, field, previousChange,
									r.isScanReportMask, preCreatedDate,
									currentChange)
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

	return violations
}

func initializePathsMap(paths []config.Paths) {
	for _, path := range paths {
		pathsMap[path.Path] = path.Fields
	}
}

func generateViolationMessage(code, path string, field string, previousChange EventFieldChange,
	isScanReportMask bool, preCreatedDate string, currentChange EventFieldChange) string {
	return fmt.Sprintf("%s:Field '%s' changed to '%s' in event id %d on %s, "+
		"but reverted back to the previous value '%s' in event id %d on %s",
		code, path+"."+field, processInputValue(previousChange.NewRecord, isScanReportMask),
		previousChange.SourceEventId, preCreatedDate,
		processInputValue(currentChange.NewRecord, isScanReportMask), currentChange.SourceEventId,
		helper.FormatTimeStamp(currentChange.CreatedDate))
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

func checkThreshold(thresholdMilliseconds, timeDifference int64) bool {
	return thresholdMilliseconds == -1 || timeDifference <= thresholdMilliseconds
}
func compareTargetFieldStrings(currentNewRecord, previousOldRecord, fieldName string) bool {
	var currentData, previousData interface{}
	helper.MustUnmarshal([]byte(currentNewRecord), &currentData)
	if currentDataSlice, ok := currentData.([]interface{}); ok {
		helper.MustUnmarshal([]byte(previousOldRecord), &previousData)
		if previousDataSlice, ok := previousData.([]interface{}); ok {
			currentFields := extractAndSortFields(currentDataSlice, fieldName)
			previousFields := extractAndSortFields(previousDataSlice, fieldName)
			return currentFields == previousFields
		}
	}
	return false
}

func extractAndSortFields(jsonArray []interface{}, fieldName string) string {
	var links []string

	for _, item := range jsonArray {
		if obj, ok := item.(map[string]interface{}); ok {
			link := extractLink(obj, fieldName)
			if link != "" {
				links = append(links, link)
			}
		}
	}

	sort.Strings(links)

	return strings.Join(links, "_")
}

func extractLink(obj map[string]interface{}, fieldName string) string {
	if fieldValue := findFieldValue(obj, fieldName); fieldValue != nil {
		if link, ok := fieldValue.(string); ok {
			return link
		}
	}
	return ""
}

func findFieldValue(item map[string]interface{}, fieldName string) interface{} {
	for key, value := range item {
		if key == fieldName {
			return value
		}
		if nestedItem, ok := value.(map[string]interface{}); ok {
			if fieldValue := findFieldValue(nestedItem, fieldName); fieldValue != nil {
				return fieldValue
			}
		}
	}
	return nil
}
