package service

import (
	"context"
	"mini-asm/internal/model"
	"mini-asm/internal/storage"
)

type mockAssetStorage struct {
	createFn      func(*model.Asset) error
	getAllFn      func() ([]*model.Asset, error)
	getByIDFn     func(string) (*model.Asset, error)
	updateFn      func(string, *model.Asset) error
	deleteFn      func(string) error
	batchDeleteFn func([]string) (*model.BatchDeleteAssetsResponse, error)
	filterFn      func(string, string) ([]*model.Asset, error)
	searchFn      func(string) ([]*model.Asset, error)
	getStatsFn    func() (*model.AssetStats, error)
	countFn       func(string, string) (*model.AssetCountResponse, error)
	batchCreateFn func([]*model.Asset) ([]string, error)
}

var _ storage.Storage = (*mockAssetStorage)(nil)

func (m *mockAssetStorage) Create(asset *model.Asset) error {
	if m.createFn != nil {
		return m.createFn(asset)
	}
	return nil
}

func (m *mockAssetStorage) GetAll() ([]*model.Asset, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, nil
}

func (m *mockAssetStorage) GetByID(id string) (*model.Asset, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, model.ErrNotFound
}

func (m *mockAssetStorage) Update(id string, asset *model.Asset) error {
	if m.updateFn != nil {
		return m.updateFn(id, asset)
	}
	return nil
}

func (m *mockAssetStorage) Delete(id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func (m *mockAssetStorage) BatchDelete(ids []string) (*model.BatchDeleteAssetsResponse, error) {
	if m.batchDeleteFn != nil {
		return m.batchDeleteFn(ids)
	}
	return &model.BatchDeleteAssetsResponse{}, nil
}

func (m *mockAssetStorage) Filter(assetType, status string) ([]*model.Asset, error) {
	if m.filterFn != nil {
		return m.filterFn(assetType, status)
	}
	return nil, nil
}

func (m *mockAssetStorage) Search(query string) ([]*model.Asset, error) {
	if m.searchFn != nil {
		return m.searchFn(query)
	}
	return nil, nil
}

func (m *mockAssetStorage) GetStats() (*model.AssetStats, error) {
	if m.getStatsFn != nil {
		return m.getStatsFn()
	}
	return &model.AssetStats{}, nil
}

func (m *mockAssetStorage) Count(assetType, status string) (*model.AssetCountResponse, error) {
	if m.countFn != nil {
		return m.countFn(assetType, status)
	}
	return &model.AssetCountResponse{}, nil
}

func (m *mockAssetStorage) BatchCreate(assets []*model.Asset) ([]string, error) {
	if m.batchCreateFn != nil {
		return m.batchCreateFn(assets)
	}
	ids := make([]string, 0, len(assets))
	for _, asset := range assets {
		ids = append(ids, asset.ID)
	}
	return ids, nil
}

type mockScanStorage struct {
	createdResults []*model.ScanResult
}

var _ storage.ScanStorage = (*mockScanStorage)(nil)

func (m *mockScanStorage) CreateScanJob(job *model.ScanJob) error { return nil }

func (m *mockScanStorage) GetScanJob(id string) (*model.ScanJob, error) {
	return nil, model.ErrNotFound
}

func (m *mockScanStorage) UpdateScanJob(job *model.ScanJob) error { return nil }

func (m *mockScanStorage) CreateScanResult(result *model.ScanResult) error {
	m.createdResults = append(m.createdResults, result)
	return nil
}

func (m *mockScanStorage) ListScanResultsByJob(scanJobID string) ([]*model.ScanResult, error) {
	return nil, nil
}

func (m *mockScanStorage) ListScanJobsByAsset(assetID string) ([]*model.ScanJob, error) {
	return nil, nil
}

func (m *mockScanStorage) ListScanResultsByAsset(assetID string) ([]*model.ScanResult, error) {
	return nil, nil
}

func (m *mockScanStorage) ListScanResultsByAssetAndTypes(assetID string, resultTypes []string) ([]*model.ScanResult, error) {
	return nil, nil
}

func newTestScanService() (*ScanService, *mockScanStorage) {
	scanStore := &mockScanStorage{}
	svc := NewScanService(&mockAssetStorage{}, scanStore)
	return svc, scanStore
}

func newTestJobAndAsset(scanType model.ScanType, assetType, name string) (*model.Asset, *model.ScanJob) {
	asset := &model.Asset{ID: "asset-1", Name: name, Type: assetType}
	job := &model.ScanJob{ID: "job-1", AssetID: asset.ID, ScanType: scanType}
	return asset, job
}

func scanWithContext(scanner Scanner, asset *model.Asset, job *model.ScanJob) (int, error) {
	return scanner.Scan(context.Background(), asset, job)
}