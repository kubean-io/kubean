package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

const KeyBits = 3072

// GenerateRSAKeyPairB64 generates RSA public-private key pair and converts both private and public keys to Base64 format.
func GenerateRSAKeyPairB64() (string, string, error) {
	privateKey, publicKey, err := generateRSAKeyPair()
	if err != nil {
		return "", "", err
	}

	privateKeyB64 := privateKeyPEMToB64(privateKey)
	publicKeyB64, err := publicKeyPEMToB64(publicKey)
	if err != nil {
		return "", "", err
	}

	return privateKeyB64, publicKeyB64, nil
}

// generateRSAKeyPair generates RSA public-private key pair.
func generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, KeyBits)
	if err != nil {
		return nil, nil, err
	}

	publicKey := &privateKey.PublicKey

	return privateKey, publicKey, nil
}

// privateKeyToPEM converts private key to PEM format.
func privateKeyToPEM(key *rsa.PrivateKey) []byte {
	// Convert private key to PKCS#1 ASN.1 DER format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	return pem.EncodeToMemory(privateKeyPEM)
}

// publicKeyToPEM converts public key to PEM format.
func publicKeyToPEM(key *rsa.PublicKey) ([]byte, error) {
	// Convert public key to PKIX ASN.1 DER format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	return pem.EncodeToMemory(publicKeyPEM), nil
}

// privateKeyPEMToB64 converts private key PEM format to Base64.
func privateKeyPEMToB64(key *rsa.PrivateKey) string {
	pemBytes := privateKeyToPEM(key)
	return base64.StdEncoding.EncodeToString(pemBytes)
}

// publicKeyPEMToB64 converts public key PEM format to Base64.
func publicKeyPEMToB64(key *rsa.PublicKey) (string, error) {
	pemBytes, err := publicKeyToPEM(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(pemBytes), nil
}
