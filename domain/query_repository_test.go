package domain

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestFindCasesByJurisdictionInImpactPeriod(t *testing.T) {
	mockDB := new(MockDB)
	queryRepo := NewQueryRepository(mockDB)

	jurisdiction := "TestJurisdiction"
	caseTypeId := "TestCaseType"
	startTime := time.Now()
	endTime := startTime.Add(time.Hour)

	expectedCases := []CaseDataEntity{
		{
			CaseId:           1,
			CaseCreatedDate:  time.Now(),
			Jurisdiction:     "TestJurisdiction",
			CaseTypeId:       "TestCaseType",
			CaseDataId:       2,
			Reference:        3,
			EventId:          4,
			EventName:        "TestEvent",
			EventCreatedDate: time.Now(),
			EventData:        "EventData",
		},
	}

	mockDB.On("Select",
		mock.AnythingOfType("*[]domain.CaseDataEntity"),
		mock.AnythingOfType("string"),
		mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			dest := args.Get(0)
			casesPtr, ok := dest.(*[]CaseDataEntity)
			if !ok {
				t.Fatalf("Invalid type for destination: %T", dest)
			}
			*casesPtr = expectedCases
		})

	c := Comparison{
		Jurisdiction:        jurisdiction,
		CaseTypeId:          caseTypeId,
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	cases, err := queryRepo.findCasesByJurisdictionInImpactPeriod(c)

	assert.NoError(t, err)
	assert.NotNil(t, cases)
	assert.Equal(t, expectedCases, cases)

	mockDB.AssertExpectations(t)
	//mockTx.AssertExpectations(t)
}

func TestFindCasesByJurisdictionInImpactPeriodReturnError(t *testing.T) {
	mockDB := new(MockDB)
	queryRepo := NewQueryRepository(mockDB)

	jurisdiction := "TestJurisdiction"
	caseTypeId := "TestCaseType"
	startTime := time.Now()
	endTime := startTime.Add(time.Hour)

	expectedError := errors.New("some error")

	mockDB.On("Select",
		mock.AnythingOfType("*[]domain.CaseDataEntity"),
		mock.AnythingOfType("string"),
		mock.Anything).
		Return(expectedError)

	c := Comparison{
		Jurisdiction:        jurisdiction,
		CaseTypeId:          caseTypeId,
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	cases, err := queryRepo.findCasesByJurisdictionInImpactPeriod(c)

	unwrappedErr := errors.Cause(err)

	assert.Error(t, err)
	assert.Nil(t, cases)
	assert.EqualError(t, unwrappedErr, expectedError.Error())

	mockDB.AssertExpectations(t)
}
