package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/slava-911/URL-shortener/internal/apperror"
	httpdto "github.com/slava-911/URL-shortener/internal/controller/http/dto"
	"github.com/slava-911/URL-shortener/internal/interf"
	"github.com/slava-911/URL-shortener/internal/jwt"
	"github.com/slava-911/URL-shortener/pkg/logging"
	"github.com/slava-911/URL-shortener/pkg/utils"
)

const (
	authURL   = "/auth"
	signupURL = "/signup"
	userURL   = "/profile"
)

type userHandler struct {
	jwtHelper   jwt.Helper
	userService interf.UserService
	validate    *validator.Validate
	logger      *logging.Logger
}

func NewUserHandler(h jwt.Helper, us interf.UserService, v *validator.Validate, l *logging.Logger) interf.Handler {
	return &userHandler{
		jwtHelper:   h,
		userService: us,
		validate:    v,
		logger:      l,
	}
}

func (h *userHandler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, authURL, apperror.Middleware(h.Auth))
	router.HandlerFunc(http.MethodPut, authURL, apperror.Middleware(h.Auth))
	router.HandlerFunc(http.MethodPost, signupURL, apperror.Middleware(h.Signup))
	router.HandlerFunc(http.MethodGet, userURL, jwt.Middleware(apperror.Middleware(h.GetUser), h.logger))
	router.HandlerFunc(http.MethodPatch, userURL, jwt.Middleware(apperror.Middleware(h.PartiallyUpdateUser), h.logger))
	router.HandlerFunc(http.MethodDelete, userURL, jwt.Middleware(apperror.Middleware(h.DeleteUser), h.logger))
}

func (h *userHandler) Signup(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var userDTO httpdto.CreateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&userDTO); err != nil {
		return apperror.BadRequestError("failed to decode data")
	}

	h.logger.Debugf("Validation for user: %s", userDTO.Name)
	if err := h.validate.Struct(userDTO); err != nil {
		return apperror.BadRequestError(utils.TranslateValidationError(err, ""))
	}

	h.logger.Debug("check password and repeat password")
	if userDTO.Password != userDTO.RepeatPassword {
		return apperror.BadRequestError("password does not match repeat password")
	}

	user, err := h.userService.Create(r.Context(), httpdto.NewUser(userDTO))
	if err != nil {
		return err
	}

	token, err := h.jwtHelper.GenerateAccessToken(user)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(token)

	return nil
}

func (h *userHandler) Auth(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	var token []byte
	var err error
	switch r.Method {
	case http.MethodPost:
		defer r.Body.Close()
		var userDTO httpdto.SigninUserDTO
		if err := json.NewDecoder(r.Body).Decode(&userDTO); err != nil {
			return apperror.BadRequestError("failed to decode data")
		}
		u, err := h.userService.GetOneByEmailAndPassword(r.Context(), userDTO.Email, userDTO.Password)
		if err != nil {
			return err
		}
		token, err = h.jwtHelper.GenerateAccessToken(u)
		if err != nil {
			return err
		}
	case http.MethodPut:
		defer r.Body.Close()
		var rt jwt.RT
		if err := json.NewDecoder(r.Body).Decode(&rt); err != nil {
			return apperror.BadRequestError("failed to decode data")
		}
		token, err = h.jwtHelper.UpdateRefreshToken(rt)
		if err != nil {
			return err
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(token)

	return err
}

func (h *userHandler) GetUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("GET USER")
	w.Header().Set("Content-Type", "application/json")

	h.logger.Debug("get id from context")
	vUserID := r.Context().Value("user_id")
	if vUserID == nil {
		h.logger.Error("there is no user_id in context")
		return apperror.ErrUnauthorized
	}
	userID := vUserID.(string)

	user, err := h.userService.GetOneByID(r.Context(), userID)
	if err != nil {
		return err
	}

	h.logger.Debug("marshal user")
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshall user. error: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(userBytes)
	return nil
}

func (h *userHandler) PartiallyUpdateUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("PARTIALLY UPDATE USER")
	w.Header().Set("Content-Type", "application/json")

	h.logger.Debug("get id from context")
	vUserID := r.Context().Value("user_id")
	if vUserID == nil {
		h.logger.Error("there is no user_id in context")
		return apperror.ErrUnauthorized
	}
	userID := vUserID.(string)

	h.logger.Debug("decode update user dto")
	var userDTO httpdto.UpdateUserDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&userDTO); err != nil {
		return apperror.BadRequestError("failed to decode data")
	}

	changedFields := make(map[string]string, 0)
	oldPassword := ""
	if userDTO.Name != nil {
		if err := h.validate.Var(*userDTO.Name, "required,min=2,max=50"); err != nil {
			return apperror.BadRequestError(utils.TranslateValidationError(err, "Name"))
		}
		changedFields["name"] = *userDTO.Name
	}
	if userDTO.Email != nil {
		if err := h.validate.Var(*userDTO.Email, "required,email"); err != nil {
			return apperror.BadRequestError(utils.TranslateValidationError(err, "Email"))
		}
		changedFields["email"] = *userDTO.Email
	}
	if userDTO.OldPassword != nil && userDTO.NewPassword != nil {
		if *userDTO.OldPassword != *userDTO.NewPassword && *userDTO.OldPassword != "" && *userDTO.NewPassword != "" {
			if err := h.validate.Var(*userDTO.NewPassword, "required,min=6"); err != nil {
				return apperror.BadRequestError(utils.TranslateValidationError(err, "New password"))
			}
			oldPassword = *userDTO.OldPassword
			changedFields["password"] = *userDTO.NewPassword
		}
	}
	if len(changedFields) == 0 {
		return apperror.BadRequestError("Nothing to update")
	}

	err := h.userService.Update(r.Context(), userID, changedFields, oldPassword)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *userHandler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("DELETE USER")
	w.Header().Set("Content-Type", "application/json")

	h.logger.Debug("get id from context")
	vUserID := r.Context().Value("user_id")
	if vUserID == nil {
		h.logger.Error("there is no user_id in context")
		return apperror.ErrUnauthorized
	}
	userID := vUserID.(string)

	err := h.userService.Delete(r.Context(), userID)
	if err != nil {
		return err
	}

	w.Header().Set("Location", signupURL)
	w.WriteHeader(http.StatusNoContent)

	return nil
}
