package service

import (
	"errors"
	"time"
)

var (
	ErrQuotaFull             = errors.New("kuota tiket sudah habis")
	ErrAlreadyRegistered     = errors.New("peserta sudah terdaftar pada event ini")
	ErrEventNotOpen          = errors.New("event tidak sedang dibuka untuk pendaftaran")
	ErrEventEnded            = errors.New("event sudah berakhir")
	ErrAlreadyCheckedIn      = errors.New("tiket sudah digunakan")
	ErrRegistrationCancelled = errors.New("tiket sudah dibatalkan")
)

// ValidateQuota memastikan masih ada sisa kuota pada jenis tiket.
func ValidateQuota(quota, sold int) error {
	if sold >= quota {
		return ErrQuotaFull
	}
	return nil
}

// ValidateEventOpen memastikan event berstatus published.
func ValidateEventOpen(status string) error {
	if status != "published" {
		return ErrEventNotOpen
	}
	return nil
}

// ValidateEventNotEnded memastikan event belum selesai.
func ValidateEventNotEnded(endAt, now time.Time) error {
	if now.After(endAt) {
		return ErrEventEnded
	}
	return nil
}

// ValidateNotRegistered menolak pendaftaran ganda pada event yang sama
// (single ticket rule: satu peserta satu tiket per event).
func ValidateNotRegistered(alreadyRegistered bool) error {
	if alreadyRegistered {
		return ErrAlreadyRegistered
	}
	return nil
}

// ValidateCheckIn memastikan tiket boleh dipakai untuk check-in.
func ValidateCheckIn(status string) error {
	switch status {
	case "checked_in":
		return ErrAlreadyCheckedIn
	case "cancelled":
		return ErrRegistrationCancelled
	}
	return nil
}
