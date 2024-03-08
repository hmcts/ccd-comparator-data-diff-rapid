package domain

import (
	"ccd-comparator-data-diff-rapid/internal/store"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type QueryRepository interface {
	findCasesByJurisdictionInImpactPeriod(caseIds []string) ([]CaseDataEntity, error)
	findCasesByEventsInImpactPeriod(comparison Comparison) ([]string, error)
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
	UserId           string    `db:"user_id"`
}

func NewQueryRepository(db store.DB) QueryRepository {
	return &queryRepository{db: db}
}

func (r queryRepository) findCasesByEventsInImpactPeriod(comparison Comparison) ([]string, error) {
	var caseIDs []string
	var query string
	var args []interface{}

	query = `SELECT cd.id FROM case_data cd
                    INNER JOIN case_event ce ON cd.id = ce.case_data_id
                    WHERE cd.jurisdiction = $1`

	args = append(args, comparison.Jurisdiction)

	if comparison.CaseTypeId != "" {
		query += " AND cd.case_type_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, comparison.CaseTypeId)
	}

	query += ` AND ce.created_date >= $` + strconv.Itoa(len(args)+1) + `
                        AND ce.created_date <= $` + strconv.Itoa(len(args)+2) + `
                    GROUP BY cd.id 
                    HAVING COUNT(ce.id) > 1
                    ORDER BY cd.id`

	args = append(args, comparison.StartTime, comparison.SearchPeriodEndTime)

	err := r.db.Select(&caseIDs, query, args...)

	if err != nil {
		return nil, errors.Wrap(err, "error while retrieving caseIDs in findCasesByJurisdictionInImpactPeriod()")
	}

	return caseIDs, nil
}

func (r queryRepository) findCasesByJurisdictionInImpactPeriod(caseIds []string) ([]CaseDataEntity, error) {
	var caseData []CaseDataEntity

	caseIDQuery := "'" + strings.Join(caseIds, "','") + "'"

	err := r.db.Select(&caseData, `SELECT cd.id as case_id, cd.created_date as case_created_date,
							cd.jurisdiction as jurisdiction, cd.case_type_id as case_type_id, cd.reference as reference,
							ce.case_data_id as case_data_id, ce.id as event_id, ce.event_id as event_name, 
							ce.user_id as user_id, ce.created_date as event_created_date, ce.data as event_data
							FROM case_data cd inner join case_event ce on cd.id = ce.case_data_id
							WHERE cd.id IN (`+caseIDQuery+`)`)

	if err != nil {
		return nil, errors.Wrap(err, "error in findCasesByJurisdictionInImpactPeriod()")
	}

	return caseData, nil
}
