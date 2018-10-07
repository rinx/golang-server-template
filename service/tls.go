package service

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	"github.com/kpango/golang-server-template/config"
	"github.com/pkg/errors"
)

var (
	// ErrTLSCertOrKeyNotFound is error variable, it's replesents tls cert or key not found error
	ErrTLSCertOrKeyNotFound = errors.New("Cert/Key path not found")
)

// NewTLSConfig returns a *tls.Config struct or error
// This function read TLS configuration and initialize *tls.Config struct.
// This function initialize TLS configuration, for example the CA certificate and key to start TLS server.
// Server and CA Certificate, and private key will read from a file from the file path definied in environment variable.
func NewTLSConfig(cfg config.TLS) (*tls.Config, error) {
	t := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP521,
			tls.CurveP384,
			tls.CurveP256,
			tls.X25519,
		},
		SessionTicketsDisabled: true,
		// PreferServerCipherSuites: true,
		// CipherSuites: []uint16{
		// tls.TLS_RSA_WITH_RC4_128_SHA,
		// tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		// tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		// tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		// tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		// tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		// tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		// tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, // Maybe this is work on TLS 1.2
		// tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA, // TLS1.3 Feature
		// tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA, // TLS1.3 Feature
		// tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
		// tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
		// },
		ClientAuth: tls.NoClientCert,
	}

	cert := os.Getenv(cfg.CertKey)
	key := os.Getenv(cfg.KeyKey)
	ca := os.Getenv(cfg.CAKey)

	if cert == "" || key == "" {
		return nil, ErrTLSCertOrKeyNotFound
	}

	crt, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	t.Certificates = make([]tls.Certificate, 1)
	t.Certificates[0] = crt

	if ca != "" {
		pool, err := NewX509CertPool(ca)
		if err != nil {
			return nil, err
		}
		t.ClientCAs = pool
		t.ClientAuth = tls.RequireAndVerifyClientCert
	}

	t.BuildNameToCertificate()
	return t, nil
}

// NewX509CertPool returns *x509.CertPool struct or error.
// The CertPool will read the certificate from the path, and append the content to the system certificate pool, and return.
func NewX509CertPool(path string) (*x509.CertPool, error) {
	var pool *x509.CertPool
	c, err := ioutil.ReadFile(path)
	if err == nil && c != nil {
		pool, err = x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(c) {
			err = errors.New("Certification Failed")
		}
	}
	return pool, err
}
