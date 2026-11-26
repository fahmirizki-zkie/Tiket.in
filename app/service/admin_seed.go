package service

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"tiket-in-backend/app/model"
	"tiket-in-backend/app/repository"
	"tiket-in-backend/helper"
)

// SeedAdmin membuat admin pertama dari ADMIN_EMAIL/ADMIN_PASSWORD/ADMIN_NAME di .env
// hanya bila belum ada admin sama sekali. Setelah itu admin lain cukup dibuat
// lewat PUT /users/:id/role (tanpa SQL manual).
func SeedAdmin(ctx context.Context, userRepo *repository.UserRepository, logger *slog.Logger) {
	email := strings.TrimSpace(strings.ToLower(os.Getenv("ADMIN_EMAIL")))
	password := os.Getenv("ADMIN_PASSWORD")
	name := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	if email == "" || password == "" {
		return
	}
	if name == "" {
		name = "Administrator"
	}

	if msg := ValidateEmail(email); msg != "" {
		logger.Warn("seed admin dilewati: ADMIN_EMAIL tidak valid", slog.String("alasan", msg))
		return
	}
	if msg := ValidatePassword(password); msg != "" {
		logger.Warn("seed admin dilewati: ADMIN_PASSWORD tidak valid", slog.String("alasan", msg))
		return
	}

	count, err := userRepo.CountByRole(ctx, "admin")
	if err != nil {
		logger.Warn("seed admin dilewati: gagal cek admin", slog.String("error", err.Error()))
		return
	}
	if count > 0 {
		return
	}

	hash, err := helper.HashPassword(password)
	if err != nil {
		logger.Warn("seed admin gagal: hash password", slog.String("error", err.Error()))
		return
	}

	admin := &model.User{Name: name, Email: email, PasswordHash: hash, Role: "admin"}
	if err := userRepo.Create(ctx, admin); err != nil {
		logger.Warn("seed admin gagal: simpan user", slog.String("error", err.Error()))
		return
	}
	logger.Info("admin pertama berhasil dibuat", slog.String("email", email))
}
