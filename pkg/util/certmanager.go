// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

// certManager helps to generate certs.
type certManager struct {
	Organizations []string      `json:"organizations"`
	EffectiveTime time.Duration `json:"effectiveTime"`
	DNSNames      []string      `json:"DNSNames"`
	CommonName    string        `json:"commonName"`
}

func NewCertManager(
	Orz []string,
	effectiveTime time.Duration,
	dnsNames []string,
	commonName string,
) *certManager {
	return &certManager{
		Organizations: Orz,
		EffectiveTime: effectiveTime,
		DNSNames:      dnsNames,
		CommonName:    commonName,
	}
}

// GenerateSelfSignedCerts return self-signed certs according to provided dns.
func (m *certManager) GenerateSelfSignedCerts() (*bytes.Buffer, *bytes.Buffer, error) {
	var serverCertPEM *bytes.Buffer
	var serverPrivateKeyPEM *bytes.Buffer
	var err error
	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization: m.Organizations,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(m.EffectiveTime),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	var caPrivateKey *rsa.PrivateKey
	caPrivateKey, err = rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	// self signed CA certificate
	var caBytes []byte
	caBytes, err = x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	// PEM encode CA cert
	caPEM := new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     m.DNSNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   m.CommonName,
			Organization: m.Organizations,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	var serverPrivateKey *rsa.PrivateKey
	serverPrivateKey, err = rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	// sign the server cert
	var serverCertBytes []byte
	serverCertBytes, err = x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	// PEM encode the server cert and key
	serverCertPEM = new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	serverPrivateKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(serverPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})

	return serverCertPEM, serverPrivateKeyPEM, err
}
