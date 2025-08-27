-- Create database
CREATE DATABASE IF NOT EXISTS pos_system;

-- Use the database
\c pos_system;

-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create products table (dengan inventaris)
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(50) UNIQUE NOT NULL,
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    price DECIMAL(12,2) NOT NULL CHECK (price >= 0),
    cost DECIMAL(12,2) DEFAULT 0 CHECK (cost >= 0),
    stock_quantity INTEGER DEFAULT 0 CHECK (stock_quantity >= 0),
    min_stock INTEGER DEFAULT 0 CHECK (min_stock >= 0),
    description TEXT,
    image_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create transactions table (header transaksi)
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    transaction_code VARCHAR(50) UNIQUE NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL CHECK (total_amount >= 0),
    payment_method VARCHAR(50) NOT NULL DEFAULT 'cash',
    payment_amount DECIMAL(12,2) NOT NULL CHECK (payment_amount >= 0),
    change_amount DECIMAL(12,2) DEFAULT 0 CHECK (change_amount >= 0),
    customer_name VARCHAR(255),
    customer_phone VARCHAR(20),
    notes TEXT,
    status VARCHAR(20) DEFAULT 'completed',
    cashier_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create transaction_items table (detail transaksi)
CREATE TABLE IF NOT EXISTS transaction_items (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER REFERENCES transactions(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE RESTRICT,
    product_name VARCHAR(255) NOT NULL, -- untuk backup jika produk dihapus
    product_sku VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(12,2) NOT NULL CHECK (unit_price >= 0),
    total_price DECIMAL(12,2) NOT NULL CHECK (total_price >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes untuk performa
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_active ON products(is_active);
CREATE INDEX idx_transactions_code ON transactions(transaction_code);
CREATE INDEX idx_transactions_date ON transactions(created_at);
CREATE INDEX idx_transaction_items_transaction ON transaction_items(transaction_id);
CREATE INDEX idx_transaction_items_product ON transaction_items(product_id);

-- Create triggers untuk update timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample categories
INSERT INTO categories (name, description) VALUES
('Makanan', 'Produk makanan dan minuman'),
('Elektronik', 'Peralatan elektronik'),
('Pakaian', 'Pakaian dan aksesoris'),
('Perawatan', 'Produk perawatan dan kesehatan')
ON CONFLICT (name) DO NOTHING;

-- Insert sample products
INSERT INTO products (name, sku, category_id, price, cost, stock_quantity, min_stock, description) VALUES
('Nasi Gudeg', 'FOOD001', 1, 15000, 8000, 50, 10, 'Nasi gudeg khas Yogyakarta'),
('Es Teh Manis', 'DRINK001', 1, 5000, 2000, 100, 20, 'Es teh manis segar'),
('Kaos Batik', 'CLOTH001', 3, 75000, 45000, 25, 5, 'Kaos batik motif tradisional'),
('Power Bank', 'ELEC001', 2, 150000, 120000, 15, 3, 'Power bank 10000mAh'),
('Hand Sanitizer', 'CARE001', 4, 12000, 7000, 80, 15, 'Hand sanitizer 60ml')
ON CONFLICT (sku) DO NOTHING;

-- Insert sample transactions
INSERT INTO transactions (transaction_code, total_amount, payment_method, payment_amount, change_amount, customer_name, cashier_name) VALUES
('TRX001', 25000, 'cash', 30000, 5000, 'Budi Santoso', 'Admin'),
('TRX002', 150000, 'card', 150000, 0, 'Siti Aminah', 'Admin')
ON CONFLICT (transaction_code) DO NOTHING;

-- Insert sample transaction items
INSERT INTO transaction_items (transaction_id, product_id, product_name, product_sku, quantity, unit_price, total_price) VALUES
(1, 1, 'Nasi Gudeg', 'FOOD001', 1, 15000, 15000),
(1, 2, 'Es Teh Manis', 'DRINK001', 2, 5000, 10000),
(2, 4, 'Power Bank', 'ELEC001', 1, 150000, 150000)
ON CONFLICT DO NOTHING;