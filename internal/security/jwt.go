package security

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"sync"
	"sync/atomic"
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

// instead of creating an interface and implementing using different structs etc,
// I decided to experiment with the approach where you have an atomic.Value
// which you set to a desired implementation depending on the context
var keyProvider atomic.Value

type KeyPair struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

type KeyProvider interface {
	GetKeys() KeyPair
}

func SetKeyProvider(provider KeyProvider) {
	keyProvider.Store(provider)
}

type awsKeyProvider struct {
	SecretName string
}

func NewAwsKeyProvider(secretName string) KeyProvider {
	return awsKeyProvider{
		SecretName: secretName,
	}
}

func (a awsKeyProvider) GetKeys() KeyPair {
	private, public := loadKeysFromSecretStore(a.SecretName)
	return KeyPair{
		PrivateKey: private,
		PublicKey:  public,
	}
}

var keys = sync.OnceValue(func() KeyPair {
	if p := keyProvider.Load(); p != nil {
		return p.(KeyProvider).GetKeys()
	}
	log.Fatalf("no key provider set")
	return KeyPair{}
})

func loadKeysFromSecretStore(secretName string) (ed25519.PrivateKey, ed25519.PublicKey) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("error loading AWS configuration: %v", err)
	}

	svc := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(ctx, input)
	if err != nil {
		log.Fatalf("failed to get secret value: %v", err)
	}

	var secretData struct {
		PrivateKey string `json:"privateKey"`
		PublicKey  string `json:"publicKey"`
	}

	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		log.Fatalf("failed to unmarshal secret data: %v", err)
	}

	privateKey, _ := pem.Decode([]byte(secretData.PrivateKey))
	if privateKey == nil {
		log.Fatalf("failed to decode private key: %v", err)
	}

	publicKey, _ := pem.Decode([]byte(secretData.PublicKey))
	if publicKey == nil {
		log.Fatalf("failed to decode public key: %v", err)
	}

	return privateKey.Bytes, publicKey.Bytes
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
	signedString, err := token.SignedString(keys().PrivateKey)
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
		return keys().PublicKey, nil
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
