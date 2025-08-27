

var productServiceURL string


package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	db := initDB()
	defer db.Close()

	productServiceURL = os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		productServiceURL = "http://localhost:8081"
	}

	transactionRepo := NewTransactionRepository(db)

	transactionService := NewTransactionService(transactionRepo)

	transactionHandler := NewTransactionHandler(transactionService)

	r := mux.NewRouter()
	r.Use(corsMiddleware)
	setupTransactionRoutes(r, transactionHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	fmt.Printf("Transaction Service running on port %s\n", port)
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

	var err error
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction Service: Successfully connected to database!")
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

// Repository Interface and Implementation
type TransactionRepository interface {
	Create(tx *Transaction) error
	CreateItem(item *TransactionItem) error
	GetAll(filters TransactionFilters) ([]*Transaction, error)
	GetByID(id int) (*Transaction, error)
	GetByCode(code string) (*Transaction, error)
	GetItemsByTransactionID(transactionID int) ([]*TransactionItem, error)
	GetSalesReport(startDate, endDate time.Time) (*SalesReport, error)
	Update(id int, tx *Transaction) error
	Delete(id int) error
}

type TransactionFilters struct {
	StartDate     *time.Time
	EndDate       *time.Time
	CustomerName  string
	PaymentMethod string
	Status        string
	Limit         int
	Offset        int
}

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(tx *Transaction) error {
	query := `INSERT INTO transactions (transaction_code, total_amount, payment_method, 
			  payment_amount, change_amount, customer_name, customer_phone, notes, cashier_name) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
			  RETURNING id, created_at, updated_at`
	
	return r.db.QueryRow(query, tx.TransactionCode, tx.TotalAmount, tx.PaymentMethod,
		tx.PaymentAmount, tx.ChangeAmount, tx.CustomerName, tx.CustomerPhone,
		tx.Notes, tx.CashierName).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)
}

func (r *transactionRepository) CreateItem(item *TransactionItem) error {
	query := `INSERT INTO transaction_items (transaction_id, product_id, product_name, 
			  product_sku, quantity, unit_price, total_price) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) 
			  RETURNING id, created_at`
	
	return r.db.QueryRow(query, item.TransactionID, item.ProductID, item.ProductName,
		item.ProductSKU, item.Quantity, item.UnitPrice, item.TotalPrice).
		Scan(&item.ID, &item.CreatedAt)
}

func (r *transactionRepository) GetAll(filters TransactionFilters) ([]*Transaction, error) {
	query := `SELECT id, transaction_code, total_amount, payment_method, payment_amount, 
			  change_amount, customer_name, customer_phone, notes, status, cashier_name, 
			  created_at, updated_at FROM transactions`
	
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filters.EndDate)
		argIndex++
	}

	if filters.CustomerName != "" {
		conditions = append(conditions, fmt.Sprintf("customer_name ILIKE $%d", argIndex))
		args = append(args, "%"+filters.CustomerName+"%")
		argIndex++
	}

	if filters.PaymentMethod != "" {
		conditions = append(conditions, fmt.Sprintf("payment_method = $%d", argIndex))
		args = append(args, filters.PaymentMethod)
		argIndex++
	}

	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filters.Status)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx := &Transaction{}
		err := rows.Scan(&tx.ID, &tx.TransactionCode, &tx.TotalAmount, &tx.PaymentMethod,
			&tx.PaymentAmount, &tx.ChangeAmount, &tx.CustomerName, &tx.CustomerPhone,
			&tx.Notes, &tx.Status, &tx.CashierName, &tx.CreatedAt, &tx.UpdatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (r *transactionRepository) GetByID(id int) (*Transaction, error) {
	query := `SELECT id, transaction_code, total_amount, payment_method, payment_amount, 
			  change_amount, customer_name, customer_phone, notes, status, cashier_name, 
			  created_at, updated_at FROM transactions WHERE id = $1`
	
	tx := &Transaction{}
	err := r.db.QueryRow(query, id).Scan(&tx.ID, &tx.TransactionCode, &tx.TotalAmount,
		&tx.PaymentMethod, &tx.PaymentAmount, &tx.ChangeAmount, &tx.CustomerName,
		&tx.CustomerPhone, &tx.Notes, &tx.Status, &tx.CashierName, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get transaction items
	items, err := r.GetItemsByTransactionID(id)
	if err != nil {
		return nil, err
	}
	tx.Items = items

	return tx, nil
}

func (r *transactionRepository) GetByCode(code string) (*Transaction, error) {
	query := `SELECT id, transaction_code, total_amount, payment_method, payment_amount, 
			  change_amount, customer_name, customer_phone, notes, status, cashier_name, 
			  created_at, updated_at FROM transactions WHERE transaction_code = $1`
	
	tx := &Transaction{}
	err := r.db.QueryRow(query, code).Scan(&tx.ID, &tx.TransactionCode, &tx.TotalAmount,
		&tx.PaymentMethod, &tx.PaymentAmount, &tx.ChangeAmount, &tx.CustomerName,
		&tx.CustomerPhone, &tx.Notes, &tx.Status, &tx.CashierName, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get transaction items
	items, err := r.GetItemsByTransactionID(tx.ID)
	if err != nil {
		return nil, err
	}
	tx.Items = items

	return tx, nil
}

func (r *transactionRepository) GetItemsByTransactionID(transactionID int) ([]*TransactionItem, error) {
	query := `SELECT id, transaction_id, product_id, product_name, product_sku, 
			  quantity, unit_price, total_price, created_at 
			  FROM transaction_items WHERE transaction_id = $1`
	
	rows, err := r.db.Query(query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*TransactionItem
	for rows.Next() {
		item := &TransactionItem{}
		err := rows.Scan(&item.ID, &item.TransactionID, &item.ProductID, &item.ProductName,
			&item.ProductSKU, &item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *transactionRepository) GetSalesReport(startDate, endDate time.Time) (*SalesReport, error) {
	// Get basic sales data
	basicQuery := `SELECT COUNT(*), COALESCE(SUM(total_amount), 0) FROM transactions 
				   WHERE created_at BETWEEN $1 AND $2 AND status = 'completed'`
	
	report := &SalesReport{}
	err := r.db.QueryRow(basicQuery, startDate, endDate).Scan(&report.TotalTransactions, &report.TotalAmount)
	if err != nil {
		return nil, err
	}

	// Get top products
	topProductsQuery := `SELECT ti.product_id, ti.product_name, SUM(ti.quantity) as total_qty, 
						 SUM(ti.total_price) as total_revenue
						 FROM transaction_items ti 
						 JOIN transactions t ON ti.transaction_id = t.id 
						 WHERE t.created_at BETWEEN $1 AND $2 AND t.status = 'completed'
						 GROUP BY ti.product_id, ti.product_name 
						 ORDER BY total_revenue DESC LIMIT 10`
	
	rows, err := r.db.Query(topProductsQuery, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topProducts []ProductSales
	for rows.Next() {
		product := ProductSales{}
		err := rows.Scan(&product.ProductID, &product.ProductName, &product.TotalQty, &product.TotalRevenue)
		if err != nil {
			return nil, err
		}
		topProducts = append(topProducts, product)
	}
	report.TopProducts = topProducts

	// Get daily sales
	dailyQuery := `SELECT DATE(created_at) as sale_date, COUNT(*) as transactions, 
				   SUM(total_amount) as amount FROM transactions 
				   WHERE created_at BETWEEN $1 AND $2 AND status = 'completed'
				   GROUP BY DATE(created_at) ORDER BY sale_date`
	
	dailyRows, err := r.db.Query(dailyQuery, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer dailyRows.Close()

	var dailySales []DailySales
	for dailyRows.Next() {
		daily := DailySales{}
		err := dailyRows.Scan(&daily.Date, &daily.Transactions, &daily.Amount)
		if err != nil {
			return nil, err
		}
		dailySales = append(dailySales, daily)
	}
	report.DailySales = dailySales

	return report, nil
}

func (r *transactionRepository) Update(id int, tx *Transaction) error {
	query := `UPDATE transactions SET payment_method = $1, payment_amount = $2, 
			  change_amount = $3, customer_name = $4, customer_phone = $5, notes = $6, 
			  status = $7, updated_at = CURRENT_TIMESTAMP WHERE id = $8`
	
	_, err := r.db.Exec(query, tx.PaymentMethod, tx.PaymentAmount, tx.ChangeAmount,
		tx.CustomerName, tx.CustomerPhone, tx.Notes, tx.Status, id)
	return err
}

func (r *transactionRepository) Delete(id int) error {
	query := `DELETE FROM transactions WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}