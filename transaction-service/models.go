package main

import "time"

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