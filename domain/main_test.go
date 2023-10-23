package domain

import (
	"ccd-comparator-data-diff-rapid/config"
	"ccd-comparator-data-diff-rapid/internal/store"
	"github.com/stretchr/testify/mock"
)

var cfg *config.Configurations

func setUp() {
	cfg = &config.Configurations{
		Database: config.Database{},
		Period: config.Period{
			SearchWindow: 1,
		},
		Rule: config.Rule{
			Active: "samevalueafterchange,fieldchangecount",
		},
		Worker: config.Worker{
			Pool: 5,
		},
		Scan: config.Scan{
			Concurrent: struct {
				Event struct {
					ThresholdMilliseconds int64
				}
			}{
				Event: struct {
					ThresholdMilliseconds int64
				}{
					ThresholdMilliseconds: 0,
				},
			},
			FieldChange: struct {
				Threshold int
			}{
				Threshold: 10,
			},
			Report: struct {
				Enabled            bool
				MaskValue          bool
				IncludeEmptyChange bool
			}{
				Enabled:            true,
				MaskValue:          false,
				IncludeEmptyChange: true,
			},
		},
	}
}

func cleanUp() {
	cfg = nil
}

type MockDB struct {
	mock.Mock
}

func (m *MockDB) MustBegin() store.Transaction {
	args := m.Called()
	return args.Get(0).(store.Transaction)
}

func (m *MockDB) Select(dest interface{}, query string, args ...interface{}) error {
	argsList := m.Called(dest, query, args)
	return argsList.Error(0)
}

type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) NamedExec(query string, arg interface{}) (interface{}, error) {
	argsList := m.Called(query, arg)
	return argsList.Get(0), argsList.Error(1)
}

func (m *MockTransaction) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Rollback() error {
	args := m.Called()
	return args.Error(0)
}
