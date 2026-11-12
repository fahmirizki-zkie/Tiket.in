package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"tiket-in-backend/config"
	"tiket-in-backend/database"
	"tiket-in-backend/route"
)

func main() {
	// 1. Memuat file .env dan logger terstruktur
	config.LoadEnv()
	logger := config.NewLogger()

	logger.Info("memulai inisialisasi server Tiket.in")

	// 2. Setup Context dengan penanganan sinyal OS (Ctrl+C atau SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Inisialisasi Koneksi Database PostgreSQL
	var pool *pgxpool.Pool
	pool, err := database.NewPool(ctx)
	if err != nil {
		logger.Warn("peringatan: koneksi PostgreSQL belum tersambung", slog.String("error", err.Error()))
	} else {
		defer pool.Close()
		logger.Info("koneksi PostgreSQL pool berhasil")
	}

	// 4. Inisialisasi Koneksi Database MongoDB
	var mongoClient *mongo.Client
	mongoClient, err = database.NewMongoClient(ctx)
	if err != nil {
		logger.Warn("peringatan: koneksi MongoDB belum tersambung", slog.String("error", err.Error()))
	} else {
		defer func() {
			disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = mongoClient.Disconnect(disconnectCtx)
		}()
		logger.Info("koneksi MongoDB client berhasil")
	}

	// 5. Inisialisasi Fiber App dengan Route Dependencies
	app := config.NewApp(logger, pool, mongoClient, route.Dependencies{
		Mongo:         mongoClient,
		EnableSwagger: config.GetEnv("SWAGGER_ENABLED", "false") == "true",
	})

	// 6. Jalankan Server HTTP di goroutine terpisah
	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server HTTP berhenti", slog.String("error", err.Error()))
		}
	}()
	logger.Info("server HTTP Tiket.in berjalan", slog.String("port", port))

	// 7. Menunggu sinyal penonaktifan (Graceful Shutdown)
	<-ctx.Done()
	logger.Info("sinyal penonaktifan diterima, memulai graceful shutdown...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("shutdown server HTTP tidak rapi", slog.String("error", err.Error()))
	}

	logger.Info("aplikasi Tiket.in berhasil berhenti dengan rapi")
}

