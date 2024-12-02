package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

type statisKeyProvider struct {
}

func (s statisKeyProvider) GetKeys() KeyPair {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate key pair: %v", err)
	}

	return KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

// make sure the key provider is set
var _ = func() KeyProvider {
	provider := statisKeyProvider{}
	SetKeyProvider(provider)
	return provider
}()

func TestGenerateAndValidateToken(t *testing.T) {

	userId := uuid.New()
	token, err := GenerateToken(userId)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	// Validate the token
	extractedUserId, err := ValidateToken(string(*token))
	assert.NoError(t, err)
	assert.Equal(t, userId, extractedUserId)
}

func TestValidateToken_InvalidSigningMethod(t *testing.T) {
	// Create a token with a different signing method (HS256 instead of EdDSA)
	userId := uuid.New()
	claims := &jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{audience},
		Subject:   userId.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	// Try to validate the token
	_, err = ValidateToken(tokenString)
	assert.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create an expired token
	userId := uuid.New()
	claims := &jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{audience},
		Subject:   userId.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour * 2)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(keys().PrivateKey)
	assert.NoError(t, err)

	// Try to validate the expired token
	_, err = ValidateToken(tokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenInvalid)
}

func TestValidateToken_InvalidSubject(t *testing.T) {
	// Create a token with invalid UUID as subject
	claims := &jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{audience},
		Subject:   "not-a-uuid",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(keys().PrivateKey)
	assert.NoError(t, err)

	// Try to validate the token
	_, err = ValidateToken(tokenString)
	assert.ErrorIs(t, err, ErrTokenSubjectInvalid)
}

func TestValidateToken_MissingSubject(t *testing.T) {
	// Create a token without a subject
	claims := &jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(keys().PrivateKey)
	assert.NoError(t, err)

	// Try to validate the token
	_, err = ValidateToken(tokenString)
	assert.ErrorIs(t, err, ErrTokenSubjectInvalid)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	// Try to validate an invalid token string
	_, err := ValidateToken("invalid-token")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenInvalid)
}

func TestValidateToken_WrongIssuer(t *testing.T) {
	// Create a token with wrong issuer
	userId := uuid.New()
	claims := &jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    "wrong-issuer",
		Audience:  jwt.ClaimStrings{audience},
		Subject:   userId.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(keys().PrivateKey)
	assert.NoError(t, err)

	// Try to validate the token
	_, err = ValidateToken(tokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenInvalid)
}

func TestValidateToken_WrongAudience(t *testing.T) {
	// Create a token with wrong audience
	userId := uuid.New()
	claims := &jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{"wrong-audience"},
		Subject:   userId.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(keys().PrivateKey)
	assert.NoError(t, err)

	// Try to validate the token
	_, err = ValidateToken(tokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenInvalid)
}
