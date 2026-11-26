package service

import (
	"testing"
	"time"
)

func TestValidateQuota(t *testing.T) {
	tests := []struct {
		name        string
		quota, sold int
		wantErr     error
	}{
		{"masih ada kuota", 10, 9, nil},
		{"kuota pas habis", 10, 10, ErrQuotaFull},
		{"melebihi kuota", 10, 11, ErrQuotaFull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateQuota(tt.quota, tt.sold); err != tt.wantErr {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEventOpen(t *testing.T) {
	tests := []struct {
		status  string
		wantErr error
	}{
		{"published", nil},
		{"draft", ErrEventNotOpen},
		{"finished", ErrEventNotOpen},
		{"cancelled", ErrEventNotOpen},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if err := ValidateEventOpen(tt.status); err != tt.wantErr {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEventNotEnded(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)

	if err := ValidateEventNotEnded(now.Add(2*time.Hour), now); err != nil {
		t.Errorf("event belum berakhir seharusnya lolos, got %v", err)
	}
	if err := ValidateEventNotEnded(now.Add(-2*time.Hour), now); err != ErrEventEnded {
		t.Errorf("got %v, want %v", err, ErrEventEnded)
	}
}

func TestValidateNotRegistered(t *testing.T) {
	if err := ValidateNotRegistered(false); err != nil {
		t.Errorf("got %v, want nil", err)
	}
	if err := ValidateNotRegistered(true); err != ErrAlreadyRegistered {
		t.Errorf("got %v, want %v", err, ErrAlreadyRegistered)
	}
}

func TestValidateCheckIn(t *testing.T) {
	tests := []struct {
		status  string
		wantErr error
	}{
		{"registered", nil},
		{"checked_in", ErrAlreadyCheckedIn},
		{"cancelled", ErrRegistrationCancelled},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if err := ValidateCheckIn(tt.status); err != tt.wantErr {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRegistrationOwner(t *testing.T) {
	if err := ValidateRegistrationOwner(7, 7); err != nil {
		t.Errorf("pemilik seharusnya lolos, got %v", err)
	}
	if err := ValidateRegistrationOwner(7, 8); err != ErrNotRegistrationOwner {
		t.Errorf("got %v, want %v", err, ErrNotRegistrationOwner)
	}
}

func TestValidateCancel(t *testing.T) {
	tests := []struct {
		status  string
		wantErr error
	}{
		{"registered", nil},
		{"checked_in", ErrAlreadyCheckedIn},
		{"cancelled", ErrRegistrationCancelled},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if err := ValidateCancel(tt.status); err != tt.wantErr {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCancelWindow(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)

	if err := ValidateCancelWindow(now.Add(24*time.Hour), now); err != nil {
		t.Errorf("event belum mulai seharusnya lolos, got %v", err)
	}
	if err := ValidateCancelWindow(now.Add(-time.Hour), now); err != ErrEventStarted {
		t.Errorf("got %v, want %v", err, ErrEventStarted)
	}
	if err := ValidateCancelWindow(now, now); err != ErrEventStarted {
		t.Errorf("tepat saat mulai seharusnya ditolak, got %v", err)
	}
}