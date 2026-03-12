package service

import (
	"mini-asm/internal/model"
	"mini-asm/internal/storage"
	"time"
	"fmt"
	"strings"
	"github.com/google/uuid"
)

// AssetService handles business logic for asset operations
// It sits between handlers (HTTP layer) and storage (data layer)
type AssetService struct {
	storage storage.Storage // Dependency injection - any Storage implementation
}

// NewAssetService creates a new asset service
// Takes Storage interface - can be memory, database, or mock
func NewAssetService(storage storage.Storage) *AssetService {
	return &AssetService{
		storage: storage,
	}
}

// CreateAsset creates a new asset with validation
// Returns the created asset or an error
func (s *AssetService) CreateAsset(name, assetType string) (*model.Asset, error) {
	// Validation - business rules enforcement
	if name == "" {
		return nil, model.ErrEmptyName
	}

	if !model.IsValidType(assetType) {
		return nil, model.ErrInvalidType
	}

	// Business logic - create asset with defaults
	asset := &model.Asset{
		ID:        uuid.New().String(), // Auto-generate UUID
		Name:      name,
		Type:      assetType,
		Status:    model.StatusActive, // Default status
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Delegate to storage layer
	if err := s.storage.Create(asset); err != nil {
		return nil, err
	}

	return asset, nil
}

// GetAllAssets retrieves all assets
func (s *AssetService) GetAllAssets() ([]*model.Asset, error) {
	return s.storage.GetAll()
}

// GetAssetByID retrieves a single asset by ID
func (s *AssetService) GetAssetByID(id string) (*model.Asset, error) {
	if id == "" {
		return nil, model.ErrInvalidInput
	}

	return s.storage.GetByID(id)
}

// UpdateAsset updates an existing asset
// Only updates provided fields (partial update)
func (s *AssetService) UpdateAsset(id string, name, assetType, status string) (*model.Asset, error) {
	// Validate ID
	if id == "" {
		return nil, model.ErrInvalidInput
	}

	// Get existing asset
	existing, err := s.storage.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Apply updates (only if provided)
	if name != "" {
		existing.Name = name
	}

	if assetType != "" {
		if !model.IsValidType(assetType) {
			return nil, model.ErrInvalidType
		}
		existing.Type = assetType
	}

	if status != "" {
		if !model.IsValidStatus(status) {
			return nil, model.ErrInvalidStatus
		}
		existing.Status = status
	}

	// Update timestamp
	existing.UpdatedAt = time.Now()

	// Save to storage
	if err := s.storage.Update(id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteAsset removes an asset
func (s *AssetService) DeleteAsset(id string) error {
	if id == "" {
		return model.ErrInvalidInput
	}

	return s.storage.Delete(id)
}

// FilterAssets returns assets matching criteria
func (s *AssetService) FilterAssets(assetType, status string) ([]*model.Asset, error) {
	// Validate filters if provided
	if assetType != "" && !model.IsValidType(assetType) {
		return nil, model.ErrInvalidType
	}

	if status != "" && !model.IsValidStatus(status) {
		return nil, model.ErrInvalidStatus
	}

	return s.storage.Filter(assetType, status)
}

// SearchAssets finds assets by name
func (s *AssetService) SearchAssets(query string) ([]*model.Asset, error) {
	query = strings.TrimSpace(query)
	
	if query == "" {
		return nil, model.ErrInvalidInput
	}

	return s.storage.Search(query)
}

func (s *AssetService) GetStats() (*model.AssetStats, error) {
	return s.storage.GetStats()
}

func (s *AssetService) Count(assetType, status string) (*model.AssetCountResponse, error) {
	if assetType != "" && !model.IsValidType(assetType) {
		return nil, model.ErrInvalidType
	}

	if status != "" && !model.IsValidStatus(status) {
		return nil, model.ErrInvalidStatus
	}

	result, err := s.storage.Count(assetType, status)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *AssetService) BatchCreateAssets(req *model.BatchCreateAssetsRequest) (*model.BatchCreateAssetsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}

	if len(req.Assets) == 0 {
		return nil, fmt.Errorf("assets list cannot be empty")
	}

	if len(req.Assets) > 100 {
		return nil, fmt.Errorf("maximum 100 assets per request")
	}

	assets := make([]*model.Asset, 0, len(req.Assets))

	for _, item := range req.Assets {
		if item.Name == "" {
			return nil, fmt.Errorf("asset name is required")
		}

		if item.Type == "" {
			return nil, fmt.Errorf("asset type is required")
		}

		if item.Type != "domain" && item.Type != "ip" && item.Type != "service" {
			return nil, fmt.Errorf("invalid asset type: %s", item.Type)
		}

		now := time.Now()

		asset := &model.Asset{
			ID:        uuid.New().String(),
			Name:      item.Name,
			Type:      item.Type,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}

		assets = append(assets, asset)
	}

	ids, err := s.storage.BatchCreate(assets)
	if err != nil {
		return nil, err
	}

	return &model.BatchCreateAssetsResponse{
		Created: len(ids),
		IDs:     ids,
	}, nil
}

func (s *AssetService) BatchDeleteAssets(ids []string) (*model.BatchDeleteAssetsResponse, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("ids query param is required")
	}

	validIDs := make([]string, 0, len(ids))
	notFoundFromInvalid := 0

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}

		if _, err := uuid.Parse(id); err != nil {
			notFoundFromInvalid++
			continue
		}

		validIDs = append(validIDs, id)
	}

	if len(validIDs) == 0 {
		return &model.BatchDeleteAssetsResponse{
			Deleted:  0,
			NotFound: notFoundFromInvalid,
		}, nil
	}

	resp, err := s.storage.BatchDelete(validIDs)
	if err != nil {
		return nil, err
	}

	resp.NotFound += notFoundFromInvalid
	return resp, nil
}


/*
🎓 NOTES:

1. Service Layer Responsibilities:
   ✅ Business logic
   ✅ Validation
   ✅ Orchestration (coordinate multiple operations)
   ✅ Default values
   ❌ HTTP concerns (status codes, JSON)
   ❌ Database details (SQL, queries)

2. Dependency Injection:
   func NewAssetService(storage storage.Storage) *AssetService

   Q: Tại sao không dùng global variable?
   A: Testability! Có thể inject mock storage trong tests

   Example:
   // Production
   service := NewAssetService(memory.NewMemoryStorage())

   // Testing
   service := NewAssetService(&MockStorage{})

3. Validation Strategy:
   - Validate BEFORE business logic
   - Return specific errors (ErrEmptyName, ErrInvalidType)
   - Handler layer maps to HTTP status codes

4. Business Logic Examples:
   - Auto-generate UUID
   - Set default status = active
   - Auto-set timestamps
   - Partial updates (only update provided fields)

5. Error Propagation:
   if err := s.storage.Create(asset); err != nil {
       return nil, err  // Let caller handle
   }

   Q: Tại sao không handle error ở đây?
   A: Service không biết context (HTTP? CLI? gRPC?)
      Handler layer sẽ decide status code

6. Comparison với "Fat Controller":

   ❌ BAD (All in handler):
   func CreateAssetHandler(w, r) {
       // Parse JSON
       // Validate
       // Generate UUID
       // Set defaults
       // Save to DB
       // Return response
   }
   → Hard to test, hard to reuse

   ✅ GOOD (Layered):
   Handler: Parse JSON, call service, return HTTP response
   Service: Validate, business logic
   Storage: Data persistence
   → Easy to test each layer, reusable logic

7. UUID Generation:
   uuid.New().String() → "550e8400-e29b-41d4-a716-446655440000"
   - Globally unique
   - No need for database auto-increment
   - Can generate offline
   - URL-safe

8. Timestamps:
   CreatedAt: time.Now()  // Set once
   UpdatedAt: time.Now()  // Update on every change

   Q: Tại sao không dùng int64 (Unix timestamp)?
   A: time.Time có timezone, human-readable, JSON support

📝 TESTING SERVICE LAYER:

func TestCreateAsset(t *testing.T) {
    // Arrange
    mockStorage := &MockStorage{}
    service := NewAssetService(mockStorage)

    // Act
    asset, err := service.CreateAsset("example.com", "domain")

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, asset.ID)
    assert.Equal(t, "active", asset.Status) // Default
}

❓ QUESTIONS TO ASK:

1. Tại sao CreateAsset return (*Asset, error) thay vì chỉ error?
   → Need to return created asset with ID to client

2. Tại sao UpdateAsset là partial update?
   → Flexible - client chỉ gửi fields muốn update

3. Service layer có nên biết về HTTP không?
   → KHÔNG! Separation of concerns

4. Làm sao test service layer mà không cần database?
   → Mock storage! (Buổi 5)
*/