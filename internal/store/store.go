package store

import (
	"fmt"
	"sync"
	"time"

	"ccd-comparator-data-diff-rapid/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

const (
	maxOpenConnection = 25
	maxIdleConnection = 25
	connMaxLifetime   = 5 * time.Minute
)

type Transaction interface {
	NamedExec(query string, arg interface{}) (interface{}, error)
	Commit() error
	Rollback() error
}

type txWrapper struct {
	tx *sqlx.Tx
}

func (t txWrapper) NamedExec(query string, arg interface{}) (interface{}, error) {
	return t.tx.NamedExec(query, arg)
}

func (t txWrapper) Commit() error {
	return t.tx.Commit()
}

func (t txWrapper) Rollback() error {
	return t.tx.Rollback()
}

type DB interface {
	MustBegin() Transaction
	Select(dest interface{}, query string, args ...interface{}) error
}

type sqlxDB struct {
	dbx *sqlx.DB
}

func (s sqlxDB) Select(dest interface{}, query string, args ...interface{}) error {
	return s.dbx.Select(dest, query, args...)
}

func (s sqlxDB) MustBegin() Transaction {
	tx := s.dbx.MustBegin()
	return txWrapper{tx}
}

var (
	dbInitOnce sync.Once
	dbInstance DB
)

func InitDatabase(configuration *config.Configurations) DB {
	dbInitOnce.Do(func() {
		var dbConfig = configuration.Database
		var err error

		var db *sqlx.DB
		db, err = initializeDB(&dbConfig)
		if err != nil {
			log.Fatal().Msgf("Failed to initialize the database: %s", err)
		}

		dbInstance = &sqlxDB{db}
	})
	return dbInstance
}

func initializeDB(config *config.Database) (*sqlx.DB, error) {
	dataSourceURL := composeDataSourceURL(config)

	db, err := sqlx.Open(config.Driver, dataSourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConnection)
	db.SetMaxIdleConns(maxIdleConnection)
	db.SetConnMaxLifetime(connMaxLifetime)

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	log.Info().Msg("Database connection is successful.")

	return db, nil
}

func composeDataSourceURL(config *config.Database) string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s", config.Username,
		config.Password, config.Host, config.Port, config.Name, config.SslMode)
}
