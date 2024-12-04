package test

import (
	"crypto/ed25519"
	"crypto/rand"
	"log"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
	"testing"
)

type MockKeyProvider struct {
	keyPair security.KeyPair
}

func (m *MockKeyProvider) GetKeys() security.KeyPair {
	return m.keyPair
}

// SetupMockKeyProvider creates a new mock key provider with a fresh key pair and sets it as the current provider
func SetupMockKeyProvider(t *testing.T) *MockKeyProvider {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate key pair: %v", err)
	}

	provider := &MockKeyProvider{
		keyPair: security.KeyPair{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		},
	}

	security.SetKeyProvider(provider)

	return provider
}
