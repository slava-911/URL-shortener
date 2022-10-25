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

type userService struct {
	storage interf.UserStorage
	logger  *logging.Logger
}

func NewUserService(userStorage interf.UserStorage, logger *logging.Logger) interf.UserService {
	return &userService{
		storage: userStorage,
		logger:  logger,
	}
}

func (s *userService) Create(ctx context.Context, u entity.User) (user entity.User, err error) {
	s.logger.Debug("generate password hash")
	if err = u.GeneratePasswordHash(); err != nil {
		s.logger.Errorf("failed to create user due to error %v", err)
		return u, err
	}

	user, err = s.storage.Create(ctx, u)
	if err != nil {
		s.logger.Error(err)
		return user, fmt.Errorf("failed to create user, error: %w", err)
	}

	return user, nil
}

func (s *userService) GetOneByEmailAndPassword(ctx context.Context, email, password string) (u entity.User, err error) {
	u, err = s.storage.FindOneByEmail(ctx, email)
	if err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return u, err
		}
		return u, fmt.Errorf("failed to find user by email, error: %w", err)
	}

	if err = u.CheckPassword(password); err != nil {
		return u, apperror.ErrNotFound
	}

	return u, nil
}

func (s *userService) GetOneByID(ctx context.Context, id string) (u entity.User, err error) {
	u, err = s.storage.FindOneByID(ctx, id)
	if err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return u, err
		}
		return u, fmt.Errorf("failed to find user by id, error: %w", err)
	}

	return u, nil
}

func (s *userService) Update(ctx context.Context, id string, chFields map[string]string, oldPass string) error {

	if oldPass != "" {
		s.logger.Debug("get user by uuid")
		user, err := s.GetOneByID(ctx, id)
		if err != nil {
			return err
		}

		s.logger.Debug("compare hash current password and old password")
		if err = user.CheckPassword(oldPass); err != nil {
			return apperror.BadRequestError("old password does not match current password")
		}

		user.Password = chFields["password"]

		s.logger.Debug("generate password hash")
		if err = user.GeneratePasswordHash(); err != nil {
			return fmt.Errorf("failed to update user, error %w", err)
		}

		chFields["password"] = user.Password
	}

	if err := s.storage.Update(ctx, id, chFields); err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to update user, error: %w", err)
	}

	return nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	if err := s.storage.Delete(ctx, id); err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete user, error: %w", err)
	}

	return nil
}
