package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lego-parser/internal/config"
	"lego-parser/internal/database"
	"lego-parser/internal/handler"
	"lego-parser/internal/repository"
	"lego-parser/internal/router"
	"lego-parser/internal/scheduler"
	"lego-parser/internal/scraper"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Load()

	// Run migrations
	logger.Info("running database migrations")
	if err := database.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Connect to database
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("connected to database")

	// Initialize repositories
	productRepo := repository.NewProductRepo(pool)
	priceRepo := repository.NewPriceRepo(pool)
	scrapeRepo := repository.NewScrapeRepo(pool)

	// Initialize orchestrator
	orchestrator := scraper.NewOrchestrator(productRepo, priceRepo, scrapeRepo, logger)

	// Register scrapers
	orchestrator.Register(scraper.NewBiblusiXSScraper(cfg, logger))
	orchestrator.Register(scraper.NewBiblusiPepelaScraper(cfg, logger))
	orchestrator.Register(scraper.NewWishlistScraper(cfg, logger))
	orchestrator.Register(scraper.NewPiccolaToysScraper(cfg, logger))
	orchestrator.Register(scraper.NewKubikiScraper(cfg, logger))
	orchestrator.Register(scraper.NewWoltXSScraper(cfg, logger))
	orchestrator.Register(scraper.NewWoltPepelaScraper(cfg, logger))
	orchestrator.Register(scraper.NewGlovoXSScraper(cfg, logger))
	orchestrator.Register(scraper.NewGlovoPepelaScraper(cfg, logger))

	// Initialize handlers
	productHandler := handler.NewProductHandler(productRepo, priceRepo)
	scrapeHandler := handler.NewScrapeHandler(orchestrator, scrapeRepo, logger)
	exportHandler := handler.NewExportHandler(productRepo)

	// Initialize router
	mux := router.New(productHandler, scrapeHandler, exportHandler)

	// Initialize scheduler
	sched, err := scheduler.New(orchestrator, cfg.CronSchedule, logger)
	if err != nil {
		logger.Error("failed to create scheduler", "error", err)
		os.Exit(1)
	}
	sched.Start()
	defer sched.Stop()

	// Start HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
