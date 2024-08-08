package security

import (
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

var (
	ErrAuthorizationHeaderMissing = errors.New("authorization header is missing")
	ErrAuthorizationHeaderEmpty   = errors.New("authorization header is empty")
	ErrInvalidTokenType           = errors.New("invalid token type")
)

func GetLoggedInUserWithError(request events.APIGatewayProxyRequest) (uuid.UUID, error) {
	authorizationHeader, present := request.Headers["Authorization"] // this is case-sensitive... https://github.com/aws/aws-lambda-go/issues/117
	if !present {
		return uuid.Nil, ErrAuthorizationHeaderMissing
	}

	authorizationHeader = strings.TrimSpace(authorizationHeader)
	if authorizationHeader == "" {
		return uuid.Nil, ErrAuthorizationHeaderEmpty
	}

	token, ok := strings.CutPrefix(authorizationHeader, "Token ")
	if !ok {
		return uuid.Nil, ErrInvalidTokenType
	}

	userId, err := ValidateToken(token)
	if err != nil {
		// ToDo @ender fix me!!!!
		return uuid.Nil, err
	}

	return userId, nil
}

func GetOptionalLoggedInUserWithError(request events.APIGatewayProxyRequest) (*uuid.UUID, error) {
	loggedInUser, err := GetLoggedInUserWithError(request)
	if err != nil {
		if errors.Is(err, ErrAuthorizationHeaderMissing) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &loggedInUser, nil
}

func GetOptionalLoggedInUser(request events.APIGatewayProxyRequest) (*uuid.UUID, *events.APIGatewayProxyResponse) {
	authorizationHeader, present := request.Headers["Authorization"] // this is case-sensitive... https://github.com/aws/aws-lambda-go/issues/117
	if !present {
		return nil, nil
	}

	userId, errResponse := getLoggedInUserFromHeader(authorizationHeader)
	if errResponse != nil {
		return nil, errResponse
	}

	return &userId, nil
}

func GetLoggedInUser(request events.APIGatewayProxyRequest) (uuid.UUID, *events.APIGatewayProxyResponse) {
	authorizationHeader, present := request.Headers["Authorization"] // this is case-sensitive... https://github.com/aws/aws-lambda-go/issues/117
	if !present {
		return uuid.Nil, &events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized, //
			Body:       "authorization header is missing",
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
	}

	userId, errResponse := getLoggedInUserFromHeader(authorizationHeader)
	if errResponse != nil {
		return uuid.Nil, errResponse
	}

	return userId, nil
}

func getLoggedInUserFromHeader(authorizationHeader string) (uuid.UUID, *events.APIGatewayProxyResponse) {
	authorizationHeader = strings.TrimSpace(authorizationHeader)
	if authorizationHeader == "" {
		return uuid.Nil, &events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized, //
			Body:       "authorization header is empty",
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
	}

	token, ok := strings.CutPrefix(authorizationHeader, "Token ")

	if !ok {
		return uuid.Nil, &events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized, //
			Body:       "invalid token type",
			Headers: map[string]string{
				"Content-Type":     "application/json",
				"WWW-Authenticate": "Bearer", // ToDo @ender optional
			},
		}
	}

	userId, err := ValidateToken(token)
	if err != nil {
		// ToDo @ender fix me!!!!
		return uuid.Nil, &events.APIGatewayProxyResponse{}
	}

	return userId, nil

}
