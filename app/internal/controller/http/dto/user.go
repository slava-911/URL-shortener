package dto

import "github.com/slava-911/URL-shortener/internal/domain/entity"

type SigninUserDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserDTO struct {
	Name           string `json:"name" validate:"required,min=2,max=50"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	RepeatPassword string `json:"repeat_password" validate:"required,min=6"`
}

func NewUser(d CreateUserDTO) entity.User {
	return entity.User{
		Name:     d.Name,
		Email:    d.Email,
		Password: d.Password,
	}
}

type UpdateUserDTO struct {
	Name        *string `json:"name,omitempty"`
	Email       *string `json:"email,omitempty"`
	OldPassword *string `json:"old_password,omitempty"`
	NewPassword *string `json:"new_password,omitempty"`
}
