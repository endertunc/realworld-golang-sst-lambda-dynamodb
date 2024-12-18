package api

import (
	samberSlog "github.com/samber/slog-http"
	veqrynslog "github.com/veqryn/slog-context/http"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/security"

	"github.com/google/uuid"
)

type Middleware func(http.Handler) http.Handler
type AuthenticatedHandlerFunc func(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token)
type OptionallyAuthenticatedHandlerFunc func(w http.ResponseWriter, r *http.Request, userId *uuid.UUID, token *domain.Token)

var DefaultMiddlewares []Middleware = []Middleware{
	veqrynslog.AttrCollection,
	samberSlog.New(slog.Default()),
	RequestIdMiddleware,
}

// WithMiddlewares applies the given middlewares to the given handler in reverse order.
// given middlewares []Middleware{m1,m2,m3}, m1(m2(m3(handler))) is returned.
func WithMiddlewares(handler http.Handler, middlewares []Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func AuthenticatedHandler(handler AuthenticatedHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userId, token, ok := security.GetLoggedInUser(ctx, w, r)
		if !ok {
			return
		}
		handler(w, r, userId, token)
	})
}

func OptionallyAuthenticatedHandler(handler OptionallyAuthenticatedHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userId, token, ok := security.GetOptionalLoggedInUser(ctx, w, r)
		if !ok {
			return
		}
		handler(w, r, userId, token)
	})
}

// RequestIdMiddleware must be added to the middleware chain before samber/slog-http middleware.
// requestID is added to the veqryn/slog-context, and it will be included in all log lines.
func RequestIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := samberSlog.GetRequestIDFromContext(r.Context())
		veqrynslog.With(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r)
	})
}
