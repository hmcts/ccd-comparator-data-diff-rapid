package store

import (
	"ccd-comparator-data-diff-rapid/config"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

const (
	maxOpenConnection = 25
	maxIdleConnection = 25
	connMaxLifetime   = 5 * time.Minute
)

var db *sqlx.DB
var once sync.Once

func LoadDB(configuration *config.Configurations) *sqlx.DB {
	once.Do(func() {
		var dbConfig = configuration.Database
		var err error

		db, err = initializeDB(&dbConfig)
		if err != nil {
			log.Fatal().Msgf("Failed to initialize the database: %s", err)
		}
	})
	return db
}

func initializeDB(config *config.Database) (*sqlx.DB, error) {
	dataSourceURL := getDataSourceURL(config)

	db, err := sqlx.Open(config.Driver, dataSourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConnection)
	db.SetMaxIdleConns(maxIdleConnection)
	db.SetConnMaxLifetime(connMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	log.Info().Msg("Database connection is successful.")

	return db, nil
}

func getDataSourceURL(config *config.Database) string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s", config.Username,
		config.Password, config.Host, config.Port, config.Name, config.SslMode)
}
