package user

import (
	"errors"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

var (
	ErrUserNotFound    = errutil.Conflict("user.not-found", errutil.WithPublicMessage("Invalid credentials"))
	ErrInvalidPassword = errutil.Conflict("user.password-invalid", errutil.WithPublicMessage("Invalid credentials"))
)
