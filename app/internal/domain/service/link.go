package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/slava-911/URL-shortener/internal/apperror"
	"github.com/slava-911/URL-shortener/internal/domain/entity"
	"github.com/slava-911/URL-shortener/internal/interf"
	"github.com/slava-911/URL-shortener/pkg/logging"
)

type linkService struct {
	storage interf.LinkStorage
	logger  *logging.Logger
}

func NewLinkService(storage interf.LinkStorage, logger *logging.Logger) interf.LinkService {
	return &linkService{
		storage: storage,
		logger:  logger,
	}
}

func (s *linkService) Create(ctx context.Context, l entity.Link) (linkID string, err error) {
	l.GenerateShortVersion(7)

	linkID, err = s.storage.Create(ctx, l)
	if err != nil {
		s.logger.Error(err)
		return linkID, fmt.Errorf("failed to create link, error: %w", err)
	}

	return linkID, nil
}

func (s *linkService) GetAllByUserID(ctx context.Context, id string) (links []entity.Link, err error) {
	links, err = s.storage.FindAllByUserID(ctx, id)
	if err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return links, err
		}
		return links, fmt.Errorf("failed to get links by user id %s, error: %w", id, err)
	}

	return links, nil
}

func (s *linkService) GetOneByID(ctx context.Context, id string) (l entity.Link, err error) {
	l, err = s.storage.FindOneByID(ctx, id)
	if err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return l, err
		}
		return l, fmt.Errorf("failed to find link by id, error: %w", err)
	}

	return l, nil
}

func (s *linkService) Update(ctx context.Context, id string, chFields map[string]string) error {
	err := s.storage.Update(ctx, id, chFields)
	if err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to update link, error: %w", err)
	}

	return nil
}

func (s *linkService) Delete(ctx context.Context, id string) error {
	if err := s.storage.Delete(ctx, id); err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete user, error: %w", err)
	}

	return nil
}

func (s *linkService) GetFullVersionByShortVersion(ctx context.Context, shortVersion string) (fv string, err error) {
	fv, err = s.storage.FindFullVersionByShortVersion(ctx, shortVersion)
	if err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return fv, err
		}
		return fv, fmt.Errorf("failed to find link by short version, error: %w", err)
	}

	return fv, nil
}
