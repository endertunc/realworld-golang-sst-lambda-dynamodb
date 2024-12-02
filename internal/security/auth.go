package security

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"strings"

	"github.com/google/uuid"
)

func GetOptionalLoggedInUser(ctx context.Context, w http.ResponseWriter, r *http.Request) (*uuid.UUID, *domain.Token, bool) {
	authorizationHeader, found := r.Header["Authorization"]
	if !found && len(authorizationHeader) == 0 {
		return nil, nil, true
	}

	userId, token, err := getLoggedInUserFromHeader(ctx, w, authorizationHeader[0])
	if err {
		return nil, nil, false
	}

	return &userId, &token, true
}

func GetLoggedInUser(ctx context.Context, w http.ResponseWriter, r *http.Request) (uuid.UUID, domain.Token, bool) {
	authorizationHeader, found := r.Header["Authorization"]
	if !found && len(authorizationHeader) == 0 {
		slog.WarnContext(ctx, "missing authorization header")
		toSimpleHTTPError(w, http.StatusUnauthorized, "authorization header is missing")
		return uuid.Nil, "", false
	}

	userId, token, err := getLoggedInUserFromHeader(ctx, w, authorizationHeader[0])
	if err {
		return uuid.Nil, "", false
	}

	return userId, token, true
}

func getLoggedInUserFromHeader(ctx context.Context, w http.ResponseWriter, authorizationHeader string) (uuid.UUID, domain.Token, bool) {
	authorizationHeader = strings.TrimSpace(authorizationHeader)
	if authorizationHeader == "" {
		slog.WarnContext(ctx, "empty authorization header")
		toSimpleHTTPError(w, http.StatusUnauthorized, "authorization header is empty")
		return uuid.Nil, "", true
	}

	token, ok := strings.CutPrefix(authorizationHeader, "Token ")
	if !ok {
		slog.WarnContext(ctx, "invalid token type")
		toSimpleHTTPError(w, http.StatusUnauthorized, "invalid token type")
		return uuid.Nil, "", true
	}

	userId, err := ValidateToken(token)
	if err != nil {
		// you should never(?) log token - this is only for development purposes
		slog.WarnContext(ctx, "invalid token", slog.Any("error", err), slog.Any("token", token))
		toSimpleHTTPError(w, http.StatusUnauthorized, "invalid token")
		return uuid.Nil, "", true
	}

	return userId, domain.Token(token), false
}

func toSimpleHTTPError(w http.ResponseWriter, statusCode int, message string) {
	body, err := json.Marshal(errutil.SimpleError{Message: message})
	if err != nil {
		slog.Error("error encoding response body", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}
