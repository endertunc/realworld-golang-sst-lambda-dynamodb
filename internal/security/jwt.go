package security

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/user"
	"time"
)

var signingKey = []byte("CHANGE_ME")

func GenerateToken(email string) (*user.Token, error) {
	// ToDo read some of the values from config
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Unix(1516239022, 0)),
		Issuer:    "realworld",
		Subject:   email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(signingKey)
	if err != nil {
		return nil, errutil.InternalError(fmt.Errorf("GenerateToken - token signing error: %w", err))
	}
	t := user.Token(signedString)
	return &t, nil
}

func validateToken() {

}
