package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"tiket-in-backend/helper"
)

// Dependencies berisi dependensi layanan yang dibutuhkan route
type Dependencies struct {
	Mongo         *mongo.Client
	EnableSwagger bool
}

// Register mendaftarkan seluruh endpoint route aplikasi
func Register(app *fiber.App, pool *pgxpool.Pool, mongoClient *mongo.Client, deps Dependencies) {
	api := app.Group("/api/v1")

	// Health Check publik yang memeriksa kesiapan server, PostgreSQL, dan MongoDB
	api.Get("/health", healthCheck(pool, mongoClient))
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

