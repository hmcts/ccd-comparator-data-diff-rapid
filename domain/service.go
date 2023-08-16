package domain

import (
	"ccd-comparator-data-diff-rapid/comparator"
	"ccd-comparator-data-diff-rapid/config"
	"ccd-comparator-data-diff-rapid/helper"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type Service interface {
	CompareEventsInImpactPeriod(jurisdiction string, caseTypeId string, startTime time.Time,
		endTime time.Time) error
}

type service struct {
	configuration *config.Configurations
	activeRules   *[]comparator.Rule
	repo          Repository
}

func NewService(configuration *config.Configurations, activeRules *[]comparator.Rule, repo Repository) Service {
	return &service{
		configuration: configuration,
		activeRules:   activeRules,
		repo:          repo,
	}
}

func (s *service) CompareEventsInImpactPeriod(jurisdiction, caseTypeId string, startTime, endTime time.Time) error {
	var wg sync.WaitGroup
	for !startTime.After(endTime) {
		transactionId := uuid.New().String()
		searchPeriodEndTime := startTime.AddDate(0, 0, s.configuration.SearchWindow)
		searchPeriodEndTime = time.Date(searchPeriodEndTime.Year(), searchPeriodEndTime.Month(),
			searchPeriodEndTime.Day(), 23, 59, 59, 999999999, searchPeriodEndTime.Location())

		if searchPeriodEndTime.After(endTime) {
			searchPeriodEndTime = endTime
		}

		wg.Add(1)
		go func(transactionId, jurisdiction, caseTypeId string, startTime, searchPeriodEndTime time.Time) {
			defer wg.Done()

			log.Info().Msgf(`tid:%s - Event comparison started: start period: %s, end period: %s with jurisdiction %s and caseType: %s`,
				transactionId, helper.FormatTimeStamp(startTime), helper.FormatTimeStamp(searchPeriodEndTime), jurisdiction, caseTypeId)

			err := s.proceedEventComparator(transactionId, jurisdiction, caseTypeId, startTime, searchPeriodEndTime, s.repo)
			if err != nil {
				log.Error().Msgf(`tid:%s - error occurred while event comparison: %s`, transactionId, err)
			}
		}(transactionId, jurisdiction, caseTypeId, startTime, searchPeriodEndTime)

		startTime = searchPeriodEndTime.Add(1 * time.Nanosecond)
	}

	wg.Wait()
	return nil
}

func (s *service) proceedEventComparator(transactionId, jurisdiction string, caseTypeId string,
	startTime time.Time, endTime time.Time, repo Repository) error {

	cases, err := repo.findCasesByJurisdictionInImpactPeriod(jurisdiction, caseTypeId, startTime, endTime)
	if err != nil {
		return errors.Wrap(err, "failed to find cases by jurisdiction and case type")
	}

	if len(cases) == 0 {
		log.Info().Msgf("No case data returned for jurisdiction=%s with caseTypeId=%s", jurisdiction, caseTypeId)
		return nil
	}

	log.Info().Msgf("tid:%s - Parsing case data with jurisdiction %s and caseType %s with %d case events",
		transactionId, jurisdiction, caseTypeId, len(cases))

	casesWithEventDetails := getCasesWithEventDetails(cases)
	if len(casesWithEventDetails) == 0 {
		log.Info().Msg("Couldn't find any case based on the search criteria provided")
		return nil
	}

	log.Info().Msgf("tid:%s - Event comparing started...", transactionId)
	eventDifferences := comparator.CompareCaseEvents(transactionId, casesWithEventDetails)
	if len(eventDifferences) == 0 {
		log.Info().Msg("No differences found in events for specified cases based on the search criteria provided")
		return nil
	}

	analyzeResult := comparator.NewEventDifferencesData(s.activeRules, eventDifferences).ProcessEventDiff()
	if !s.configuration.Report.Enabled {
		log.Info().Msgf("Analysis completed without saving the report. "+
			"Total records in analyzeResult: %d. Total number of field change: %d", analyzeResult.Size(), len(eventDifferences))
		return nil
	}

	err = s.saveReport(transactionId, analyzeResult, eventDifferences)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) saveReport(transactionId string, analyzeResult *comparator.AnalyzeResult, eventDifferences comparator.EventFieldDifferences) error {
	if analyzeResult.IsNotEmpty() || s.configuration.Report.IncludeEmptyChange {
		eventDataReportEntities, err := comparator.SaveReport(eventDifferences, analyzeResult, s.configuration)
		if err != nil {
			return errors.Wrap(err, "failed to process SaveReport")
		}

		numberOfRecord := len(eventDataReportEntities)
		if numberOfRecord == 0 {
			log.Info().Msg("No record returned from SaveReport")
			return nil
		}

		log.Info().Msgf("tid:%s - Saving report data to the database. Total record number: %d", transactionId, numberOfRecord)

		err = s.repo.saveAllEventDataReport(eventDataReportEntities)
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
