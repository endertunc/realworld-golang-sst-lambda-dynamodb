package main

import (
	"errors"
	"fmt"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
)

func main() {
	rootErr := errors.New("fuck you asshole")

	err := security.ErrTokenSubjectInvalid.Errorf("%w", rootErr)
	app := errutil.AppError{}
	result1 := errors.As(err, &app)

	result2 := errors.Is(app, security.ErrTokenSubjectInvalid)

	fmt.Println(result1)
	fmt.Println(result2)

	fmt.Println(app.Error())
	fmt.Println(app.Message)
	fmt.Println(app.InternalMessage)
	fmt.Println(app.Cause.Error())

}
