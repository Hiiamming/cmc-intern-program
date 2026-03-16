package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"mini-asm/internal/model"
)

type TechScanner struct {
	svc *ScanService
}

func NewTechScanner(svc *ScanService) *TechScanner {
	return &TechScanner{svc: svc}
}

type detectedTech struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Version    string `json:"version"`
	Confidence int    `json:"confidence"`
}

func (s *TechScanner) Scan(ctx context.Context, asset *model.Asset, job *model.ScanJob) (int, error) {
	url := "https://" + strings.TrimSpace(asset.Name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := s.svc.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	body := string(bodyBytes)

	headers := map[string]string{}
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[strings.ToLower(k)] = v[0]
		}
	}

	metaTags := map[string]string{}
	reGenerator := regexp.MustCompile(`(?i)<meta[^>]+name=["']generator["'][^>]+content=["']([^"']+)["']`)
	reViewport := regexp.MustCompile(`(?i)<meta[^>]+name=["']viewport["'][^>]+content=["']([^"']+)["']`)

	if m := reGenerator.FindStringSubmatch(body); len(m) > 1 {
		metaTags["generator"] = m[1]
	}
	if m := reViewport.FindStringSubmatch(body); len(m) > 1 {
		metaTags["viewport"] = m[1]
	}

	techs := make([]detectedTech, 0)

	server := strings.ToLower(headers["server"])
	xPoweredBy := strings.ToLower(headers["x-powered-by"])

	if strings.Contains(server, "nginx") {
		techs = append(techs, detectedTech{"nginx", "Web Server", headers["server"], 100})
	}
	if strings.Contains(server, "apache") {
		techs = append(techs, detectedTech{"Apache", "Web Server", headers["server"], 100})
	}
	if strings.Contains(server, "cloudflare") || headers["cf-ray"] != "" {
		techs = append(techs, detectedTech{"Cloudflare", "CDN", "", 100})
	}
	if strings.Contains(xPoweredBy, "express") {
		techs = append(techs, detectedTech{"Express", "Backend Framework", headers["x-powered-by"], 95})
	}
	if strings.Contains(xPoweredBy, "php") {
		techs = append(techs, detectedTech{"PHP", "Programming Language", headers["x-powered-by"], 95})
	}
	if strings.Contains(body, "__NEXT_DATA__") || strings.Contains(body, "/_next/") {
		techs = append(techs, detectedTech{"Next.js", "JavaScript Framework", "", 90})
	}
	if strings.Contains(strings.ToLower(body), "react") || strings.Contains(body, "data-reactroot") {
		techs = append(techs, detectedTech{"React", "JavaScript Framework", "", 80})
	}
	if strings.Contains(body, "wp-content") || strings.Contains(strings.ToLower(metaTags["generator"]), "wordpress") {
		techs = append(techs, detectedTech{"WordPress", "CMS", metaTags["generator"], 95})
	}

	payload := map[string]interface{}{
		"domain":       asset.Name,
		"technologies": techs,
		"headers":      headers,
		"meta_tags":    metaTags,
		"created_at":   time.Now().UTC(),
	}

	if err := s.svc.saveResult(job, "tech", payload); err != nil {
		return 0, err
	}

	return 1, nil
}