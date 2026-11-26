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

type UserService struct {
	userRepo *repository.UserRepository
	logger   *slog.Logger
}

func NewUserService(
	userRepo *repository.UserRepository,
	logger *slog.Logger,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// List: GET /api/v1/users (admin). Query: page, limit, search, sort, order, role.
func (s *UserService) List(c *fiber.Ctx) error {
	q := helper.ParseListQuery(c)

	role := strings.TrimSpace(strings.ToLower(c.Query("role")))
	if role != "" {
		if msg := ValidateAssignableRole(role); msg != "" {
			return helper.Validation(map[string]string{"role": msg})
		}
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	users, total, err := s.userRepo.FindAll(ctx, q, role)
	if err != nil {
		s.logger.Error("user list: gagal ambil data", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	totalPages := total / q.Limit
	if total%q.Limit != 0 {
		totalPages++
	}

	return helper.SuccessList(c, "daftar user berhasil diambil", users, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// UpdateRole: PUT /api/v1/users/:id/role (admin).
func (s *UserService) UpdateRole(c *fiber.Ctx) error {
	targetID, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("ID user tidak valid")
	}

	actorID, ok := c.Locals("userID").(int64)
	if !ok {
		return helper.Unauthorized("sesi tidak valid")
	}

	var req model.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	req.Role = strings.TrimSpace(strings.ToLower(req.Role))

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}
	if msg := ValidateAssignableRole(req.Role); msg != "" {
		return helper.Validation(map[string]string{"role": msg})
	}
	if msg := CanChangeRole(actorID, int64(targetID)); msg != "" {
		return helper.Forbidden(msg)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.userRepo.UpdateRole(ctx, int64(targetID), req.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.NotFound("user tidak ditemukan")
		}
		s.logger.Error("user update role: gagal update", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	s.logger.Info("role user diubah",
		slog.Int64("actor_id", actorID),
		slog.Int64("target_id", user.ID),
		slog.String("role_baru", user.Role),
	)

	return helper.Success(c, fiber.StatusOK, "role user berhasil diperbarui", user)
}
