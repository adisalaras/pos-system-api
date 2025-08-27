package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CategoryHandler struct {
	service CategoryService
}

func NewCategoryHandler(service CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.GetAllCategories()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "categori gagal diambil", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Kategori berhasil diambil", categories)
}

func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID kategori tidak valid", err)
		return
	}

	category, err := h.service.GetCategoryByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Kategori tidak ditemukan", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Kategori berhasil diambil", category)
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "request body tidak valid", err)
		return
	}

	category, err := h.service.CreateCategory(req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Gagal membuat kategori", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, "Kategori berhasil dibuat", category)
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID kategori tidak valid", err)
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "request body tidak valid", err)
		return
	}

	category, err := h.service.UpdateCategory(id, req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "kategori gagal diperbarui", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Kategori berhasil diperbarui", category)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID kategori tidak valid", err)
		return
	}

	err = h.service.DeleteCategory(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal menghapus kategori", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Kategori berhasil dihapus", nil)
}

// Helper functions
func respondWithJSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: false,
		Message: message,
		Error:   err.Error(),
	}

	json.NewEncoder(w).Encode(response)
}