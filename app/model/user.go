package model

import "time"

type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Role     string `json:"role"     validate:"required,oneof=peserta panitia"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UpdateRoleRequest dipakai admin untuk mengubah role user (PUT /users/:id/role)
type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=peserta panitia admin"`
}

// UpdateProfileRequest dipakai user untuk mengubah profilnya sendiri (PUT /auth/me)
type UpdateProfileRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email,max=100"`
}

// ChangePasswordRequest dipakai user untuk mengganti password (PUT /auth/me/password)
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}
