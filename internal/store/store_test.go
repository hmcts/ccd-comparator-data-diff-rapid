package store

import (
	"ccd-comparator-data-diff-rapid/config"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestInitializeDBFailure(t *testing.T) {
	dbInitOnce = sync.Once{} // Reset the sync.Once to allow reinitialization for this test case
	dbConfig := config.Configurations{Database: config.Database{}}
	db, err := initializeDB(&dbConfig.Database)
	assert.Nil(t, db)
	assert.Error(t, err)
}

func TestGetDataSourceURL(t *testing.T) {
	appConfig := config.GetConfigurations("../..", "config_test")

	expectedURL := "user=ccd password=ccd host=localhost port=5050 dbname=ccd_data sslmode=disable"
	url := composeDataSourceURL(&appConfig.Database)
	assert.Equal(t, expectedURL, url)
}
