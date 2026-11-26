package service

import (
	"regexp"
	"strings"
	"unicode"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return "email wajib diisi"
	}
	if len(email) > 100 {
		return "email maksimal 100 karakter"
	}
	if !emailRegex.MatchString(email) {
		return "format email tidak valid"
	}
	return ""
}

func ValidatePassword(password string) string {
	if password == "" {
		return "password wajib diisi"
	}
	if len(password) < 8 {
		return "password minimal 8 karakter"
	}
	if len(password) > 72 {
		return "password maksimal 72 karakter"
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper {
		return "password harus mengandung huruf besar"
	}
	if !hasLower {
		return "password harus mengandung huruf kecil"
	}
	if !hasDigit {
		return "password harus mengandung angka"
	}

	return ""
}

func ValidateRole(role string) string {
	role = strings.TrimSpace(strings.ToLower(role))
	switch role {
	case "peserta", "panitia":
		return ""
	default:
		return "role harus peserta atau panitia"
	}
}

// ValidateAssignableRole memeriksa role yang boleh diberikan admin.
// Berbeda dengan ValidateRole (registrasi publik) yang menolak "admin".
func ValidateAssignableRole(role string) string {
	role = strings.TrimSpace(strings.ToLower(role))
	switch role {
	case "peserta", "panitia", "admin":
		return ""
	default:
		return "role harus peserta, panitia, atau admin"
	}
}

// CanChangeRole: admin tidak boleh mengubah role dirinya sendiri,
// supaya sistem tidak pernah kehilangan admin terakhir.
func CanChangeRole(actorID, targetID int64) string {
	if actorID == targetID {
		return "admin tidak dapat mengubah role dirinya sendiri"
	}
	return ""
}

// ValidateName memvalidasi nama tampilan user (dipakai update profil)
func ValidateName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "nama wajib diisi"
	}
	n := len([]rune(name))
	if n < 2 {
		return "nama minimal 2 karakter"
	}
	if n > 100 {
		return "nama maksimal 100 karakter"
	}
	return ""
}

// ValidatePasswordChange memeriksa pasangan password lama dan baru
func ValidatePasswordChange(oldPassword, newPassword string) string {
	if oldPassword == "" {
		return "password lama wajib diisi"
	}
	if msg := ValidatePassword(newPassword); msg != "" {
		return msg
	}
	if oldPassword == newPassword {
		return "password baru tidak boleh sama dengan password lama"
	}
	return ""
}
