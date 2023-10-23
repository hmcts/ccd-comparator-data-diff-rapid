package domain

import (
	"ccd-comparator-data-diff-rapid/comparator"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockQueryRepository struct {
	mock.Mock
}

func (m *MockQueryRepository) findCasesByJurisdictionInImpactPeriod(comparison Comparison) ([]CaseDataEntity, error) {
	args := m.Called(comparison)
	return args.Get(0).([]CaseDataEntity), args.Error(1)
}

type MockSaveRepository struct {
	mock.Mock
}

func (m *MockSaveRepository) saveAllEventDataReport(eventDataReportEntities []comparator.EventDataReportEntity) error {
	args := m.Called(eventDataReportEntities)
	return args.Error(0)
}

func TestService_CompareEventsInImpactPeriodHappyPath(t *testing.T) {
	setUp()
	defer cleanUp()

	mockQueryRepo := new(MockQueryRepository)
	mockSaveRepo := new(MockSaveRepository)

	enabledRuleList := comparator.NewRuleFactory(cfg).GetEnabledRuleList()

	service := NewService(cfg, &enabledRuleList, mockQueryRepo, mockSaveRepo)

	startTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)

	mockQueryRepo.On("findCasesByJurisdictionInImpactPeriod", Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}).
		Return([]CaseDataEntity{
			{
				Reference:        1,
				EventId:          1,
				EventName:        "Event1",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"John\", \"age\": 30, \"city\": \"New York\"}",
				CaseDataId:       100,
			},
			{
				Reference:        1,
				EventId:          2,
				EventName:        "Event2",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"Mary\", \"age\": 30, \"city\": \"Los Angeles\"}",
				CaseDataId:       101,
			},
			{
				Reference:        1,
				EventId:          3,
				EventName:        "Event3",
				EventCreatedDate: time.Now(),
				EventData:        "{\"age\": 30, \"city\": \"Los Angeles\", \"extra\": \"test\"}",
				CaseDataId:       102,
			},
			{
				Reference:        1,
				EventId:          4,
				EventName:        "Event4",
				EventCreatedDate: time.Now(),
				EventData:        "{\"age\": 30, \"city\": \"New York\", \"extra\": \"test\"}",
				CaseDataId:       103,
			},
		}, nil)

	mockSaveRepo.On("saveAllEventDataReport", mock.Anything).Return(nil)

	c := Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	service.CompareEventsInImpactPeriod(c)

	mockQueryRepo.AssertExpectations(t)
	mockSaveRepo.AssertExpectations(t)
}

func TestService_CompareEventsInImpactIgnoreSaveReport(t *testing.T) {
	setUp()
	defer cleanUp()

	mockQueryRepo := new(MockQueryRepository)
	mockSaveRepo := new(MockSaveRepository)

	enabledRuleList := comparator.NewRuleFactory(cfg).GetEnabledRuleList()

	cfg.Report.Enabled = false
	service := NewService(cfg, &enabledRuleList, mockQueryRepo, mockSaveRepo)

	startTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	mockQueryRepo.On("findCasesByJurisdictionInImpactPeriod", Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}).
		Return([]CaseDataEntity{
			{
				Reference:        1,
				EventId:          1,
				EventName:        "Event1",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"John\", \"age\": 30, \"city\": \"New York\"}",
				CaseDataId:       100,
			},
			{
				Reference:        1,
				EventId:          2,
				EventName:        "Event2",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"Mary\", \"age\": 30, \"city\": \"Los Angeles\"}",
				CaseDataId:       101,
			},
			{
				Reference:        1,
				EventId:          3,
				EventName:        "Event3",
				EventCreatedDate: time.Now(),
				EventData:        "{\"age\": 30, \"city\": \"Los Angeles\", \"extra\": \"test\"}",
				CaseDataId:       102,
			},
			{
				Reference:        1,
				EventId:          4,
				EventName:        "Event4",
				EventCreatedDate: time.Now(),
				EventData:        "{\"age\": 30, \"city\": \"New York\", \"extra\": \"test\"}",
				CaseDataId:       103,
			},
		}, nil)

	c := Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	service.CompareEventsInImpactPeriod(c)

	mockQueryRepo.AssertExpectations(t)
	mockSaveRepo.AssertNotCalled(t, "saveAllEventDataReport")
}

func TestService_CompareEventsInImpactNoCaseData(t *testing.T) {
	setUp()
	defer cleanUp()

	mockQueryRepo := new(MockQueryRepository)
	mockSaveRepo := new(MockSaveRepository)

	enabledRuleList := comparator.NewRuleFactory(cfg).GetEnabledRuleList()

	cfg.Report.Enabled = false
	service := NewService(cfg, &enabledRuleList, mockQueryRepo, mockSaveRepo)

	startTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	mockQueryRepo.On("findCasesByJurisdictionInImpactPeriod", Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}).
		Return([]CaseDataEntity{}, nil)

	mockSaveRepo.On("saveAllEventDataReport", mock.Anything).Return(nil)

	c := Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	service.CompareEventsInImpactPeriod(c)

	mockQueryRepo.AssertExpectations(t)
	mockSaveRepo.AssertNotCalled(t, "saveAllEventDataReport")
}

func TestService_CompareEventsInImpactEmptyEventFieldChanges(t *testing.T) {
	setUp()
	defer cleanUp()

	mockQueryRepo := new(MockQueryRepository)
	mockSaveRepo := new(MockSaveRepository)

	enabledRuleList := comparator.NewRuleFactory(cfg).GetEnabledRuleList()

	cfg.Report.Enabled = false
	service := NewService(cfg, &enabledRuleList, mockQueryRepo, mockSaveRepo)

	startTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	mockQueryRepo.On("findCasesByJurisdictionInImpactPeriod", Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}).
		Return([]CaseDataEntity{
			{
				EventId:          1,
				EventName:        "Event1",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"John\", \"age\": 30, \"city\": \"New York\"}",
				CaseDataId:       100,
			},
		}, nil)

	c := Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	service.CompareEventsInImpactPeriod(c)

	mockQueryRepo.AssertExpectations(t)
}

func TestService_CompareEventsInImpactErrorFromFindCasesByJurisdictionInImpactPeriod(t *testing.T) {
	setUp()
	defer cleanUp()

	mockQueryRepo := new(MockQueryRepository)
	mockSaveRepo := new(MockSaveRepository)

	enabledRuleList := comparator.NewRuleFactory(cfg).GetEnabledRuleList()

	cfg.Report.Enabled = false
	service := NewService(cfg, &enabledRuleList, mockQueryRepo, mockSaveRepo)

	startTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	mockQueryRepo.On("findCasesByJurisdictionInImpactPeriod", Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}).
		Return([]CaseDataEntity{}, errors.New("error occured"))

	c := Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	service.CompareEventsInImpactPeriod(c)

	mockQueryRepo.AssertExpectations(t)
}

func TestService_CompareEventsInImpactErrorFromSaveAllEventDataReport(t *testing.T) {
	setUp()
	defer cleanUp()

	mockQueryRepo := new(MockQueryRepository)
	mockSaveRepo := new(MockSaveRepository)

	enabledRuleList := comparator.NewRuleFactory(cfg).GetEnabledRuleList()

	service := NewService(cfg, &enabledRuleList, mockQueryRepo, mockSaveRepo)

	startTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	mockQueryRepo.On("findCasesByJurisdictionInImpactPeriod", Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}).
		Return([]CaseDataEntity{
			{
				Reference:        1,
				EventId:          1,
				EventName:        "Event1",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"John\", \"age\": 30, \"city\": \"New York\"}",
				CaseDataId:       100,
			},
			{
				Reference:        1,
				EventId:          2,
				EventName:        "Event2",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"Mary\", \"age\": 30, \"city\": \"Los Angeles\"}",
				CaseDataId:       101,
			},
			{
				Reference:        1,
				EventId:          3,
				EventName:        "Event3",
				EventCreatedDate: time.Now(),
				EventData:        "{\"age\": 30, \"city\": \"Los Angeles\", \"extra\": \"test\"}",
				CaseDataId:       102,
			},
			{
				Reference:        1,
				EventId:          4,
				EventName:        "Event4",
				EventCreatedDate: time.Now(),
				EventData:        "{\"age\": 30, \"city\": \"New York\", \"extra\": \"test\"}",
				CaseDataId:       103,
			},
		}, nil)

	mockSaveRepo.On("saveAllEventDataReport", mock.Anything).Return(errors.New("error occured"))

	c := Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	service.CompareEventsInImpactPeriod(c)

	mockQueryRepo.AssertExpectations(t)
	mockSaveRepo.AssertExpectations(t)
}

func TestService_CompareEventsInImpactPeriodNoFieldChange(t *testing.T) {
	setUp()
	defer cleanUp()

	mockQueryRepo := new(MockQueryRepository)
	mockSaveRepo := new(MockSaveRepository)

	enabledRuleList := comparator.NewRuleFactory(cfg).GetEnabledRuleList()
	cfg.Concurrent.Event.ThresholdMilliseconds = 5000
	service := NewService(cfg, &enabledRuleList, mockQueryRepo, mockSaveRepo)

	startTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	mockQueryRepo.On("findCasesByJurisdictionInImpactPeriod", Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}).
		Return([]CaseDataEntity{
			{
				Reference:        1,
				EventId:          1,
				EventName:        "Event1",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"John\", \"age\": 30, \"city\": \"New York\"}",
				CaseDataId:       100,
			},
			{
				Reference:        1,
				EventId:          2,
				EventName:        "Event2",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"Mary\", \"age\": 30, \"city\": \"Los Angeles\"}",
				CaseDataId:       101,
			},
		}, nil)

	mockSaveRepo.On("saveAllEventDataReport", mock.Anything).Return(nil)

	c := Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	service.CompareEventsInImpactPeriod(c)

	mockQueryRepo.AssertExpectations(t)
}

func TestService_CompareEventsInImpactPeriodSkipEmptyChange(t *testing.T) {
	setUp()
	defer cleanUp()

	// Create mock objects for dependencies
	mockQueryRepo := new(MockQueryRepository)
	mockSaveRepo := new(MockSaveRepository)

	enabledRuleList := comparator.NewRuleFactory(cfg).GetEnabledRuleList()
	cfg.Concurrent.Event.ThresholdMilliseconds = 5000
	cfg.Report.IncludeEmptyChange = false
	service := NewService(cfg, &enabledRuleList, mockQueryRepo, mockSaveRepo)

	startTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	mockQueryRepo.On("findCasesByJurisdictionInImpactPeriod", Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}).
		Return([]CaseDataEntity{
			{
				Reference:        1,
				EventId:          1,
				EventName:        "Event1",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"John\", \"age\": 30, \"city\": \"New York\"}",
				CaseDataId:       100,
			},
			{
				Reference:        1,
				EventId:          2,
				EventName:        "Event2",
				EventCreatedDate: time.Now(),
				EventData:        "{\"name\": \"Mary\", \"age\": 30, \"city\": \"Los Angeles\"}",
				CaseDataId:       101,
			},
		}, nil)

	mockSaveRepo.On("saveAllEventDataReport", mock.Anything).Return(nil)

	c := Comparison{
		Jurisdiction:        "jurisdiction",
		CaseTypeId:          "caseType",
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	service.CompareEventsInImpactPeriod(c)

	mockQueryRepo.AssertExpectations(t)
}
