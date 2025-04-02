package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"
)

var (
	certCache   = make(map[string]*tls.Certificate)
	certCacheMu sync.Mutex
)

// getOrGenerateCert - возвращает сертификат (tls.Certificate) для данного домена.
// Если уже генерировали ранее - берём из кэша. Иначе создаём.
func getOrGenerateCert(host string) (*tls.Certificate, error) {
	certCacheMu.Lock()
	c, ok := certCache[host]
	certCacheMu.Unlock()
	if ok {
		return c, nil
	}

	newCert, err := generateCertForHost(host)
	if err != nil {
		return nil, err
	}

	certCacheMu.Lock()
	certCache[host] = newCert
	certCacheMu.Unlock()

	return newCert, nil
}

func generateCertForHost(host string) (*tls.Certificate, error) {
	if caCert == nil || caKey == nil {
		return nil, fmt.Errorf("CA не загружен")
	}

	// Генерируем RSA-ключ для нового сертификата
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("ошибка GenerateKey: %w", err)
	}

	// Заполняем шаблон
	serialNumber, _ := rand.Int(rand.Reader, big.NewInt(1<<60))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"MyMITM Proxy"},
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour * 30), // 30 дней

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// добавить SubjectAltName = host
	// (DNSNames: []string{host}) - чтобы не было проблем с SNI
	// template.DNSNames = []string{host}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &priv.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка CreateCertificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("ошибка X509KeyPair: %w", err)
	}

	return &tlsCert, nil
}
