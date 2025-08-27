package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	service ProductService
}


func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	filters := ProductFilters{}

	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		if id, err := strconv.Atoi(categoryID); err == nil {
			filters.CategoryID = &id
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = search
	}

	if isActive := r.URL.Query().Get("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters.IsActive = &active
		}
	}

	if lowStock := r.URL.Query().Get("low_stock"); lowStock == "true" {
		filters.LowStock = true
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters.Limit = l
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filters.Offset = o
		}
	}

	products, err := h.service.GetAllProducts(filters)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "gagal mengambil data produk", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Produk berhasil diambil", products)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "id produk tidak valid", err)
		return
	}

	product, err := h.service.GetProductByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Product tidak ditemukan", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Product berhasil diambil", product)
}

func (h *ProductHandler) GetProductBySKU(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sku := vars["sku"]

	product, err := h.service.GetProductBySKU(sku)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Product tidak ditemukan", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Product berhasil diambil", product)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	product, err := h.service.CreateProduct(req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Gagal membuat produk", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, "Produk berhasil dibuat", product)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID produk tidak valid", err)
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "request body tidak valid", err)
		return
	}

	product, err := h.service.UpdateProduct(id, req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Gagal memperbarui produk", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Produk berhasil diperbarui", product)
}

func (h *ProductHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID produk tidak valid", err)
		return
	}

	var req StockUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "request body tidak valid", err)
		return
	}

	err = h.service.UpdateStock(id, req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "stok gagal di update", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Stok berhasil diperbarui", nil)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID produk tidak valid", err)
		return
	}

	err = h.service.DeleteProduct(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "produk gagal dihapus", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Produk berhasil dihapus", nil)
}

func (h *ProductHandler) GetLowStockProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.GetLowStockProducts()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengambil produk dengan stok rendah", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Produk dengan stok rendah berhasil diambil", products)
}
