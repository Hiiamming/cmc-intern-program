package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"mini-asm/internal/model"
)

type IPScanner struct {
	svc *ScanService
}

func NewIPScanner(svc *ScanService) *IPScanner {
	return &IPScanner{svc: svc}
}

type ipAPIResponse struct {
	Status     string  `json:"status"`
	Country    string  `json:"country"`
	CountryCode string `json:"countryCode"`
	RegionName string  `json:"regionName"`
	City       string  `json:"city"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	ISP        string  `json:"isp"`
	Org        string  `json:"org"`
	AS         string  `json:"as"`
	ASName     string  `json:"asname"`
	Reverse    string  `json:"reverse"`
	Message    string  `json:"message"`
}

func (s *IPScanner) Scan(ctx context.Context, asset *model.Asset, job *model.ScanJob) (int, error) {
	ipStr := strings.TrimSpace(asset.Name)
	if net.ParseIP(ipStr) == nil {
		return 0, fmt.Errorf("invalid IP address")
	}

	reverseDNS := ""
	names, err := net.LookupAddr(ipStr)
	if err == nil && len(names) > 0 {
		reverseDNS = strings.TrimSuffix(names[0], ".")
	}

	// private/localhost: không gọi API ngoài
	if isPrivateOrLocalIP(ipStr) {
		payload := map[string]interface{}{
			"ip_address": ipStr,
			"geolocation": map[string]interface{}{
				"country":      "Local / Private",
				"country_code": "LOCAL",
				"city":         "",
				"region":       "",
				"latitude":     0,
				"longitude":    0,
				"isp":          "Local Network",
				"org":          "Local Network",
			},
			"asn": map[string]interface{}{
				"number":      0,
				"name":        "PRIVATE",
				"description": "Private or loopback address",
			},
			"reverse_dns": reverseDNS,
			"created_at":  time.Now().UTC(),
		}

		if err := s.svc.saveResult(job, "ip", payload); err != nil {
			return 0, err
		}
		return 1, nil
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,countryCode,regionName,city,lat,lon,isp,org,as,asname,reverse", ipStr),
		nil,
	)
	if err != nil {
		return 0, err
	}

	resp, err := s.svc.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("ip lookup failed: %w", err)
	}
	defer resp.Body.Close()

	var data ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("decode ip lookup failed: %w", err)
	}

	if data.Status != "success" {
		return 0, fmt.Errorf("ip lookup failed: %s", data.Message)
	}

	asnNumber := 0
	asnName := ""
	if strings.HasPrefix(data.AS, "AS") {
		var num int
		fmt.Sscanf(data.AS, "AS%d", &num)
		asnNumber = num
	}
	asnName = data.ASName
	if asnName == "" {
		asnName = data.AS
	}

	payload := map[string]interface{}{
		"ip_address": ipStr,
		"geolocation": map[string]interface{}{
			"country":      data.Country,
			"country_code": data.CountryCode,
			"city":         data.City,
			"region":       data.RegionName,
			"latitude":     data.Lat,
			"longitude":    data.Lon,
			"isp":          data.ISP,
			"org":          data.Org,
		},
		"asn": map[string]interface{}{
			"number":      asnNumber,
			"name":        asnName,
			"description": data.AS,
		},
		"reverse_dns": reverseDNS,
		"created_at":  time.Now().UTC(),
	}

	if err := s.svc.saveResult(job, "ip", payload); err != nil {
		return 0, err
	}

	return 1, nil
}