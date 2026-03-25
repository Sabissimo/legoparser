package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"lego-parser/internal/scraper"
)

type Scheduler struct {
	cron         *cron.Cron
	orchestrator *scraper.Orchestrator
	logger       *slog.Logger
}

func New(orchestrator *scraper.Orchestrator, schedule string, logger *slog.Logger) (*Scheduler, error) {
	c := cron.New(cron.WithSeconds())

	s := &Scheduler{
		cron:         c,
		orchestrator: orchestrator,
		logger:       logger,
	}

	_, err := c.AddFunc(schedule, func() {
		logger.Info("scheduled scrape starting")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()

		if err := orchestrator.RunAll(ctx, "scheduled"); err != nil {
			logger.Error("scheduled scrape failed", "error", err)
		}
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
	s.logger.Info("scheduler started")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("scheduler stopped")
}
