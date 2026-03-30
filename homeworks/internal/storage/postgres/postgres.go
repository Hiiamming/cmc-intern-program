package postgres

import (
	"database/sql"
	"fmt"

	"mini-asm/internal/config"
	"mini-asm/internal/model"

	"github.com/lib/pq" // PostgreSQL driver
)

// PostgresStorage implements the Storage interface using PostgreSQL
// This is a concrete implementation that can be swapped with MemoryStorage
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func NewPostgresStorageFromConfig(config *config.PostgresConfig) (*PostgresStorage, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost,
		config.DBPort,
		config.DBUser,
		config.DBPassword,
		config.DBName,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &PostgresStorage{db: db}, nil
}

// Create inserts a new asset into the database
func (p *PostgresStorage) Create(asset *model.Asset) error {
	query := `
		INSERT INTO assets (id, name, type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := p.db.Exec(
		query,
		asset.ID,
		asset.Name,
		asset.Type,
		asset.Status,
		asset.CreatedAt,
		asset.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}

	return nil
}

// GetAll retrieves all assets from the database
func (p *PostgresStorage) GetAll() ([]*model.Asset, error) {
	query := `
		SELECT id, name, type, status, created_at, updated_at
		FROM assets
		ORDER BY created_at DESC
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query assets: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		asset := &model.Asset{}
		err := rows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.Type,
			&asset.Status,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, asset)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return assets, nil
}

// GetByID retrieves a single asset by its ID
func (p *PostgresStorage) GetByID(id string) (*model.Asset, error) {
	query := `
		SELECT id, name, type, status, created_at, updated_at
		FROM assets
		WHERE id = $1
	`

	asset := &model.Asset{}
	err := p.db.QueryRow(query, id).Scan(
		&asset.ID,
		&asset.Name,
		&asset.Type,
		&asset.Status,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return asset, nil
}

// Update modifies an existing asset in the database
func (p *PostgresStorage) Update(id string, asset *model.Asset) error {
	query := `
		UPDATE assets
		SET name = $1, type = $2, status = $3, updated_at = $4
		WHERE id = $5
	`

	result, err := p.db.Exec(
		query,
		asset.Name,
		asset.Type,
		asset.Status,
		asset.UpdatedAt,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

// Delete removes an asset from the database
func (p *PostgresStorage) Delete(id string) error {
	query := `DELETE FROM assets WHERE id = $1`

	result, err := p.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

// Filter returns assets matching the given criteria
// Filter returns assets matching the given criteria
func (p *PostgresStorage) Filter(assetType, status string) ([]*model.Asset, error) {
	var (
		query string
		rows  *sql.Rows
		err   error
	)

	switch {
	case assetType != "" && status != "":
		query = `
			SELECT id, name, type, status, created_at, updated_at
			FROM assets
			WHERE type = $1 AND status = $2
			ORDER BY created_at DESC
		`
		rows, err = p.db.Query(query, assetType, status)

	case assetType != "":
		query = `
			SELECT id, name, type, status, created_at, updated_at
			FROM assets
			WHERE type = $1
			ORDER BY created_at DESC
		`
		rows, err = p.db.Query(query, assetType)

	case status != "":
		query = `
			SELECT id, name, type, status, created_at, updated_at
			FROM assets
			WHERE status = $1
			ORDER BY created_at DESC
		`
		rows, err = p.db.Query(query, status)

	default:
		query = `
			SELECT id, name, type, status, created_at, updated_at
			FROM assets
			ORDER BY created_at DESC
		`
		rows, err = p.db.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to filter assets: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		asset := &model.Asset{}
		err := rows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.Type,
			&asset.Status,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, asset)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return assets, nil
}

// Search finds assets by partial name match
func (p *PostgresStorage) Search(query string) ([]*model.Asset, error) {

	sqlQuery := `
		SELECT id, name, type, status, created_at, updated_at
		FROM assets
		WHERE name ILIKE $1
		LIMIT 100
	`

	searchPattern := "%" + query + "%"

	rows, err := p.db.Query(sqlQuery, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search assets: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset

	for rows.Next() {
		asset := &model.Asset{}

		err := rows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.Type,
			&asset.Status,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

// GetStats returns aggregated statistics about assets
func (p *PostgresStorage) GetStats() (*model.AssetStats, error) {
	stats := &model.AssetStats{
		ByType:   make(map[string]int),
		ByStatus: make(map[string]int),
	}

	err := p.db.QueryRow(`SELECT COUNT(*) FROM assets`).Scan(&stats.Total)
	if err != nil {
		return nil, fmt.Errorf("failed to count total assets: %w", err)
	}

	typeRows, err := p.db.Query(`
		SELECT type, COUNT(*)
		FROM assets
		GROUP BY type
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to group assets by type: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var assetType string
		var count int
		if err := typeRows.Scan(&assetType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type stats: %w", err)
		}
		stats.ByType[assetType] = count
	}

	if err := typeRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating type stats: %w", err)
	}

	statusRows, err := p.db.Query(`
		SELECT status, COUNT(*)
		FROM assets
		GROUP BY status
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to group assets by status: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status stats: %w", err)
		}
		stats.ByStatus[status] = count
	}

	if err := statusRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating status stats: %w", err)
	}

	return stats, nil
}

// Count returns the total number of assets matching criteria
// Count returns the total number of assets matching criteria
func (p *PostgresStorage) Count(assetType, status string) (*model.AssetCountResponse, error) {
	var (
		query string
		count int
		err   error
	)

	switch {
	case assetType != "" && status != "":
		query = `
			SELECT COUNT(*)
			FROM assets
			WHERE type = $1 AND status = $2
		`
		err = p.db.QueryRow(query, assetType, status).Scan(&count)

	case assetType != "":
		query = `
			SELECT COUNT(*)
			FROM assets
			WHERE type = $1
		`
		err = p.db.QueryRow(query, assetType).Scan(&count)

	case status != "":
		query = `
			SELECT COUNT(*)
			FROM assets
			WHERE status = $1
		`
		err = p.db.QueryRow(query, status).Scan(&count)

	default:
		query = `
			SELECT COUNT(*)
			FROM assets
		`
		err = p.db.QueryRow(query).Scan(&count)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	return &model.AssetCountResponse{
		Count: count,
		Filters: model.AssetCountFilters{
			Type:   assetType,
			Status: status,
		},
	}, nil
}

func (p *PostgresStorage) BatchCreate(assets []*model.Asset) ([]string, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `
		INSERT INTO assets (id, name, type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	ids := make([]string, 0, len(assets))

	for _, asset := range assets {
		_, err = tx.Exec(
			query,
			asset.ID,
			asset.Name,
			asset.Type,
			asset.Status,
			asset.CreatedAt,
			asset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert asset %s: %w", asset.Name, err)
		}

		ids = append(ids, asset.ID)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ids, nil
}

func (p *PostgresStorage) BatchDelete(ids []string) (*model.BatchDeleteAssetsResponse, error) {
	result, err := p.db.Exec(`DELETE FROM assets WHERE id = ANY($1::uuid[])`, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to batch delete assets: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get affected rows: %w", err)
	}

	deleted := int(rows)

	return &model.BatchDeleteAssetsResponse{
		Deleted:  deleted,
		NotFound: len(ids) - deleted,
	}, nil
}

func (p *PostgresStorage) CreateScanJob(job *model.ScanJob) error {
	query := `
		INSERT INTO scan_jobs (id, asset_id, scan_type, status, started_at, ended_at, error, results, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := p.db.Exec(
		query,
		job.ID,
		job.AssetID,
		job.ScanType,
		job.Status,
		job.StartedAt,
		job.EndedAt,
		job.Error,
		job.Results,
		job.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create scan job: %w", err)
	}

	return nil
}

func (p *PostgresStorage) GetScanJob(id string) (*model.ScanJob, error) {
	query := `
		SELECT id, asset_id, scan_type, status, started_at, ended_at, error, results, created_at
		FROM scan_jobs
		WHERE id = $1
	`

	job := &model.ScanJob{}
	err := p.db.QueryRow(query, id).Scan(
		&job.ID,
		&job.AssetID,
		&job.ScanType,
		&job.Status,
		&job.StartedAt,
		&job.EndedAt,
		&job.Error,
		&job.Results,
		&job.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scan job: %w", err)
	}

	return job, nil
}

func (p *PostgresStorage) UpdateScanJob(job *model.ScanJob) error {
	query := `
		UPDATE scan_jobs
		SET status = $1, ended_at = $2, error = $3, results = $4
		WHERE id = $5
	`

	result, err := p.db.Exec(
		query,
		job.Status,
		job.EndedAt,
		job.Error,
		job.Results,
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update scan job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (p *PostgresStorage) CreateScanResult(resultItem *model.ScanResult) error {
	query := `
		INSERT INTO scan_results (id, scan_job_id, asset_id, result_type, data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := p.db.Exec(
		query,
		resultItem.ID,
		resultItem.ScanJobID,
		resultItem.AssetID,
		resultItem.ResultType,
		resultItem.Data,
		resultItem.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create scan result: %w", err)
	}

	return nil
}

func (p *PostgresStorage) ListScanResultsByJob(scanJobID string) ([]*model.ScanResult, error) {
	query := `
		SELECT id, scan_job_id, asset_id, result_type, data, created_at
		FROM scan_results
		WHERE scan_job_id = $1
		ORDER BY created_at ASC
	`

	rows, err := p.db.Query(query, scanJobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query scan results: %w", err)
	}
	defer rows.Close()

	results := make([]*model.ScanResult, 0)
	for rows.Next() {
		item := &model.ScanResult{}
		var rawData []byte

		if err := rows.Scan(
			&item.ID,
			&item.ScanJobID,
			&item.AssetID,
			&item.ResultType,
			&rawData,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan scan result: %w", err)
		}

		item.Data = rawData
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating scan results: %w", err)
	}

	return results, nil
}

func (p *PostgresStorage) ListScanJobsByAsset(assetID string) ([]*model.ScanJob, error) {
	query := `
		SELECT id, asset_id, scan_type, status, started_at, ended_at, error, results, created_at
		FROM scan_jobs
		WHERE asset_id = $1
		ORDER BY created_at DESC
	`
	rows, err := p.db.Query(query, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query scan jobs by asset: %w", err)
	}
	defer rows.Close()

	var jobs []*model.ScanJob
	for rows.Next() {
		job := &model.ScanJob{}
		if err := rows.Scan(
			&job.ID, &job.AssetID, &job.ScanType, &job.Status,
			&job.StartedAt, &job.EndedAt, &job.Error, &job.Results, &job.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan scan job: %w", err)
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (p *PostgresStorage) ListScanResultsByAsset(assetID string) ([]*model.ScanResult, error) {
	query := `
		SELECT id, scan_job_id, asset_id, result_type, data, created_at
		FROM scan_results
		WHERE asset_id = $1
		ORDER BY created_at DESC
	`
	rows, err := p.db.Query(query, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query scan results by asset: %w", err)
	}
	defer rows.Close()

	var results []*model.ScanResult
	for rows.Next() {
		item := &model.ScanResult{}
		var rawData []byte
		if err := rows.Scan(
			&item.ID, &item.ScanJobID, &item.AssetID, &item.ResultType, &rawData, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan scan result: %w", err)
		}
		item.Data = rawData
		results = append(results, item)
	}
	return results, rows.Err()
}

func (p *PostgresStorage) ListScanResultsByAssetAndTypes(assetID string, resultTypes []string) ([]*model.ScanResult, error) {
	query := `
		SELECT id, scan_job_id, asset_id, result_type, data, created_at
		FROM scan_results
		WHERE asset_id = $1
		  AND result_type = ANY($2)
		ORDER BY created_at DESC
	`

	rows, err := p.db.Query(query, assetID, pq.Array(resultTypes))
	if err != nil {
		return nil, fmt.Errorf("failed to query scan results by asset and type: %w", err)
	}
	defer rows.Close()

	var results []*model.ScanResult
	for rows.Next() {
		result := &model.ScanResult{}
		var rawData []byte

		err := rows.Scan(
			&result.ID,
			&result.ScanJobID,
			&result.AssetID,
			&result.ResultType,
			&rawData,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan filtered scan result: %w", err)
		}

		result.Data = rawData
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating filtered scan results rows: %w", err)
	}

	return results, nil
}
