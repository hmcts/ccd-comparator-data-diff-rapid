package config

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"strings"
	"sync"
)

type Configurations struct {
	Database
	Period
	Worker
	Rule
	Scan
	Log
}

type Database struct {
	Username       string
	Password       string
	Host           string
	Port           int
	Name           string
	Driver         string
	SslMode        string
	BatchSize      int
	EventDataTable string
}

type Period struct {
	StartTime    string
	EndTime      string
	SearchWindow int
}

type Worker struct {
	Pool int
}

type Rule struct {
	Active string
}

type Scan struct {
	Jurisdiction         string
	CaseType             string
	CaseId               string
	MaxEventProcessCount int
	BatchSize            int
	Concurrent           struct {
		Event struct {
			ThresholdMilliseconds int64
		}
	}
	FieldChange struct {
		Threshold int
	}

	Report struct {
		Enabled            bool
		MaskValue          bool
		IncludeEmptyChange bool
		IncludeNoChange    bool
	}
}

type Log struct {
	Level string
	Type  string
}

var once sync.Once
var appConfigs *Configurations

func GetConfigurations(path, fileName string) *Configurations {
	once.Do(func() {
		appConfigs = loadConfig(path, fileName)
	})
	return appConfigs
}

func loadConfig(path, fileName string) *Configurations {
	viper.SetConfigName(fileName)
	viper.AddConfigPath(path)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Msgf("err reading config file: %s", err)
	}

	if err := bindEnvironmentVariables(); err != nil {
		log.Fatal().Err(err)
	}

	defaultBindings()

	var appConfigs *Configurations

	if err := viper.Unmarshal(&appConfigs); err != nil {
		log.Fatal().Msgf("Failed to unmarshal appConfigs: %s", err)
	}

	return appConfigs
}

func defaultBindings() {
	viper.SetDefault("database.sslmode", "disable")
}

func bindEnvironmentVariables() error {
	for _, key := range viper.AllKeys() {
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		err := viper.BindEnv(key, envKey)
		if err != nil {
			return errors.Errorf("config: unable to bind env: %s\n", err.Error())
		}
	}

	return nil
}
