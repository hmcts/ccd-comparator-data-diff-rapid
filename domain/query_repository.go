package domain

import (
	"ccd-comparator-data-diff-rapid/internal/store"
	"github.com/pkg/errors"
	"time"
)

type QueryRepository interface {
	findCasesByJurisdictionInImpactPeriod(comparison Comparison) ([]CaseDataEntity, error)
}

type queryRepository struct {
	db store.DB
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

func NewQueryRepository(db store.DB) QueryRepository {
	return &queryRepository{db: db}
}

func (r queryRepository) findCasesByJurisdictionInImpactPeriod(comparison Comparison) ([]CaseDataEntity, error) {
	var caseData []CaseDataEntity

	err := r.db.Select(&caseData, `
			SELECT 
				cd.id as case_id,
				cd.created_date as case_created_date,
				cd.jurisdiction as jurisdiction,
				cd.case_type_id as case_type_id,
				cd.reference as reference,
				ce.case_data_id as case_data_id,
				ce.id as event_id,
				ce.event_id as event_name,
				ce.created_date as event_created_date,
				ce.data as event_data
			FROM 
				case_data cd 
			INNER JOIN 
				case_event ce ON cd.id = ce.case_data_id
			WHERE 
				cd.id IN (
					SELECT DISTINCT cd.id
					FROM case_data cd 
					INNER JOIN case_event ce ON cd.id = ce.case_data_id
					WHERE cd.jurisdiction = $1
					AND cd.case_type_id = $2
					AND ce.created_date >= $3
					AND ce.created_date <= $4)`,
		comparison.Jurisdiction, comparison.CaseTypeId, comparison.StartTime, comparison.SearchPeriodEndTime)

	if err != nil {
		return nil, errors.Wrap(err, "error in findCasesByJurisdictionInImpactPeriod()")
	}

	return caseData, nil
}
