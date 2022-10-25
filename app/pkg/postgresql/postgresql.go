package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/slava-911/URL-shortener/pkg/logging"
	"github.com/slava-911/URL-shortener/pkg/utils"
)

type Client interface {
	Begin(context.Context) (pgx.Tx, error)
	BeginFunc(ctx context.Context, f func(pgx.Tx) error) error
	BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Close()
}

type dbConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
}

// NewDBConfig creates new pg config instance
func NewDBConfig(username string, password string, host string, port string, database string) *dbConfig {
	return &dbConfig{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		Database: database,
	}
}

// NewClient establishes a database connection and returns postgresql Client interface
func NewClient(maxAttempts int, delay time.Duration, cfg *dbConfig, logger *logging.Logger) (p Client, err error) {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	err = utils.DoWithAttempts(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), delay)
		defer cancel()

		pgxCfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			logger.Fatalf("Unable to parse config: %v\n", err)
		}

		p, err = pgxpool.ConnectConfig(ctx, pgxCfg)
		if err != nil {
			logger.Error("Failed to connect to postgres... Going to do the next attempt")

			return err
		}

		return nil
	}, maxAttempts, delay)

	if err != nil {
		logger.Fatal("All attempts are exceeded. Unable to connect to postgres")
	}

	return p, nil
}

func DetailedPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		pgErr = err.(*pgconn.PgError)
		newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		return newErr
	}
	return nil
}
