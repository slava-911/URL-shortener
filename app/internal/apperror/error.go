package apperror

import (
	"encoding/json"
	"fmt"
)

var (
	ErrNotFound     = NewAppError("not found", "US-010", "")
	ErrUnauthorized = NewAppError("unauthorized", "US-003", "")
)

type AppError struct {
	Err              error  `json:"-"`
	Message          string `json:"message,omitempty"`
	DeveloperMessage string `json:"developer_message,omitempty"`
	Code             string `json:"code,omitempty"`
}

func NewAppError(message, code, developerMessage string) *AppError {
	return &AppError{
		Err:              fmt.Errorf(message),
		Code:             code,
		Message:          message,
		DeveloperMessage: developerMessage,
	}
}

func (e *AppError) Error() string {
	return e.Err.Error()
}

func (e *AppError) Unwrap() error { return e.Err }

func (e *AppError) Marshal() []byte {
	bytes, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return bytes
}

func BadRequestError(message string) *AppError {
	return NewAppError(message, "US-002", "something wrong with user data")
}

func systemError(developerMessage string) *AppError {
	return NewAppError("system error", "US-001", developerMessage)
}
