package comparator

import (
	"ccd-comparator-data-diff-rapid/helper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEventDifferencesData_ProcessEventDiff(t *testing.T) {
	rule1 := NewSameValueAfterChangeRule(0, false)
	rule2 := NewFieldChangeCountRule(0)
	activeRules := []Rule{rule1, rule2}

	timeNow := time.Now()
	fieldDifferences1 := []EventFieldChange{
		{
			OldRecord:       "old_record1",
			NewRecord:       "new_record1",
			CreatedDate:     timeNow,
			SourceEventId:   1,
			SourceEventName: "event1",
			OperationType:   helper.Modified,
		},
		{
			OldRecord:       "new_record1",
			NewRecord:       "old_record1",
			CreatedDate:     timeNow,
			SourceEventId:   1,
			SourceEventName: "event1",
			OperationType:   helper.Modified,
		},
	}
	fieldDifferences2 := []EventFieldChange{
		{
			OldRecord:       "old_record2",
			NewRecord:       "new_record2",
			CreatedDate:     timeNow,
			SourceEventId:   2,
			SourceEventName: "event2",
			OperationType:   helper.Modified,
		},
	}

	eventDifferences := map[string][]EventFieldChange{
		"combined_reference_1->field1": fieldDifferences1,
		"combined_reference_2->field2": fieldDifferences2,
	}

	eventData := NewEventChangesAnalyze(&activeRules, eventDifferences)
	analyzeResult := eventData.AnalyzeEventFieldChanges()

	if analyzeResult.IsEmpty() {
		t.Error("Analyze result should not be empty.")
	}
	expectedSize := 2
	if analyzeResult.Size() != expectedSize {
		t.Errorf("Expected size : %d, but got: %d", expectedSize, analyzeResult.Size())
	}

	expectedResultMessage1 := helper.FormatTimeStamp(timeNow) + "->field field1 changed in event id 1 with the value new_record1 on " + helper.FormatTimeStamp(timeNow) + ". " +
		"The field field1 was updated back to old value old_record1 in event id 1 on " + helper.FormatTimeStamp(timeNow) + "\nJsonNode field change " +
		"threshold 0 exceeded for field field1."
	if expectedResultMessage1 != analyzeResult.Get("combined_reference_1->field1", 1) {
		t.Errorf("Incorrect result message. Expected message: %s, but got: %s", expectedResultMessage1,
			analyzeResult.Get("combined_reference_1->field1", 1))
	}
}

func TestEventDifferencesData_AppendMessages(t *testing.T) {
	existingMessage := "Existing Message"
	newMessage := "New Message"

	pair := &Pair[int64, string]{Left: 123, Right: newMessage}
	result := appendMessages("", pair.Right)
	assert.Equal(t, newMessage, result, "Appended message should be equal to the new message")

	// Test when existingMessage is not empty
	pair = &Pair[int64, string]{Left: 123, Right: newMessage}
	result = appendMessages(existingMessage, pair.Right)
	assert.Equal(t, existingMessage+"\n"+newMessage, result, "Appended message should have both existing and new messages")
}
