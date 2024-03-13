package main

import (
	"bufio"
	"ccd-comparator-data-diff-rapid/comparator"
	"ccd-comparator-data-diff-rapid/config"
	"ccd-comparator-data-diff-rapid/domain"
	"ccd-comparator-data-diff-rapid/helper"
	"ccd-comparator-data-diff-rapid/internal/store"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var cpuProfile = flag.Bool("cpu-profile", false, "write cpu profile to `file`")
var memoryProfile = flag.Bool("mem-profile", false, "write memory profile to `file`")
var configFile = flag.String("configFile", "./config", "Configuration file")
var sourceFile = flag.String("sourceFile", "", "File contains existing case types")

func main() {
	fmt.Println("Starting...")

	flag.Parse()

	fmt.Printf("Using the target configuration file: %s\n", *configFile)

	configurations := loadConfig(*configFile)
	validateConfigurations(*configurations)
	printConfigurations(*configurations)

	initiateLogger(configurations.Level, configurations.Type)
	defer elapsed("Execution")()

	enableAndManageProfiles()

	activeRules := comparator.NewRuleFactory(configurations).GetEnabledRuleList()
	db := store.InitDatabase(configurations)
	queryRepo := domain.NewQueryRepository(db)
	saveRepo := domain.NewSaveRepository(db)
	service := domain.NewService(configurations, &activeRules, queryRepo, saveRepo)

	orchestrateEventComparisons(service, configurations)
}

func validateConfigurations(c config.Configurations) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal().Msgf("Validation error: %s", r)
		}
	}()

	// Check period start time
	if isEmpty(c.Period.StartTime) {
		log.Fatal().Msg("Validation error: Period start time is empty. Please provide a valid start time.")
	} else {
		helper.MustParseTime("", c.Period.StartTime)
	}

	// Check Jurisdiction and CaseId
	if isEmpty(c.Jurisdiction) && isEmpty(c.CaseId) {
		log.Fatal().Msg("Validation error: Either Jurisdiction or CaseId must be set. Please provide one of them.")
	}
}

func isEmpty(value string) bool {
	if len(strings.TrimSpace(value)) == 0 {
		return true
	}
	return false
}

func loadConfig(configFile string) *config.Configurations {
	dirPath := filepath.Dir(configFile)
	fileName := filepath.Base(configFile)
	fileExt := filepath.Ext(fileName)
	fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)
	return config.GetConfigurations(dirPath, fileNameWithoutExt)
}

func orchestrateEventComparisons(service *domain.Service, configurations *config.Configurations) {
	log.Info().Msgf("Enabled roles: %s", configurations.Active)

	startTime := helper.MustParseTime("", configurations.StartTime)

	var endTime time.Time
	if strings.TrimSpace(configurations.EndTime) == "" {
		endTime = time.Now()
	} else {
		endTime = helper.MustParseTime("", configurations.EndTime)
	}

	if *sourceFile != "" {
		jurisdictionWithCaseTypes := readCSVIntoMap(*sourceFile)

		var sortedJurisdictions []string
		for jurisdiction := range jurisdictionWithCaseTypes {
			sortedJurisdictions = append(sortedJurisdictions, jurisdiction)
		}
		sort.Strings(sortedJurisdictions)

		for _, jurisdiction := range sortedJurisdictions {
			caseTypes := jurisdictionWithCaseTypes[jurisdiction]
			for _, caseType := range caseTypes {
				log.Info().Msgf("Scanning - jurisdiction: %s and caseType: %s", jurisdiction, caseType)
				performEventComparisonByJurisdiction(service, jurisdiction, caseType, startTime, endTime)
			}
		}
		return
	}

	performEventComparisonByJurisdiction(service, configurations.Jurisdiction, configurations.CaseType, startTime, endTime)
}

func performEventComparisonByJurisdiction(service *domain.Service, jurisdiction string, caseType string, startTime time.Time, endTime time.Time) {
	comparison := domain.Comparison{
		Jurisdiction:        jurisdiction,
		CaseTypeId:          caseType,
		StartTime:           startTime,
		SearchPeriodEndTime: endTime,
	}
	service.CompareEventsInImpactPeriod(comparison)
}

func enableAndManageProfiles() {
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
}

func initiateLogger(logLevel, logType string) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Warn().Msgf("Invalid log level '%s', defaulting to 'info'", logLevel)
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006/01/02 15:04:05"})

	if logType == "file" {
		fileWriter, err := openLogFile()
		if err != nil {
			log.Warn().Msgf("err opening log file: %s, continuing with 'info'", err)
		} else {
			log.Logger = zerolog.New(fileWriter).With().Timestamp().Logger()
		}
	}
}

func openLogFile() (*os.File, error) {
	logFileName := fmt.Sprintf("log_%s.log", time.Now().Format("2006-01-02_15-04-05"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	fmt.Printf("Logs will be saved to %s\n", logFileName)

	return logFile, nil
}

func readCSVIntoMap(csvFilePath string) map[string][]string {
	csvDataMap := make(map[string][]string)

	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatal().Msgf("err opening CSV file: %s", err)
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
		log.Fatal().Msgf("err reading CSV file: %s", err)
	}

	return csvDataMap
}

func maskPassword(config string) string {
	pattern := `"Password":\s*"[^"]*"`

	r, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	return r.ReplaceAllString(config, `"Password": "***"`)
}

func printConfigurations(s interface{}) {
	var p []byte
	p, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	configString := fmt.Sprintf("%s", p)
	maskedConfigString := maskPassword(configString)
	fmt.Println(maskedConfigString)
}

func elapsed(msg string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v", msg, time.Since(start))
	}
}
