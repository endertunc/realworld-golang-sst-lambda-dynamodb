package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// you can use this script to generate a private/public key pair for JWT
//
//nolint:all
func main() {
	// Generate a new key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("Failed to generate key pair: %v\n", err)
		os.Exit(1)
	}

	// Create keys directory if it doesn't exist
	keysDir := "keys"
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		fmt.Printf("Failed to create keys directory: %v\n", err)
		os.Exit(1)
	}

	// Save private key
	privateKeyPath := filepath.Join(keysDir, "private.pem")
	privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKey,
	}
	privatePEM := pem.EncodeToMemory(privateKeyBlock)
	if err := os.WriteFile(privateKeyPath, privatePEM, 0600); err != nil {
		fmt.Printf("Failed to save private key: %v\n", err)
		os.Exit(1)
	}

	// Save public key
	publicKeyPath := filepath.Join(keysDir, "public.pem")
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	}
	publicPEM := pem.EncodeToMemory(publicKeyBlock)
	if err := os.WriteFile(publicKeyPath, publicPEM, 0644); err != nil {
		fmt.Printf("Failed to save public key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Keys successfully generated and saved in %s directory\n", keysDir)
	fmt.Printf("Private key: %s\n", privateKeyPath)
	fmt.Printf("Public key: %s\n", publicKeyPath)
}
