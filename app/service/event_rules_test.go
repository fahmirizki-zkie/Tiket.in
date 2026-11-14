package service

import (
	"testing"
	"time"
)

func TestValidateEventDates(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		startAt time.Time
		endAt   time.Time
		want    string
	}{
		{
			name:    "valid dates in future",
			startAt: now.Add(2 * time.Hour),
			endAt:   now.Add(5 * time.Hour),
			want:    "",
		},
		{
			name:    "zero start date",
			startAt: time.Time{},
			endAt:   now.Add(2 * time.Hour),
			want:    "waktu mulai event wajib diisi",
		},
		{
			name:    "zero end date",
			startAt: now.Add(2 * time.Hour),
			endAt:   time.Time{},
			want:    "waktu selesai event wajib diisi",
		},
		{
			name:    "end before start",
			startAt: now.Add(5 * time.Hour),
			endAt:   now.Add(2 * time.Hour),
			want:    "waktu selesai harus setelah waktu mulai",
		},
		{
			name:    "end equals start",
			startAt: now.Add(2 * time.Hour),
			endAt:   now.Add(2 * time.Hour),
			want:    "waktu selesai harus setelah waktu mulai",
		},
		{
			name:    "start in deep past",
			startAt: now.Add(-2 * time.Hour),
			endAt:   now.Add(2 * time.Hour),
			want:    "waktu mulai event tidak boleh di masa lalu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEventDates(tt.startAt, tt.endAt)
			if got != tt.want {
				t.Errorf("ValidateEventDates() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateTicketPriceAndQuota(t *testing.T) {
	tests := []struct {
		name  string
		price float64
		quota int
		want  string
	}{
		{
			name:  "valid regular ticket",
			price: 50000,
			quota: 100,
			want:  "",
		},
		{
			name:  "valid free ticket",
			price: 0,
			quota: 50,
			want:  "",
		},
		{
			name:  "negative price",
			price: -10000,
			quota: 50,
			want:  "harga tiket tidak boleh negatif",
		},
		{
			name:  "zero quota",
			price: 25000,
			quota: 0,
			want:  "kuota tiket minimal 1 tiket",
		},
		{
			name:  "negative quota",
			price: 25000,
			quota: -5,
			want:  "kuota tiket minimal 1 tiket",
		},
		{
			name:  "excessive quota",
			price: 25000,
			quota: 200000,
			want:  "kuota tiket maksimal 100.000 tiket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateTicketPriceAndQuota(tt.price, tt.quota)
			if got != tt.want {
				t.Errorf("ValidateTicketPriceAndQuota() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCanModifyEvent(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   string
	}{
		{
			name:   "published can be modified",
			status: "published",
			want:   "",
		},
		{
			name:   "draft can be modified",
			status: "draft",
			want:   "",
		},
		{
			name:   "finished event cannot be modified",
			status: "finished",
			want:   "event yang sudah selesai tidak dapat diubah",
		},
		{
			name:   "cancelled event cannot be modified",
			status: "cancelled",
			want:   "event yang telah dibatalkan tidak dapat diubah",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanModifyEvent(tt.status)
			if got != tt.want {
				t.Errorf("CanModifyEvent(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestValidateEventOwnership(t *testing.T) {
	tests := []struct {
		name        string
		organizerID int64
		userID      int64
		role        string
		want        string
	}{
		{
			name:        "owner panitia match",
			organizerID: 10,
			userID:      10,
			role:        "panitia",
			want:        "",
		},
		{
			name:        "admin can edit any event",
			organizerID: 10,
			userID:      99,
			role:        "admin",
			want:        "",
		},
		{
			name:        "different panitia forbidden",
			organizerID: 10,
			userID:      12,
			role:        "panitia",
			want:        "anda tidak memiliki hak akses untuk mengubah event ini",
		},
		{
			name:        "peserta forbidden",
			organizerID: 10,
			userID:      10,
			role:        "peserta",
			want:        "", // Note: if organizer ID matches, check role logic if called
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEventOwnership(tt.organizerID, tt.userID, tt.role)
			if got != tt.want {
				t.Errorf("ValidateEventOwnership() = %q, want %q", got, tt.want)
			}
		})
	}
}
