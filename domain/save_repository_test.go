package domain

import (
	"ccd-comparator-data-diff-rapid/comparator"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type result struct {
	lastInsertId int64
	rowsAffected int64
}

func (r result) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func TestSaveAllEventDataReport(t *testing.T) {
	mockTx := new(MockTransaction)
	mockDB := new(MockDB)
	mockDB.On("MustBegin", mock.Anything).Return(mockTx, nil)

	saveRepo := NewSaveRepository(mockDB)

	// Successful scenario
	mockTx.On("NamedExec", mock.Anything, mock.Anything).Return(result{}, nil)
	mockTx.On("Commit").Return(nil).Once()

	eventDataReportEntities := make([]comparator.EventDataReportEntity, 201)

	err := saveRepo.saveAllEventDataReport(0, eventDataReportEntities)
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestSaveAllEventDataReportInsertError(t *testing.T) {
	mockTx := new(MockTransaction)
	mockDB := new(MockDB)
	mockDB.On("MustBegin", mock.Anything).Return(mockTx, nil)

	saveRepo := NewSaveRepository(mockDB)

	mockTx.On("NamedExec", mock.Anything, mock.Anything).
		Return(nil, errors.New("insert error")).Once()
	mockTx.On("Rollback").Return(nil).Once()

	eventDataReportEntities := make([]comparator.EventDataReportEntity, 100)

	err := saveRepo.saveAllEventDataReport(0, eventDataReportEntities)
	assert.Error(t, err)
	assert.EqualError(t, err, "Failed while batch inserting report: insert error")
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestSaveAllEventDataReportCommitError(t *testing.T) {
	mockTx := new(MockTransaction)
	mockDB := new(MockDB)
	mockDB.On("MustBegin", mock.Anything).Return(mockTx, nil)

	saveRepo := NewSaveRepository(mockDB)

	// err during commit
	mockTx.On("NamedExec", mock.Anything, mock.Anything).Return(result{}, nil).Once()
	mockTx.On("Commit").Return(errors.New("commit error")).Once()
	mockTx.On("Rollback").Return(nil).Once()

	eventDataReportEntities := make([]comparator.EventDataReportEntity, 100)

	err := saveRepo.saveAllEventDataReport(0, eventDataReportEntities)
	assert.Error(t, err)
	assert.EqualError(t, err, "Failed while committing the transaction: commit error")
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}
