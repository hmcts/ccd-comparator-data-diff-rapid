package comparator

import (
	"ccd-comparator-data-diff-rapid/config"
	"ccd-comparator-data-diff-rapid/helper"
	"fmt"
	"strings"
	"time"
)

type EventDataReportEntity struct {
	ID                       int64     `db:"id"`
	EventID                  int64     `db:"event_id"`
	EventName                string    `db:"event_name"`
	CaseTypeID               string    `db:"case_type_id"`
	Reference                string    `db:"reference"`
	FieldName                string    `db:"field_name"`
	ChangeType               string    `db:"change_type"`
	OldRecord                string    `db:"old_record"`
	NewRecord                string    `db:"new_record"`
	PreviousEventCreatedDate time.Time `db:"previous_event_created_date"`
	EventCreatedDate         time.Time `db:"event_created_date"`
	AnalyzeResult            string    `db:"analyze_result"`
	PotentialRisk            bool      `db:"potential_risk"`
}

func PrepareReportEntities(eventDifferences map[string][]EventFieldChange, analyzeResult *AnalyzeResult, configurations *config.Configurations) ([]EventDataReportEntity, error) {
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
			if configurations.Report.IncludeNoChange || eventFieldDiff.OperationType != helper.NoChange {
				entity := EventDataReportEntity{}
				combinedResult := analyzeResult.Get(combinedReference, eventFieldDiff.SourceEventId)

				var previousEventCreatedDate time.Time
				var message string
				if combinedResult != "" {
					combinedResults := strings.Split(combinedResult, "->")
					if len(combinedResults) == 2 {
						previousEventCreatedDate = helper.MustParseTime("", combinedResults[0])
						message = combinedResults[1]
					} else {
						message = combinedResults[0]
					}
				}

				if configurations.Report.IncludeEmptyChange || message != "" {
					var oldRecord, newRecord string
					if !configurations.Report.MaskValue {
						oldRecord = eventFieldDiff.OldRecord
						newRecord = eventFieldDiff.NewRecord
					}
					entity.EventID = eventFieldDiff.SourceEventId
					entity.EventName = eventFieldDiff.SourceEventName
					entity.CaseTypeID = configurations.CaseType
					entity.Reference = caseReference
					entity.FieldName = fieldName
					entity.ChangeType = string(eventFieldDiff.OperationType)
					entity.OldRecord = oldRecord
					entity.NewRecord = newRecord
					entity.PreviousEventCreatedDate = previousEventCreatedDate
					entity.EventCreatedDate = eventFieldDiff.CreatedDate
					entity.AnalyzeResult = message
					entity.PotentialRisk = message != ""
					eventDataReportEntities = append(eventDataReportEntities, entity)
				}
			}
		}
	}

	return eventDataReportEntities, nil
}
