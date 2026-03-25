package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"lego-parser/internal/models"
	"lego-parser/internal/repository"

	"github.com/xuri/excelize/v2"
)

type ExportHandler struct {
	productRepo *repository.ProductRepo
}

func NewExportHandler(productRepo *repository.ProductRepo) *ExportHandler {
	return &ExportHandler{productRepo: productRepo}
}

func (h *ExportHandler) ExportProducts(w http.ResponseWriter, r *http.Request) {
	params := models.QueryParams{Page: 1, PerPage: 10000, SortBy: "name_ka", SortOrder: "asc"}
	if s := r.URL.Query().Get("search"); s != "" {
		params.Search = s
	}

	products, _, err := h.productRepo.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	f := excelize.NewFile()
	sheet := "Products"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"ID", "MCode", "Barcode", "LEGO ID",
		"Name", "Name EN",
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for row, p := range products {
		r := row + 2
		f.SetCellValue(sheet, cellName(1, r), p.ID)
		f.SetCellValue(sheet, cellName(2, r), derefStr(p.MCode))
		f.SetCellValue(sheet, cellName(3, r), derefStr(p.Barcode))
		f.SetCellValue(sheet, cellName(4, r), derefStr(p.InvoiceCode))
		f.SetCellValue(sheet, cellName(5, r), p.NameKa)
		f.SetCellValue(sheet, cellName(6, r), derefStr(p.NameEn))
	}

	lastCell := cellName(len(headers), len(products)+1)
	f.AutoFilter(sheet, "A1:"+lastCell, nil)

	filename := fmt.Sprintf("products_%s.xlsx", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(w)
}

func (h *ExportHandler) ExportComparison(w http.ResponseWriter, r *http.Request) {
	params := models.QueryParams{Page: 1, PerPage: 10000, SortBy: "name_ka", SortOrder: "asc"}
	if s := r.URL.Query().Get("search"); s != "" {
		params.Search = s
	}

	results, _, err := h.productRepo.GetComparison(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	f := excelize.NewFile()
	sheet := "Price Comparison"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"MCode", "Barcode", "LEGO ID",
		"Name", "Name EN",
		"Biblusi XS ₾", "Biblusi Pepela ₾",
		"Wolt XS ₾", "Wolt Pepela ₾",
		"Glovo XS ₾", "Glovo Pepela ₾",
		"Wishlist ₾", "PiccolaToys ₾", "Kubiki ₾",
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for i, h := range headers {
		cell := cellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for row, c := range results {
		r := row + 2
		f.SetCellValue(sheet, cellName(1, r), derefStr(c.MCode))
		f.SetCellValue(sheet, cellName(2, r), derefStr(c.Barcode))
		f.SetCellValue(sheet, cellName(3, r), derefStr(c.InvoiceCode))
		f.SetCellValue(sheet, cellName(4, r), c.NameKa)
		f.SetCellValue(sheet, cellName(5, r), derefStr(c.NameEn))
		setCellFloat(f, sheet, 6, r, c.BiblusiXSPrice)
		setCellFloat(f, sheet, 7, r, c.BiblusiPepelaPrice)
		setCellFloat(f, sheet, 8, r, c.WoltXSPrice)
		setCellFloat(f, sheet, 9, r, c.WoltPepelaPrice)
		setCellFloat(f, sheet, 10, r, c.GlovoXSPrice)
		setCellFloat(f, sheet, 11, r, c.GlovoPepelaPrice)
		setCellFloat(f, sheet, 12, r, c.WishlistPrice)
		setCellFloat(f, sheet, 13, r, c.PiccolaToysPrice)
		setCellFloat(f, sheet, 14, r, c.KubikiPrice)
	}

	lastCell := cellName(len(headers), len(results)+1)
	f.AutoFilter(sheet, "A1:"+lastCell, nil)

	filename := fmt.Sprintf("price_comparison_%s.xlsx", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(w)
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func setCellFloat(f *excelize.File, sheet string, col, row int, val *float64) {
	cell := cellName(col, row)
	if val != nil {
		f.SetCellValue(sheet, cell, *val)
	}
}

func floatStr(val *float64) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *val)
}

// ExportProductsCSV exports products as CSV (Qlik-compatible, tab-delimited with UTF-8 BOM)
func (h *ExportHandler) ExportProductsCSV(w http.ResponseWriter, r *http.Request) {
	params := models.QueryParams{Page: 1, PerPage: 100000, SortBy: "name_ka", SortOrder: "asc"}
	if s := r.URL.Query().Get("search"); s != "" {
		params.Search = s
	}

	products, _, err := h.productRepo.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("products_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	// UTF-8 BOM for Excel/Qlik compatibility
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	writer.Comma = '\t'

	writer.Write([]string{"ID", "MCode", "Barcode", "LegoID", "Name", "NameEN", "Sources"})
	for _, p := range products {
		sources := ""
		if len(p.Sources) > 0 {
			for i, s := range p.Sources {
				if i > 0 {
					sources += ","
				}
				sources += s
			}
		}
		writer.Write([]string{
			fmt.Sprintf("%d", p.ID),
			derefStr(p.MCode),
			derefStr(p.Barcode),
			derefStr(p.InvoiceCode),
			p.NameKa,
			derefStr(p.NameEn),
			sources,
		})
	}
	writer.Flush()
}

// ExportComparisonCSV exports price comparison as CSV (Qlik-compatible)
func (h *ExportHandler) ExportComparisonCSV(w http.ResponseWriter, r *http.Request) {
	params := models.QueryParams{Page: 1, PerPage: 100000, SortBy: "name_ka", SortOrder: "asc"}
	if s := r.URL.Query().Get("search"); s != "" {
		params.Search = s
	}

	results, _, err := h.productRepo.GetComparison(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("price_comparison_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	writer.Comma = '\t'

	writer.Write([]string{
		"MCode", "Barcode", "LegoID", "Name", "NameEN",
		"BiblusiXS", "BiblusiPepela", "WoltXS", "WoltPepela",
		"GlovoXS", "GlovoPepela", "Wishlist", "PiccolaToys", "Kubiki",
	})
	for _, c := range results {
		writer.Write([]string{
			derefStr(c.MCode),
			derefStr(c.Barcode),
			derefStr(c.InvoiceCode),
			c.NameKa,
			derefStr(c.NameEn),
			floatStr(c.BiblusiXSPrice),
			floatStr(c.BiblusiPepelaPrice),
			floatStr(c.WoltXSPrice),
			floatStr(c.WoltPepelaPrice),
			floatStr(c.GlovoXSPrice),
			floatStr(c.GlovoPepelaPrice),
			floatStr(c.WishlistPrice),
			floatStr(c.PiccolaToysPrice),
			floatStr(c.KubikiPrice),
		})
	}
	writer.Flush()
}
