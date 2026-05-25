package inventory

import (
	"context"
	"log/slog"
	"time"
)

// Janitor periodically frees stock held by expired reservations. It exists because a MongoDB TTL
// index would only delete the expired document and leak the held stock; expiry has to actively
// return inventory to available, which is a business action, not a deletion.
type Janitor struct {
	svc      *Service
	interval time.Duration
	batch    int64
	log      *slog.Logger
}

func NewJanitor(svc *Service, interval time.Duration, log *slog.Logger) *Janitor {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Janitor{svc: svc, interval: interval, batch: 200, log: log}
}

// Run blocks until ctx is cancelled, sweeping expired reservations on each tick.
func (j *Janitor) Run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	j.log.Info("janitor started", "interval", j.interval.String())

	for {
		select {
		case <-ctx.Done():
			j.log.Info("janitor stopped")
			return
		case <-ticker.C:
			released, err := j.svc.ReleaseExpired(ctx, j.batch)
			if err != nil {
				j.log.Error("janitor sweep failed", "error", err)
				continue
			}
			if released > 0 {
				j.log.Info("janitor released expired reservations", "count", released)
			}
		}
	}
}
