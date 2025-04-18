package server_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-obvious/server"
	"github.com/go-obvious/server/internal/middleware/apicaller"
	"github.com/go-obvious/server/internal/middleware/panic"
	"github.com/go-obvious/server/internal/middleware/requestid"
	"github.com/stretchr/testify/assert"
)

type mockAPI struct {
	name string
}

func (m *mockAPI) Name() string {
	return m.name
}

func (m *mockAPI) Register(app server.Server) error {
	r := app.Router().(*chi.Mux)
	r.Get("/mock", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return nil
}

func TestNew(t *testing.T) {
	version := &server.ServerVersion{
		Revision: "1.0.0",
	}

	middleware := []server.Middleware{
		apicaller.Middleware,
		panic.Middleware,
		requestid.Middleware,
	}

	apis := []server.API{
		&mockAPI{name: "mockAPI"},
	}

	srv := server.New(version, middleware, apis...)

	assert.NotNil(t, srv)
	assert.IsType(t, &server.ServerVersion{}, version)

	router := srv.Router().(*chi.Mux)
	assert.NotNil(t, router)

	// Test built-in routes
	req, _ := http.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	req, _ = http.NewRequest("GET", "/healthz", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Test custom API route
	req, _ = http.NewRequest("GET", "/mock", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRun(t *testing.T) {
	version := &server.ServerVersion{
		Revision: "1.0.0",
	}

	middleware := []server.Middleware{
		apicaller.Middleware,
		panic.Middleware,
		requestid.Middleware,
	}

	apis := []server.API{
		&mockAPI{name: "mockAPI"},
	}

	srv := server.New(version, middleware, apis...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		srv.Run(ctx)
	}()

	// Test if the server is running by making a request to a known route
	req, _ := http.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()
	srv.Router().(*chi.Mux).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewWithNilMiddleware(t *testing.T) {
	version := &server.ServerVersion{
		Revision: "1.0.0",
	}

	var middleware []server.Middleware

	apis := []server.API{
		&mockAPI{name: "mockAPI"},
	}

	srv := server.New(version, middleware, apis...)

	assert.NotNil(t, srv)
	assert.IsType(t, &server.ServerVersion{}, version)

	router := srv.Router().(*chi.Mux)
	assert.NotNil(t, router)

	// Test built-in routes
	req, _ := http.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	req, _ = http.NewRequest("GET", "/healthz", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Test custom API route
	req, _ = http.NewRequest("GET", "/mock", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewWithNilAPI(t *testing.T) {
	version := &server.ServerVersion{
		Revision: "1.0.0",
	}

	middleware := []server.Middleware{
		apicaller.Middleware,
		panic.Middleware,
		requestid.Middleware,
	}

	var apis []server.API

	srv := server.New(version, middleware, apis...)

	assert.NotNil(t, srv)
	assert.IsType(t, &server.ServerVersion{}, version)

	router := srv.Router().(*chi.Mux)
	assert.NotNil(t, router)

	// Test built-in routes
	req, _ := http.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	req, _ = http.NewRequest("GET", "/healthz", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRunWithTLS(t *testing.T) {
	// Generate self-signed TLS certificates for testing
	cert, key, err := generateSelfSignedCert("localhost")
	assert.NoError(t, err)

	// Write the certificates to temporary files
	certFile, err := os.CreateTemp("", "cert.pem")
	assert.NoError(t, err)
	defer os.Remove(certFile.Name())
	_, err = certFile.Write(cert)
	assert.NoError(t, err)
	certFile.Close()

	keyFile, err := os.CreateTemp("", "key.pem")
	assert.NoError(t, err)
	defer os.Remove(keyFile.Name())
	_, err = keyFile.Write(key)
	assert.NoError(t, err)
	keyFile.Close()

	// this ensures the server start reads these files
	os.Setenv("SERVER_CERTIFICATE_CERT_FILE", certFile.Name())
	os.Setenv("SERVER_CERTIFICATE_KEY_FILE", keyFile.Name())
	os.Setenv("SERVER_MODE", "https")
	os.Setenv("SERVER_DOMAIN", "localhost")
	os.Setenv("SERVER_PORT", "443")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	version := &server.ServerVersion{
		Revision: "1.0.0",
	}

	middleware := []server.Middleware{
		apicaller.Middleware,
		panic.Middleware,
		requestid.Middleware,
	}

	apis := []server.API{
		&mockAPI{name: "mockAPI"},
	}

	srv := server.New(version, middleware, apis...)
	go srv.Run(ctx)

	// Sleep to allow svr.Run to initialize anything it needs.
	time.Sleep(100 * time.Millisecond)

	// Test if the server is running with TLS by making a request to a known route
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, _ := http.NewRequest("GET", "https://localhost:443/about", nil)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func generateSelfSignedCert(host string) ([]byte, []byte, error) {
	// Generate a private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add the host as a DNS name
	template.DNSNames = []string{host}

	// Generate the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	// Encode the certificate and private key to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certPEM, keyPEM, nil
}
