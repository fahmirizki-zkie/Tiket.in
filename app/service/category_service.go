package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"tiket-in-backend/app/model"
	"tiket-in-backend/app/repository"
	"tiket-in-backend/helper"
)

type CategoryService struct {
	catRepo *repository.CategoryRepository
	logger  *slog.Logger
}

func NewCategoryService(
	catRepo *repository.CategoryRepository,
	logger *slog.Logger,
) *CategoryService {
	return &CategoryService{
		catRepo: catRepo,
		logger:  logger,
	}
}

func (s *CategoryService) Create(c *fiber.Ctx) error {
	var req model.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	req.Name = strings.TrimSpace(req.Name)

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	exists, err := s.catRepo.ExistsByName(ctx, req.Name, 0)
	if err != nil {
		s.logger.Error("category create: gagal cek nama", slog.String("error", err.Error()))
		return helper.Internal(err)
	}
	if exists {
		return helper.Conflict("nama kategori sudah digunakan")
	}

	cat := &model.Category{Name: req.Name}
	if err := s.catRepo.Create(ctx, cat); err != nil {
		s.logger.Error("category create: gagal simpan", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.Created(c, "kategori berhasil dibuat", cat, fmt.Sprintf("/api/v1/categories/%d", cat.ID))
}

func (s *CategoryService) List(c *fiber.Ctx) error {
	q := helper.ParseListQuery(c)

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	categories, total, err := s.catRepo.FindAll(ctx, q)
	if err != nil {
		s.logger.Error("category list: gagal ambil data", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	if categories == nil {
		categories = []model.Category{}
	}

	totalPages := total / q.Limit
	if total%q.Limit != 0 {
		totalPages++
	}

	return helper.SuccessList(c, "daftar kategori berhasil diambil", categories, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (s *CategoryService) Update(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("ID kategori tidak valid")
	}

	var req model.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	req.Name = strings.TrimSpace(req.Name)

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	exists, err := s.catRepo.ExistsByName(ctx, req.Name, id)
	if err != nil {
		s.logger.Error("category update: gagal cek nama", slog.String("error", err.Error()))
		return helper.Internal(err)
	}
	if exists {
		return helper.Conflict("nama kategori sudah digunakan")
	}

	cat, err := s.catRepo.Update(ctx, id, req.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.NotFound("kategori tidak ditemukan")
		}
		s.logger.Error("category update: gagal update", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "kategori berhasil diperbarui", cat)
}

func (s *CategoryService) Delete(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("ID kategori tidak valid")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.catRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.NotFound("kategori tidak ditemukan")
		}
		s.logger.Error("category delete: gagal hapus", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}
