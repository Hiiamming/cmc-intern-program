package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"mini-asm/internal/model"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIPScanner_Scan_PrivateIP(t *testing.T) {
	svc, scanStore := newTestScanService()
	scanner := NewIPScanner(svc)
	asset, job := newTestJobAndAsset(model.ScanTypeIP, model.TypeIP, "127.0.0.1")

	n, err := scanWithContext(scanner, asset, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 result, got %d", n)
	}
	if len(scanStore.createdResults) != 1 {
		t.Fatalf("expected 1 stored result, got %d", len(scanStore.createdResults))
	}
	if scanStore.createdResults[0].ResultType != "ip" {
		t.Fatalf("expected result type ip, got %s", scanStore.createdResults[0].ResultType)
	}
}

func TestPortScanner_Scan_Localhost(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Skipf("cannot bind localhost:8080: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	svc, scanStore := newTestScanService()
	scanner := NewPortScanner(svc)
	asset, job := newTestJobAndAsset(model.ScanTypePort, model.TypeIP, "127.0.0.1")

	n, err := scanWithContext(scanner, asset, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 result, got %d", n)
	}
	if len(scanStore.createdResults) != 1 {
		t.Fatalf("expected 1 stored result, got %d", len(scanStore.createdResults))
	}

	var payload map[string]any
	if err := json.Unmarshal(scanStore.createdResults[0].Data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	openPorts, ok := payload["open_ports"].([]any)
	if !ok || len(openPorts) == 0 {
		t.Fatalf("expected at least one open port in payload, got %#v", payload["open_ports"])
	}
}

func TestSSLScanner_Scan_LocalTLSServer(t *testing.T) {
	certPEM, keyPEM, err := generateSelfSignedCert("127.0.0.1")
	if err != nil {
		t.Fatalf("generate cert: %v", err)
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("load key pair: %v", err)
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:443", &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		t.Skipf("cannot bind localhost:443: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(2 * time.Second))
				buf := make([]byte, 1)
				_, _ = c.Read(buf)
			}(conn)
		}
	}()

	svc, scanStore := newTestScanService()
	scanner := NewSSLScanner(svc)
	asset, job := newTestJobAndAsset(model.ScanTypeSSL, model.TypeDomain, "127.0.0.1")

	n, err := scanWithContext(scanner, asset, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 result, got %d", n)
	}
	if len(scanStore.createdResults) != 1 || scanStore.createdResults[0].ResultType != "ssl" {
		t.Fatalf("unexpected stored results: %+v", scanStore.createdResults)
	}
}

func TestTechScanner_Scan(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "nginx/1.25")
		w.Header().Set("X-Powered-By", "Express")
		_, _ = w.Write([]byte(`<html><head><meta name="generator" content="WordPress 6.0"></head><body>__NEXT_DATA__ react wp-content</body></html>`))
	}))
	defer ts.Close()

	svc, scanStore := newTestScanService()

	testTransport := ts.Client().Transport.(*http.Transport).Clone()
	testTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	targetAddr := strings.TrimPrefix(ts.URL, "https://")
	testTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, targetAddr)
	}
	svc.httpClient = &http.Client{Transport: testTransport}

	scanner := NewTechScanner(svc)
	asset, job := newTestJobAndAsset(model.ScanTypeTech, model.TypeDomain, "example.com")

	n, err := scanWithContext(scanner, asset, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 result, got %d", n)
	}
	if len(scanStore.createdResults) != 1 || scanStore.createdResults[0].ResultType != "tech" {
		t.Fatalf("unexpected stored results: %+v", scanStore.createdResults)
	}
}

func generateSelfSignedCert(host string) ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return nil, nil, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: host},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     []string{host},
		IPAddresses:  []net.IP{net.ParseIP(host)},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pemEncode("CERTIFICATE", der)
	keyPEM := pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(priv))
	return certPEM, keyPEM, nil
}

func pemEncode(typ string, der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der})
}