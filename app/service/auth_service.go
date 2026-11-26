package service

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"tiket-in-backend/app/model"
	"tiket-in-backend/app/repository"
	"tiket-in-backend/helper"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtMgr   *helper.JWTManager
	logger   *slog.Logger
}

func NewAuthService(
	userRepo *repository.UserRepository,
	jwtMgr *helper.JWTManager,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtMgr:   jwtMgr,
		logger:   logger,
	}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Role = strings.TrimSpace(strings.ToLower(req.Role))

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	if msg := ValidateEmail(req.Email); msg != "" {
		return helper.Validation(map[string]string{"email": msg})
	}
	if msg := ValidatePassword(req.Password); msg != "" {
		return helper.Validation(map[string]string{"password": msg})
	}
	if msg := ValidateRole(req.Role); msg != "" {
		return helper.Validation(map[string]string{"role": msg})
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Error("register: gagal cek email", slog.String("error", err.Error()))
		return helper.Internal(err)
	}
	if existing != nil {
		return helper.Conflict("email sudah terdaftar")
	}

	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("register: gagal hash password", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	user := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return helper.Conflict("email sudah terdaftar")
		}
		s.logger.Error("register: gagal simpan user", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	token, err := s.jwtMgr.GenerateToken(user.ID, user.Role)
	if err != nil {
		s.logger.Error("register: gagal buat token", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.Created(c, "registrasi berhasil", model.AuthResponse{
		Token: token,
		User:  *user,
	}, "")
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.Unauthorized("email atau password salah")
		}
		s.logger.Error("login: gagal cari user", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	if !helper.VerifyPassword(user.PasswordHash, req.Password) {
		return helper.Unauthorized("email atau password salah")
	}

	token, err := s.jwtMgr.GenerateToken(user.ID, user.Role)
	if err != nil {
		s.logger.Error("login: gagal buat token", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", model.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (s *AuthService) GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return helper.Unauthorized("sesi tidak valid")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.NotFound("user tidak ditemukan")
		}
		s.logger.Error("getme: gagal cari user", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "data user berhasil diambil", user)
}

// UpdateProfile: PUT /api/v1/auth/me (nama dan email milik sendiri).
func (s *AuthService) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return helper.Unauthorized("sesi tidak valid")
	}

	var req model.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}
	if msg := ValidateName(req.Name); msg != "" {
		return helper.Validation(map[string]string{"name": msg})
	}
	if msg := ValidateEmail(req.Email); msg != "" {
		return helper.Validation(map[string]string{"email": msg})
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	taken, err := s.userRepo.ExistsByEmail(ctx, req.Email, userID)
	if err != nil {
		s.logger.Error("update profil: gagal cek email", slog.String("error", err.Error()))
		return helper.Internal(err)
	}
	if taken {
		return helper.Conflict("email sudah digunakan user lain")
	}

	user, err := s.userRepo.UpdateProfile(ctx, userID, req.Name, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.NotFound("user tidak ditemukan")
		}
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return helper.Conflict("email sudah digunakan user lain")
		}
		s.logger.Error("update profil: gagal simpan", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diperbarui", user)
}

// ChangePassword: PUT /api/v1/auth/me/password.
func (s *AuthService) ChangePassword(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return helper.Unauthorized("sesi tidak valid")
	}

	var req model.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}
	if msg := ValidatePasswordChange(req.OldPassword, req.NewPassword); msg != "" {
		return helper.Validation(map[string]string{"new_password": msg})
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.NotFound("user tidak ditemukan")
		}
		s.logger.Error("ganti password: gagal cari user", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	if !helper.VerifyPassword(user.PasswordHash, req.OldPassword) {
		return helper.Unauthorized("password lama salah")
	}

	hash, err := helper.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error("ganti password: gagal hash", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.NotFound("user tidak ditemukan")
		}
		s.logger.Error("ganti password: gagal simpan", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "password berhasil diganti", nil)
}
