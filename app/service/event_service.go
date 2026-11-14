package service

import (
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/gofiber/fiber/v2"

	"tiket-in-backend/app/model"
	"tiket-in-backend/app/repository"
	"tiket-in-backend/helper"
)

type EventService struct {
	eventRepo      *repository.EventRepository
	ticketTypeRepo *repository.TicketTypeRepository
	catRepo        *repository.CategoryRepository
	logger         *slog.Logger
}

func NewEventService(
	eventRepo *repository.EventRepository,
	ticketTypeRepo *repository.TicketTypeRepository,
	catRepo *repository.CategoryRepository,
	logger *slog.Logger,
) *EventService {
	return &EventService{
		eventRepo:      eventRepo,
		ticketTypeRepo: ticketTypeRepo,
		catRepo:        catRepo,
		logger:         logger,
	}
}

// List mengambil katalog publik event dengan pencarian, kategori, status, dan pagination
func (s *EventService) List(c *fiber.Ctx) error {
	q := model.EventListQuery{
		Page:       c.QueryInt("page", 1),
		Limit:      c.QueryInt("limit", 10),
		Search:     strings.TrimSpace(c.Query("search")),
		CategoryID: int64(c.QueryInt("category_id", 0)),
		Status:     strings.TrimSpace(c.Query("status")),
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	events, total, err := s.eventRepo.FindAll(ctx, q)
	if err != nil {
		s.logger.Error("event list: gagal query events", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	if events == nil {
		events = []model.Event{}
	}

	totalPages := 0
	if q.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(q.Limit)))
	}

	meta := &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return helper.SuccessList(c, "daftar event berhasil diambil", events, meta)
}

// Get mengambil detail satu event beserta daftar tiket dan informasi penyelenggara
func (s *EventService) Get(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id event tidak valid")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	event, err := s.eventRepo.FindByID(ctx, int64(id))
	if err != nil {
		s.logger.Error("event get: gagal query event", slog.String("error", err.Error()))
		return helper.Internal(err)
	}
	if event == nil {
		return helper.NotFound("event tidak ditemukan")
	}

	tickets, err := s.ticketTypeRepo.FindByEventID(ctx, event.ID)
	if err != nil {
		s.logger.Error("event get: gagal query ticket types", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	detail := model.EventDetailResponse{
		Event:       *event,
		TicketTypes: tickets,
	}

	return helper.Success(c, fiber.StatusOK, "detail event berhasil diambil", detail)
}

// Create membuat event baru oleh panitia atau admin
func (s *EventService) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok || userID <= 0 {
		return helper.Unauthorized("sesi pengguna tidak valid")
	}

	var req model.CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Location = strings.TrimSpace(req.Location)

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	if msg := ValidateEventDates(req.StartAt, req.EndAt); msg != "" {
		return helper.BadRequest(msg)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if s.catRepo != nil {
		cat, err := s.catRepo.FindByID(ctx, int(req.CategoryID))
		if err != nil || cat == nil {
			return helper.BadRequest("kategori yang dipilih tidak valid atau tidak ditemukan")
		}
	}

	event := model.Event{
		OrganizerID: userID,
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		Status:      "published",
	}

	if err := s.eventRepo.Create(ctx, &event); err != nil {
		s.logger.Error("event create: gagal simpan", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.Created(c, "event berhasil dibuat", event, fmt.Sprintf("/api/v1/events/%d", event.ID))
}

// Update memperbarui data event oleh panitia penyelenggara atau admin
func (s *EventService) Update(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id event tidak valid")
	}

	userID, _ := c.Locals("userID").(int64)
	userRole, _ := c.Locals("userRole").(string)

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	existing, err := s.eventRepo.FindByID(ctx, int64(id))
	if err != nil {
		s.logger.Error("event update: gagal query", slog.String("error", err.Error()))
		return helper.Internal(err)
	}
	if existing == nil {
		return helper.NotFound("event tidak ditemukan")
	}

	if msg := ValidateEventOwnership(existing.OrganizerID, userID, userRole); msg != "" {
		return helper.Forbidden(msg)
	}

	if msg := CanModifyEvent(existing.Status); msg != "" {
		return helper.BadRequest(msg)
	}

	var req model.UpdateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	if req.Title != "" {
		existing.Title = strings.TrimSpace(req.Title)
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.Location != "" {
		existing.Location = strings.TrimSpace(req.Location)
	}
	if req.CategoryID > 0 {
		if s.catRepo != nil {
			cat, err := s.catRepo.FindByID(ctx, int(req.CategoryID))
			if err != nil || cat == nil {
				return helper.BadRequest("kategori tidak ditemukan")
			}
		}
		existing.CategoryID = req.CategoryID
	}
	if req.StartAt != nil {
		existing.StartAt = *req.StartAt
	}
	if req.EndAt != nil {
		existing.EndAt = *req.EndAt
	}
	if req.Status != "" {
		existing.Status = req.Status
	}

	if msg := ValidateEventDates(existing.StartAt, existing.EndAt); msg != "" {
		return helper.BadRequest(msg)
	}

	if err := s.eventRepo.Update(ctx, existing); err != nil {
		s.logger.Error("event update: gagal simpan", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "event berhasil diperbarui", existing)
}

// CreateTicket menambahkan jenis tiket pada event oleh panitia penyelenggara
func (s *EventService) CreateTicket(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id event tidak valid")
	}

	userID, _ := c.Locals("userID").(int64)
	userRole, _ := c.Locals("userRole").(string)

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	event, err := s.eventRepo.FindByID(ctx, int64(id))
	if err != nil {
		s.logger.Error("create ticket: gagal query event", slog.String("error", err.Error()))
		return helper.Internal(err)
	}
	if event == nil {
		return helper.NotFound("event tidak ditemukan")
	}

	if msg := ValidateEventOwnership(event.OrganizerID, userID, userRole); msg != "" {
		return helper.Forbidden(msg)
	}

	if msg := CanModifyEvent(event.Status); msg != "" {
		return helper.BadRequest(msg)
	}

	var req model.CreateTicketTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	if msg := ValidateTicketPriceAndQuota(req.Price, req.Quota); msg != "" {
		return helper.BadRequest(msg)
	}

	existingTicket, err := s.ticketTypeRepo.FindByNameAndEventID(ctx, event.ID, req.Name)
	if err != nil {
		s.logger.Error("create ticket: gagal cek nama", slog.String("error", err.Error()))
		return helper.Internal(err)
	}
	if existingTicket != nil {
		return helper.Conflict("jenis tiket dengan nama tersebut sudah ada pada event ini")
	}

	ticket := model.TicketType{
		EventID: event.ID,
		Name:    req.Name,
		Price:   req.Price,
		Quota:   req.Quota,
		Sold:    0,
	}

	if err := s.ticketTypeRepo.Create(ctx, &ticket); err != nil {
		s.logger.Error("create ticket: gagal simpan", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	ticket.AvailableSeats = ticket.Quota - ticket.Sold

	return helper.Created(c, "jenis tiket berhasil ditambahkan", ticket, fmt.Sprintf("/api/v1/events/%d/tickets", event.ID))
}

// ListTickets mengambil daftar jenis tiket yang tersedia di suatu event
func (s *EventService) ListTickets(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id event tidak valid")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	event, err := s.eventRepo.FindByID(ctx, int64(id))
	if err != nil {
		s.logger.Error("list tickets: gagal query event", slog.String("error", err.Error()))
		return helper.Internal(err)
	}
	if event == nil {
		return helper.NotFound("event tidak ditemukan")
	}

	tickets, err := s.ticketTypeRepo.FindByEventID(ctx, int64(id))
	if err != nil {
		s.logger.Error("list tickets: gagal query ticket types", slog.String("error", err.Error()))
		return helper.Internal(err)
	}

	if tickets == nil {
		tickets = []model.TicketType{}
	}

	return helper.Success(c, fiber.StatusOK, "daftar jenis tiket berhasil diambil", tickets)
}
