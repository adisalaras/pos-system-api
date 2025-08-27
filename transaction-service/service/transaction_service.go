package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Transaction struct {
	ID              int                `json:"id"`
	TransactionCode string             `json:"transaction_code"`
	TotalAmount     float64            `json:"total_amount"`
	PaymentMethod   string             `json:"payment_method"`
	PaymentAmount   float64            `json:"payment_amount"`
	ChangeAmount    float64            `json:"change_amount"`
	CustomerName    *string            `json:"customer_name"`
	CustomerPhone   *string            `json:"customer_phone"`
	Notes           *string            `json:"notes"`
	Status          string             `json:"status"`
	CashierName     *string            `json:"cashier_name"`
	Items           []*TransactionItem `json:"items,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type TransactionItem struct {
	ID            int       `json:"id"`
	TransactionID int       `json:"transaction_id"`
	ProductID     int       `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ProductSKU    string    `json:"product_sku"`
	Quantity      int       `json:"quantity"`
	UnitPrice     float64   `json:"unit_price"`
	TotalPrice    float64   `json:"total_price"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateTransactionRequest struct {
	PaymentMethod string                         `json:"payment_method"`
	PaymentAmount float64                        `json:"payment_amount"`
	CustomerName  *string                        `json:"customer_name"`
	CustomerPhone *string                        `json:"customer_phone"`
	Notes         *string                        `json:"notes"`
	CashierName   *string                        `json:"cashier_name"`
	Items         []CreateTransactionItemRequest `json:"items"`
}

type CreateTransactionItemRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type SalesReportRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type SalesReport struct {
	TotalTransactions int            `json:"total_transactions"`
	TotalAmount       float64        `json:"total_amount"`
	TotalProfit       float64        `json:"total_profit"`
	TopProducts       []ProductSales `json:"top_products"`
	DailySales        []DailySales   `json:"daily_sales"`
}

type ProductSales struct {
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	TotalQty     int     `json:"total_quantity"`
	TotalRevenue float64 `json:"total_revenue"`
}

type DailySales struct {
	Date         string  `json:"date"`
	Transactions int     `json:"transactions"`
	Amount       float64 `json:"amount"`
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

type TransactionService interface {
	CreateTransaction(req CreateTransactionRequest) (*Transaction, error)
	GetAllTransactions(filters TransactionFilters) ([]*Transaction, error)
	GetTransactionByID(id int) (*Transaction, error)
	GetTransactionByCode(code string) (*Transaction, error)
	GetSalesReport(req SalesReportRequest) (*SalesReport, error)
	UpdateTransaction(id int, req UpdateTransactionRequest) (*Transaction, error)
	DeleteTransaction(id int) error
}

type UpdateTransactionRequest struct {
	PaymentMethod *string `json:"payment_method"`
	CustomerName  *string `json:"customer_name"`
	CustomerPhone *string `json:"customer_phone"`
	Notes         *string `json:"notes"`
	Status        *string `json:"status"`
}

type transactionService struct {
	repo TransactionRepository
}

type Product struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Price         float64 `json:"price"`
	Cost          float64 `json:"cost"`
	StockQuantity int     `json:"stock_quantity"`
}

type ProductResponse struct {
	Success bool    `json:"success"`
	Data    Product `json:"data"`
}

func NewTransactionService(repo TransactionRepository) TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) CreateTransaction(req CreateTransactionRequest) (*Transaction, error) {
	// Validasi
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("transaksi harus memiliki setidaknya satu item")
	}
	if req.PaymentAmount < 0 {
		return nil, fmt.Errorf("jumlah pembayaran tidak boleh negatif")
	}

	// Validasi produk dan hitung total
	var totalAmount float64
	var validatedItems []CreateTransactionItemRequest

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("Quantity item harus positif")
		}
		// cari data produk yang cocok dengan request id
		product, err := s.getProductByID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("produk dengan ID %d tidak ditemukan: %w", item.ProductID, err)
		}

		// cekk stock
		if product.StockQuantity < item.Quantity {
			return nil, fmt.Errorf("stok tidak mencukupi untuk produk %s (tersedia: %d, diminta: %d)",
				product.Name, product.StockQuantity, item.Quantity)
		}

		totalAmount += product.Price * float64(item.Quantity)
		validatedItems = append(validatedItems, item)
	}

	// hitung kembalian
	changeAmount := req.PaymentAmount - totalAmount
	if changeAmount < 0 {
		return nil, fmt.Errorf("jumlah pembayaran tidak mencukupi")
	}

	// Generate code transaksi
	transactionCode := s.generateTransactionCode()

	// buat transaksi
	transaction := &Transaction{
		TransactionCode: transactionCode,
		TotalAmount:     totalAmount,
		PaymentMethod:   req.PaymentMethod,
		PaymentAmount:   req.PaymentAmount,
		ChangeAmount:    changeAmount,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		Notes:           req.Notes,
		Status:          "completed",
		CashierName:     req.CashierName,
	}

	err := s.repo.Create(transaction)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat transaksi: %w", err)
	}

	// buat item transaksi
	var items []*TransactionItem
	for _, itemReq := range validatedItems {
		product, _ := s.getProductByID(itemReq.ProductID)

		item := &TransactionItem{
			TransactionID: transaction.ID,
			ProductID:     itemReq.ProductID,
			ProductName:   product.Name,
			ProductSKU:    product.SKU,
			Quantity:      itemReq.Quantity,
			UnitPrice:     product.Price,
			TotalPrice:    product.Price * float64(itemReq.Quantity),
		}

		err := s.repo.CreateItem(item)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat transaksi: %w", err)
		}

		// Update stok produk
		err = s.updateProductStock(itemReq.ProductID, itemReq.Quantity, "subtract")
		if err != nil {
			return nil, fmt.Errorf("gagal memperbarui stok produk: %w", err)
		}

		items = append(items, item)
	}

	transaction.Items = items
	return transaction, nil
}

func (s *transactionService) GetAllTransactions(filters TransactionFilters) ([]*Transaction, error) {
	transactions, err := s.repo.GetAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	// ambil item transaksi untuk setiap transaksi
	for _, tx := range transactions {
		items, err := s.repo.GetItemsByTransactionID(tx.ID)
		if err != nil {
			return nil, fmt.Errorf("gagal mendapatkan item transaksi: %w", err)
		}
		tx.Items = items
	}

	return transactions, nil
}

func (s *transactionService) GetTransactionByID(id int) (*Transaction, error) {
	return s.repo.GetByID(id)
}

func (s *transactionService) GetTransactionByCode(code string) (*Transaction, error) {
	return s.repo.GetByCode(code)
}

func (s *transactionService) GetSalesReport(req SalesReportRequest) (*SalesReport, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("tanggal mulai tidak valid")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("tanggal akhir tidak valid")
	}

	// set waktu akhir ke akhir hari
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	return s.repo.GetSalesReport(startDate, endDate)
}

func (s *transactionService) UpdateTransaction(id int, req UpdateTransactionRequest) (*Transaction, error) {
	
	existingTx, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("transaksi tidak ditemukan: %w", err)
	}

	// perbarui field yang diubah
	if req.PaymentMethod != nil {
		existingTx.PaymentMethod = *req.PaymentMethod
	}
	if req.CustomerName != nil {
		existingTx.CustomerName = req.CustomerName
	}
	if req.CustomerPhone != nil {
		existingTx.CustomerPhone = req.CustomerPhone
	}
	if req.Notes != nil {
		existingTx.Notes = req.Notes
	}
	if req.Status != nil {
		existingTx.Status = *req.Status
	}

	err = s.repo.Update(id, existingTx)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	return s.repo.GetByID(id)
}

func (s *transactionService) DeleteTransaction(id int) error {
	return s.repo.Delete(id)
}


func (s *transactionService) generateTransactionCode() string {
	now := time.Now()
	return fmt.Sprintf("TRX%s%d", now.Format("20060102"), now.Unix()%10000)
}

func (s *transactionService) getProductByID(id int) (*Product, error) {
	url := fmt.Sprintf("%s/api/products/%d", productServiceURL, id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gagal mendapatkan produk: %d", resp.StatusCode)
	}

	var response ProductResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("gagal mendapatkan produk: %d", resp.StatusCode)
	}

	return &response.Data, nil
}

func (s *transactionService) updateProductStock(productID int, quantity int, operation string) error {
	url := fmt.Sprintf("%s/api/products/%d/stock", productServiceURL, productID)

	stockReq := map[string]interface{}{
		"quantity": quantity,
		"type":     operation,
		"notes":    "Stok diperbarui dari transaksi",
	}

	jsonData, err := json.Marshal(stockReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gagal memperbarui stok produk, status: %d", resp.StatusCode)
	}

	return nil
}
