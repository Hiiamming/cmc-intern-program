package postgres

import (
	"database/sql"
	"fmt"

	"mini-asm/internal/config"
	"mini-asm/internal/model"

	_ "github.com/lib/pq" // PostgreSQL driver
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
func (p *PostgresStorage) Filter(assetType, status string) ([]*model.Asset, error) {
	query := `
		SELECT id, name, type, status, created_at, updated_at
		FROM assets
		WHERE 1=1
	`

	var args []interface{}
	argCount := 1

	if assetType != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, assetType)
		argCount++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	rows, err := p.db.Query(query, args...)
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
func (p *PostgresStorage) Count(assetType, status string) (*model.AssetCountResponse, error) {
	query := `
		SELECT COUNT(*)
		FROM assets
		WHERE 1=1
	`

	var args []interface{}
	argCount := 1

	if assetType != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, assetType)
		argCount++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	var count int
	if err := p.db.QueryRow(query, args...).Scan(&count); err != nil {
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
	tx, err := p.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	resp := &model.BatchDeleteAssetsResponse{}

	for _, id := range ids {
		result, execErr := tx.Exec(`DELETE FROM assets WHERE id = $1`, id)
		if execErr != nil {
			err = fmt.Errorf("failed to delete asset %s: %w", id, execErr)
			return nil, err
		}

		rows, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			err = fmt.Errorf("failed to get affected rows for %s: %w", id, rowsErr)
			return nil, err
		}

		if rows == 0 {
			resp.NotFound++
		} else {
			resp.Deleted++
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return resp, nil
}

