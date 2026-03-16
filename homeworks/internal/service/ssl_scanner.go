package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"
	"crypto/x509/pkix"

	"mini-asm/internal/model"
)

type SSLScanner struct {
	svc *ScanService
}

func NewSSLScanner(svc *ScanService) *SSLScanner {
	return &SSLScanner{svc: svc}
}

func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("unknown(%d)", v)
	}
}

func cipherSuiteName(id uint16) string {
	cs := tls.CipherSuiteName(id)
	if cs == "" {
		return fmt.Sprintf("unknown(%d)", id)
	}
	return cs
}

func certName(n pkix.Name) string {
	return n.String()
}

func isSelfSigned(cert *x509.Certificate) bool {
	return cert.Subject.String() == cert.Issuer.String()
}

func calcGrade(cert *x509.Certificate, tlsVersion string) (string, []string) {
	issues := []string{}
	now := time.Now()

	if now.After(cert.NotAfter) {
		issues = append(issues, "certificate expired")
	}
	if isSelfSigned(cert) {
		issues = append(issues, "self-signed certificate")
	}
	if tlsVersion == "TLS 1.0" || tlsVersion == "TLS 1.1" {
		issues = append(issues, "old TLS version")
	}

	switch {
	case len(issues) == 0:
		return "A", issues
	case len(issues) == 1 && issues[0] == "self-signed certificate":
		return "B", issues
	default:
		return "C", issues
	}
}

func (s *SSLScanner) Scan(ctx context.Context, asset *model.Asset, job *model.ScanJob) (int, error) {
	host := strings.TrimSpace(asset.Name)
	if host == "" {
		return 0, fmt.Errorf("empty host")
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, "443"), &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return 0, fmt.Errorf("tls dial failed: %w", err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return 0, fmt.Errorf("no peer certificate")
	}

	cert := state.PeerCertificates[0]
	now := time.Now()
	daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)

	grade, issues := calcGrade(cert, tlsVersionName(state.Version))

	payload := map[string]interface{}{
		"domain": host,
		"certificate": map[string]interface{}{
			"subject":           cert.Subject.String(),
			"issuer":            cert.Issuer.String(),
			"serial_number":     cert.SerialNumber.Text(16),
			"valid_from":        cert.NotBefore.UTC(),
			"valid_until":       cert.NotAfter.UTC(),
			"days_until_expiry": daysUntilExpiry,
			"is_expired":        now.After(cert.NotAfter),
			"is_self_signed":    isSelfSigned(cert),
			"san":               cert.DNSNames,
		},
		"connection": map[string]interface{}{
			"tls_version":  tlsVersionName(state.Version),
			"cipher_suite": cipherSuiteName(state.CipherSuite),
			"key_exchange": "",
		},
		"grade":      grade,
		"issues":     issues,
		"created_at": time.Now().UTC(),
	}

	if err := s.svc.saveResult(job, "ssl", payload); err != nil {
		return 0, err
	}

	return 1, nil
}