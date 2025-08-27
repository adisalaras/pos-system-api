package main

import (
	"github.com/gorilla/mux"
)

func setupRoutes(r *mux.Router, categoryHandler *CategoryHandler, productHandler *ProductHandler) {
	// semua routes akan diawali dengan /api, contoh : /api/categories
	api := r.PathPrefix("/api").Subrouter()

	categories := api.PathPrefix("/categories").Subrouter()
	categories.HandleFunc("", categoryHandler.GetCategories).Methods("GET")
	categories.HandleFunc("", categoryHandler.CreateCategory).Methods("POST")
	categories.HandleFunc("/{id:[0-9]+}", categoryHandler.GetCategory).Methods("GET")
	categories.HandleFunc("/{id:[0-9]+}", categoryHandler.UpdateCategory).Methods("PUT")
	categories.HandleFunc("/{id:[0-9]+}", categoryHandler.DeleteCategory).Methods("DELETE")

	products := api.PathPrefix("/products").Subrouter()
	products.HandleFunc("", productHandler.GetProducts).Methods("GET")
	products.HandleFunc("", productHandler.CreateProduct).Methods("POST")
	products.HandleFunc("/{id:[0-9]+}", productHandler.GetProduct).Methods("GET")
	products.HandleFunc("/{id:[0-9]+}", productHandler.UpdateProduct).Methods("PUT")
	products.HandleFunc("/{id:[0-9]+}", productHandler.DeleteProduct).Methods("DELETE")
	
	products.HandleFunc("/{id:[0-9]+}/stock", productHandler.UpdateStock).Methods("PUT")
	
	products.HandleFunc("/sku/{sku}", productHandler.GetProductBySKU).Methods("GET")
	products.HandleFunc("/low-stock", productHandler.GetLowStockProducts).Methods("GET")

	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "product-service",
		})
	}).Methods("GET")
}