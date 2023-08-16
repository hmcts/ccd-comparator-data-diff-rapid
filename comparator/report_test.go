package comparator

import (
	"ccd-comparator-data-diff-rapid/config"
	"ccd-comparator-data-diff-rapid/helper"
	"testing"
	"time"
)

func TestSaveReport(t *testing.T) {
	eventDifferences := make(map[string][]EventFieldDifference)
	analyzeResult := NewAnalyzeResult()
	configurations := &config.Configurations{
		Scan: config.Scan{
			CaseType: "YOUR_CASE_TYPE_ID",
			Report: struct {
				Enabled            bool
				MaskValue          bool
				IncludeEmptyChange bool
			}{true, false, true},
		},
	}

	entities, err := SaveReport(eventDifferences, analyzeResult, configurations)
	if err != nil {
		t.Errorf("TestSaveReport: Expected no error, but got: %v", err)
	}
	if len(entities) > 0 {
		t.Errorf("TestSaveReport: Expected no entities, but got %d", len(entities))
	}

	timeTest := time.Now()
	eventDifferences = map[string][]EventFieldDifference{
		"Case1->Field1": {
			{
				OldRecord:       "old_value",
				NewRecord:       "new_value",
				CreatedDate:     timeTest,
				SourceEventId:   123,
				SourceEventName: "Event1",
				OperationType:   helper.Modified,
			},
		},
	}
	analyzeResult.Put("Case1->Field1", int64(123), helper.FormatTimeStamp(timeTest)+"->Field1 is different")

	configurations.Report.IncludeEmptyChange = false
	entities, err = SaveReport(eventDifferences, analyzeResult, configurations)
	if err != nil {
		t.Errorf("TestSaveReport: Expected no error, but got: %v", err)
	}
	if len(entities) != 1 {
		t.Errorf("TestSaveReport: Expected 1 entity, but got %d", len(entities))
	}

	entity := entities[0]
	if entity.EventID != 123 {
		t.Errorf("TestSaveReport: Expected EventID 123, but got %d", entity.EventID)
	}
	if entity.EventName != "Event1" {
		t.Errorf("TestSaveReport: Expected EventName 'Event1', but got '%s'", entity.EventName)
	}
	if entity.CaseTypeID != "YOUR_CASE_TYPE_ID" { // Replace with your expected case type ID
		t.Errorf("TestSaveReport: Expected CaseTypeID 'YOUR_CASE_TYPE_ID', but got '%s'", entity.CaseTypeID)
	}
	if entity.Reference != "Case1" {
		t.Errorf("TestSaveReport: Expected Reference 'Case1', but got '%s'", entity.Reference)
	}
	if entity.FieldName != "Field1" {
		t.Errorf("TestSaveReport: Expected FieldName 'Field1', but got '%s'", entity.FieldName)
	}
	if entity.ChangeType != string(helper.Modified) {
		t.Errorf("TestSaveReport: Expected ChangeType 'Modified', but got '%s'", entity.ChangeType)
	}
	if entity.OldRecord != "old_value" {
		t.Errorf("TestSaveReport: Expected OldRecord 'old_value', but got '%s'", entity.OldRecord)
	}
	if entity.NewRecord != "new_value" {
		t.Errorf("TestSaveReport: Expected NewRecord 'new_value', but got '%s'", entity.NewRecord)
	}
	if entity.AnalyzeResult != "Field1 is different" {
		t.Errorf("TestSaveReport: Unexpected AnalyzeResult, got '%s'", entity.AnalyzeResult)
	}
	if !entity.PotentialRisk {
		t.Errorf("TestSaveReport: Expected PotentialRisk true, but got false")
	}
}

func TestSaveReportInvalidCombinedReference(t *testing.T) {
	eventDifferences := map[string][]EventFieldDifference{
		"InvalidCombinedReference": {
			{
				OldRecord:       "old_value",
				NewRecord:       "new_value",
				CreatedDate:     time.Now(),
				SourceEventId:   123,
				SourceEventName: "Event1",
				OperationType:   helper.Modified,
			},
		},
	}
	analyzeResult := NewAnalyzeResult()
	analyzeResult.Put("Case1_Field1", int64(123), helper.FormatTimeStamp(time.Now())+"->Field1 is different")
	configurations := &config.Configurations{
		Scan: config.Scan{
			Report: struct {
				Enabled            bool
				MaskValue          bool
				IncludeEmptyChange bool
			}{true, false, true},
		},
	}

	_, err := SaveReport(eventDifferences, analyzeResult, configurations)
	if err == nil {
		t.Error("TestSaveReportInvalidCombinedReference: Expected an error, but got nil")
	}
	expectedErrorMsg := "invalid combinedReference format: InvalidCombinedReference"
	if err.Error() != expectedErrorMsg {
		t.Errorf("TestSaveReportInvalidCombinedReference: Expected error message '%s', but got '%v'", expectedErrorMsg, err)
	}
}
