package main

import (
	"github.com/gorilla/mux"
)

func setupTransactionRoutes(r *mux.Router, transactionHandler *TransactionHandler) {
	api := r.PathPrefix("/api").Subrouter()

	// Transaction routes
	transactions := api.PathPrefix("/transactions").Subrouter()
	transactions.HandleFunc("", transactionHandler.GetTransactions).Methods("GET")
	transactions.HandleFunc("", transactionHandler.CreateTransaction).Methods("POST")
	transactions.HandleFunc("/{id:[0-9]+}", transactionHandler.GetTransaction).Methods("GET")
	transactions.HandleFunc("/{id:[0-9]+}", transactionHandler.UpdateTransaction).Methods("PUT")
	transactions.HandleFunc("/{id:[0-9]+}", transactionHandler.DeleteTransaction).Methods("DELETE")
	transactions.HandleFunc("/code/{code}", transactionHandler.GetTransactionByCode).Methods("GET")

	// Sales report routes
	sales := api.PathPrefix("/sales").Subrouter()
	sales.HandleFunc("/report", transactionHandler.GetSalesReport).Methods("GET", "POST")

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "transaction-service",
		})
	}).Methods("GET")
}
