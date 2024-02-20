package domain

import (
	"ccd-comparator-data-diff-rapid/comparator"
	"ccd-comparator-data-diff-rapid/internal/store"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const defaultBatchSize = 100

type SaveRepository interface {
	saveAllEventDataReport(batchSize int, eventDataReportEntities []comparator.EventDataReportEntity) error
}

type saveRepository struct {
	db store.DB
}

func NewSaveRepository(db store.DB) SaveRepository {
	return &saveRepository{db: db}
}

func (s saveRepository) saveAllEventDataReport(batchSize int, eventDataReportEntities []comparator.EventDataReportEntity) error {
	totalEntities := len(eventDataReportEntities)
	if batchSize == 0 {
		batchSize = defaultBatchSize
	}
	tx := s.db.MustBegin()
	for i := 0; i < totalEntities; i += batchSize {
		end := i + batchSize
		if end > totalEntities {
			end = totalEntities
		}

		batch := eventDataReportEntities[i:end]

		res, err := tx.NamedExec(`INSERT INTO event_data_report_v2 (
			event_id, event_name, case_type_id, reference, field_name, change_type,
			old_record, new_record, previous_event_created_date, event_created_date,
			analyze_result_detail, potential_risk)
		VALUES (:event_id, :event_name, :case_type_id, :reference, :field_name, :change_type, :old_record, :new_record,
			:previous_event_created_date, :event_created_date, :analyze_result, :potential_risk)`, batch)

		if err != nil {
			_ = tx.Rollback()
			return errors.Wrap(err, "Failed while batch inserting report")
		}

		if log.Debug().Enabled() {
			count, _ := res.(sql.Result).RowsAffected()
			log.Info().Msgf("%d records persisted successfully", count)
		}
	}
	err := tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrap(err, "Failed while committing the transaction")
	}

	return nil
}
