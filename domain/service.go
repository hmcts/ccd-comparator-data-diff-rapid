package domain

import (
	"ccd-comparator-data-diff-rapid/comparator"
	"ccd-comparator-data-diff-rapid/config"
	"ccd-comparator-data-diff-rapid/helper"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

const (
	hoursInDay      = 24
	minutesInHour   = 60
	secondsInMinute = 60
	nanoseconds     = 999999999
)

type Service struct {
	configuration *config.Configurations
	activeRules   *[]comparator.Rule
	queryRepo     QueryRepository
	saveRepo      SaveRepository
}

func NewService(configuration *config.Configurations, activeRules *[]comparator.Rule,
	queryRepo QueryRepository, saveRepo SaveRepository) *Service {
	return &Service{
		configuration: configuration,
		activeRules:   activeRules,
		queryRepo:     queryRepo,
		saveRepo:      saveRepo,
	}
}

type ComparisonResult struct {
	TransactionID string
	Result        string
	Error         error
}

func (s Service) CompareEventsInImpactPeriod(jurisdiction, caseTypeId string, startTime, endTime time.Time) {
	var wg sync.WaitGroup

	resultChan := make(chan ComparisonResult)

	for !startTime.After(endTime) {
		transactionId := uuid.New().String()
		searchPeriodEndTime := calculateSearchPeriodEndTime(startTime, endTime, s.configuration.SearchWindow)

		wg.Add(1)
		go s.compareAndSaveEvents(&wg, resultChan, transactionId, jurisdiction, caseTypeId, startTime, searchPeriodEndTime)

		startTime = searchPeriodEndTime.Add(1 * time.Nanosecond)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		if result.Error != nil {
			log.Error().Msgf("tid:%s - ERROR: %s", result.TransactionID, result.Error)
		} else {
			log.Info().Msgf("tid:%s - Result: %s", result.TransactionID, result.Result)
		}
	}
}

func calculateSearchPeriodEndTime(startTime, endTime time.Time, searchWindow int) time.Time {
	searchPeriodEndTime := startTime.AddDate(0, 0, searchWindow)
	searchPeriodEndTime = time.Date(searchPeriodEndTime.Year(), searchPeriodEndTime.Month(),
		searchPeriodEndTime.Day(), hoursInDay-1, minutesInHour-1, secondsInMinute-1, nanoseconds, searchPeriodEndTime.Location())

	if searchPeriodEndTime.After(endTime) {
		searchPeriodEndTime = endTime
	}

	return searchPeriodEndTime
}

func (s Service) compareAndSaveEvents(wg *sync.WaitGroup, resultChan chan ComparisonResult, transactionId, jurisdiction,
	caseTypeId string, startTime, endTime time.Time) {
	defer wg.Done()

	logEventComparisonStart(transactionId, jurisdiction, caseTypeId, startTime, endTime)

	cases, err := s.queryRepo.findCasesByJurisdictionInImpactPeriod(jurisdiction, caseTypeId, startTime, endTime)
	if err != nil {
		handleError(resultChan, transactionId, err, "finding cases")
		return
	}

	if len(cases) == 0 {
		noDataMessage := fmt.Sprintf("No case data returned for jurisdiction: %s with caseTypeId: %s", jurisdiction, caseTypeId)
		sendResult(resultChan, transactionId, noDataMessage)
		return
	}

	logParsingCaseData(transactionId, jurisdiction, caseTypeId, len(cases))

	casesWithEventDetails := getCasesWithEventDetails(cases)

	logEventComparisonStarted(transactionId)

	// Compare events by case reference
	eventFieldChanges := comparator.CompareEventsByCaseReference(transactionId, casesWithEventDetails)
	if len(eventFieldChanges) == 0 {
		resultMessage := fmt.Sprintf("No differences found in events for specified cases based on the search criteria provided")
		sendResult(resultChan, transactionId, resultMessage)
		return
	}

	analyzeResult := comparator.NewEventChangesAnalyze(s.activeRules, eventFieldChanges).AnalyzeEventFieldChanges()

	if !s.configuration.Report.Enabled {
		resultMessage := fmt.Sprintf("Analysis completed without saving the report. Total records in analyzeResult: %d. Total number of field change: %d",
			analyzeResult.Size(), len(eventFieldChanges))
		sendResult(resultChan, transactionId, resultMessage)
		return
	}

	if err := s.saveReport(transactionId, analyzeResult, eventFieldChanges); err != nil {
		handleError(resultChan, transactionId, err, "saving the report")
		return
	}

	sendResult(resultChan, transactionId, "Operation completed successfully.")
}

func logEventComparisonStart(transactionID, jurisdiction, caseTypeID string, startTime, endTime time.Time) {
	log.Info().Msgf("tid:%s - Event comparison started: start period: %s, end period: %s with jurisdiction: %s and caseType: %s",
		transactionID, helper.FormatTimeStamp(startTime), helper.FormatTimeStamp(endTime), jurisdiction, caseTypeID)
}

func logParsingCaseData(transactionID, jurisdiction, caseTypeID string, numCases int) {
	log.Info().Msgf("tid:%s - Parsing case data with jurisdiction %s and caseType %s with %d case events",
		transactionID, jurisdiction, caseTypeID, numCases)
}

func logEventComparisonStarted(transactionID string) {
	log.Info().Msgf("tid:%s - Event comparing started...", transactionID)
}

func handleError(resultChan chan ComparisonResult, transactionID string, err error, context string) {
	sendError(resultChan, transactionID, errors.Wrap(err, fmt.Sprintf("error occurred while %s", context)))
}

func sendResult(resultChan chan ComparisonResult, transactionId, resultFormat string, args ...interface{}) {
	result := ComparisonResult{
		TransactionID: transactionId,
		Result:        fmt.Sprintf(resultFormat, args...),
	}
	resultChan <- result
}

func sendError(resultChan chan ComparisonResult, transactionId string, error error) {
	result := ComparisonResult{
		TransactionID: transactionId,
		Error:         error,
	}
	resultChan <- result
}

func (s Service) saveReport(transactionId string, analyzeResult *comparator.AnalyzeResult, eventDifferences comparator.EventFieldChanges) error {
	if analyzeResult.IsNotEmpty() || s.configuration.Report.IncludeEmptyChange {
		eventDataReportEntities, err := comparator.PrepareReportEntities(eventDifferences, analyzeResult, s.configuration)
		if err != nil {
			return errors.Wrap(err, "failed to process PrepareReportEntities")
		}

		numberOfRecord := len(eventDataReportEntities)
		if numberOfRecord == 0 {
			log.Info().Msg("No record returned from PrepareReportEntities")
			return nil
		}

		log.Info().Msgf("tid:%s - Saving report data to the database. Total record number: %d", transactionId, numberOfRecord)

		err = s.saveRepo.saveAllEventDataReport(eventDataReportEntities)
		if err != nil {
			return errors.Wrap(err, "failed to save report data")
		}

		log.Info().Msgf("tid:%s - Records successfully saved to the database", transactionId)
	} else {
		log.Info().Msgf("tid:%s - Saving the report has been skipped", transactionId)
	}
	return nil
}

func getCasesWithEventDetails(cases []CaseDataEntity) comparator.CasesWithEventDetails {
	casesWithEventDetails := make(comparator.CasesWithEventDetails)

	for _, caseData := range cases {
		if _, ok := casesWithEventDetails[caseData.Reference]; !ok {
			casesWithEventDetails[caseData.Reference] = make(map[int64]comparator.EventDetails)
		}

		casesWithEventDetails[caseData.Reference][caseData.EventId] = comparator.EventDetails{
			Id:          caseData.EventId,
			Name:        caseData.EventName,
			CreatedDate: caseData.EventCreatedDate,
			Data:        caseData.EventData,
			CaseDataId:  caseData.CaseDataId,
		}
	}

	return casesWithEventDetails
}
