package security

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

var (
	ErrAuthorizationHeaderMissing = errors.New("authorization header is missing")
	ErrAuthorizationHeaderEmpty   = errors.New("authorization header is empty")
	ErrInvalidTokenType           = errors.New("invalid token type")
)

//func GetLoggedInUserWithError(request events.APIGatewayProxyRequest) (uuid.UUID, error) {
//	authorizationHeader, present := request.Headers["Authorization"] // this is case-sensitive... https://github.com/aws/aws-lambda-go/issues/117
//	if !present {
//		return uuid.Nil, ErrAuthorizationHeaderMissing
//	}
//
//	authorizationHeader = strings.TrimSpace(authorizationHeader)
//	if authorizationHeader == "" {
//		return uuid.Nil, ErrAuthorizationHeaderEmpty
//	}
//
//	token, ok := strings.CutPrefix(authorizationHeader, "Token ")
//	if !ok {
//		return uuid.Nil, ErrInvalidTokenType
//	}
//
//	userId, err := ValidateToken(token)
//	if err != nil {
//		// ToDo @ender fix me!!!!
//		return uuid.Nil, err
//	}
//
//	return userId, nil
//}

//func GetOptionalLoggedInUserWithError(request events.APIGatewayProxyRequest) (*uuid.UUID, error) {
//	loggedInUser, err := GetLoggedInUserWithError(request)
//	if err != nil {
//		if errors.Is(err, ErrAuthorizationHeaderMissing) {
//			return nil, nil
//		} else {
//			return nil, err
//		}
//	}
//	return &loggedInUser, nil
//}

func GetOptionalLoggedInUser(ctx context.Context, request events.APIGatewayProxyRequest) (*uuid.UUID, *domain.Token, *events.APIGatewayProxyResponse) {
	authorizationHeader, present := request.Headers["authorization"] // this is case-sensitive... https://github.com/aws/aws-lambda-go/issues/117
	if !present {
		return nil, nil, nil
	}

	userId, token, errResponse := getLoggedInUserFromHeader(ctx, authorizationHeader)
	if errResponse != nil {
		return nil, nil, errResponse
	}

	return &userId, &token, nil
}

func GetLoggedInUser(ctx context.Context, request events.APIGatewayProxyRequest) (uuid.UUID, domain.Token, *events.APIGatewayProxyResponse) {
	authorizationHeader, present := request.Headers["authorization"] // this is case-sensitive... https://github.com/aws/aws-lambda-go/issues/117
	if !present {
		slog.WarnContext(ctx, "missing authorization header")
		response := toSimpleError(ctx, http.StatusUnauthorized, "authorization header is missing")
		return uuid.Nil, "", &response
	}

	userId, token, errResponse := getLoggedInUserFromHeader(ctx, authorizationHeader)
	if errResponse != nil {
		return uuid.Nil, "", errResponse
	}

	return userId, token, nil
}

func getLoggedInUserFromHeader(ctx context.Context, authorizationHeader string) (uuid.UUID, domain.Token, *events.APIGatewayProxyResponse) {
	authorizationHeader = strings.TrimSpace(authorizationHeader)
	if authorizationHeader == "" {
		slog.WarnContext(ctx, "empty authorization header")
		response := toSimpleError(ctx, http.StatusUnauthorized, "authorization header is empty")
		return uuid.Nil, "", &response
	}

	token, ok := strings.CutPrefix(authorizationHeader, "Token ")

	if !ok {
		slog.WarnContext(ctx, "invalid token type")
		response := toSimpleError(ctx, http.StatusUnauthorized, "invalid token type")
		return uuid.Nil, "", &response
	}

	userId, err := ValidateToken(token)
	if err != nil {
		// you should never(?) log token - this is only for development purposes
		slog.WarnContext(ctx, "invalid token", slog.Any("error", err), slog.Any("token", token))
		response := toSimpleError(ctx, http.StatusUnauthorized, "invalid token")
		return uuid.Nil, "", &response
	}

	return userId, domain.Token(token), nil
}

// ToDo @ender this is a duplicate of the one in api/response-helpers. We can't reference it due to the cyclic dependency
func toSimpleError(ctx context.Context, statusCode int, message string) events.APIGatewayProxyResponse {
	body, err := json.Marshal(errutil.SimpleError{Message: message})
	if err != nil {
		slog.ErrorContext(ctx, "error encoding response body", slog.Any("error", err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "internal server error",
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

// HTTP-specific authentication functions

func GetOptionalLoggedInUserHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) (*uuid.UUID, *domain.Token, bool) {
	authorizationHeader, found := r.Header["Authorization"]
	if !found && len(authorizationHeader) == 0 {
		return nil, nil, true
	}

	userId, token, err := getLoggedInUserFromHeaderHTTP(ctx, w, authorizationHeader[0])
	if err {
		return nil, nil, false
	}

	return &userId, &token, true
}

func GetLoggedInUserHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) (uuid.UUID, domain.Token, bool) {
	authorizationHeader, found := r.Header["Authorization"]
	if !found && len(authorizationHeader) == 0 {
		slog.WarnContext(ctx, "missing authorization header")
		toSimpleHTTPError(w, http.StatusUnauthorized, "authorization header is missing")
		return uuid.Nil, "", false
	}

	userId, token, err := getLoggedInUserFromHeaderHTTP(ctx, w, authorizationHeader[0])
	if err {
		return uuid.Nil, "", false
	}

	return userId, token, true
}

func getLoggedInUserFromHeaderHTTP(ctx context.Context, w http.ResponseWriter, authorizationHeader string) (uuid.UUID, domain.Token, bool) {
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
	w.Write(body)
}
