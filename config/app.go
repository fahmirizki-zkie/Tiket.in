package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"tiket-in-backend/app/model"
	"tiket-in-backend/helper"
	"tiket-in-backend/middleware"
	"tiket-in-backend/route"
)

// NewApp merakit instance Fiber, middleware, dan mendaftarkan route
func NewApp(
	logger *slog.Logger,
	pool *pgxpool.Pool,
	mongoClient *mongo.Client,
	deps route.Dependencies,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Tiket.in Backend API"),
		ErrorHandler: newErrorHandler(logger),
		BodyLimit:    2 * 1024 * 1024, // 2MB
	})

	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", "*"))
	route.Register(app, pool, mongoClient, deps)

	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	return app
}

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID, _ := c.Locals("requestid").(string)

		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):
			// Error yang sudah terencana

		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}

		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status:  fiberErr.Code,
					Code:    "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		if appErr.Status >= fiber.StatusInternalServerError {
			causeMsg := ""
			if cause := appErr.Unwrap(); cause != nil {
				causeMsg = cause.Error()
			} else {
				causeMsg = appErr.Error()
			}

			logger.Error("request_failed",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", causeMsg),
			)
		} else {
			logger.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
			)
		}

		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success: false,
			Code:    appErr.Code,
			Message: appErr.Message,
			Fields:  appErr.Fields,
		})
	}
}

