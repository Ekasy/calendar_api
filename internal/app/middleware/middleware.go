package middleware

import (
	"context"
	"net/http"
	"nocalendar/internal/app/auth"

	"github.com/sirupsen/logrus"
)

type contextKey string

const ContextUserKey contextKey = "user_key"

func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

type AuthMiddleware struct {
	authUsecase auth.AuthUsecase
	logger      *logrus.Logger
}

func NewAuthMiddleware(authUsecase auth.AuthUsecase, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authUsecase: authUsecase,
		logger:      logger,
	}
}

func (am *AuthMiddleware) TokenChecking(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorize")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "unauthorized"}`))
			return
		}

		usr, err := am.authUsecase.GetUserByToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "unauthorized"}`))
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserKey, usr)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
