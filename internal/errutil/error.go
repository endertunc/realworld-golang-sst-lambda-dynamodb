package errutil

import (
	"errors"
)

type GenericError struct {
	Message string `json:"message"`
}

//type AppError struct {
//	Code            int    `json:"code"`
//	Message         string `json:"message"`
//	Operation       string `json:"-"`
//	InternalMessage string `json:"-"`
//	Cause           error  `json:"-"`
//}
//
//func (e AppError) Error() string {
//	return e.Message
//}
//
//func (e AppError) Unwrap() error {
//	return e.Cause
//}
//
//func (e AppError) WithInternalMessage(message string) AppError {
//	e.InternalMessage = message
//	return e
//}
//
//func BadRequestError(operation, message string) AppError {
//	return AppError{
//		Code:      http.StatusBadRequest,
//		Message:   message,
//		Operation: operation,
//	}
//}
//
//func UnauthorizedError(operation, message string) AppError {
//	return AppError{
//		Code:      http.StatusUnauthorized,
//		Message:   message,
//		Operation: operation,
//	}
//}
//
//func UserNotFound(operation, message string, cause error) AppError {
//	return AppError{
//		Code:      http.StatusBadRequest,
//		Message:   message,
//		Operation: operation,
//		Cause:     cause,
//	}
//}
//
//func NotFoundError(operation, message string) AppError {
//	return AppError{
//		Code:      http.StatusNotFound,
//		Message:   message,
//		Operation: operation,
//	}
//}
//
//func ConflictError(operation string, message string) AppError {
//	return AppError{
//		Code:      http.StatusConflict,
//		Message:   message,
//		Operation: operation,
//	}
//}
//
//func InternalError(operation string) AppError {
//	return AppError{
//		Code:      http.StatusInternalServerError,
//		Message:   "internal error",
//		Operation: operation,
//	}
//}
//
//func (e AppError) Errorf(format string, args ...any) AppError {
//	err := fmt.Errorf(format, args...)
//
//	return AppError{
//		Code:            e.Code,
//		Message:         e.Message,
//		Operation:       e.Operation,
//		InternalMessage: err.Error(),
//		Cause:           fmt.Errorf("%w: %w", e, err),
//	}
//}
//
//func (e AppError) Errorf(format string, args ...any) AppError {
//	err := fmt.Errorf(format, args...)
//
//	return AppError{
//		Code:            e.Code,
//		InternalMessage: err.Error(),
//		Message:         e.Message,
//		Operation:       e.Operation,
//		Cause:           errors.Unwrap(err),
//	}
//}

//func ErrPathParamMissing(operation, message string) AppError {
//	return BadRequestError(operation, message)
//}

// ToAPIGatewayProxyResponse ToDo @ender this function should log the errors
// ToDo @ender this function and api.ToErrorAPIGatewayProxyResponse are doing the same thing
//func ToAPIGatewayProxyResponse(context context.Context, handlerName string, error error) events.APIGatewayProxyResponse {
//
//	// var appErr *AppError // This doesn't work???
//	appErr := AppError{}
//	ok := errors.As(error, &appErr)
//	if !ok {
//		slog.Error("unexpected error", slog.String("handler", handlerName), slog.Any("error", error))
//
//		internalServerError, err := json.Marshal(AppError{
//			Code:    http.StatusInternalServerError,
//			Message: "internal error",
//			Cause:   nil,
//		})
//		if err != nil {
//			return events.APIGatewayProxyResponse{
//				StatusCode: http.StatusInternalServerError,
//				Body:       string(internalServerError),
//				Headers:    map[string]string{"Content-Type": "application/json"},
//			}
//		}
//		return events.APIGatewayProxyResponse{
//			StatusCode: http.StatusInternalServerError,
//			Body:       "internal error",
//		}
//	}
//
//	responseBody, err := json.Marshal(appErr)
//	if err != nil {
//		return events.APIGatewayProxyResponse{
//			StatusCode: http.StatusInternalServerError,
//			Body:       "internal error",
//		}
//	}
//
//	//slog.Error("", slog.String("handler", handlerName), slog.Any("error", appErr.Cause))
//	return events.APIGatewayProxyResponse{
//		StatusCode: appErr.Code,
//		Body:       string(responseBody),
//		Headers:    map[string]string{"Content-Type": "application/json"},
//	}
//}

// note on where to define errors (and I am not still sure which one I like better):
// - define errors in where they are returned: errors are closely related to the functions that return them.
// - define errors in a single place: easy to find and manage all errors in one place.
var (
	//ErrUserNotFound = NotFoundError("user.not_found", "user not found")
	//ErrUsernameAlreadyExists = ConflictError("user.already_exists.username", "username already exists")
	//ErrEmailAlreadyExists    = ConflictError("user.already_exists.email", "email already exists")
	//ErrDynamoQuery = InternalError("dynamodb.query")
	//ErrDynamoMapping = InternalError("dynamodb.mapping")
	//ErrHashPassword          = BadRequestError("password.hash", "invalid credentials")
	//ErrTokenGenerate         = InternalError("token.generate")
	//ErrJsonDecode            = BadRequestError("json.decoder", "error decoding request body")
	//ErrJsonEncode            = InternalError("json.encode")
	ErrJsonDecode            = errors.New("json decode failed")
	ErrJsonEncode            = errors.New("json encode failed")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrArticleNotFound       = errors.New("article not found")
	// ErrHashPassword this will be mapped to InternalServerError anyway, so I might as well remove this.
	ErrHashPassword            = errors.New("hash password failed")
	ErrTokenGenerate           = errors.New("generate token failed")
	ErrDynamoQuery             = errors.New("dynamodb query failed")
	ErrDynamoMapping           = errors.New("dynamodb mapping failed")
	ErrDynamoMarshalling       = errors.New("dynamodb marshalling failed")
	ErrCantFollowYourself      = errors.New("cannot follow yourself")
	ErrCantDeleteOthersComment = errors.New("cannot delete other's comment")
	ErrCommentNotFound         = errors.New("comment not found")
	ErrAlreadyFavorited        = errors.New("already favorited")
	ErrAlreadyUnfavorited      = errors.New("already unfavorited")
)
