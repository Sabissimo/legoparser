package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"lego-parser/internal/models"
	"lego-parser/internal/repository"
)

type ProductHandler struct {
	productRepo *repository.ProductRepo
	priceRepo   *repository.PriceRepo
}

func NewProductHandler(productRepo *repository.ProductRepo, priceRepo *repository.PriceRepo) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
		priceRepo:   priceRepo,
	}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	params := models.ParseQueryParams(r)
	products, total, err := h.productRepo.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":        products,
		"total":       total,
		"page":        params.Page,
		"per_page":    params.PerPage,
		"total_pages": totalPages,
	})
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	product, err := h.productRepo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if product == nil {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	// Attach latest prices
	prices, err := h.priceRepo.GetLatestByProductID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	product.LatestPrices = prices

	writeJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Comparison(w http.ResponseWriter, r *http.Request) {
	params := models.ParseQueryParams(r)
	results, total, err := h.productRepo.GetComparison(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":        results,
		"total":       total,
		"page":        params.Page,
		"per_page":    params.PerPage,
		"total_pages": totalPages,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
