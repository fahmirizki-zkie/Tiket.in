package service

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"tiket-in-backend/app/model"
	"tiket-in-backend/app/repository"
	"tiket-in-backend/helper"
)

type AttendanceService struct {
	attendanceRepo   *repository.AttendanceRepository
	registrationRepo *repository.RegistrationRepository
	eventRepo        *repository.EventRepository
	txManager        *repository.TxManager
	logger           *slog.Logger
}

func NewAttendanceService(
	attendanceRepo *repository.AttendanceRepository,
	registrationRepo *repository.RegistrationRepository,
	eventRepo *repository.EventRepository,
	txManager *repository.TxManager,
	logger *slog.Logger,
) *AttendanceService {
	return &AttendanceService{
		attendanceRepo:   attendanceRepo,
		registrationRepo: registrationRepo,
		eventRepo:        eventRepo,
		txManager:        txManager,
		logger:           logger,
	}
}

// CheckIn memvalidasi kode tiket lalu mencatat kehadiran peserta.
// Hanya panitia penyelenggara event tersebut (atau admin) yang boleh check-in.
func (s *AttendanceService) CheckIn(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok || userID <= 0 {
		return helper.Unauthorized("sesi pengguna tidak valid")
	}
	userRole, _ := c.Locals("userRole").(string)

	var req model.CheckInRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}
	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var result model.Attendance

	err := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Kunci registrasi berdasarkan kode tiket
		reg, err := s.registrationRepo.FindByTicketCodeForUpdate(ctx, tx, req.TicketCode)
		if err != nil {
			return err
		}
		if reg == nil {
			return helper.NotFound("kode tiket tidak ditemukan")
		}

		// 2. Pastikan panitia berhak atas event ini
		event, err := s.eventRepo.FindByID(ctx, reg.EventID)
		if err != nil {
			return err
		}
		if event == nil {
			return helper.NotFound("event tidak ditemukan")
		}
		if msg := ValidateEventOwnership(event.OrganizerID, userID, userRole); msg != "" {
			return helper.Forbidden(msg)
		}

		// 3. Tiket belum dipakai dan tidak dibatalkan
		if err := ValidateCheckIn(reg.Status); err != nil {
			return err
		}

		// 4. Catat kehadiran dan ubah status registrasi
		att := model.Attendance{
			RegistrationID: reg.ID,
			CheckedInBy:    userID,
		}
		if err := s.attendanceRepo.Create(ctx, tx, &att); err != nil {
			return err
		}
		if err := s.registrationRepo.UpdateStatus(ctx, tx, reg.ID, "checked_in"); err != nil {
			return err
		}

		att.TicketCode = reg.TicketCode
		att.EventTitle = event.Title
		result = att
		return nil
	})
	if err != nil {
		return s.mapError("attendance check-in", err)
	}

	return helper.Created(c, "check-in berhasil", result, "")
}

func (s *AttendanceService) mapError(op string, err error) error {
	var appErr *helper.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, ErrAlreadyCheckedIn),
		errors.Is(err, repository.ErrDuplicateAttendance):
		return helper.Conflict(ErrAlreadyCheckedIn.Error())
	case errors.Is(err, ErrRegistrationCancelled):
		return helper.BadRequest(err.Error())
	}

	s.logger.Error(op+": gagal", slog.String("error", err.Error()))
	return helper.Internal(err)
}