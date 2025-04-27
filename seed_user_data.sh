#!/bin/bash

# Connect to MySQL and execute commands
docker exec -i brokerapp-mysql mysql -u brokerapp -pbrokerapp brokerapp << EOF

-- Delete any existing data for this user
DELETE FROM positions WHERE user_id = 5;
DELETE FROM orders WHERE user_id = 5;
DELETE FROM holdings WHERE user_id = 5;

-- Sample holdings
INSERT INTO holdings (user_id, symbol, quantity, price, value, created_at, updated_at) VALUES
(5, 'AAPL', 100, 150.25, 15025.00, NOW(), NOW()),
(5, 'GOOGL', 50, 2800.75, 140037.50, NOW(), NOW()),
(5, 'MSFT', 75, 300.50, 22537.50, NOW(), NOW()),
(5, 'AMZN', 25, 3500.00, 87500.00, NOW(), NOW()),
(5, 'TSLA', 10, 700.25, 7002.50, NOW(), NOW());

-- Sample orders
INSERT INTO orders (user_id, symbol, side, price, quantity, status, created_at, updated_at) VALUES
(5, 'AAPL', 'buy', 150.25, 100, 'filled', NOW(), NOW()),
(5, 'GOOGL', 'buy', 2800.75, 50, 'filled', NOW(), NOW()),
(5, 'MSFT', 'buy', 300.50, 75, 'filled', NOW(), NOW()),
(5, 'AMZN', 'buy', 3500.00, 25, 'filled', NOW(), NOW()),
(5, 'TSLA', 'buy', 700.25, 10, 'filled', NOW(), NOW()),
(5, 'AAPL', 'sell', 155.00, 50, 'pending', NOW(), NOW()),
(5, 'MSFT', 'sell', 305.00, 25, 'pending', NOW(), NOW());

-- Sample positions
INSERT INTO positions (user_id, symbol, quantity, entry_price, current_price, unrealized_pnl, realized_pnl, total_pnl, pnl_percentage, created_at, updated_at) VALUES
(5, 'AAPL', 100, 150.25, 155.00, 475.00, 0.00, 475.00, 3.16, NOW(), NOW()),
(5, 'GOOGL', 50, 2800.75, 2850.00, 2462.50, 0.00, 2462.50, 0.88, NOW(), NOW()),
(5, 'MSFT', 75, 300.50, 305.00, 337.50, 0.00, 337.50, 1.12, NOW(), NOW()),
(5, 'AMZN', 25, 3500.00, 3550.00, 1250.00, 0.00, 1250.00, 0.36, NOW(), NOW()),
(5, 'TSLA', 10, 700.25, 710.00, 97.50, 0.00, 97.50, 1.39, NOW(), NOW());

EOF

echo "Sample data has been added successfully for user test@example.com!"
echo "You can now view your holdings, orders, and positions using the API endpoints." 