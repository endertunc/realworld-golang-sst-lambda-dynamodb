package errutil

import (
	"errors"
	"fmt"
	"net/http"
)

// Base represents the static information about a specific error.
// Always use [NewBase] to create new instances of Base.
type Base struct {
	// Because Base is typically instantiated as a package or global
	// variable, having private members reduces the probability of a
	// bug messing with the error base.
	status        int
	messageID     string
	publicMessage string
}

// NewBase initializes a [Base] that is used to construct [Error].
// The reason is used to determine the status code that should be
// returned for the error, and the msgID is passed to the caller
// to serve as the base for user facing error messages.
//
// msgID should be structured as component.errorBrief, for example
//
//	login.failedAuthentication
//	dashboards.validationError
//	dashboards.uidAlreadyExists
func NewBase(status int, msgID string, opts ...BaseOpt) Base {
	b := Base{
		status:    status,
		messageID: msgID,
	}

	for _, opt := range opts {
		b = opt(b)
	}

	return b
}

// NotFound initializes a new [Base] error with reason StatusNotFound
// that is used to construct [Error]. The msgID is passed to the caller
// to serve as the base for user facing error messages.
//
// msgID should be structured as component.errorBrief, for example
//
//	folder.notFound
//	plugin.notRegistered
func NotFound(msgID string, opts ...BaseOpt) Base {
	return NewBase(http.StatusNotFound, msgID, opts...)
}

// UnprocessableEntity initializes a new [Base] error with reason StatusUnprocessableEntity
// that is used to construct [Error]. The msgID is passed to the caller
// to serve as the base for user facing error messages.
//
// msgID should be structured as component.errorBrief, for example
//
//	plugin.checksumMismatch
func UnprocessableEntity(msgID string, opts ...BaseOpt) Base {
	return NewBase(http.StatusUnprocessableEntity, msgID, opts...)
}

// Conflict initializes a new [Base] error with reason StatusConflict
// that is used to construct [Error]. The msgID is passed to the caller
// to serve as the base for user facing error messages.
//
// msgID should be structured as component.errorBrief, for example
//
//	folder.alreadyExists
func Conflict(msgID string, opts ...BaseOpt) Base {
	return NewBase(http.StatusConflict, msgID, opts...)
}

// BadRequest initializes a new [Base] error with reason StatusBadRequest
// that is used to construct [Error]. The msgID is passed to the caller
// to serve as the base for user facing error messages.
//
// msgID should be structured as component.errorBrief, for example
//
//	query.invalidDatasourceId
//	sse.dataQueryError
func BadRequest(msgID string, opts ...BaseOpt) Base {
	return NewBase(http.StatusBadRequest, msgID, opts...)
}

// Internal initializes a new [Base] error with reason StatusInternal
// that is used to construct [Error]. The msgID is passed to the caller
// to serve as the base for user facing error messages.
//
// msgID should be structured as component.errorBrief, for example
//
//	sqleng.connectionError
//	plugin.downstreamError
func Internal(msgID string, opts ...BaseOpt) Base {
	return NewBase(http.StatusInternalServerError, msgID, opts...)
}

// Unauthorized initializes a new [Base] error with reason StatusUnauthorized
// that is used to construct [Error]. The msgID is passed to the caller
// to serve as the base for user facing error messages.
//
// msgID should be structured as component.errorBrief, for example
//
//	auth.unauthorized
func Unauthorized(msgID string, opts ...BaseOpt) Base {
	return NewBase(http.StatusUnauthorized, msgID, opts...)
}

type BaseOpt func(Base) Base

// WithPublicMessage sets the default public message that will be used
// for errors based on this [Base].
//
// Used as a functional option to [NewBase].
func WithPublicMessage(message string) BaseOpt {
	return func(b Base) Base {
		b.publicMessage = message
		return b
	}
}

// Errorf creates a new [Error] with Reason and MessageID from [Base],
// and Message and Underlying will be populated using the rules of
// [fmt.Errorf].
func (b Base) Errorf(format string, args ...any) Error {
	err := fmt.Errorf(format, args...)

	return Error{
		Status:        b.status,
		LogMessage:    err.Error(),
		PublicMessage: b.publicMessage,
		MessageID:     b.messageID,
		Underlying:    errors.Unwrap(err),
	}
}

// Error makes Base implement the error type. Relying on this is
// discouraged, as the Error type can carry additional information
// that's valuable when debugging.
func (b Base) Error() string {
	return b.Errorf("").Error()
}

// Is validates that an [Error] has the same reason and messageID as the
// Base.
//
// Implements the interface used by [errors.Is].
func (b Base) Is(err error) bool {
	// The linter complains that it wants to use errors.As because it
	// handles unwrapping, we don't want to do that here since we want
	// to validate the equality between the two objects.
	// errors.Is handles the unwrapping, should you want it.
	//nolint:errorlint
	base, isBase := err.(Base)
	//nolint:errorlint
	gfErr, isOurError := err.(Error)

	switch {
	case isOurError:
		return b.status == gfErr.Status && b.messageID == gfErr.MessageID
	case isBase:
		return b.status == base.status && b.messageID == base.messageID
	default:
		return false
	}
}

// Error is the error type for errors within Grafana, extending
// the Go error type with Grafana specific metadata to reduce
// boilerplate error handling for status codes and internationalization
// support.
//
// Use [Base.Errorf] or [Template.Build] to construct errors:
//
//	// package-level
//	var errMonthlyQuota = NewBase(errutil.StatusTooManyRequests, "service.monthlyQuotaReached")
//	// in function
//	err := errMonthlyQuota.Errorf("user '%s' reached their monthly quota for service", userUID)
//
// or
//
//	// package-level
//	var errRateLimited = NewBase(errutil.StatusTooManyRequests, "service.backoff").MustTemplate(
//		"quota reached for user {{ .Private.user }}, rate limited until {{ .Public.time }}",
//		errutil.WithPublic("Too many requests, try again after {{ .Public.time }}"),
//	)
//	// in function
//	err := errRateLimited.Build(TemplateData{
//		Private: map[string]interface{ "user": userUID },
//		Public: map[string]interface{ "time": rateLimitUntil },
//	})
//
// Error implements Unwrap and Is to natively support Go 1.13 style
// errors as described in https://go.dev/blog/go1.13-errors .
type Error struct {
	Status        int
	MessageID     string
	LogMessage    string
	Underlying    error
	PublicMessage string
	PublicPayload map[string]any
}

// Error implements the error interface.
func (e Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.MessageID, e.LogMessage)
}

// Unwrap is used by errors.As to iterate over the sequence of
// underlying errors until a matching type is found.
func (e Error) Unwrap() error {
	return e.Underlying
}

// Is checks whether an error is derived from the error passed as an
// argument.
//
// Implements the interface used by [errors.Is].
func (e Error) Is(other error) bool {
	// The linter complains that it wants to use errors.As because it
	// handles unwrapping, we don't want to do that here since we want
	// to validate the equality between the two objects.
	// errors.Is handles the unwrapping, should you want it.
	//nolint:errorlint
	o, isOurError := other.(Error)
	//nolint:errorlint
	base, isBase := other.(Base)

	switch {
	case isOurError:
		return o.Status == e.Status && o.MessageID == e.MessageID && o.Error() == e.Error()
	case isBase:
		return base.Is(e)
	default:
		return false
	}
}

//// PublicError is derived from Error and only contains information
//// available to the end user.
//type PublicError struct {
//	StatusCode int            `json:"statusCode"`
//	MessageID  string         `json:"messageId"`
//	Message    string         `json:"message,omitempty"`
//	Extra      map[string]any `json:"extra,omitempty"`
//}
//
//// Public returns a subset of the error with non-sensitive information
//// that may be relayed to the caller.
//func (e Error) Public() PublicError {
//	message := e.PublicMessage
//	if message == "" {
//		if e.Reason == StatusUnknown {
//			// The unknown status is equal to the empty string.
//			message = string(StatusInternal)
//		} else {
//			message = string(e.Reason.Status())
//		}
//	}
//
//	return PublicError{
//		StatusCode: e.Reason.Status().HTTPStatus(),
//		MessageID:  e.MessageID,
//		Message:    message,
//		Extra:      e.PublicPayload,
//	}
//}
//
//// Error implements the error interface.
//func (p PublicError) Error() string {
//	return fmt.Sprintf("[%s] %s", p.MessageID, p.Message)
//}

type AppError struct {
	Code    int
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func BadRequestError(message string, cause error) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Cause:   cause,
	}
}

func NotFoundError(message string, cause error) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: message,
		Cause:   cause,
	}
}

func InternalError(cause error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "internal error",
		Cause:   cause,
	}
}

// AppError should implement the marshaler interface?
//func (e *AppError) MarshalJSON() ([]byte, error) {
//	return json.Marshal(struct {
//		Code    int    `json:"code"`
//		Message string `json:"message"`
//	}{

// WriteToResponse convert given AppError to http response
func WriteToResponse(err error, w http.ResponseWriter) {
	// ToDo log AppErr
	var appErr *AppError
	ok := errors.As(err, &appErr)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(appErr.Code)
	_, err = w.Write([]byte(appErr.Message))
	if err != nil {
		// todo log error
		return
	}
}
