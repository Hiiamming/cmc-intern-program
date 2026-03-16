package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mini-asm/internal/model"
	"mini-asm/internal/service"
)

func TestCreateAssetHandler(t *testing.T) {
	mockStorage := &serviceTestStorage{}
	handler := NewAssetHandler(service.NewAssetService(mockStorage))

	body := `{"name":"test.com","type":"domain"}`
	req := httptest.NewRequest(http.MethodPost, "/assets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreateAsset(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

type serviceTestStorage struct{}

func (s *serviceTestStorage) Create(asset *model.Asset) error { return nil }
func (s *serviceTestStorage) GetAll() ([]*model.Asset, error) { return nil, nil }
func (s *serviceTestStorage) GetByID(id string) (*model.Asset, error) { return nil, nil }
func (s *serviceTestStorage) Update(id string, asset *model.Asset) error { return nil }
func (s *serviceTestStorage) Delete(id string) error { return nil }

func (s *serviceTestStorage) BatchDelete(ids []string) (*model.BatchDeleteAssetsResponse, error) {
	return &model.BatchDeleteAssetsResponse{}, nil
}

func (s *serviceTestStorage) Filter(assetType, status string) ([]*model.Asset, error) {
	return nil, nil
}

func (s *serviceTestStorage) Search(query string) ([]*model.Asset, error) { return nil, nil }

func (s *serviceTestStorage) GetStats() (*model.AssetStats, error) {
	return &model.AssetStats{}, nil
}

func (s *serviceTestStorage) Count(assetType, status string) (*model.AssetCountResponse, error) {
	return &model.AssetCountResponse{}, nil
}

func (s *serviceTestStorage) BatchCreate(assets []*model.Asset) ([]string, error) { return nil, nil }