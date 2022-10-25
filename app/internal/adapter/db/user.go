package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/slava-911/URL-shortener/internal/apperror"
	"github.com/slava-911/URL-shortener/internal/domain/entity"
	"github.com/slava-911/URL-shortener/internal/interf"
	"github.com/slava-911/URL-shortener/pkg/logging"
	"github.com/slava-911/URL-shortener/pkg/postgresql"
	"github.com/slava-911/URL-shortener/pkg/utils"
)

type userStorage struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewUserStorage(client postgresql.Client, logger *logging.Logger) interf.UserStorage {
	return &userStorage{
		client: client,
		logger: logger,
	}
}

func (s *userStorage) Create(ctx context.Context, u entity.User) (user entity.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		INSERT INTO users
			(name, email, password)
		VALUES
			($1, $2, $3)
		RETURNING id
	`
	s.logger.Tracef("SQL Query: %s", utils.FormatQuery(q))

	row := s.client.QueryRow(ctx, q, u.Name, u.Email, u.Password)
	if err = row.Scan(&u.ID); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return u, detErr
		}
		return u, err
	}

	return u, nil
}

func (s *userStorage) FindOneByID(ctx context.Context, id string) (u entity.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		SELECT
		    u.id, u.name, u.email, u.password
		FROM
		    users u
		WHERE
		    u.id = $1
	`
	s.logger.Tracef("SQL Query: %s", utils.FormatQuery(q))

	row := s.client.QueryRow(ctx, q, id)
	if err = row.Scan(&u.ID, &u.Name, &u.Email, &u.Password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return u, apperror.ErrNotFound
		}
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return u, detErr
		}
		return u, err
	}
	return u, nil
}

func (s *userStorage) FindOneByEmail(ctx context.Context, email string) (u entity.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		SELECT
		    u.id, u.name, u.email, u.password
		FROM
		    users u
		WHERE
		    u.email = $1
	`
	s.logger.Tracef("SQL Query: %s", utils.FormatQuery(q))

	row := s.client.QueryRow(ctx, q, email)
	if err = row.Scan(&u.ID, &u.Name, &u.Email, &u.Password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return u, apperror.ErrNotFound
		}
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return u, detErr
		}
		return u, err
	}
	return u, nil
}

func (s *userStorage) Update(ctx context.Context, id string, chFields map[string]string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fields := make([]string, 0)
	params := make([]interface{}, 0)
	paramNum := 1

	for k, v := range chFields {
		fields = append(fields, fmt.Sprintf("%s=$%d", k, paramNum))
		params = append(params, v)
		paramNum++
	}

	fieldsToSet := strings.Join(fields, ", ")

	q := `
		UPDATE
		    users u
		SET
		    %s
		WHERE
		    u.id = $%d
	`
	q = fmt.Sprintf(q, fieldsToSet, paramNum)
	s.logger.Tracef("SQL Query: %s", utils.FormatQuery(q))

	params = append(params, id)
	s.logger.Tracef("params: %s", params)

	if _, err := s.client.Exec(ctx, q, params...); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}

	return nil
}

func (s *userStorage) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		DELETE FROM
		    users u
		WHERE
		    u.id = $1
	`
	s.logger.Tracef("SQL Query: %s", utils.FormatQuery(q))

	if _, err := s.client.Exec(ctx, q, id); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}

	return nil
}
