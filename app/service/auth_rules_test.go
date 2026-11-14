package service

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"valid", "user@example.com", ""},
		{"empty", "", "email wajib diisi"},
		{"no_at", "userexample.com", "format email tidak valid"},
		{"no_domain", "user@", "format email tidak valid"},
		{"too_long", string(make([]byte, 101)) + "@x.com", "email maksimal 100 karakter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEmail(tt.email)
			if got != tt.want {
				t.Errorf("ValidateEmail(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{"valid", "Secret123", ""},
		{"empty", "", "password wajib diisi"},
		{"too_short", "Ab1", "password minimal 8 karakter"},
		{"no_upper", "secret123", "password harus mengandung huruf besar"},
		{"no_lower", "SECRET123", "password harus mengandung huruf kecil"},
		{"no_digit", "SecretPass", "password harus mengandung angka"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePassword(tt.password)
			if got != tt.want {
				t.Errorf("ValidatePassword(%q) = %q, want %q", tt.password, got, tt.want)
			}
		})
	}
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want string
	}{
		{"peserta", "peserta", ""},
		{"panitia", "panitia", ""},
		{"admin", "admin", "role harus peserta atau panitia"},
		{"empty", "", "role harus peserta atau panitia"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateRole(tt.role)
			if got != tt.want {
				t.Errorf("ValidateRole(%q) = %q, want %q", tt.role, got, tt.want)
			}
		})
	}
}
