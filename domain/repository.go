package domain

import (
	"ccd-comparator-data-diff-rapid/comparator"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

const batchSize = 100

type Repository interface {
	findCasesByJurisdictionInImpactPeriod(jurisdiction string, caseTypeId string, startTime time.Time,
		endTime time.Time) ([]CaseDataEntity, error)
	saveAllEventDataReport(eventDataReportEntities []comparator.EventDataReportEntity) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

type CaseDataEntity struct {
	CaseId           int64     `db:"case_id"`
	CaseCreatedDate  time.Time `db:"case_created_date"`
	Jurisdiction     string    `db:"jurisdiction"`
	CaseTypeId       string    `db:"case_type_id"`
	CaseDataId       int64     `db:"case_data_id"`
	Reference        int64     `db:"reference"`
	EventId          int64     `db:"event_id"`
	EventName        string    `db:"event_name"`
	EventCreatedDate time.Time `db:"event_created_date"`
	EventData        string    `db:"event_data"`
}

func (r *repository) findCasesByJurisdictionInImpactPeriod(jurisdiction string, caseTypeId string, startTime time.Time,
	endTime time.Time) ([]CaseDataEntity, error) {

	var caseData []CaseDataEntity
	err := r.db.Select(&caseData, `SELECT cd.id as case_id, cd.created_date as case_created_date,
							cd.jurisdiction as jurisdiction, cd.case_type_id as case_type_id, cd.reference as reference,
							ce.case_data_id as case_data_id, ce.id as event_id, ce.event_id as event_name, 
							ce.created_date as event_created_date, ce.data as event_data
							FROM case_data cd inner join case_event ce on cd.id = ce.case_data_id
							WHERE cd.jurisdiction = $1 --and cd.reference = 1681802449198475
								AND cd.case_type_id =  $2
								AND cd.created_date >= $3
								AND cd.created_date <= $4`,
		jurisdiction, caseTypeId, startTime, endTime)

	if err != nil {
		return nil, errors.WithMessage(err, " : error in findCasesByJurisdictionInImpactPeriod()")
	}

	return caseData, nil
}

func (r *repository) saveAllEventDataReport(eventDataReportEntities []comparator.EventDataReportEntity) error {
	totalEntities := len(eventDataReportEntities)

	tx := r.db.MustBegin()
	for i := 0; i < totalEntities; i += batchSize {
		end := i + batchSize
		if end > totalEntities {
			end = totalEntities
		}

		batch := eventDataReportEntities[i:end]

		_, err := tx.NamedExec(`INSERT INTO event_data_report (
			event_id, event_name, case_type_id, reference, field_name, change_type,
			old_record, new_record, previous_event_created_date, event_created_date,
			analyze_result, potential_risk)
		VALUES (:event_id, :event_name, :case_type_id, :reference, :field_name, :change_type, :old_record, :new_record,
			:previous_event_created_date, :event_created_date, :analyze_result, :potential_risk)`, batch)

		if err != nil {
			return errors.WithMessage(err, "Failed while batch inserting report:")
		}
	}
	err := tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return errors.WithMessage(err, "Failed while committing the transaction:")
	}
	return nil
}
