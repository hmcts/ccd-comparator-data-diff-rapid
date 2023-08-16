package main

import (
	"bufio"
	"ccd-comparator-data-diff-rapid/comparator"
	"ccd-comparator-data-diff-rapid/config"
	"ccd-comparator-data-diff-rapid/domain"
	"ccd-comparator-data-diff-rapid/helper"
	"ccd-comparator-data-diff-rapid/internal/store"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var cpuProfile = flag.Bool("cpu-profile", false, "write cpu profile to `file`")
var memoryProfile = flag.Bool("mem-profile", false, "write memory profile to `file`")
var configFile = flag.String("configFile", "./config", "Configuration file")
var sourceFile = flag.String("sourceFile", "", "File contains existing casetypes")

func main() {
	fmt.Println("Starting...")

	flag.Parse()

	dirPath := filepath.Dir(*configFile)
	fileName := filepath.Base(*configFile)
	fileExt := filepath.Ext(fileName)
	fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)
	configurations := config.GetConfigurations(dirPath, fileNameWithoutExt)

	initiateLogger(configurations.Level, configurations.Type)
	defer elapsed("Execution")()

	if *cpuProfile {
		fmt.Println("Enabled CPUTrace")
		f := helper.StartCPUTrace()
		defer f()
	}

	if *memoryProfile {
		fmt.Println("Enabled MemoryProfile")
		f := helper.StartMemoryProfile()
		defer f()
	}

	dbx := store.LoadDB(configurations)
	repository := domain.NewRepository(dbx)

	log.Info().Msgf("The following roles have been enabled and will be applied: %s", configurations.Active)

	activeRules := comparator.NewRuleFactory(configurations).GetEnabledRuleList()
	service := domain.NewService(configurations, &activeRules, repository)

	startTime := helper.MustParseTime("", configurations.StartTime)
	endTime := helper.MustParseTime("", configurations.EndTime)

	var jurisdictionWithCaseTypes map[string][]string
	if *sourceFile != "" {
		jurisdictionWithCaseTypes = readCSVIntoMap(*sourceFile)

		var sortedJurisdictions []string
		for jurisdiction := range jurisdictionWithCaseTypes {
			sortedJurisdictions = append(sortedJurisdictions, jurisdiction)
		}
		sort.Strings(sortedJurisdictions)

		for _, jurisdiction := range sortedJurisdictions {
			caseTypes := jurisdictionWithCaseTypes[jurisdiction]
			for _, caseType := range caseTypes {
				log.Info().Msgf("Scanning for jurisdiction: %s and caseType: %s", jurisdiction, caseType)
				performEventComparison(service, jurisdiction, caseType, startTime, endTime)
			}
		}
		return
	} else {
		jurisdiction := configurations.Jurisdiction
		caseType := configurations.CaseType
		performEventComparison(service, jurisdiction, caseType, startTime, endTime)
	}

}

func initiateLogger(logLevel, logType string) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if logType == "file" {
		fileWriter, _ := openLogFile()
		log.Logger = zerolog.New(fileWriter).With().Timestamp().Logger()
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006/01/02 15:04:05"})
	}
}

func openLogFile() (*os.File, error) {
	logFileName := fmt.Sprintf("log_%s.log", time.Now().Format("2006-01-02_15-04-05"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal().Msgf("failed to open log file: %s", err)
	}
	fmt.Printf("Logs will be saved to %s", logFileName)
	return logFile, nil
}

func performEventComparison(service domain.Service, jurisdiction string, caseType string, startTime time.Time, endTime time.Time) {
	err := service.CompareEventsInImpactPeriod(jurisdiction, caseType, startTime, endTime)
	if err != nil {
		log.Fatal().Err(err)
	}
}

func readCSVIntoMap(csvFilePath string) map[string][]string {
	csvDataMap := make(map[string][]string)

	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatal().Msgf("Error opening CSV file: %s", err)
	}
	defer csvFile.Close()

	scanner := bufio.NewScanner(csvFile)

	// Skip header
	if scanner.Scan() {
	}

	for scanner.Scan() {
		line := scanner.Text()
		rowData := strings.Split(line, ",")

		if len(rowData) < 2 {
			continue
		}

		key := rowData[0]
		value := rowData[1]

		if _, found := csvDataMap[key]; !found {
			csvDataMap[key] = make([]string, 0)
		}

		csvDataMap[key] = append(csvDataMap[key], value)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal().Msgf("Error reading CSV file: %s", err)
	}

	return csvDataMap
}

func elapsed(msg string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v", msg, time.Since(start))
	}
}
