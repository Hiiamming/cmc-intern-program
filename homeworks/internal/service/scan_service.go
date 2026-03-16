package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
	"io"
	"errors"
	

	"mini-asm/internal/model"
	"mini-asm/internal/storage"

	"github.com/google/uuid"
)

type ScanService struct {
	assetStorage storage.Storage
	scanStorage  storage.ScanStorage
	httpClient   *http.Client
	scanners     map[model.ScanType]Scanner
}

// helper chung
func isPrivateOrLocalIP(ipStr string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}

	privateCIDRs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range privateCIDRs {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func NewScanService(assetStorage storage.Storage, scanStorage storage.ScanStorage) *ScanService {
	s := &ScanService{
		assetStorage: assetStorage,
		scanStorage:  scanStorage,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		scanners: make(map[model.ScanType]Scanner),
	}

	// existing scanners giữ nguyên logic cũ trong service
	s.scanners[model.ScanTypeIP] = NewIPScanner(s)
	s.scanners[model.ScanTypePort] = NewPortScanner(s)
	s.scanners[model.ScanTypeSSL] = NewSSLScanner(s)
	s.scanners[model.ScanTypeTech] = NewTechScanner(s)

	return s
}

func (s *ScanService) StartScan(assetID string, scanType model.ScanType) (*model.ScanJob, error) {
	asset, err := s.assetStorage.GetByID(assetID)
	if err != nil {
		return nil, err
	}

	if !model.IsValidScanType(scanType) {
		return nil, fmt.Errorf("invalid scan type: %s", scanType)
	}

	if !isScanAllowedForAssetType(scanType, asset.Type) {
		return nil, fmt.Errorf("scan type %s does not support asset type %s", scanType, asset.Type)
	}

	now := time.Now().UTC()
	job := &model.ScanJob{
		ID:        uuid.New().String(),
		AssetID:   assetID,
		ScanType:  scanType,
		Status:    model.ScanStatusPending,
		StartedAt: now,
		CreatedAt: now,
		Error:     "",
		Results:   0,
	}

	if err := s.scanStorage.CreateScanJob(job); err != nil {
		return nil, err
	}

	go s.performScan(asset, job)

	return job, nil
}


func (s *ScanService) GetScanJob(jobID string) (*model.ScanJob, error) {
	if strings.TrimSpace(jobID) == "" {
		return nil, model.ErrInvalidInput
	}
	return s.scanStorage.GetScanJob(jobID)
}

func (s *ScanService) GetScanResults(jobID string) (*model.ScanResultsResponse, error) {
	if strings.TrimSpace(jobID) == "" {
		return nil, model.ErrInvalidInput
	}

	job, err := s.scanStorage.GetScanJob(jobID)
	if err != nil {
		return nil, err
	}

	results, err := s.scanStorage.ListScanResultsByJob(jobID)
	if err != nil {
		return nil, err
	}

	return &model.ScanResultsResponse{
		Job:     job,
		Results: results,
	}, nil
}

func (s *ScanService) ListAssetScans(assetID string) ([]*model.ScanJob, error) {
	if strings.TrimSpace(assetID) == "" {
		return nil, model.ErrInvalidInput
	}

	if _, err := s.assetStorage.GetByID(assetID); err != nil {
		return nil, err
	}

	return s.scanStorage.ListScanJobsByAsset(assetID)
}

func (s *ScanService) ListAssetResults(assetID string) ([]*model.ScanResult, error) {
	if strings.TrimSpace(assetID) == "" {
		return nil, model.ErrInvalidInput
	}

	if _, err := s.assetStorage.GetByID(assetID); err != nil {
		return nil, err
	}

	return s.scanStorage.ListScanResultsByAsset(assetID)
}

func (s *ScanService) ListAssetDNSResults(assetID string) ([]*model.ScanResult, error) {
	if strings.TrimSpace(assetID) == "" {
		return nil, model.ErrInvalidInput
	}

	if _, err := s.assetStorage.GetByID(assetID); err != nil {
		return nil, err
	}

	return s.scanStorage.ListScanResultsByAssetAndTypes(assetID, []string{
		"dns_a",
		"dns_mx",
		"dns_ns",
	})
}

func (s *ScanService) ListAssetWhoisResults(assetID string) ([]*model.ScanResult, error) {
	if strings.TrimSpace(assetID) == "" {
		return nil, model.ErrInvalidInput
	}

	if _, err := s.assetStorage.GetByID(assetID); err != nil {
		return nil, err
	}

	return s.scanStorage.ListScanResultsByAssetAndTypes(assetID, []string{
		"whois",
	})
}

func (s *ScanService) ListAssetSubdomainResults(assetID string) ([]*model.ScanResult, error) {
	if strings.TrimSpace(assetID) == "" {
		return nil, model.ErrInvalidInput
	}

	if _, err := s.assetStorage.GetByID(assetID); err != nil {
		return nil, err
	}

	return s.scanStorage.ListScanResultsByAssetAndTypes(assetID, []string{
		"subdomain",
	})
}

func (s *ScanService) performScan(asset *model.Asset, job *model.ScanJob) {
	job.Status = model.ScanStatusRunning
	job.Error = ""
	_ = s.scanStorage.UpdateScanJob(job)

	log.Printf("[scan] start job=%s type=%s asset=%s", job.ID, job.ScanType, asset.Name)

	var totalResults int
	var errs []string

	run := func(fn func(*model.Asset, *model.ScanJob) (int, error)) {
		n, err := fn(asset, job)
		totalResults += n
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	switch job.ScanType {
	case model.ScanTypeDNS:
		run(s.runDNSScan)
	case model.ScanTypeWHOIS:
		run(s.runWHOISScan)
	case model.ScanTypeSubdomain:
		run(s.runSubdomainScan)
	case model.ScanTypeCertTrans:
		run(s.runCertTransScan)
	case model.ScanTypeASN:
		run(s.runASNScan)
	case model.ScanTypeIP, model.ScanTypePort, model.ScanTypeSSL, model.ScanTypeTech:
		scanner, ok := s.scanners[job.ScanType]
		if !ok {
			errs = append(errs, fmt.Sprintf("scanner not registered for %s", job.ScanType))
			break
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		n, err := scanner.Scan(ctx, asset, job)
		totalResults += n
		if err != nil {
			errs = append(errs, err.Error())
		}
	case model.ScanTypeAll:
		run(s.runDNSScan)
		run(s.runWHOISScan)
		run(s.runSubdomainScan)
		run(s.runCertTransScan)
	default:
		errs = append(errs, fmt.Sprintf("unsupported scan type: %s", job.ScanType))
	}

	now := time.Now().UTC()
	job.EndedAt = &now
	job.Results = totalResults

	switch {
	case len(errs) == 0:
		job.Status = model.ScanStatusCompleted
		job.Error = ""
	case totalResults > 0:
		job.Status = model.ScanStatusPartial
		job.Error = strings.Join(errs, "; ")
	default:
		job.Status = model.ScanStatusFailed
		job.Error = strings.Join(errs, "; ")
	}

	_ = s.scanStorage.UpdateScanJob(job)
	log.Printf("[scan] finish job=%s type=%s status=%s results=%d err=%s", job.ID, job.ScanType, job.Status, job.Results, job.Error)
}

func isScanAllowedForAssetType(scanType model.ScanType, assetType string) bool {
	switch scanType {
	case model.ScanTypeASN, model.ScanTypeIP, model.ScanTypePort:
		return assetType == model.TypeIP
	case model.ScanTypeDNS, model.ScanTypeWHOIS, model.ScanTypeSubdomain, model.ScanTypeCertTrans, model.ScanTypeSSL, model.ScanTypeTech, model.ScanTypeAll:
		return assetType == model.TypeDomain
	default:
		return false
	}
}

func (s *ScanService) saveResult(job *model.ScanJob, resultType string, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.scanStorage.CreateScanResult(&model.ScanResult{
		ID:         uuid.New().String(),
		ScanJobID:  job.ID,
		AssetID:    job.AssetID,
		ResultType: resultType,
		Data:       b,
		CreatedAt:  time.Now().UTC(),
	})
}

func (s *ScanService) runDNSScan(asset *model.Asset, job *model.ScanJob) (int, error) {
	var total int
	var errs []string

	ips, err := net.LookupIP(asset.Name)
	if err == nil {
		for _, ip := range ips {
			if saveErr := s.saveResult(job, "dns_a", map[string]interface{}{
				"name":  asset.Name,
				"value": ip.String(),
			}); saveErr != nil {
				errs = append(errs, saveErr.Error())
			} else {
				total++
			}
		}
	} else {
		errs = append(errs, "lookup A/AAAA failed: "+err.Error())
	}

	mxs, err := net.LookupMX(asset.Name)
	if err == nil {
		for _, mx := range mxs {
			if saveErr := s.saveResult(job, "dns_mx", map[string]interface{}{
				"name":     asset.Name,
				"host":     strings.TrimSuffix(mx.Host, "."),
				"pref":     mx.Pref,
				"priority": mx.Pref,
			}); saveErr != nil {
				errs = append(errs, saveErr.Error())
			} else {
				total++
			}
		}
	} else {
		errs = append(errs, "lookup MX failed: "+err.Error())
	}

	nss, err := net.LookupNS(asset.Name)
	if err == nil {
		for _, ns := range nss {
			if saveErr := s.saveResult(job, "dns_ns", map[string]interface{}{
				"name": asset.Name,
				"host": strings.TrimSuffix(ns.Host, "."),
			}); saveErr != nil {
				errs = append(errs, saveErr.Error())
			} else {
				total++
			}
		}
	} else {
		errs = append(errs, "lookup NS failed: "+err.Error())
	}

	if total == 0 && len(errs) > 0 {
		return 0, errors.New(strings.Join(errs, "; "))
	}
	if len(errs) > 0 {
		return total, errors.New(strings.Join(errs, "; "))
	}

	return total, nil
}

func (s *ScanService) runWHOISScan(asset *model.Asset, job *model.ScanJob) (int, error) {
	server := getWhoisServer(asset.Name)
	raw, err := queryWhois(server, asset.Name)
	if err != nil {
		return 0, err
	}

	payload := map[string]interface{}{
		"domain":    asset.Name,
		"server":    server,
		"registrar": extractWhoisField(raw, []string{"Registrar:", "registrar:"}),
		"raw_data":  raw,
	}

	if err := s.saveResult(job, "whois", payload); err != nil {
		return 0, err
	}

	return 1, nil
}

func (s *ScanService) runSubdomainScan(asset *model.Asset, job *model.ScanJob) (int, error) {
	wordlist := []string{
		"www", "api", "mail", "dev", "staging", "app", "admin", "blog", "test", "portal",
	}

	total := 0
	for _, sub := range wordlist {
		name := sub + "." + asset.Name
		ips, err := net.LookupIP(name)
		if err != nil || len(ips) == 0 {
			continue
		}

		values := make([]string, 0, len(ips))
		for _, ip := range ips {
			values = append(values, ip.String())
		}

		if err := s.saveResult(job, "subdomain", map[string]interface{}{
			"name": name,
			"ips":  values,
		}); err != nil {
			continue
		}
		total++
	}

	if total == 0 {
		return 0, fmt.Errorf("no subdomains found")
	}

	return total, nil
}

type crtShItem struct {
	NameValue string `json:"name_value"`
	Issuer    string `json:"issuer_name"`
}

func (s *ScanService) runCertTransScan(asset *model.Asset, job *model.ScanJob) (int, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", asset.Name)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("crt.sh returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var items []crtShItem
	if err := json.Unmarshal(body, &items); err != nil {
		return 0, err
	}

	seen := map[string]struct{}{}
	var names []string

	for _, item := range items {
		parts := strings.Split(item.NameValue, "\n")
		for _, part := range parts {
			name := strings.TrimSpace(strings.TrimPrefix(part, "*."))
			if name == "" {
				continue
			}
			if !strings.HasSuffix(name, asset.Name) {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}

	sort.Strings(names)

	total := 0
	for _, name := range names {
		if err := s.saveResult(job, "cert_trans", map[string]interface{}{
			"name":   name,
			"source": "crt.sh",
		}); err != nil {
			continue
		}
		total++
	}

	if total == 0 {
		return 0, fmt.Errorf("no certificate transparency results found")
	}

	return total, nil
}

func (s *ScanService) runASNScan(asset *model.Asset, job *model.ScanJob) (int, error) {
	raw, err := queryWhois("whois.cymru.com", "begin\nverbose\n"+asset.Name+"\nend\n")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected ASN response")
	}

	var dataLine string
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line != "" {
			dataLine = line
			break
		}
	}

	if dataLine == "" {
		return 0, fmt.Errorf("no ASN data found")
	}

	parts := strings.Split(dataLine, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	payload := map[string]interface{}{
		"raw": raw,
	}

	if len(parts) >= 7 {
		payload["asn"] = parts[0]
		payload["ip"] = parts[1]
		payload["bgp_prefix"] = parts[2]
		payload["cc"] = parts[3]
		payload["registry"] = parts[4]
		payload["allocated"] = parts[5]
		payload["as_name"] = parts[6]
	}

	if err := s.saveResult(job, "asn", payload); err != nil {
		return 0, err
	}

	return 1, nil
}

func getWhoisServer(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return "whois.iana.org"
	}

	tld := parts[len(parts)-1]
	switch tld {
	case "com", "net":
		return "whois.verisign-grs.com"
	case "org":
		return "whois.pir.org"
	case "io":
		return "whois.nic.io"
	case "dev", "app":
		return "whois.nic.google"
	case "info":
		return "whois.afilias.net"
	default:
		return "whois.iana.org"
	}
}

func queryWhois(server, query string) (string, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(server, "43"), 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to connect WHOIS server %s: %w", server, err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))

	if _, err := conn.Write([]byte(query + "\r\n")); err != nil {
		return "", fmt.Errorf("failed to send WHOIS query: %w", err)
	}

	b, err := io.ReadAll(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read WHOIS response: %w", err)
	}

	return string(b), nil
}

func extractWhoisField(raw string, prefixes []string) string {
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		for _, prefix := range prefixes {
			if strings.HasPrefix(strings.TrimSpace(line), prefix) {
				return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), prefix))
			}
		}
	}
	return ""
}