package proxy

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

var (
	caCert *x509.Certificate
	caKey  *rsa.PrivateKey
)

// LoadCA - загружает корневой сертификат (ca.crt) и приватный ключ (ca.key).
func LoadCA(caCertPath, caKeyPath string) error {

	certPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("ошибка чтения ca.crt: %w", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("не удалось pem.Decode ca.crt")
	}
	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("ошибка x509.ParseCertificate: %w", err)
	}

	keyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать ca.key: %w", err)
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("не удалось pem.Decode ca.key")
	}
	parsedKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("ошибка ParsePKCS1PrivateKey: %w", err)
	}

	caCert = parsedCert
	caKey = parsedKey

	return nil
}
