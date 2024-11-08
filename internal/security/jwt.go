package security

import (
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"time"
)

const (
	issuer   = "realworld"
	audience = "realworld"
)

var (
	ErrTokenInvalid              = errors.New("invalid JWT token")
	ErrTokenSigningMethodInvalid = errors.New("invalid signing method")
	ErrTokenInvalidSubjectType   = errors.New("invalid subject type")
	ErrTokenSubjectInvalid       = errors.New("invalid uuid subject")
)

// Load keys from files
var publicKey, privateKey = loadKeys()

func loadKeys() (ed25519.PublicKey, ed25519.PrivateKey) {
	keysDirectory := os.Getenv("JWT_KEYS_DIRECTORY")
	// Read private key
	privateKeyPath := filepath.Join(keysDirectory, "private.pem")
	privatePEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read private key: %v", err))
	}

	privateBlock, _ := pem.Decode(privatePEM)
	if privateBlock == nil {
		panic("failed to decode private key PEM")
	}

	if len(privateBlock.Bytes) != ed25519.PrivateKeySize {
		panic("invalid private key size")
	}

	privateKey := ed25519.PrivateKey(privateBlock.Bytes)

	// Read public key
	publicKeyPath := filepath.Join(keysDirectory, "public.pem")
	publicPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read public key: %v", err))
	}

	publicBlock, _ := pem.Decode(publicPEM)
	if publicBlock == nil {
		panic("failed to decode public key PEM")
	}

	if len(publicBlock.Bytes) != ed25519.PublicKeySize {
		panic("invalid public key size")
	}

	publicKey := ed25519.PublicKey(publicBlock.Bytes)

	return publicKey, privateKey
}

func GenerateToken(userId uuid.UUID) (*domain.Token, error) {
	// ToDo read some of the values from config
	nowInUTC := time.Now().UTC()

	claims := &jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{audience},
		Subject:   userId.String(),
		ExpiresAt: jwt.NewNumericDate(nowInUTC.Add(time.Minute * 60)),
		IssuedAt:  jwt.NewNumericDate(nowInUTC),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signedString, err := token.SignedString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrTokenGenerate, err)
	}
	t := domain.Token(signedString)
	return &t, nil
}

func ValidateToken(tokenString string) (uuid.UUID, error) {
	var parserOptions = []jwt.ParserOption{
		jwt.WithIssuedAt(),
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	}, parserOptions...)

	if err != nil {
		// no need to wrap ErrTokenSigningMethodInvalid which we throw ourselves
		if errors.Is(err, ErrTokenSigningMethodInvalid) {
			return uuid.Nil, err
		}
		return uuid.Nil, fmt.Errorf("%w: %w", ErrTokenInvalid, err)
	}

	// if the subject is not a string, jwt-go will return an invalid type error
	userIdAsString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %w", ErrTokenInvalidSubjectType, err)
	}

	userId, err := uuid.Parse(userIdAsString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %w", ErrTokenSubjectInvalid, err)
	}

	return userId, nil
}
