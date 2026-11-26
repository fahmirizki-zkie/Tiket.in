package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"tiket-in-backend/app/repository"
	"tiket-in-backend/app/service"
	"tiket-in-backend/config"
	"tiket-in-backend/database"
	"tiket-in-backend/route"
	"tiket-in-backend/worker"
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

	// 3b. Inisialisasi Redis Client & Outbox Relay Worker (Modul 11 & 12)
	rdb, err := database.NewRedisClient(ctx)
	if err != nil {
		logger.Warn("peringatan: koneksi Redis belum tersambung (outbox relay standby)", slog.String("error", err.Error()))
	} else {
		defer rdb.Close()
		logger.Info("koneksi Redis client berhasil tersambung")
	}

	// 3c. Seed admin pertama dari .env (hanya jika belum ada admin)
	if pool != nil {
		service.SeedAdmin(ctx, repository.NewUserRepository(pool), logger)
	}

	outboxRepo := repository.NewOutboxRepository(pool)
	relay := worker.NewOutboxRelay(outboxRepo, rdb, logger)
	go relay.Start(ctx)

	// 4. Inisialisasi Fiber App dengan Route Dependencies
	app := config.NewApp(logger, pool, route.Dependencies{
		EnableSwagger: config.GetEnv("SWAGGER_ENABLED", "false") == "true",
		RedisClient:   rdb,
	})

	// 5. Jalankan Server HTTP di goroutine terpisah
	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server HTTP berhenti", slog.String("error", err.Error()))
		}
	}()
	logger.Info("server HTTP Tiket.in berjalan", slog.String("port", port))

	// 6. Menunggu sinyal penonaktifan (Graceful Shutdown)
	<-ctx.Done()
	logger.Info("sinyal shutdown diterima, menghentikan server...")

	// 7. Beri waktu toleransi pembersihan resource
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("gagal mematikan server dengan aman", slog.String("error", err.Error()))
	} else {
		logger.Info("server berhasil dimatikan dengan aman")
	}
}
