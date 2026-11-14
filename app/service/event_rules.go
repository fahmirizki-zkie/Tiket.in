package service

import (
	"strings"
	"time"
)

// ValidateEventDates memvalidasi kelayakan jadwal waktu pelaksanaan event
func ValidateEventDates(startAt, endAt time.Time) string {
	if startAt.IsZero() {
		return "waktu mulai event wajib diisi"
	}
	if endAt.IsZero() {
		return "waktu selesai event wajib diisi"
	}
	if !endAt.After(startAt) {
		return "waktu selesai harus setelah waktu mulai"
	}
	// Toleransi 5 menit untuk variasi clock drift jaringan
	if startAt.Before(time.Now().Add(-5 * time.Minute)) {
		return "waktu mulai event tidak boleh di masa lalu"
	}
	return ""
}

// ValidateTicketPriceAndQuota memvalidasi kelayakan batasan kuota dan harga jenis tiket
func ValidateTicketPriceAndQuota(price float64, quota int) string {
	if price < 0 {
		return "harga tiket tidak boleh negatif"
	}
	if quota <= 0 {
		return "kuota tiket minimal 1 tiket"
	}
	if quota > 100000 {
		return "kuota tiket maksimal 100.000 tiket"
	}
	return ""
}

// CanModifyEvent memastikan event yang sudah selesai atau dibatalkan tidak dapat diubah
func CanModifyEvent(eventStatus string) string {
	status := strings.ToLower(strings.TrimSpace(eventStatus))
	switch status {
	case "finished":
		return "event yang sudah selesai tidak dapat diubah"
	case "cancelled":
		return "event yang telah dibatalkan tidak dapat diubah"
	default:
		return ""
	}
}

// ValidateEventOwnership memvalidasi apakah user memiliki kewenangan mengubah event (pemilik atau admin)
func ValidateEventOwnership(eventOrganizerID, currentUserID int64, currentUserRole string) string {
	if strings.ToLower(currentUserRole) == "admin" {
		return ""
	}
	if eventOrganizerID != currentUserID {
		return "anda tidak memiliki hak akses untuk mengubah event ini"
	}
	return ""
}
