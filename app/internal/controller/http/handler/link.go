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
	linksURL     = "/links"
	linkURL      = "/links/:id"
	shortLinkURL = "/s/:short_version"
)

type linkHandler struct {
	linkService interf.LinkService
	validate    *validator.Validate
	logger      *logging.Logger
}

func NewLinkHandler(ls interf.LinkService, v *validator.Validate, l *logging.Logger) interf.Handler {
	return &linkHandler{
		linkService: ls,
		validate:    v,
		logger:      l,
	}
}

func (h *linkHandler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, linksURL, jwt.Middleware(apperror.Middleware(h.CreateLink), h.logger))
	router.HandlerFunc(http.MethodGet, linksURL, jwt.Middleware(apperror.Middleware(h.GetUserLinks), h.logger))
	router.HandlerFunc(http.MethodGet, linkURL, jwt.Middleware(apperror.Middleware(h.GetLink), h.logger))
	router.HandlerFunc(http.MethodPatch, linkURL, jwt.Middleware(apperror.Middleware(h.PartiallyUpdateLink), h.logger))
	router.HandlerFunc(http.MethodDelete, linkURL, jwt.Middleware(apperror.Middleware(h.DeleteLink), h.logger))
	router.HandlerFunc(http.MethodGet, shortLinkURL, apperror.Middleware(h.ClickOnLink))
}

func (h *linkHandler) CreateLink(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("CREATE LINK")
	w.Header().Set("Content-Type", "application/json")

	vUserID := r.Context().Value("user_id")
	if vUserID == nil {
		h.logger.Error("there is no user_id in context")
		return apperror.ErrUnauthorized
	}
	userID := vUserID.(string)

	h.logger.Debug("decode create link dto")
	var linkDTO httpdto.CreateLinkDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&linkDTO); err != nil {
		return apperror.BadRequestError("invalid data")
	}
	linkDTO.UserID = userID

	h.logger.Debugf("Validation for link: %s", linkDTO.FullVersion)
	if err := h.validate.Struct(linkDTO); err != nil {
		return apperror.BadRequestError(utils.TranslateValidationError(err, ""))
	}
	if !httpdto.ValidLink(linkDTO.FullVersion) {
		return apperror.BadRequestError("Need an absolute path link to create a short link. Ex: https://p.com/")
	}

	linkID, err := h.linkService.Create(r.Context(), httpdto.NewLink(linkDTO))
	if err != nil {
		return err
	}
	w.Header().Set("Location", fmt.Sprintf("%s/%s", linksURL, linkID))
	w.WriteHeader(http.StatusCreated)

	return nil
}

func (h *linkHandler) GetUserLinks(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("GET USER LINKS")
	w.Header().Set("Content-Type", "application/json")

	vUserID := r.Context().Value("user_id")
	if vUserID == nil {
		h.logger.Error("there is no user_id in context")
		return apperror.ErrUnauthorized
	}
	userID := vUserID.(string)

	links, err := h.linkService.GetAllByUserID(r.Context(), userID)
	if err != nil {
		return err
	}

	linksBytes, err := json.Marshal(links)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(linksBytes)

	return nil
}

func (h *linkHandler) GetLink(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("GET LINK")
	w.Header().Set("Content-Type", "application/json")

	vUserID := r.Context().Value("user_id")
	if vUserID == nil {
		h.logger.Error("there is no user_id in context")
		return apperror.ErrUnauthorized
	}

	h.logger.Debug("get id from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	linkID := params.ByName("id")
	if linkID == "" {
		return apperror.BadRequestError("id query parameter is required")
	}

	link, err := h.linkService.GetOneByID(r.Context(), linkID)
	if err != nil {
		return err
	}

	linkBytes, err := json.Marshal(link)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(linkBytes)

	return nil
}

func (h *linkHandler) PartiallyUpdateLink(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("PARTIALLY UPDATE USER")
	w.Header().Set("Content-Type", "application/json")

	h.logger.Debug("get id from context")
	//params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	//userID := params.ByName("id")
	vUserID := r.Context().Value("user_id")
	if vUserID == nil {
		h.logger.Error("there is no user_id in context")
		return apperror.ErrUnauthorized
	}

	h.logger.Debug("get id from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	linkID := params.ByName("id")
	if linkID == "" {
		return apperror.BadRequestError("id query parameter is required")
	}

	h.logger.Debug("decode update user dto")
	var linkDTO httpdto.UpdateLinkDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&linkDTO); err != nil {
		return apperror.BadRequestError("failed to decode data")
	}

	changedFields := make(map[string]string, 0)
	if linkDTO.FullVersion != nil {
		if err := h.validate.Var(*linkDTO.FullVersion, "required,min=3,max=2000"); err != nil {
			return apperror.BadRequestError(utils.TranslateValidationError(err, "Link full version"))
		}
		if !httpdto.ValidLink(*linkDTO.FullVersion) {
			return apperror.BadRequestError("Need an absolute path link to create a short link. Ex: https://p.com/")
		}
		changedFields["full_version"] = *linkDTO.FullVersion
	}
	if linkDTO.Description != nil {
		changedFields["description"] = *linkDTO.Description
	}
	if len(changedFields) == 0 {
		return apperror.BadRequestError("Nothing to update")
	}

	err := h.linkService.Update(r.Context(), linkID, changedFields)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *linkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("DELETE LINK")
	w.Header().Set("Content-Type", "application/json")

	vUserID := r.Context().Value("user_id")
	if vUserID == nil {
		h.logger.Error("there is no user_id in context")
		return apperror.ErrUnauthorized
	}

	h.logger.Debug("get id from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	linkID := params.ByName("id")
	if linkID == "" {
		return apperror.BadRequestError("id query parameter is required")
	}

	err := h.linkService.Delete(r.Context(), linkID)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *linkHandler) ClickOnLink(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("CLICK ON THE LINK")
	w.Header().Set("Content-Type", "application/json")

	h.logger.Debug("get short_version from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	shortVersion := params.ByName("short_version")
	if shortVersion == "" {
		return apperror.BadRequestError("short_version query parameter is required")
	}

	fullLink, err := h.linkService.GetFullVersionByShortVersion(r.Context(), shortVersion)
	if err != nil {
		return err
	}

	h.logger.Infof("Redirected from link short version: %s", shortVersion)
	w.Header().Set("Location", fullLink)
	http.Redirect(w, r, fullLink, http.StatusTemporaryRedirect)

	return nil
}
