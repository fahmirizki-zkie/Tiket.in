package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"tiket-in-backend/app/repository"
)

type OutboxRelay struct {
	outboxRepo *repository.OutboxRepository
	rdb        *redis.Client
	logger     *slog.Logger
	interval   time.Duration
}

func NewOutboxRelay(outboxRepo *repository.OutboxRepository, rdb *redis.Client, logger *slog.Logger) *OutboxRelay {
	return &OutboxRelay{
		outboxRepo: outboxRepo,
		rdb:        rdb,
		logger:     logger,
		interval:   3 * time.Second,
	}
}

func (w *OutboxRelay) Start(ctx context.Context) {
	if w.outboxRepo == nil || w.rdb == nil {
		w.logger.Info("outbox relay tidak aktif (repository atau Redis client belum tersambung)")
		return
	}

	w.logger.Info("outbox relay worker aktif berjalan di background", slog.Duration("interval", w.interval))
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("outbox relay worker dihentikan dengan aman")
			return
		case <-ticker.C:
			w.processPending(ctx)
		}
	}
}

func (w *OutboxRelay) processPending(ctx context.Context) {
	items, err := w.outboxRepo.FetchPending(ctx, 20)
	if err != nil {
		w.logger.Warn("outbox relay: gagal membaca pending outbox", slog.String("error", err.Error()))
		return
	}

	for _, item := range items {
		// Publikasi event ke channel Redis pub/sub dan list antrean
		err := w.rdb.Publish(ctx, item.EventType, string(item.Payload)).Err()
		if err == nil {
			_ = w.rdb.LPush(ctx, "events:outbox", string(item.Payload)).Err()
			if markErr := w.outboxRepo.MarkCompleted(ctx, item.ID); markErr != nil {
				w.logger.Warn("outbox relay: gagal update status outbox", slog.Int64("id", item.ID), slog.String("error", markErr.Error()))
			} else {
				w.logger.Info("outbox event diteruskan ke Redis", slog.String("event_type", item.EventType), slog.Int64("id", item.ID))
			}
		} else {
			w.logger.Warn("outbox relay: gagal publish event ke Redis", slog.Int64("id", item.ID), slog.String("error", err.Error()))
			_ = w.outboxRepo.IncrementRetry(ctx, item.ID)
		}
	}
}
