package model

import "testing"

func TestAssetValidation(t *testing.T) {
	tests := []struct {
		name       string
		asset      Asset
		wantType   bool
		wantStatus bool
	}{
		{name: "valid domain asset", asset: Asset{Name: "example.com", Type: TypeDomain, Status: StatusActive}, wantType: true, wantStatus: true},
		{name: "valid ip asset", asset: Asset{Name: "127.0.0.1", Type: TypeIP, Status: StatusInactive}, wantType: true, wantStatus: true},
		{name: "valid service asset", asset: Asset{Name: "ssh", Type: TypeService, Status: StatusActive}, wantType: true, wantStatus: true},
		{name: "invalid asset type", asset: Asset{Name: "test", Type: "invalid", Status: StatusActive}, wantType: false, wantStatus: true},
		{name: "invalid asset status", asset: Asset{Name: "test", Type: TypeDomain, Status: "weird"}, wantType: true, wantStatus: false},
		{name: "empty values", asset: Asset{}, wantType: false, wantStatus: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidType(tt.asset.Type); got != tt.wantType {
				t.Fatalf("IsValidType(%q) = %v, want %v", tt.asset.Type, got, tt.wantType)
			}
			if got := IsValidStatus(tt.asset.Status); got != tt.wantStatus {
				t.Fatalf("IsValidStatus(%q) = %v, want %v", tt.asset.Status, got, tt.wantStatus)
			}
		})
	}
}