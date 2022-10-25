package jwt

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v4"
	"github.com/slava-911/URL-shortener/internal/config"
	"github.com/slava-911/URL-shortener/pkg/logging"
)

func Middleware(h http.HandlerFunc, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			logger.Error("Malformed token")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("The correct token is required for authorization"))
			return
		}
		logger.Debug("create jwt verifier")
		jwtToken := authHeader[1]
		key := []byte(config.GetConfig().JWT.Secret)
		verifier, err := jwt.NewVerifierHS(jwt.HS256, key)
		if err != nil {
			unauthorized(w, err, logger)
			return
		}
		logger.Debug("parse and verify token")
		newToken, err := jwt.Parse([]byte(jwtToken), verifier)
		if err != nil {
			unauthorized(w, err, logger)
			return
		}

		logger.Debug("parse user claims")
		var uc UserClaims
		err = json.Unmarshal(newToken.Claims(), &uc)
		if err != nil {
			unauthorized(w, err, logger)
			return
		}
		if valid := uc.IsValidAt(time.Now()); !valid {
			logger.Error("token has been expired")
			unauthorized(w, err, logger)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", uc.ID)
		h(w, r.WithContext(ctx))
	}
}

func unauthorized(w http.ResponseWriter, err error, logger *logging.Logger) {
	logger.Error(err)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("unauthorized"))
}
