package service

import (
	"errors"
	"mini-asm/internal/model"
	"testing"
)

func TestAssetService_Create(t *testing.T) {
	tests := []struct {
		name       string
		inputName  string
		inputType  string
		storageErr error
		wantErr    error
	}{
		{name: "valid asset", inputName: "example.com", inputType: model.TypeDomain},
		{name: "empty name", inputName: "", inputType: model.TypeDomain, wantErr: model.ErrEmptyName},
		{name: "invalid type", inputName: "example.com", inputType: "weird", wantErr: model.ErrInvalidType},
		{name: "storage error bubbles up", inputName: "example.com", inputType: model.TypeDomain, storageErr: errors.New("db down")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockAssetStorage{
				createFn: func(asset *model.Asset) error {
					if tt.storageErr != nil {
						return tt.storageErr
					}
					if asset.ID == "" {
						t.Fatal("expected generated ID")
					}
					if asset.Status != model.StatusActive {
						t.Fatalf("expected default status %q, got %q", model.StatusActive, asset.Status)
					}
					return nil
				},
			}
			svc := NewAssetService(store)

			asset, err := svc.CreateAsset(tt.inputName, tt.inputType)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if tt.storageErr != nil {
				if !errors.Is(err, tt.storageErr) {
					t.Fatalf("expected storage error %v, got %v", tt.storageErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if asset == nil || asset.Name != tt.inputName || asset.Type != tt.inputType {
				t.Fatalf("unexpected asset: %+v", asset)
			}
		})
	}
}