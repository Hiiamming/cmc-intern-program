package model

import (
	"encoding/json"
	"time"
)

type ScanType string

const (
	ScanTypeDNS       ScanType = "dns"
	ScanTypeWHOIS     ScanType = "whois"
	ScanTypeSubdomain ScanType = "subdomain"
	ScanTypeCertTrans ScanType = "cert_trans"
	ScanTypeASN       ScanType = "asn"

	// NEW
	ScanTypeIP   ScanType = "ip"
	ScanTypePort ScanType = "port"
	ScanTypeSSL  ScanType = "ssl"
	ScanTypeTech ScanType = "tech"

	ScanTypeAll ScanType = "all"
)

type ScanStatus string

const (
	ScanStatusPending   ScanStatus = "pending"
	ScanStatusRunning   ScanStatus = "running"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
	ScanStatusPartial   ScanStatus = "partial"
)

type ScanJob struct {
	ID        string     `json:"id"`
	AssetID   string     `json:"asset_id"`
	ScanType  ScanType   `json:"scan_type"`
	Status    ScanStatus `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	Error     string     `json:"error"`
	Results   int        `json:"results"`
	CreatedAt time.Time  `json:"created_at"`
}

type ScanResult struct {
	ID         string          `json:"id"`
	ScanJobID  string          `json:"scan_job_id"`
	AssetID    string          `json:"asset_id"`
	ResultType string          `json:"result_type"`
	Data       json.RawMessage `json:"data"`
	CreatedAt  time.Time       `json:"created_at"`
}

type ScanResultsResponse struct {
	Job     *ScanJob      `json:"job"`
	Results []*ScanResult `json:"results"`
}

type StartScanRequest struct {
	ScanType ScanType `json:"scan_type"`
}

func IsValidScanType(t ScanType) bool {
	switch t {
	case ScanTypeDNS,
		ScanTypeWHOIS,
		ScanTypeSubdomain,
		ScanTypeCertTrans,
		ScanTypeASN,
		ScanTypeIP,
		ScanTypePort,
		ScanTypeSSL,
		ScanTypeTech,
		ScanTypeAll:
		return true
	default:
		return false
	}
}