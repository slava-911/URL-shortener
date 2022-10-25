package interf

import (
	"context"

	"github.com/julienschmidt/httprouter"
	"github.com/slava-911/URL-shortener/internal/domain/entity"
)

type Handler interface {
	Register(router *httprouter.Router)
}

type UserStorage interface {
	Create(ctx context.Context, u entity.User) (entity.User, error)
	FindOneByEmail(ctx context.Context, email string) (entity.User, error)
	FindOneByID(ctx context.Context, id string) (entity.User, error)
	Update(ctx context.Context, id string, chFields map[string]string) error
	Delete(ctx context.Context, id string) error
}

type UserService interface {
	Create(ctx context.Context, u entity.User) (entity.User, error)
	GetOneByEmailAndPassword(ctx context.Context, email, password string) (entity.User, error)
	GetOneByID(ctx context.Context, id string) (entity.User, error)
	Update(ctx context.Context, id string, chFields map[string]string, oldPass string) error
	Delete(ctx context.Context, id string) error
}

type LinkStorage interface {
	Create(ctx context.Context, l entity.Link) (string, error)
	FindAllByUserID(ctx context.Context, id string) ([]entity.Link, error)
	FindOneByID(ctx context.Context, id string) (entity.Link, error)
	Update(ctx context.Context, id string, chFields map[string]string) error
	Delete(ctx context.Context, id string) error
	FindFullVersionByShortVersion(ctx context.Context, shortVersion string) (string, error)
}

type LinkService interface {
	Create(ctx context.Context, l entity.Link) (string, error)
	GetAllByUserID(ctx context.Context, id string) ([]entity.Link, error)
	GetOneByID(ctx context.Context, id string) (entity.Link, error)
	Update(ctx context.Context, id string, chFields map[string]string) error
	Delete(ctx context.Context, id string) error
	GetFullVersionByShortVersion(ctx context.Context, shortVersion string) (string, error)
}
