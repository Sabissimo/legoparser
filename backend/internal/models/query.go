package models

import (
	"net/http"
	"strconv"
)

type QueryParams struct {
	Page      int
	PerPage   int
	SortBy    string
	SortOrder string
	Search   string
	Source   string
	MinPrice  *float64
	MaxPrice  *float64
	InStock   *bool
}

func ParseQueryParams(r *http.Request) QueryParams {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	sortBy := q.Get("sort_by")
	if sortBy == "" {
		sortBy = "name_ka"
	}

	sortOrder := q.Get("sort_order")
	if sortOrder != "desc" {
		sortOrder = "asc"
	}

	params := QueryParams{
		Page:      page,
		PerPage:   perPage,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Search: q.Get("search"),
		Source: q.Get("source"),
	}

	if v := q.Get("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			params.MinPrice = &f
		}
	}
	if v := q.Get("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			params.MaxPrice = &f
		}
	}
	if v := q.Get("in_stock"); v != "" {
		b := v == "true"
		params.InStock = &b
	}

	return params
}
