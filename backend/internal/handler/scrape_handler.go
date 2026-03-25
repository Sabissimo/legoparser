package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"lego-parser/internal/repository"
	"lego-parser/internal/scraper"
)

type ScrapeHandler struct {
	orchestrator *scraper.Orchestrator
	scrapeRepo   *repository.ScrapeRepo
	logger       *slog.Logger
}

func NewScrapeHandler(orchestrator *scraper.Orchestrator, scrapeRepo *repository.ScrapeRepo, logger *slog.Logger) *ScrapeHandler {
	return &ScrapeHandler{
		orchestrator: orchestrator,
		scrapeRepo:   scrapeRepo,
		logger:       logger,
	}
}

type runRequest struct {
	Source string `json:"source"` // "all", "biblusi", "wishlist", etc.
}

func (h *ScrapeHandler) Run(w http.ResponseWriter, r *http.Request) {
	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Source == "" {
		req.Source = "all"
	}

	if h.orchestrator.IsRunning() {
		writeError(w, http.StatusConflict, "scraper is already running")
		return
	}

	// Run in background goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()

		var err error
		if req.Source == "all" {
			err = h.orchestrator.RunAll(ctx, "manual")
		} else {
			err = h.orchestrator.RunSource(ctx, req.Source, "manual")
		}

		if err != nil {
			h.logger.Error("scrape run failed", "source", req.Source, "error", err)
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "scrape started",
		"source":  req.Source,
	})
}

func (h *ScrapeHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := h.scrapeRepo.List(r.Context(), 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (h *ScrapeHandler) Status(w http.ResponseWriter, r *http.Request) {
	running, err := h.scrapeRepo.GetRunning(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"is_running": h.orchestrator.IsRunning(),
		"runs":       running,
	})
}
