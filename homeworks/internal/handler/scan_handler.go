package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mini-asm/internal/model"
	"mini-asm/internal/service"

	"github.com/google/uuid"
)

type ScanHandler struct {
	scanService *service.ScanService
}

func parseAssetIDFromPath(r *http.Request) (string, error) {
	assetID := strings.Trim(strings.TrimSpace(r.PathValue("id")), `"`)
	if assetID == "" {
		return "", model.ErrInvalidInput
	}
	if _, err := uuid.Parse(assetID); err != nil {
		return "", model.ErrInvalidInput
	}
	return assetID, nil
}

func decodeScanResults(results []*model.ScanResult) []interface{} {
	items := make([]interface{}, 0, len(results))
	for _, result := range results {
		var item interface{}
		if err := json.Unmarshal(result.Data, &item); err != nil {
			item = map[string]interface{}{
				"id":          result.ID,
				"scan_job_id": result.ScanJobID,
				"asset_id":    result.AssetID,
				"result_type": result.ResultType,
				"data":        string(result.Data),
				"created_at":  result.CreatedAt,
			}
		}
		items = append(items, item)
	}
	return items
}

func NewScanHandler(scanService *service.ScanService) *ScanHandler {
	return &ScanHandler{
		scanService: scanService,
	}
}

func (h *ScanHandler) StartScan(w http.ResponseWriter, r *http.Request) {
	assetID := strings.Trim(strings.TrimSpace(r.PathValue("id")), `"`)
	if assetID == "" {
		writeJSONError(w, http.StatusBadRequest, "asset ID is required")
		return
	}

	if _, err := uuid.Parse(assetID); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid asset ID format")
		return
	}

	var req model.StartScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	job, err := h.scanService.StartScan(assetID, req.ScanType)
	if err != nil {
		writeJSONError(w, mapErrorToStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, job)
}

func (h *ScanHandler) GetScanJob(w http.ResponseWriter, r *http.Request) {
	jobID := strings.Trim(strings.TrimSpace(r.PathValue("id")), `"`)
	if jobID == "" {
		writeJSONError(w, http.StatusBadRequest, "job ID is required")
		return
	}

	if _, err := uuid.Parse(jobID); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid job ID format")
		return
	}

	job, err := h.scanService.GetScanJob(jobID)
	if err != nil {
		writeJSONError(w, mapErrorToStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func (h *ScanHandler) GetScanResults(w http.ResponseWriter, r *http.Request) {
	jobID := strings.Trim(strings.TrimSpace(r.PathValue("id")), `"`)
	if jobID == "" {
		writeJSONError(w, http.StatusBadRequest, "job ID is required")
		return
	}

	if _, err := uuid.Parse(jobID); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid job ID format")
		return
	}

	resp, err := h.scanService.GetScanResults(jobID)
	if err != nil {
		writeJSONError(w, mapErrorToStatus(err), err.Error())
		return
	}

	out := map[string]interface{}{
		"job_id":    resp.Job.ID,
		"scan_type": resp.Job.ScanType,
		"results":   []interface{}{},
	}

	items := make([]interface{}, 0, len(resp.Results))
	for _, r := range resp.Results {
		var item interface{}
		if err := json.Unmarshal(r.Data, &item); err != nil {
			item = map[string]interface{}{
				"raw": string(r.Data),
			}
		}
		items = append(items, item)
	}
	out["results"] = items

	writeJSON(w, http.StatusOK, out)
}

func (h *ScanHandler) ListAssetScans(w http.ResponseWriter, r *http.Request) {
	assetID, err := parseAssetIDFromPath(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid asset ID format")
		return
	}

	scans, err := h.scanService.ListAssetScans(assetID)
	if err != nil {
		writeJSONError(w, mapErrorToStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"asset_id": assetID,
		"scans":    scans,
	})
}

func (h *ScanHandler) ListAssetResults(w http.ResponseWriter, r *http.Request) {
	assetID, err := parseAssetIDFromPath(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid asset ID format")
		return
	}

	results, err := h.scanService.ListAssetResults(assetID)
	if err != nil {
		writeJSONError(w, mapErrorToStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"asset_id": assetID,
		"results":  decodeScanResults(results),
	})
}

func (h *ScanHandler) GetAssetDNS(w http.ResponseWriter, r *http.Request) {
	assetID, err := parseAssetIDFromPath(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid asset ID format")
		return
	}

	results, err := h.scanService.ListAssetDNSResults(assetID)
	if err != nil {
		writeJSONError(w, mapErrorToStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"asset_id": assetID,
		"records":  decodeScanResults(results),
	})
}

func (h *ScanHandler) GetAssetWhois(w http.ResponseWriter, r *http.Request) {
	assetID, err := parseAssetIDFromPath(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid asset ID format")
		return
	}

	results, err := h.scanService.ListAssetWhoisResults(assetID)
	if err != nil {
		writeJSONError(w, mapErrorToStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"asset_id": assetID,
		"whois":    decodeScanResults(results),
	})
}

func (h *ScanHandler) GetAssetSubdomains(w http.ResponseWriter, r *http.Request) {
	assetID, err := parseAssetIDFromPath(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid asset ID format")
		return
	}

	results, err := h.scanService.ListAssetSubdomainResults(assetID)
	if err != nil {
		writeJSONError(w, mapErrorToStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"asset_id":   assetID,
		"subdomains": decodeScanResults(results),
	})
}