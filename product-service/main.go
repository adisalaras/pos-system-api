// main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	// koneksi database
	db := initDB()
	defer db.Close()

	// Inisialisasi repository
	categoryRepo := NewCategoryRepository(db)
	productRepo := NewProductRepository(db)

	// Inisialisasi services
	categoryService := NewCategoryService(categoryRepo)
	productService := NewProductService(productRepo)

	// Inisialisasi handlers
	categoryHandler := NewCategoryHandler(categoryService)
	productHandler := NewProductHandler(productService)

	// Setup routes
	r := mux.NewRouter()
	
	// Tambah CORS middleware
	r.Use(corsMiddleware)

	// Setup routes
	setupRoutes(r, categoryHandler, productHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Printf("Product Service running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initDB() *sql.DB {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Koneksi database berhasil!")
	return db
}

func corsMiddleware(next http.Handler) http.Handler {
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

