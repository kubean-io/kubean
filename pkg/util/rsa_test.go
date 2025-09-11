package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestGenerateRSAKeyPairB64(t *testing.T) {
	privateKeyB64, publicKeyB64, err := GenerateRSAKeyPairB64()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairB64() failed: %v", err)
	}

	if privateKeyB64 == "" {
		t.Error("Private key B64 should not be empty")
	}

	if publicKeyB64 == "" {
		t.Error("Public key B64 should not be empty")
	}

	// Verify the keys can be decoded back
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		t.Errorf("Failed to decode private key B64: %v", err)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		t.Errorf("Failed to decode public key B64: %v", err)
	}

	// Verify PEM structure
	privateBlock, _ := pem.Decode(privateKeyBytes)
	if privateBlock == nil || privateBlock.Type != "RSA PRIVATE KEY" {
		t.Error("Invalid private key PEM format")
	}

	publicBlock, _ := pem.Decode(publicKeyBytes)
	if publicBlock == nil || publicBlock.Type != "PUBLIC KEY" {
		t.Error("Invalid public key PEM format")
	}
}

func TestGenerateRSAKeyPair(t *testing.T) {
	privateKey, publicKey, err := generateRSAKeyPair()
	if err != nil {
		t.Fatalf("generateRSAKeyPair() failed: %v", err)
	}

	if privateKey == nil {
		t.Error("Private key should not be nil")
		return
	}

	if publicKey == nil {
		t.Error("Public key should not be nil")
		return
	}

	// Verify key size
	if privateKey.Size() != KeyBits/8 {
		t.Errorf("Expected key size %d bytes, got %d", KeyBits/8, privateKey.Size())
	}

	// Verify public key matches private key
	if !privateKey.PublicKey.Equal(publicKey) {
		t.Error("Public key should match private key's public key")
	}
}

func TestPrivateKeyToPEM(t *testing.T) {
	privateKey, _, err := generateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	pemBytes := privateKeyToPEM(privateKey)
	if len(pemBytes) == 0 {
		t.Error("PEM bytes should not be empty")
	}

	// Verify PEM format
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		t.Error("Failed to decode PEM block")
		return
	}

	if block.Type != "RSA PRIVATE KEY" {
		t.Errorf("Expected PEM type 'RSA PRIVATE KEY', got '%s'", block.Type)
	}

	// Verify we can parse it back
	parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Errorf("Failed to parse private key from PEM: %v", err)
	}

	if !privateKey.Equal(parsedKey) {
		t.Error("Parsed key should match original key")
	}
}

func TestPublicKeyToPEM(t *testing.T) {
	_, publicKey, err := generateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	pemBytes, err := publicKeyToPEM(publicKey)
	if err != nil {
		t.Fatalf("publicKeyToPEM() failed: %v", err)
	}

	if len(pemBytes) == 0 {
		t.Error("PEM bytes should not be empty")
	}

	// Verify PEM format
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		t.Error("Failed to decode PEM block")
		return
	}

	if block.Type != "PUBLIC KEY" {
		t.Errorf("Expected PEM type 'PUBLIC KEY', got '%s'", block.Type)
	}

	// Verify we can parse it back
	parsedKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Errorf("Failed to parse public key from PEM: %v", err)
	}

	parsedKey, ok := parsedKeyInterface.(*rsa.PublicKey)
	if !ok {
		t.Error("Parsed key should be RSA public key")
	}

	if !publicKey.Equal(parsedKey) {
		t.Error("Parsed key should match original key")
	}
}

func TestPrivateKeyPEMToB64(t *testing.T) {
	privateKey, _, err := generateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	b64Key := privateKeyPEMToB64(privateKey)
	if b64Key == "" {
		t.Error("B64 key should not be empty")
	}

	// Verify we can decode it back to PEM
	pemBytes, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		t.Errorf("Failed to decode B64 key: %v", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		t.Error("Invalid PEM format after B64 decode")
	}

	parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Errorf("Failed to parse private key from PEM: %v", err)
	}

	if !privateKey.Equal(parsedKey) {
		t.Error("Parsed key should match original key")
	}
}

func TestPublicKeyPEMToB64(t *testing.T) {
	_, publicKey, err := generateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	b64Key, err := publicKeyPEMToB64(publicKey)
	if err != nil {
		t.Fatalf("publicKeyPEMToB64() failed: %v", err)
	}

	if b64Key == "" {
		t.Error("B64 key should not be empty")
	}

	// Verify we can decode it back to PEM
	pemBytes, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		t.Errorf("Failed to decode B64 key: %v", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		t.Error("Invalid PEM format after B64 decode")
	}

	parsedKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Errorf("Failed to parse public key from PEM: %v", err)
	}

	parsedKey, ok := parsedKeyInterface.(*rsa.PublicKey)
	if !ok {
		t.Error("Parsed key should be RSA public key")
	}

	if !publicKey.Equal(parsedKey) {
		t.Error("Parsed key should match original key")
	}
}

func TestKeyBitsConstant(t *testing.T) {
	if KeyBits != 3072 {
		t.Errorf("Expected KeyBits to be 3072, got %d", KeyBits)
	}

	// Verify generated keys actually use this bit size
	privateKey, _, err := generateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	expectedSize := KeyBits / 8
	if privateKey.Size() != expectedSize {
		t.Errorf("Generated key size %d bytes doesn't match expected %d bytes for %d bits",
			privateKey.Size(), expectedSize, KeyBits)
	}
}

func BenchmarkGenerateRSAKeyPairB64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := GenerateRSAKeyPairB64()
		if err != nil {
			b.Fatalf("GenerateRSAKeyPairB64() failed: %v", err)
		}
	}
}
