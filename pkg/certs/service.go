package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func CreateCertificates() error {
	cert, priv, err := CreateCertificate("ca", nil, nil)
	if err != nil {
		return fmt.Errorf("Failed to create CA self-signed certificate: %v", err)
	}
	_, _, err = CreateCertificate("server", cert, priv)
	if err != nil {
		return fmt.Errorf("Failed to create server certificate: %v", err)
	}
	_, _, err = CreateCertificate("client", cert, priv)
	if err != nil {
		return fmt.Errorf("Failed to create client certificate: %v", err)
	}
	return nil
}

func CreateCertificate(name string, parent []byte, parent_pk any) ([]byte, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate RSA key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate serial number: %v", err)
	}

	var derBytes []byte

	if len(parent) == 0 {
		// CA self-signed certificate
		ca := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization:       []string{"Company, INC."},
				OrganizationalUnit: []string{"tonx22"},
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().AddDate(10, 0, 0),
			IsCA:                  true,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
		}

		derBytes, err = x509.CreateCertificate(rand.Reader, &ca, &ca, &privateKey.PublicKey, privateKey)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to create certificate: %v", err)
		}

	} else {
		// signed by other certificate
		template := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{name},
			},
			DNSNames:              []string{"localhost"},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().AddDate(10, 0, 0),
			KeyUsage:              x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		parent_cert, err := x509.ParseCertificate(parent)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to parse parent certificate: %v", err)
		}

		derBytes, err = x509.CreateCertificate(rand.Reader, &template, parent_cert, &privateKey.PublicKey, parent_pk)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to create certificate: %v", err)
		}
	}

	// dump certificate to file
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		return nil, nil, fmt.Errorf("Failed to encode certificate to PEM")
	}
	if err := os.WriteFile(fmt.Sprintf("%s-cert.pem", name), pemCert, 0644); err != nil {
		return nil, nil, fmt.Errorf("Failed to write certificate file: %v", err)
	}

	// dump private key to file
	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	pemKey := pem.EncodeToMemory(privateKeyBlock)
	if pemKey == nil {
		return nil, nil, fmt.Errorf("Failed to encode key to PEM")
	}
	if err := os.WriteFile(fmt.Sprintf("%s-key.pem", name), pemKey, 0600); err != nil {
		return nil, nil, fmt.Errorf("Failed to write private key file: %v", err)
	}
	return derBytes, privateKey, nil
}
