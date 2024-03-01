package comparator

import (
	"ccd-comparator-data-diff-rapid/helper"
	"fmt"
	"testing"
	"time"
)

const ruleSetLayout = "2006-01-02T15:04:05.000"

func TestSameValueAfterChangeRule_CreateViolationInThreshold(t *testing.T) {
	createdDateBase := helper.MustParseTime(ruleSetLayout, "2023-01-01T00:00:00.000")
	createdDateBaseSecond := createdDateBase.Add(10000 * time.Millisecond)
	createdDateBaseThird := createdDateBase.Add(19000 * time.Millisecond)

	fieldDifferences := []EventFieldChange{
		{
			OldRecord:       "value1",
			NewRecord:       "value1",
			CreatedDate:     createdDateBase,
			SourceEventId:   1,
			SourceEventName: "Event1",
			OperationType:   Modified,
		},
		{
			OldRecord:       "value1",
			NewRecord:       "value2",
			CreatedDate:     createdDateBaseSecond,
			SourceEventId:   2,
			SourceEventName: "Event2",
			OperationType:   Modified,
		},
		{
			OldRecord:       "value2",
			NewRecord:       "value1",
			CreatedDate:     createdDateBaseThird,
			SourceEventId:   3,
			SourceEventName: "Event3",
			OperationType:   Modified,
		},
	}

	rule := NewStaticFieldChangeRule(10000, false)
	fieldName := "field1"
	result := rule.CheckForViolation(fieldName, fieldDifferences)

	if result == nil {
		t.Errorf("Expected 1 violation, but got 0")
	}

	expectedSourceEventId := 3
	expectedMessage := helper.FormatTimeStamp(createdDateBaseSecond) + "->field field1 changed in event id 2 with the value value2 on " +
		helper.FormatTimeStamp(createdDateBaseSecond) + ". The field field1 was updated back to old value value1 " +
		"in event id 3 on " + helper.FormatTimeStamp(createdDateBaseThird)

	if result[0].sourceEventId != 3 {
		t.Errorf("Incorrect SourceEventId. Expected: %d, Got: %d", expectedSourceEventId, result[0].sourceEventId)
	}
	if result[0].message != expectedMessage {
		t.Errorf("Incorrect violation message. Expected: %s, Got: %s", expectedMessage, result[0].message)
	}
}

func TestSameValueAfterChangeRule_IgnoreViolationExceedThreshold(t *testing.T) {
	createdDateBase := helper.MustParseTime(ruleSetLayout, "2023-01-01T00:00:00.000")
	createdDateBaseSecond := createdDateBase.Add(10 * time.Second)
	createdDateBaseThird := createdDateBase.Add(21 * time.Second)

	fieldDifferences := []EventFieldChange{
		{
			OldRecord:       "value1",
			NewRecord:       "value1",
			CreatedDate:     createdDateBase,
			SourceEventId:   1,
			SourceEventName: "Event1",
			OperationType:   Modified,
		},
		{
			OldRecord:       "value1",
			NewRecord:       "value2",
			CreatedDate:     createdDateBaseSecond,
			SourceEventId:   2,
			SourceEventName: "Event2",
			OperationType:   Modified,
		},
		{
			OldRecord:       "value2",
			NewRecord:       "value1",
			CreatedDate:     createdDateBaseThird,
			SourceEventId:   3,
			SourceEventName: "Event3",
			OperationType:   Modified,
		},
	}

	rule := NewStaticFieldChangeRule(10, false)
	fieldName := "field1"
	result := rule.CheckForViolation(fieldName, fieldDifferences)

	if result != nil {
		t.Errorf("Expected 0 violation, but got 1")
	}
}

func TestSameValueAfterChangeRule_ViolationWith0Threshold(t *testing.T) {
	createdDateBase := helper.MustParseTime(ruleSetLayout, "2023-01-01T00:00:00.000")
	createdDateBaseSecond := createdDateBase.Add(10 * time.Second)
	createdDateBaseThird := createdDateBase.Add(120 * time.Second)

	fieldDifferences := []EventFieldChange{
		{
			OldRecord:       "value1",
			NewRecord:       "value1",
			CreatedDate:     createdDateBase,
			SourceEventId:   1,
			SourceEventName: "Event1",
			OperationType:   Modified,
		},
		{
			OldRecord:       "value1",
			NewRecord:       "value2",
			CreatedDate:     createdDateBaseSecond,
			SourceEventId:   2,
			SourceEventName: "Event2",
			OperationType:   Modified,
		},
		{
			OldRecord:       "value2",
			NewRecord:       "value1",
			CreatedDate:     createdDateBaseThird,
			SourceEventId:   3,
			SourceEventName: "Event3",
			OperationType:   Modified,
		},
	}

	rule := NewStaticFieldChangeRule(0, false)
	fieldName := "field1"
	result := rule.CheckForViolation(fieldName, fieldDifferences)

	if result == nil {
		t.Errorf("Expected 1 violation, but got 0")
	}

	expectedSourceEventId := 3
	expectedMessage := helper.FormatTimeStamp(createdDateBaseSecond) + "->field field1 changed in event id 2 with the value value2 on " +
		helper.FormatTimeStamp(createdDateBaseSecond) + ". The field field1 was updated back to old value value1 " +
		"in event id 3 on " + helper.FormatTimeStamp(createdDateBaseThird)

	if result[0].sourceEventId != 3 {
		t.Errorf("Incorrect SourceEventId. Expected: %d, Got: %d", expectedSourceEventId, result[0].sourceEventId)
	}
	if result[0].message != expectedMessage {
		t.Errorf("Incorrect violation message. Expected: %s, Got: %s", expectedMessage, result[0].message)
	}
}

func TestSameValueAfterChangeRule_CreateViolationInThresholdAndMask(t *testing.T) {
	createdDateBase := helper.MustParseTime(ruleSetLayout, "2023-01-01T00:00:00.000")
	createdDateBaseSecond := createdDateBase.Add(10000 * time.Millisecond)
	createdDateBaseThird := createdDateBase.Add(19000 * time.Millisecond)

	fieldDifferences := []EventFieldChange{
		{
			OldRecord:       "value1",
			NewRecord:       "value1",
			CreatedDate:     createdDateBase,
			SourceEventId:   1,
			SourceEventName: "Event1",
			OperationType:   Modified,
		},
		{
			OldRecord:       "value1",
			NewRecord:       "value2",
			CreatedDate:     createdDateBaseSecond,
			SourceEventId:   2,
			SourceEventName: "Event2",
			OperationType:   Modified,
		},
		{
			OldRecord:       "value2",
			NewRecord:       "value1",
			CreatedDate:     createdDateBaseThird,
			SourceEventId:   3,
			SourceEventName: "Event3",
			OperationType:   Modified,
		},
	}

	rule := NewStaticFieldChangeRule(10000, true)
	fieldName := "field1"
	result := rule.CheckForViolation(fieldName, fieldDifferences)

	if result == nil {
		t.Errorf("Expected 1 violation, but got 0")
	}

	expectedSourceEventId := 3
	expectedMessage := helper.FormatTimeStamp(createdDateBaseSecond) + "->field field1 changed in event id 2 with the value *** on " +
		helper.FormatTimeStamp(createdDateBaseSecond) + ". The field field1 was updated back to old value *** " +
		"in event id 3 on " + helper.FormatTimeStamp(createdDateBaseThird)

	if result[0].sourceEventId != 3 {
		t.Errorf("Incorrect SourceEventId. Expected: %d, Got: %d", expectedSourceEventId, result[0].sourceEventId)
	}
	if result[0].message != expectedMessage {
		t.Errorf("Incorrect violation message. Expected: %s, Got: %s", expectedMessage, result[0].message)
	}
}

func TestSameValueAfterChangeRule_CreateViolationInThresholdOldValueEmpty(t *testing.T) {
	createdDateBase := helper.MustParseTime(ruleSetLayout, "2023-01-01T00:00:00.000")
	createdDateBaseSecond := createdDateBase.Add(10000 * time.Millisecond)
	createdDateBaseThird := createdDateBase.Add(19000 * time.Millisecond)

	value25Char := "value1value1value1value1value1"
	fieldDifferences := []EventFieldChange{
		{
			OldRecord:       "",
			NewRecord:       value25Char,
			CreatedDate:     createdDateBase,
			SourceEventId:   1,
			SourceEventName: "Event1",
			OperationType:   Modified,
		},
		{
			OldRecord:       value25Char,
			NewRecord:       "value2",
			CreatedDate:     createdDateBaseSecond,
			SourceEventId:   2,
			SourceEventName: "Event2",
			OperationType:   Deleted,
		},
		{
			OldRecord:       "value2",
			NewRecord:       value25Char,
			CreatedDate:     createdDateBaseThird,
			SourceEventId:   3,
			SourceEventName: "Event3",
			OperationType:   Modified,
		},
	}

	rule := NewStaticFieldChangeRule(10000, false)
	fieldName := "field1"
	result := rule.CheckForViolation(fieldName, fieldDifferences)

	if result == nil {
		t.Errorf("Expected 1 violation, but got 0")
	}

	expectedSourceEventId := 3
	expectedMessage := helper.FormatTimeStamp(createdDateBaseSecond) + "->field field1 changed in event id 2 with the value value2 on " +
		helper.FormatTimeStamp(createdDateBaseSecond) + ". The field field1 was updated back to old value value1value1value1value1v " +
		"in event id 3 on " + helper.FormatTimeStamp(createdDateBaseThird)

	if result[0].sourceEventId != int64(expectedSourceEventId) {
		t.Errorf("Incorrect SourceEventId. Expected: %d, Got: %d", expectedSourceEventId, result[0].sourceEventId)
	}
	if result[0].message != expectedMessage {
		t.Errorf("Incorrect violation message. Expected: %s, Got: %s", expectedMessage, result[0].message)
	}
}

func TestFieldChangeCountRule_CheckForViolation(t *testing.T) {
	rule := NewFieldChangeCountRule(3)

	fieldName := "myField"
	differences := []EventFieldChange{
		{SourceEventId: 1, OldRecord: "oldValue1", NewRecord: "newValue1"},
		{SourceEventId: 2, OldRecord: "oldValue2", NewRecord: "newValue2"},
		{SourceEventId: 3, OldRecord: "oldValue3", NewRecord: "newValue3"},
	}

	result := rule.CheckForViolation(fieldName, differences)
	if result != nil {
		t.Errorf("Expected nil, but got %v", result)
	}

	differences = append(differences, EventFieldChange{SourceEventId: 4, OldRecord: "oldValue4", NewRecord: "newValue4"})
	result = rule.CheckForViolation(fieldName, differences)
	expectedMessage := fmt.Sprintf("JsonNode field change threshold %d exceeded for field %s.", rule.fieldChangeThreshold, fieldName)
	expectedSourceEventId := 4
	if result[0].sourceEventId != 4 {
		t.Errorf("Incorrect SourceEventId. Expected: %d, Got: %d", expectedSourceEventId, result[0].sourceEventId)
	}
	if result[0].message != expectedMessage {
		t.Errorf("Incorrect violation message. Expected: %s, Got: %s", expectedMessage, result[0].message)
	}
}
