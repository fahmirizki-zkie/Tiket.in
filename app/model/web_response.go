package model

// WebResponse adalah amplop baku untuk semua respon sukses dan gagal aplikasi Tiket.in
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Meta memuat metadata pagination
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListQuery menampung parameter query string pagination dan filter
type ListQuery struct {
	Page   int
	Limit  int
	Search string
	Sort   string
	Order  string
}

// ErrorResponse memuat struktur JSON error terstandar
type ErrorResponse struct {
	Success bool              `json:"success"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}
