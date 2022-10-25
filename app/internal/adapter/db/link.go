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

type linkStorage struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewLinkStorage(client postgresql.Client, logger *logging.Logger) interf.LinkStorage {
	return &linkStorage{
		client: client,
		logger: logger,
	}
}

func (s *linkStorage) Create(ctx context.Context, l entity.Link) (linkID string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		INSERT INTO links
			(full_version, short_version, description, clicked, user_id)
		VALUES
			($1, $2, $3, $4, $5)
		RETURNING id
	`
	s.logger.Tracef("SQL Query: %s", utils.FormatQuery(q))

	row := s.client.QueryRow(ctx, q, l.FullVersion, l.ShortVersion, l.Description, 0, l.UserID)
	if err = row.Scan(&linkID); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return linkID, detErr
		}
		return linkID, err
	}

	return linkID, nil
}

func (s *linkStorage) FindAllByUserID(ctx context.Context, userID string) (links []entity.Link, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		SELECT
		    l.id, l.full_version, l.short_version, COALESCE(l.description, ''), l.created_at, COALESCE(l.clicked, 0), l.user_id
		FROM
		    links l
		WHERE
		    l.user_id = $1
	`
	s.logger.Tracef("SQL Query: %s", utils.FormatQuery(q))

	rows, err := s.client.Query(ctx, q, userID)
	if err != nil {
		return links, err
	}

	for rows.Next() {
		var l entity.Link
		err = rows.Scan(&l.ID, &l.FullVersion, &l.ShortVersion, &l.Description, &l.CreatedAt, &l.Clicked, &l.UserID)
		if err != nil {
			if detErr := postgresql.DetailedPgError(err); detErr != nil {
				return links, detErr
			}
			return links, err
		}
		links = append(links, l)
	}

	if err = rows.Err(); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return links, detErr
		}
		return links, err
	}

	return links, nil
}

func (s *linkStorage) FindOneByID(ctx context.Context, id string) (l entity.Link, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		SELECT
		    l.id, l.full_version, l.short_version, COALESCE(l.description, ''), l.created_at, COALESCE(l.clicked, 0), l.user_id
		FROM
		    links l
		WHERE
		    l.id = $1
	`
	s.logger.Tracef("SQL Query: %s", utils.FormatQuery(q))

	row := s.client.QueryRow(ctx, q, id)
	err = row.Scan(&l.ID, &l.FullVersion, &l.ShortVersion, &l.Description, &l.CreatedAt, &l.Clicked, &l.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return l, apperror.ErrNotFound
		}
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return l, detErr
		}
		return l, err
	}
	return l, nil
}

func (s *linkStorage) Update(ctx context.Context, id string, chFields map[string]string) error {
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
		    links l
		SET
		    %s
		WHERE
		    l.id = $%d
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

func (s *linkStorage) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		DELETE FROM
		    links l
		WHERE
		    l.id = $1
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

func (s *linkStorage) FindFullVersionByShortVersion(ctx context.Context, sv string) (fv string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		UPDATE
		    links
   		SET
   		    clicked = clicked + 1
   		WHERE
		    short_version = $1
		RETURNING full_version
	`
	s.logger.Tracef("SQL Query: %s", utils.FormatQuery(q))

	row := s.client.QueryRow(ctx, q, sv)
	if err = row.Scan(&fv); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fv, apperror.ErrNotFound
		}
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return fv, detErr
		}
		return fv, err
	}

	return fv, nil
}
