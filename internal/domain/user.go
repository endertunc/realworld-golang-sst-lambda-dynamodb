package domain

import (
	"time"
)

type User struct {
	Email          string
	HashedPassword string
	Username       string
	Bio            *string
	Image          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

//var (
//	ErrUserNotFound    = fmt.Errorf("user not found")
//	ErrInvalidPassword = fmt.Errorf("invalid password")
//	ErrQueryFailed     = fmt.Errorf("query failed")
//	ErrInternal        = fmt.Errorf("internal error")
//	ErrRepository      = fmt.Errorf("repository error")
//)

//type UserNotFoundError struct {
//	Email string
//}
//
//func (err *UserNotFoundError) Error() string {
//	return fmt.Sprintf("user with email [%s] is not found", err.Email)
//}

//type InvalidPasswordError struct {
//	Inner error
//}
//
//func (err *InvalidPasswordError) Error() string {
//	return fmt.Sprintf("invalid password: %v", err.Inner)
//}

//type UnexpectedError struct {
//	Inner   error
//	Message string
//}
//
//func (err *UnexpectedError) Error() string {
//	return fmt.Sprintf("unexpected error: %v, message %s", err.Inner, err.Message)
//}
//
//func (err *UnexpectedError) Unwrap() error {
//	return err.Inner
//}
