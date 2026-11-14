package route

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"tiket-in-backend/app/repository"
	"tiket-in-backend/app/service"
	"tiket-in-backend/helper"
	mw "tiket-in-backend/middleware"
)

type Dependencies struct {
	Mongo         *mongo.Client
	EnableSwagger bool
}

func Register(app *fiber.App, pool *pgxpool.Pool, mongoClient *mongo.Client, deps Dependencies) {
	logger := slog.Default()
	jwtMgr := helper.NewJWTManager()

	userRepo := repository.NewUserRepository(pool)
	catRepo := repository.NewCategoryRepository(pool)
	eventRepo := repository.NewEventRepository(pool)
	ticketTypeRepo := repository.NewTicketTypeRepository(pool)

	authSvc := service.NewAuthService(userRepo, jwtMgr, logger)
	catSvc := service.NewCategoryService(catRepo, logger)
	eventSvc := service.NewEventService(eventRepo, ticketTypeRepo, catRepo, logger)

	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool, mongoClient))

	auth := api.Group("/auth", mw.RequireJSON)
	auth.Post("/register", authSvc.Register)
	auth.Post("/login", authSvc.Login)

	api.Get("/auth/me", mw.AuthRequired(jwtMgr), authSvc.GetMe)

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
	events.Post("/:id/tickets", mw.RequireJSON, mw.AuthRequired(jwtMgr), mw.RequireRole("panitia", "admin"), eventSvc.CreateTicket)
	events.Get("/:id/tickets", eventSvc.ListTickets)
}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if pool != nil {
			if err := pool.Ping(ctx); err != nil {
				return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
			}
		}

		if mongoClient != nil {
			if err := mongoClient.Ping(ctx, nil); err != nil {
				return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
			}
		}

		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", fiber.Map{
			"status":   "healthy",
			"postgres": "connected",
			"mongodb":  "connected",
		})
	}
}
