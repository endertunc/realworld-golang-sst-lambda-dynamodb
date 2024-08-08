package security

import (
	"crypto/ed25519"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"time"
)

// ToDo @ender read from config
// ToDo @ender do not ignore the error here
var privateKey, _, _ = ed25519.GenerateKey(nil)

var (
	ErrTokenInvalid              = errutil.UnauthorizedError("token.invalid", "invalid JWT token")
	ErrTokenSigningMethodInvalid = errutil.UnauthorizedError("token.signing_method_invalid", "invalid signing method")
	ErrTokenSubjectMissing       = errutil.UnauthorizedError("token.subject_missing", "missing subject")
	ErrTokenSubjectInvalid       = errutil.UnauthorizedError("token.subject_invalid", "invalid uuid subject")
)

func GenerateToken(userId uuid.UUID) (*domain.Token, error) {
	// ToDo read some of the values from config

	nowInUTC := time.Now().UTC()
	claims := &jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    "realworld",
		Audience:  jwt.ClaimStrings{"realworld"},
		Subject:   userId.String(),
		ExpiresAt: jwt.NewNumericDate(nowInUTC.Add(time.Minute * 60)),
		IssuedAt:  jwt.NewNumericDate(nowInUTC),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signedString, err := token.SignedString(privateKey)
	if err != nil {
		return nil, errutil.ErrTokenGenerate.Errorf("GenerateToken - token signing error: %w", err)
	}
	t := domain.Token(signedString)
	return &t, nil
}

func ValidateToken(tokenString string) (uuid.UUID, error) {
	var parserOptions = []jwt.ParserOption{
		jwt.WithIssuer("issuer"),
		jwt.WithAudience("audience"),
		jwt.WithSubject("subject"),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect: https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			// ToDo @ender token.Header["alg"] what if there is no alg?
			return uuid.Nil, ErrTokenSigningMethodInvalid.Errorf("ValidateToken - unexpected signing method: %v", token.Header["alg"])
		}
		return privateKey, nil
	}, parserOptions...)

	// ToDo @ender if err is ErrTokenSigningMethodInvalid do not wrap it

	if err != nil {
		return uuid.Nil, ErrTokenInvalid.Errorf("ValidateToken - token validation error: %v", err)
	}

	userIdAsString, err := token.Claims.GetSubject()

	if err != nil {
		return uuid.Nil, ErrTokenSubjectMissing.Errorf("ValidateToken - subject claim is missing: %v", err)
	}

	userId, err := uuid.Parse(userIdAsString)
	if err != nil {
		return uuid.Nil, ErrTokenSubjectInvalid.Errorf("ValidateToken - invalid subject claim: %v", err)
	}

	return userId, nil

}
