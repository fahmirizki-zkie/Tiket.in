package route

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"tiket-in-backend/app/repository"
	"tiket-in-backend/app/service"
	"tiket-in-backend/helper"
	mw "tiket-in-backend/middleware"
)

type Dependencies struct {
	EnableSwagger bool
	RedisClient   *redis.Client
}

func Register(app *fiber.App, pool *pgxpool.Pool, deps Dependencies) {
	logger := slog.Default()
	jwtMgr := helper.NewJWTManager()

	userRepo := repository.NewUserRepository(pool)
	catRepo := repository.NewCategoryRepository(pool)
	eventRepo := repository.NewEventRepository(pool)
	ticketTypeRepo := repository.NewTicketTypeRepository(pool)
	registrationRepo := repository.NewRegistrationRepository(pool)
	attendanceRepo := repository.NewAttendanceRepository(pool)
	outboxRepo := repository.NewOutboxRepository(pool)
	txManager := repository.NewTxManager(pool)

	authSvc := service.NewAuthService(userRepo, jwtMgr, logger)
	userSvc := service.NewUserService(userRepo, logger)
	catSvc := service.NewCategoryService(catRepo, logger)
	eventSvc := service.NewEventService(eventRepo, ticketTypeRepo, catRepo, logger)
	registrationSvc := service.NewRegistrationService(registrationRepo, ticketTypeRepo, eventRepo, txManager, logger)
	registrationSvc.SetOutboxRepository(outboxRepo)
	attendanceSvc := service.NewAttendanceService(attendanceRepo, registrationRepo, eventRepo, txManager, logger)

	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool, deps.RedisClient))

	auth := api.Group("/auth", mw.RequireJSON)
	auth.Post("/register", authSvc.Register)
	auth.Post("/login", authSvc.Login)

	api.Get("/auth/me", mw.AuthRequired(jwtMgr), authSvc.GetMe)
	api.Put("/auth/me", mw.RequireJSON, mw.AuthRequired(jwtMgr), authSvc.UpdateProfile)
	api.Put("/auth/me/password", mw.RequireJSON, mw.AuthRequired(jwtMgr), authSvc.ChangePassword)

	users := api.Group("/users")
	users.Get("/", mw.AuthRequired(jwtMgr), mw.RequireRole("admin"), userSvc.List)
	users.Put("/:id/role", mw.RequireJSON, mw.AuthRequired(jwtMgr), mw.RequireRole("admin"), userSvc.UpdateRole)

	categories := api.Group("/categories")
	categories.Get("/", catSvc.List)
	categories.Post("/", mw.RequireJSON, mw.AuthRequired(jwtMgr), mw.RequireRole("admin"), catSvc.Create)
	categories.Put("/:id", mw.RequireJSON, mw.AuthRequired(jwtMgr), mw.RequireRole("admin"), catSvc.Update)
	categories.Delete("/:id", mw.AuthRequired(jwtMgr), mw.RequireRole("admin"), catSvc.Delete)

	events := api.Group("/events")
	events.Get("/", eventSvc.List)
	events.Get("/:id", eventSvc.Get)
	events.Post("/", mw.RequireJSON, mw.AuthRequired(jwtMgr), mw.RequireRole("panitia", "admin"), eventSvc.Create)
	events.Put("/:id", mw.RequireJSON, mw.AuthRequired(jwtMgr), mw.RequireRole("panitia", "admin"), eventSvc.Update)
	events.Put("/:id/cancel", mw.AuthRequired(jwtMgr), mw.RequireRole("panitia", "admin"), eventSvc.Cancel)
	events.Post("/:id/tickets", mw.RequireJSON, mw.AuthRequired(jwtMgr), mw.RequireRole("panitia", "admin"), eventSvc.CreateTicket)
	events.Get("/:id/tickets", eventSvc.ListTickets)
	events.Put("/:id/tickets/:ticket_id", mw.RequireJSON, mw.AuthRequired(jwtMgr), mw.RequireRole("panitia", "admin"), eventSvc.UpdateTicket)
	events.Delete("/:id/tickets/:ticket_id", mw.AuthRequired(jwtMgr), mw.RequireRole("panitia", "admin"), eventSvc.DeleteTicket)

	registrations := api.Group("/registrations")
	registrations.Post("/", mw.RequireJSON, mw.AuthRequired(jwtMgr), registrationSvc.Create)
	registrations.Get("/me", mw.AuthRequired(jwtMgr), registrationSvc.GetMine)

	attendances := api.Group("/attendances")
	attendances.Post("/check-in", mw.RequireJSON, mw.AuthRequired(jwtMgr), mw.RequireRole("panitia", "admin"), attendanceSvc.CheckIn)
}

func healthCheck(pool *pgxpool.Pool, rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if pool != nil {
			if err := pool.Ping(ctx); err != nil {
				return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
			}
		}

		redisStatus := "disconnected"
		if rdb != nil {
			if err := rdb.Ping(ctx).Err(); err == nil {
				redisStatus = "connected"
			}
		}

		return helper.Success(c, fiber.StatusOK, "server berjalan dengan baik", fiber.Map{
			"status":   "healthy",
			"postgres": "connected",
			"redis":    redisStatus,
		})
	}
}
