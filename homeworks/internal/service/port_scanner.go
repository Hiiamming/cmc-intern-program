package service

import (
	"context"
	"fmt"
	"net"
	"time"

	"mini-asm/internal/model"
)

type PortScanner struct {
	svc *ScanService
}

func NewPortScanner(svc *ScanService) *PortScanner {
	return &PortScanner{svc: svc}
}

type openPortResult struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	State    string `json:"state"`
	Service  string `json:"service"`
	Version  string `json:"version"`
}

func guessService(port int) string {
	switch port {
	case 22:
		return "ssh"
	case 80:
		return "http"
	case 443:
		return "https"
	case 3306:
		return "mysql"
	case 5432:
		return "postgresql"
	case 6379:
		return "redis"
	case 8080:
		return "http-alt"
	default:
		return ""
	}
}

func (s *PortScanner) Scan(ctx context.Context, asset *model.Asset, job *model.ScanJob) (int, error) {
	ipStr := asset.Name
	if net.ParseIP(ipStr) == nil {
		return 0, fmt.Errorf("invalid IP address")
	}

	if !isPrivateOrLocalIP(ipStr) {
		return 0, model.ErrForbiddenScan
	}

	start := time.Now()
	ports := []int{22, 80, 443, 8080, 3306, 5432, 6379}
	openPorts := make([]openPortResult, 0)

	dialer := &net.Dialer{Timeout: 700 * time.Millisecond}

	for _, port := range ports {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", ipStr, port))
		if err == nil {
			_ = conn.Close()
			openPorts = append(openPorts, openPortResult{
				Port:     port,
				Protocol: "tcp",
				State:    "open",
				Service:  guessService(port),
				Version:  "",
			})
		}
	}

	payload := map[string]interface{}{
		"ip_address":       ipStr,
		"open_ports":       openPorts,
		"closed_ports":     len(ports) - len(openPorts),
		"total_scanned":    len(ports),
		"scan_duration_ms": time.Since(start).Milliseconds(),
		"created_at":       time.Now().UTC(),
	}

	if err := s.svc.saveResult(job, "port", payload); err != nil {
		return 0, err
	}

	return 1, nil
}