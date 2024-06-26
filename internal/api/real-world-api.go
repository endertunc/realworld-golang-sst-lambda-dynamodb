package api

import (
	"realworld-aws-lambda-dynamodb-golang/internal/article"
	"realworld-aws-lambda-dynamodb-golang/internal/comment"
	"realworld-aws-lambda-dynamodb-golang/internal/follower"
	"realworld-aws-lambda-dynamodb-golang/internal/user"
)

// RealWorldApi Make sure we conform to ServerInterface
type RealWorldApi struct {
	article.ArticlesApi
	user.UserApi
	comment.CommentsApi
	follower.FollowerApi
}

var _ ServerInterface = (*RealWorldApi)(nil)

//var (
//	ErrNotFound   = errors.New("not found")
//	ErrBadRequest = errors.New("bad request")
//	ErrInternal   = errors.New("internal error")
//)
//
//func ToGenericErrorResponse(err error) GenericErrorJSONResponse {
//	return GenericErrorJSONResponse{
//		Errors: struct {
//			Body []string `json:"body"`
//		}(struct{ Body []string }{
//			Body: []string{err.Error()},
//		}),
//	}
//}
//
//type errKind int
//
//const (
//	_ errKind = iota
//	InternalError
//	NotFound
//	InvalidPassword
//)
//
//type AppError struct {
//	Kind    errKind
//	Message string
//	err     error
//}
//
//func (e *AppError) Error() string {
//	switch e.Kind {
//	case InternalError:
//		return e.Message
//	case NotFound:
//		return e.Message
//	case InvalidPassword:
//		return e.Message
//	default:
//		return e.err.Error()
//	}
//}
//
//func (e *AppError) Unwrap() error {
//	return e.err
//}
//
//func (e *AppError) Is(target error) bool {
//	var t *AppError
//	ok := errors.As(target, &t)
//	if !ok {
//		return false
//	}
//	return e.Kind == t.Kind
//}
//
//var (
//	ErrUserNotFound    = &AppError{Kind: NotFound, Message: "user not found"}
//	ErrInvalidPassword = &AppError{Kind: InvalidPassword, Message: "invalid password"}
//	ErrInternalError   = &AppError{Kind: InternalError, Message: "internal error"}
//	ErrInvalidParam    = &AppError{Kind: InternalError, Message: "internal error"}
//)
//
//func HelloWorld(err error) {
//
//	if errors.As(err, &ErrUserNotFound) {
//		ErrUserNotFound.Message
//	}
//
//	if errors.Is(err, ErrUserNotFound) {
//		fmt.Println("User not found")
//	}
//}
