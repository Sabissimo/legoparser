package router

import (
	"net/http"

	"lego-parser/internal/handler"
)

func New(productHandler *handler.ProductHandler, scrapeHandler *handler.ScrapeHandler, exportHandler *handler.ExportHandler) http.Handler {
	mux := http.NewServeMux()

	// CORS middleware
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Products
	mux.HandleFunc("GET /api/products", productHandler.List)
	mux.HandleFunc("GET /api/products/comparison", productHandler.Comparison)
	mux.HandleFunc("GET /api/products/{id}", productHandler.GetByID)

	// Scraper
	mux.HandleFunc("POST /api/scrape/run", scrapeHandler.Run)
	mux.HandleFunc("GET /api/scrape/runs", scrapeHandler.ListRuns)
	mux.HandleFunc("GET /api/scrape/status", scrapeHandler.Status)

	// Export (Excel)
	mux.HandleFunc("GET /api/export/products", exportHandler.ExportProducts)
	mux.HandleFunc("GET /api/export/comparison", exportHandler.ExportComparison)

	// Export (CSV/QVD - Qlik compatible, tab-delimited UTF-8)
	mux.HandleFunc("GET /api/export/products/csv", exportHandler.ExportProductsCSV)
	mux.HandleFunc("GET /api/export/comparison/csv", exportHandler.ExportComparisonCSV)

	return corsHandler(mux)
}
