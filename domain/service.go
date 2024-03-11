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

type Comparison struct {
	Jurisdiction        string
	CaseTypeId          string
	StartTime           time.Time
	SearchPeriodEndTime time.Time
}

type comparisonWork struct {
	transactionId string
	comparison    Comparison
	caseIds       []string
}

type comparisonResult struct {
	transactionId string
	result        string
	err           error
}

func (s Service) CompareEventsInImpactPeriod(comparison Comparison) {
	resultChan := make(chan comparisonResult)
	defer func() {
		close(resultChan)
	}()

	processResults(resultChan)

	s.startComparisonWorkers(comparison, resultChan)
}

func processResults(resultChan <-chan comparisonResult) {
	go func() {
		for result := range resultChan {
			if result.err != nil {
				log.Error().Msgf("tid:%s - ERROR: %s", result.transactionId, result.err)
			} else {
				log.Info().Msgf("tid:%s - result: %s", result.transactionId, result.result)
			}
		}
	}()
}

func (s Service) startComparisonWorkers(comparison Comparison, resultChan chan<- comparisonResult) {
	var wg sync.WaitGroup
	s.processComparisonWork(&wg, comparison, resultChan)
	wg.Wait()
}

func (s Service) processComparisonWork(wg *sync.WaitGroup, comparison Comparison, resultChan chan<- comparisonResult) {
	numberOfWorker := s.configuration.Worker.Pool
	workers := make(chan comparisonWork, numberOfWorker)
	defer closeWorkers(workers)

	caseIds, err := s.retrieveCaseIdsByEvents(comparison)
	if err != nil {
		log.Error().Msgf("Couldn't retrieve caseIds. ERROR: %s", err)
		return
	}

	countOfCaseIds := len(caseIds)
	if countOfCaseIds == 0 {
		log.Warn().Msgf("Couldn't retrieved any case: start period: %s, "+
			"end period: %s with jurisdiction: %s and caseType: %s",
			helper.FormatTimeStamp(comparison.StartTime), helper.FormatTimeStamp(comparison.SearchPeriodEndTime),
			comparison.Jurisdiction, comparison.CaseTypeId)
		return
	}

	log.Info().Msgf("'%d' case retrieved: start period: %s, "+
		"end period: %s with jurisdiction: %s and caseType: %s",
		countOfCaseIds, helper.FormatTimeStamp(comparison.StartTime), helper.FormatTimeStamp(comparison.SearchPeriodEndTime),
		comparison.Jurisdiction, comparison.CaseTypeId)

	for wid := 1; wid <= numberOfWorker; wid++ {
		wg.Add(1)
		go s.compareAndSaveEvents(wid, wg, workers, resultChan)
	}

	batchSize := s.configuration.Scan.BatchSize
	for i := 0; i < countOfCaseIds; i += batchSize {
		end := i + batchSize
		if end > countOfCaseIds {
			end = countOfCaseIds
		}

		batch := caseIds[i:end]

		transactionId := uuid.New().String()
		workers <- comparisonWork{
			transactionId: transactionId,
			caseIds:       batch,
			comparison:    comparison,
		}
	}
}

func closeWorkers(workers chan comparisonWork) {
	log.Info().Msgf("All jobs have been sent successfully to the workers")
	close(workers)
}

func (s Service) retrieveCaseIdsByEvents(comparison Comparison) ([]string, error) {
	return s.queryRepo.findCasesByEventsInImpactPeriod(comparison)
}

func (s Service) compareAndSaveEvents(workerId int, wg *sync.WaitGroup, workers <-chan comparisonWork, resultChan chan<- comparisonResult) {
	defer func(id int) {
		log.Info().Msgf("Worker %d has completed its work and is being deferred.", id)
		wg.Done()
	}(workerId)

	for w := range workers {
		logEventComparisonStart(workerId, w)
		cases, err := s.queryRepo.findCasesByJurisdictionInImpactPeriod(w.caseIds)
		if err != nil {
			handleError(resultChan, w.transactionId, err, "finding cases")
			continue
		}

		if len(cases) == 0 {
			noDataMessage := fmt.Sprintf("No case data returned for jurisdiction: %s with caseTypeId: %s",
				w.comparison.Jurisdiction, w.comparison.CaseTypeId)
			sendResult(resultChan, w.transactionId, noDataMessage)
			continue
		}

		logParsingCaseData(w.transactionId, w.comparison.Jurisdiction, w.comparison.CaseTypeId, len(cases))

		casesWithEventDetails := getCasesWithEventDetails(cases)

		logEventComparisonStarted(w.transactionId)

		// Compare events by case reference
		eventFieldChanges := comparator.CompareEventsByCaseReference(w.transactionId, casesWithEventDetails)
		if len(eventFieldChanges) == 0 {
			resultMessage := fmt.Sprintf("No differences found in events for specified cases based on the search criteria provided")
			sendResult(resultChan, w.transactionId, resultMessage)
			continue
		}

		analyzeResult := comparator.NewEventChangesAnalyze(s.activeRules, eventFieldChanges).AnalyzeEventFieldChanges()

		if !s.configuration.Report.Enabled {
			resultMessage := fmt.Sprintf("Analysis completed without saving the report. Total records in analyzeResult: %d. Total number of field change: %d",
				analyzeResult.Size(), len(eventFieldChanges))
			sendResult(resultChan, w.transactionId, resultMessage)
			continue
		}

		if err := s.saveReport(w.transactionId, analyzeResult, eventFieldChanges); err != nil {
			handleError(resultChan, w.transactionId, err, "saving the report")
			continue
		}

		sendResult(resultChan, w.transactionId, "Operation completed successfully.")
	}
}

func logEventComparisonStart(workerId int, w comparisonWork) {
	log.Info().Msgf("tid:%s -- Event comparison started in worker %d: caseIds: %s",
		w.transactionId, workerId, w.caseIds)
}

func logParsingCaseData(transactionId, jurisdiction, caseTypeID string, numCases int) {
	log.Info().Msgf("tid:%s - Parsing case data with jurisdiction %s and caseType %s with %d case events",
		transactionId, jurisdiction, caseTypeID, numCases)
}

func logEventComparisonStarted(transactionId string) {
	log.Info().Msgf("tid:%s - Event comparing started...", transactionId)
}

func handleError(resultChan chan<- comparisonResult, transactionId string, err error, context string) {
	sendError(resultChan, transactionId, errors.Wrap(err, fmt.Sprintf("error occurred while %s", context)))
}

func sendResult(resultChan chan<- comparisonResult, transactionId, resultFormat string, args ...interface{}) {
	result := comparisonResult{
		transactionId: transactionId,
		result:        fmt.Sprintf(resultFormat, args...),
	}
	resultChan <- result
}

func sendError(resultChan chan<- comparisonResult, transactionId string, error error) {
	result := comparisonResult{
		transactionId: transactionId,
		err:           error,
	}
	resultChan <- result
}

func (s Service) saveReport(transactionId string, analyzeResult *comparator.AnalyzeResult,
	eventDifferences comparator.EventFieldChanges) error {

	if analyzeResult.IsNotEmpty() || s.configuration.Report.IncludeEmptyChange {
		eventDataReportEntities, err := comparator.PrepareReportEntities(eventDifferences, analyzeResult,
			s.configuration)
		if err != nil {
			return errors.Wrap(err, "failed to process PrepareReportEntities")
		}

		numberOfRecord := len(eventDataReportEntities)
		if numberOfRecord == 0 {
			log.Info().Msg("No records were returned from PrepareReportEntities.")
			return nil
		}

		log.Info().Msgf("tid:%s - Saving report data to the database. Total record number: %d", transactionId, numberOfRecord)

		err = s.saveRepo.saveAllEventDataReport(s.configuration.Database.BatchSize,
			s.configuration.Database.EventDataTable, eventDataReportEntities)
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
			UserId:      caseData.UserId,
			CaseTypeId:  caseData.CaseTypeId,
		}
	}

	return casesWithEventDetails
}
