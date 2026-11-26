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

func ulangiString(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func TestValidateEmail_Tambahan(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"spasi_saja", "   ", "email wajib diisi"},
		{"tld_satu_huruf", "user@example.c", "format email tidak valid"},
		{"ada_spasi_di_tengah", "us er@example.com", "format email tidak valid"},
		{"subdomain_valid", "user@mail.unair.ac.id", ""},
		{"plus_tag_valid", "user+tiket@example.com", ""},
		{"tanpa_local_part", "@example.com", "format email tidak valid"},
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

func TestValidatePassword_Tambahan(t *testing.T) {
	// 72 karakter valid, 73 karakter ditolak
	p72 := "Aa1" + ulangiString("x", 69)
	p73 := p72 + "x"

	tests := []struct {
		name     string
		password string
		want     string
	}{
		{"tepat_72", p72, ""},
		{"tepat_73", p73, "password maksimal 72 karakter"},
		{"tepat_8", "Abcdef12", ""},
		{"tujuh_karakter", "Abcde12", "password minimal 8 karakter"},
		{"simbol_diizinkan", "Abcdef1!@#", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePassword(tt.password)
			if got != tt.want {
				t.Errorf("ValidatePassword(len=%d) = %q, want %q", len(tt.password), got, tt.want)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"valid", "Fatin", ""},
		{"empty", "", "nama wajib diisi"},
		{"spasi_saja", "   ", "nama wajib diisi"},
		{"satu_huruf", "A", "nama minimal 2 karakter"},
		{"dua_huruf", "Al", ""},
		{"tepat_100", ulangiString("a", 100), ""},
		{"lebih_100", ulangiString("a", 101), "nama maksimal 100 karakter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateName(tt.input)
			if got != tt.want {
				t.Errorf("ValidateName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateAssignableRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want string
	}{
		{"peserta", "peserta", ""},
		{"panitia", "panitia", ""},
		{"admin", "admin", ""},
		{"huruf_besar", "ADMIN", ""},
		{"dengan_spasi", "  panitia  ", ""},
		{"tidak_dikenal", "superadmin", "role harus peserta, panitia, atau admin"},
		{"empty", "", "role harus peserta, panitia, atau admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateAssignableRole(tt.role)
			if got != tt.want {
				t.Errorf("ValidateAssignableRole(%q) = %q, want %q", tt.role, got, tt.want)
			}
		})
	}
}

func TestCanChangeRole(t *testing.T) {
	tests := []struct {
		name     string
		actorID  int64
		targetID int64
		want     string
	}{
		{"user_lain", 1, 2, ""},
		{"diri_sendiri", 5, 5, "admin tidak dapat mengubah role dirinya sendiri"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanChangeRole(tt.actorID, tt.targetID)
			if got != tt.want {
				t.Errorf("CanChangeRole(%d, %d) = %q, want %q", tt.actorID, tt.targetID, got, tt.want)
			}
		})
	}
}

func TestValidatePasswordChange(t *testing.T) {
	tests := []struct {
		name string
		old  string
		new  string
		want string
	}{
		{"valid", "Lama1234", "Baru12345", ""},
		{"lama_kosong", "", "Baru12345", "password lama wajib diisi"},
		{"baru_kosong", "Lama1234", "", "password wajib diisi"},
		{"baru_terlalu_pendek", "Lama1234", "Ab1", "password minimal 8 karakter"},
		{"baru_tanpa_huruf_besar", "Lama1234", "baru12345", "password harus mengandung huruf besar"},
		{"baru_tanpa_angka", "Lama1234", "BaruBaruBaru", "password harus mengandung angka"},
		{"baru_sama_dengan_lama", "Sama12345", "Sama12345", "password baru tidak boleh sama dengan password lama"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePasswordChange(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("ValidatePasswordChange(%q, %q) = %q, want %q", tt.old, tt.new, got, tt.want)
			}
		})
	}
}
