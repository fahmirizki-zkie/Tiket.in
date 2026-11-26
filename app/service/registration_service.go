package service

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"tiket-in-backend/app/model"
	"tiket-in-backend/app/repository"
	"tiket-in-backend/helper"
)

type RegistrationService struct {
	registrationRepo *repository.RegistrationRepository
	ticketTypeRepo   *repository.TicketTypeRepository
	eventRepo        *repository.EventRepository
	txManager        *repository.TxManager
	outboxRepo       *repository.OutboxRepository
	logger           *slog.Logger
}

func NewRegistrationService(
	registrationRepo *repository.RegistrationRepository,
	ticketTypeRepo *repository.TicketTypeRepository,
	eventRepo *repository.EventRepository,
	txManager *repository.TxManager,
	logger *slog.Logger,
) *RegistrationService {
	return &RegistrationService{
		registrationRepo: registrationRepo,
		ticketTypeRepo:   ticketTypeRepo,
		eventRepo:        eventRepo,
		txManager:        txManager,
		logger:           logger,
	}
}

// SetOutboxRepository menyematkan OutboxRepository untuk Transactional Outbox pattern
func (s *RegistrationService) SetOutboxRepository(outboxRepo *repository.OutboxRepository) {
	s.outboxRepo = outboxRepo
}

// Create mendaftarkan peserta ke event dan menghasilkan kode tiket unik.
// Semua langkah berjalan di satu transaksi dengan row lock pada jenis tiket,
// sehingga kuota aman dari race condition.
func (s *RegistrationService) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok || userID <= 0 {
		return helper.Unauthorized("sesi pengguna tidak valid")
	}

	var req model.CreateRegistrationRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}
	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var result model.Registration

	err := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Kunci baris jenis tiket
		tt, err := s.ticketTypeRepo.FindByIDForUpdate(ctx, tx, req.TicketTypeID)
		if err != nil {
			return err
		}
		if tt == nil {
			return helper.NotFound("jenis tiket tidak ditemukan")
		}

		// 2. Cek event
		event, err := s.eventRepo.FindByID(ctx, tt.EventID)
		if err != nil {
			return err
		}
		if event == nil {
			return helper.NotFound("event tidak ditemukan")
		}
		if err := ValidateEventOpen(event.Status); err != nil {
			return err
		}
		if err := ValidateEventNotEnded(event.EndAt, time.Now()); err != nil {
			return err
		}

		// 3. Satu peserta satu tiket per event
		exists, err := s.registrationRepo.ExistsByUserAndEvent(ctx, tx, userID, event.ID)
		if err != nil {
			return err
		}
		if err := ValidateNotRegistered(exists); err != nil {
			return err
		}

		// 4. Cek kuota
		if err := ValidateQuota(tt.Quota, tt.Sold); err != nil {
			return err
		}

		// 5. Buat kode tiket dan simpan pendaftaran
		code, err := helper.GenerateTicketCode()
		if err != nil {
			return err
		}

		reg := model.Registration{
			UserID:       userID,
			EventID:      event.ID,
			TicketTypeID: tt.ID,
			TicketCode:   code,
			Status:       "registered",
		}
		if err := s.registrationRepo.Create(ctx, tx, &reg); err != nil {
			return err
		}

		// 6. Tambah counter tiket terjual
		if err := s.ticketTypeRepo.IncrementSold(ctx, tx, tt.ID, 1); err != nil {
			return err
		}

		// 7. Simpan event Transactional Outbox (Modul 11 & 12)
		if s.outboxRepo != nil {
			outboxPayload := map[string]any{
				"registration_id": reg.ID,
				"user_id":         reg.UserID,
				"event_id":        reg.EventID,
				"ticket_code":     reg.TicketCode,
				"ticket_type_id":  reg.TicketTypeID,
				"registered_at":   time.Now(),
			}
			if err := s.outboxRepo.CreateWithTx(ctx, tx, "ticket.created", outboxPayload); err != nil {
				return err
			}
		}

		reg.EventTitle = event.Title
		reg.EventLocation = event.Location
		reg.EventStartAt = event.StartAt
		reg.TicketTypeName = tt.Name
		result = reg
		return nil
	})
	if err != nil {
		return s.mapError("registration create", err)
	}

	return helper.Created(c, "pendaftaran tiket berhasil", result, "/api/v1/registrations/me")
}

// GetMine mengambil daftar tiket milik peserta yang sedang login.
func (s *RegistrationService) GetMine(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok || userID <= 0 {
		return helper.Unauthorized("sesi pengguna tidak valid")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	status := strings.TrimSpace(c.Query("status"))

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	regs, err := s.registrationRepo.FindByUserID(ctx, userID, status, limit, (page-1)*limit)
	if err != nil {
		return s.mapError("registration get mine", err)
	}

	return helper.Success(c, fiber.StatusOK, "daftar tiket berhasil diambil", regs)
}

// mapError mengubah error aturan bisnis dan repository menjadi AppError HTTP.
func (s *RegistrationService) mapError(op string, err error) error {
	var appErr *helper.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, ErrQuotaFull):
		return helper.Conflict(err.Error())
	case errors.Is(err, ErrAlreadyRegistered),
		errors.Is(err, repository.ErrDuplicateRegistration):
		return helper.Conflict(ErrAlreadyRegistered.Error())
	case errors.Is(err, ErrEventNotOpen), errors.Is(err, ErrEventEnded):
		return helper.BadRequest(err.Error())
	case errors.Is(err, ErrNotRegistrationOwner):
		return helper.Forbidden(err.Error())
	case errors.Is(err, ErrAlreadyCheckedIn), errors.Is(err, ErrRegistrationCancelled):
		return helper.Conflict(err.Error())
	case errors.Is(err, ErrEventStarted):
		return helper.BadRequest(err.Error())
	}

	s.logger.Error(op+": gagal", slog.String("error", err.Error()))
	return helper.Internal(err)
}

// Cancel membatalkan tiket milik peserta dan mengembalikan kuota.
// Registrasi dikunci lebih dulu, lalu counter sold dikurangi dalam transaksi yang sama.
func (s *RegistrationService) Cancel(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok || userID <= 0 {
		return helper.Unauthorized("sesi pengguna tidak valid")
	}

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id pendaftaran tidak valid")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var result model.Registration

	err := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Kunci registrasi
		reg, err := s.registrationRepo.FindByIDForUpdate(ctx, tx, int64(id))
		if err != nil {
			return err
		}
		if reg == nil {
			return helper.NotFound("pendaftaran tidak ditemukan")
		}

		// 2. Hanya pemilik, dan statusnya masih registered
		if err := ValidateRegistrationOwner(reg.UserID, userID); err != nil {
			return err
		}
		if err := ValidateCancel(reg.Status); err != nil {
			return err
		}

		// 3. Event belum dimulai
		event, err := s.eventRepo.FindByID(ctx, reg.EventID)
		if err != nil {
			return err
		}
		if event == nil {
			return helper.NotFound("event tidak ditemukan")
		}
		if err := ValidateCancelWindow(event.StartAt, time.Now()); err != nil {
			return err
		}

		// 4. Ubah status dan kembalikan kuota
		if err := s.registrationRepo.UpdateStatus(ctx, tx, reg.ID, "cancelled"); err != nil {
			return err
		}
		if err := s.ticketTypeRepo.DecrementSold(ctx, tx, reg.TicketTypeID, 1); err != nil {
			return err
		}

		reg.Status = "cancelled"
		reg.EventTitle = event.Title
		result = *reg
		return nil
	})
	if err != nil {
		return s.mapError("registration cancel", err)
	}

	return helper.Success(c, fiber.StatusOK, "pendaftaran tiket berhasil dibatalkan", result)
}

