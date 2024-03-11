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
	ID                       int64         `db:"id"`
	EventID                  int64         `db:"event_id"`
	EventName                string        `db:"event_name"`
	CaseTypeID               string        `db:"case_type_id"`
	Reference                string        `db:"reference"`
	FieldName                string        `db:"field_name"`
	ChangeType               string        `db:"change_type"`
	OldRecord                string        `db:"old_record"`
	NewRecord                string        `db:"new_record"`
	PreviousEventCreatedDate time.Time     `db:"previous_event_created_date"`
	EventCreatedDate         time.Time     `db:"event_created_date"`
	AnalyzeResult            string        `db:"analyze_result"`
	PotentialRisk            bool          `db:"potential_risk"`
	PreviousEventUserId      string        `db:"previous_event_user_id"`
	EventUserId              string        `db:"event_user_id"`
	EventDelta               time.Duration `db:"event_delta"`
}

func PrepareReportEntities(eventDifferences map[string][]EventFieldChange, analyzeResult *AnalyzeResult,
	configurations *config.Configurations, caseTypeId string) ([]EventDataReportEntity, error) {

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

		for _, eventFieldDiff := range fieldDifferences {
			if configurations.Report.IncludeNoChange || eventFieldDiff.OperationType != NoChange {
				entity := EventDataReportEntity{}
				violation := analyzeResult.Get(combinedReference, eventFieldDiff.SourceEventId)

				var previousEventCreatedDate time.Time
				var previousUserId string
				var message string
				var delta time.Duration

				if violation.sourceEventId != 0 {
					previousEventCreatedDate = helper.MustParseTime("", violation.previousEventCreatedDate)
					previousUserId = violation.previousEventUserId
					message = violation.message
					delta = time.Duration(eventFieldDiff.CreatedDate.Sub(previousEventCreatedDate).Milliseconds())
				}

				if configurations.Report.IncludeEmptyChange || message != "" {
					var oldRecord, newRecord string
					if !configurations.Report.MaskValue {
						oldRecord = eventFieldDiff.OldRecord
						newRecord = eventFieldDiff.NewRecord
					}
					entity.EventID = eventFieldDiff.SourceEventId
					entity.EventName = eventFieldDiff.SourceEventName
					entity.CaseTypeID = caseTypeId
					entity.Reference = caseReference
					entity.FieldName = fieldName
					entity.ChangeType = string(eventFieldDiff.OperationType)
					entity.OldRecord = stripBytes(oldRecord)
					entity.NewRecord = stripBytes(newRecord)
					entity.PreviousEventCreatedDate = previousEventCreatedDate
					entity.EventCreatedDate = eventFieldDiff.CreatedDate
					entity.AnalyzeResult = stripBytes(message)
					entity.PotentialRisk = message != ""
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

	return string(data)
}
