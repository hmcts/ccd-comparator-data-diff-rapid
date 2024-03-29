package comparator

import (
	"bytes"
	"ccd-comparator-data-diff-rapid/config"
	"ccd-comparator-data-diff-rapid/helper"
	"fmt"
	"strings"
	"time"
)

type EventDataReportEntity struct {
	Id                       int64         `db:"id"`
	EventId                  int64         `db:"event_id"`
	PreviousEventId          int64         `db:"previous_event_id"`
	EventName                string        `db:"event_name"`
	PreviousEventName        string        `db:"previous_event_name"`
	CaseTypeId               string        `db:"case_type_id"`
	Reference                string        `db:"reference"`
	FieldName                string        `db:"field_name"`
	ChangeType               string        `db:"change_type"`
	OldRecord                string        `db:"old_record"`
	NewRecord                string        `db:"new_record"`
	ArrayChangeRecord        string        `db:"array_change_record"`
	PreviousEventCreatedDate time.Time     `db:"previous_event_created_date"`
	EventCreatedDate         time.Time     `db:"event_created_date"`
	AnalyzeResult            string        `db:"analyze_result"`
	RuleMatched              bool          `db:"rule_matched"`
	PreviousEventUserId      string        `db:"previous_event_user_id"`
	EventUserId              string        `db:"event_user_id"`
	EventDelta               time.Duration `db:"event_delta"`
}

func PrepareReportEntities(eventDifferences map[string][]EventFieldChange, analyzeResult *AnalyzeResult,
	configurations *config.Configurations) ([]EventDataReportEntity, error) {

	if analyzeResult.IsEmpty() && !configurations.Report.IncludeEmptyChange {
		return nil, nil
	}

	var eventDataReportEntities []EventDataReportEntity

	for combinedReference, fieldDifferences := range eventDifferences {
		parts := strings.Split(combinedReference, "->")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid combinedReference format: %s", combinedReference)
		}
		caseReference := parts[0]
		fieldName := parts[1]

		var changeIndex int
		for i, eventFieldDiff := range fieldDifferences {
			if configurations.Report.IncludeNoChange || eventFieldDiff.OperationType != NoChange {
				violation := analyzeResult.Get(combinedReference, eventFieldDiff.SourceEventId)

				var previousEventCreatedDate time.Time
				var previousUserId string
				var previousEventId int64
				var previousEventName string
				var message string
				var delta time.Duration
				isArrayChange := false

				if violation.sourceEventId != 0 {
					previousEventCreatedDate = helper.MustParseTime("", violation.previousEventCreatedDate)
					previousUserId = violation.previousEventUserId
					previousEventId = violation.previousEventId
					previousEventName = violation.previousEventName
					message = string(violation.ruleType) + ":" + violation.message
					delta = time.Duration(eventFieldDiff.CreatedDate.Sub(previousEventCreatedDate).Milliseconds())
					if violation.ruleType == RuleTypeArrayFieldChange {
						isArrayChange = true
					}
				} else if i > 0 {
					previousChange := fieldDifferences[changeIndex]
					previousEventCreatedDate = previousChange.CreatedDate
					previousUserId = previousChange.UserId
					previousEventId = previousChange.SourceEventId
					previousEventName = previousChange.SourceEventName
					delta = time.Duration(eventFieldDiff.CreatedDate.Sub(previousEventCreatedDate).Milliseconds())
				}

				changeIndex = i

				if configurations.Report.IncludeEmptyChange || message != "" {
					var oldRecord, newRecord string
					if !configurations.Report.MaskValue {
						oldRecord = eventFieldDiff.OldRecord
						newRecord = eventFieldDiff.NewRecord
					}

					entity := EventDataReportEntity{}
					entity.EventId = eventFieldDiff.SourceEventId
					entity.PreviousEventId = previousEventId
					entity.PreviousEventName = previousEventName
					entity.EventName = eventFieldDiff.SourceEventName
					entity.CaseTypeId = eventFieldDiff.CaseTypeId
					entity.Reference = caseReference
					entity.FieldName = fieldName
					entity.ChangeType = string(eventFieldDiff.OperationType)
					if isArrayChange {
						entity.ArrayChangeRecord = stripBytes(newRecord)
					} else {
						entity.OldRecord = stripBytes(oldRecord)
						entity.NewRecord = stripBytes(newRecord)
					}
					entity.PreviousEventCreatedDate = previousEventCreatedDate
					entity.EventCreatedDate = eventFieldDiff.CreatedDate
					entity.AnalyzeResult = stripBytes(message)
					entity.RuleMatched = message != ""
					entity.EventUserId = eventFieldDiff.UserId
					entity.PreviousEventUserId = previousUserId
					entity.EventDelta = delta
					eventDataReportEntities = append(eventDataReportEntities, entity)
				}
			}
		}
	}

	return eventDataReportEntities, nil
}

func stripBytes(value string) string {
	data := []byte(value)
	data = bytes.Replace(data, []byte{0xe2, 0x27, 0x20}, []byte{}, -1)
	data = bytes.Replace(data, []byte{0xe2, 0x80, 0x27}, []byte{}, -1)
	data = bytes.Replace(data, []byte{0xc2, 0x27}, []byte{}, -1)

	return string(data)
}
