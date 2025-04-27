#!/bin/bash

# Connect to MySQL and execute commands
docker exec -i brokerapp-mysql mysql -u root -ppassword brokerapp << EOF

-- Delete existing data in correct order to handle foreign key constraints
DELETE FROM positions;
DELETE FROM orders;
DELETE FROM holdings;
DELETE FROM refresh_tokens;
DELETE FROM users;

-- Reset auto-increment counters
ALTER TABLE positions AUTO_INCREMENT = 1;
ALTER TABLE orders AUTO_INCREMENT = 1;
ALTER TABLE holdings AUTO_INCREMENT = 1;
ALTER TABLE refresh_tokens AUTO_INCREMENT = 1;
ALTER TABLE users AUTO_INCREMENT = 1;

-- Create sample users with proper bcrypt hashed passwords
-- Password for all users is: password123
-- Hash generated with bcrypt.DefaultCost (10)
INSERT INTO users (email, password, created_at, updated_at) VALUES
('john.doe@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', NOW(), NOW()),
('jane.smith@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', NOW(), NOW()),
('bob.wilson@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', NOW(), NOW());

-- Wait for users to be created
SELECT SLEEP(1);

-- Sample holdings
INSERT INTO holdings (user_id, symbol, quantity, price, value, created_at, updated_at) VALUES
(1, 'AAPL', 100, 150.25, 15025.00, NOW(), NOW()),
(1, 'GOOGL', 50, 2800.75, 140037.50, NOW(), NOW()),
(2, 'MSFT', 75, 300.50, 22537.50, NOW(), NOW()),
(2, 'AMZN', 25, 3500.00, 87500.00, NOW(), NOW()),
(3, 'TSLA', 10, 700.25, 7002.50, NOW(), NOW()),
(3, 'NVDA', 30, 800.75, 24022.50, NOW(), NOW());

-- Sample orders
INSERT INTO orders (user_id, symbol, side, price, quantity, status, created_at, updated_at) VALUES
(1, 'AAPL', 'buy', 150.25, 100, 'filled', NOW(), NOW()),
(1, 'GOOGL', 'buy', 2800.75, 50, 'filled', NOW(), NOW()),
(2, 'MSFT', 'buy', 300.50, 75, 'filled', NOW(), NOW()),
(2, 'AMZN', 'buy', 3500.00, 25, 'filled', NOW(), NOW()),
(3, 'TSLA', 'buy', 700.25, 10, 'filled', NOW(), NOW()),
(3, 'NVDA', 'buy', 800.75, 30, 'filled', NOW(), NOW()),
(1, 'AAPL', 'sell', 155.00, 50, 'pending', NOW(), NOW()),
(2, 'MSFT', 'sell', 305.00, 25, 'pending', NOW(), NOW());

-- Sample positions
INSERT INTO positions (user_id, symbol, quantity, entry_price, current_price, unrealized_pnl, realized_pnl, total_pnl, pnl_percentage, created_at, updated_at) VALUES
(1, 'AAPL', 100, 150.25, 155.00, 475.00, 0.00, 475.00, 3.16, NOW(), NOW()),
(1, 'GOOGL', 50, 2800.75, 2850.00, 2462.50, 0.00, 2462.50, 0.88, NOW(), NOW()),
(2, 'MSFT', 75, 300.50, 305.00, 337.50, 0.00, 337.50, 1.12, NOW(), NOW()),
(2, 'AMZN', 25, 3500.00, 3550.00, 1250.00, 0.00, 1250.00, 0.36, NOW(), NOW()),
(3, 'TSLA', 10, 700.25, 710.00, 97.50, 0.00, 97.50, 1.39, NOW(), NOW()),
(3, 'NVDA', 30, 800.75, 810.00, 277.50, 0.00, 277.50, 0.35, NOW(), NOW());

EOF

echo "Sample data has been reset successfully!"
echo "You can now log in with any of these users using password: password123"
echo "Users:"
echo "1. john.doe@example.com"
echo "2. jane.smith@example.com"
echo "3. bob.wilson@example.com" 