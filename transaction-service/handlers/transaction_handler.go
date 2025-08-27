package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type TransactionHandler struct {
	service TransactionService
}

func NewTransactionHandler(service TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	transaction, err := h.service.CreateTransaction(req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to create transaction", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, "Transaction created successfully", transaction)
}

func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	filters := TransactionFilters{}

	// parsing parameter (contoh: ?start_date=2023-01-01&end_date=2023-01-31&customer=John)
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filters.StartDate = &date
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			date = date.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filters.EndDate = &date
		}
	}

	if customer := r.URL.Query().Get("customer"); customer != "" {
		filters.CustomerName = customer
	}

	if payment := r.URL.Query().Get("payment_method"); payment != "" {
		filters.PaymentMethod = payment
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = status
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

	transactions, err := h.service.GetAllTransactions(filters)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "gagal mendapatkan transaksi", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Transaksi berhasil diambil", transactions)
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "id transaksi tidak valid", err)
		return
	}

	transaction, err := h.service.GetTransactionByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Transaksi tidak ditemukan", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Transaksi berhasil diambil", transaction)
}

func (h *TransactionHandler) GetTransactionByCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	transaction, err := h.service.GetTransactionByCode(code)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Transaksi tidak ditemukan", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Transaksi berhasil diambil", transaction)
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID transaksi tidak valid", err)
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "request body tidak valid", err)
		return
	}

	transaction, err := h.service.UpdateTransaction(id, req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Gagal memperbarui transaksi", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Transaksi berhasil diperbarui", transaction)
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID transaksi tidak valid", err)
		return
	}

	err = h.service.DeleteTransaction(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal menghapus transaksi", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Transaksi berhasil dihapus", nil)
}

func (h *TransactionHandler) GetSalesReport(w http.ResponseWriter, r *http.Request) {
	var req SalesReportRequest


	if r.Method == "GET" {
		req.StartDate = r.URL.Query().Get("start_date")
		req.EndDate = r.URL.Query().Get("end_date")

		// default 30 hari terakhir jika start date null
		if req.StartDate == "" {
			req.StartDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		}
		if req.EndDate == "" {
			req.EndDate = time.Now().Format("2006-01-02")
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "request body tidak valid", err)
			return
		}
	}

	report, err := h.service.GetSalesReport(req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "gagal mendapatkan laporan sales", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "laporan sales berhasil didapatkan", report)
}

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

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}
