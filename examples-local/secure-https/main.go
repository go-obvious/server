package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"github.com/go-obvious/server"
)

// SecureAPI demonstrates HTTPS server with TLS security features
type SecureAPI struct{}

func (api *SecureAPI) Name() string {
	return "secure-api"
}

func (api *SecureAPI) Register(app server.Server) error {
	router := app.Router().(*chi.Mux)
	
	router.Route("/api", func(r chi.Router) {
		r.Get("/secure", api.handleSecure)
		r.Get("/tls-info", api.handleTLSInfo)
		r.Post("/protected", api.handleProtected)
	})
	
	return nil
}

func (api *SecureAPI) handleSecure(w http.ResponseWriter, r *http.Request) {
	zlog.Info().Msg("Secure endpoint called")
	
	response := map[string]interface{}{
		"message": "This is a secure HTTPS endpoint",
		"tls_enabled": r.TLS != nil,
		"security_features": []string{
			"TLS 1.2+ enforcement",
			"Secure cipher suites",
			"HSTS headers",
			"Security headers",
			"CORS protection",
		},
	}
	
	render.JSON(w, r, response)
}

func (api *SecureAPI) handleTLSInfo(w http.ResponseWriter, r *http.Request) {
	zlog.Info().Msg("TLS info endpoint called")
	
	if r.TLS == nil {
		http.Error(w, "TLS not enabled", http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"tls_version": getTLSVersionString(r.TLS.Version),
		"cipher_suite": tls.CipherSuiteName(r.TLS.CipherSuite),
		"server_name": r.TLS.ServerName,
		"negotiated_protocol": r.TLS.NegotiatedProtocol,
		"peer_certificates": len(r.TLS.PeerCertificates),
	}
	
	render.JSON(w, r, response)
}

func (api *SecureAPI) handleProtected(w http.ResponseWriter, r *http.Request) {
	zlog.Info().Msg("Protected endpoint called")
	
	var body map[string]interface{}
	if err := render.DecodeJSON(r.Body, &body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"message": "Data received securely over HTTPS",
		"data": body,
		"security_headers": map[string]string{
			"hsts": w.Header().Get("Strict-Transport-Security"),
			"csp": w.Header().Get("Content-Security-Policy"),
			"frame_options": w.Header().Get("X-Frame-Options"),
		},
	}
	
	render.JSON(w, r, response)
}

func getTLSVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

func generateSelfSignedCert() error {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"go-obvious/server"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"Example"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{"localhost"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// Save certificate
	certOut, err := os.Create("server.crt")
	if err != nil {
		return err
	}
	defer certOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Save private key
	keyOut, err := os.Create("server.key")
	if err != nil {
		return err
	}
	defer keyOut.Close()

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return nil
}

func main() {
	// Set up structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Info().Msg("Starting secure HTTPS server example")

	// Generate self-signed certificate if it doesn't exist
	if _, err := os.Stat("server.crt"); os.IsNotExist(err) {
		zlog.Info().Msg("Generating self-signed certificate...")
		if err := generateSelfSignedCert(); err != nil {
			log.Fatalf("Failed to generate certificate: %v", err)
		}
		zlog.Info().Msg("Certificate generated: server.crt, server.key")
	}

	// Set environment variables for HTTPS mode
	os.Setenv("SERVER_MODE", "https")
	os.Setenv("SERVER_PORT", "8443")
	os.Setenv("SERVER_CERTIFICATE_CERT_FILE", "server.crt")
	os.Setenv("SERVER_CERTIFICATE_KEY_FILE", "server.key")
	os.Setenv("SERVER_TLS_MIN_VERSION", "1.2")
	os.Setenv("SERVER_SECURITY_HEADERS_ENABLED", "true")
	os.Setenv("SERVER_HSTS_MAX_AGE", "31536000")

	// Create server with version info
	version := &server.ServerVersion{
		Revision: "v1.0.0",
		Tag:      "secure-https-example",
		Time:     "2024-01-01T00:00:00Z",
	}

	// Create and configure API
	api := &SecureAPI{}

	// Start server with API
	srv := server.New(version).WithAPIs(api)

	zlog.Info().Msg("HTTPS server starting on https://localhost:8443")
	zlog.Info().Msg("Try: curl -k https://localhost:8443/api/secure")
	zlog.Info().Msg("Note: Using -k flag because of self-signed certificate")

	srv.Run(context.Background())
}