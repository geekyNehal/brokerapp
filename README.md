# BrokerApp

A Go-based trading platform with user authentication, holdings management, orderbook, and positions tracking.

## Prerequisites

- Docker and Docker Compose
- Go 1.21 or later (for local development)

## Quick Start

1. Clone the repository:
```bash
git clone <repository-url>
cd brokerapp
```

2. Start the services:
```bash
docker-compose up --build
```

This will start:
- The BrokerApp service on port 8080
- MySQL database on port 3306

## Database Initialization

The database is automatically initialized with:
- Required tables (users, holdings, orders, positions)
- Sample data for testing

## API Documentation

### Authentication

#### Sign Up
```http
POST /api/signup
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "your-password"
}
```

#### Login
```http
POST /api/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "your-password"
}
```

Response:
```json
{
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Refresh Token
```http
POST /api/refresh
Content-Type: application/json

{
    "refresh_token": "your-refresh-token"
}
```

### Protected Endpoints

All protected endpoints require the `Authorization` header:
```
Authorization: Bearer <access_token>
```

#### Get User Profile
```http
GET /api/profile
```

Response:
```json
{
    "id": 1,
    "email": "user@example.com",
    "created_at": "2024-02-20T12:00:00Z"
}
```

#### Holdings Management

Get Holdings:
```http
GET /api/holdings
```

Response:
```json
[
    {
        "symbol": "AAPL",
        "quantity": 10,
        "price": 150.50,
        "value": 1505.00
    }
]
```

Create Holding:
```http
POST /api/holdings
Content-Type: application/json

{
    "symbol": "AAPL",
    "quantity": 10,
    "price": 150.50
}
```

#### Orderbook
```http
GET /api/orderbook
```

Response:
```json
{
    "orders": [
        {
            "id": 1,
            "symbol": "AAPL",
            "type": "buy",
            "quantity": 10,
            "price": 150.50,
            "status": "open"
        }
    ],
    "pnl": {
        "unrealized": 0,
        "realized": 0,
        "total": 0
    }
}
```

#### Positions
```http
GET /api/positions
```

Response:
```json
[
    {
        "symbol": "AAPL",
        "quantity": 10,
        "average_price": 150.50,
        "current_price": 152.00,
        "unrealized_pnl": 15.00
    }
]
```

## Development

### Local Development Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Run the application:
```bash
go run cmd/brokerapp/main.go
```

### Running Tests

```bash
go test ./...
```

## Environment Variables

- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 3306)
- `DB_USER`: Database user (default: root)
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name (default: brokerapp)
- `JWT_SECRET`: Secret key for JWT token generation
- `SERVER_PORT`: Server port (default: 8080)

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Holdings Table
```sql
CREATE TABLE holdings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    value DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### Orders Table
```sql
CREATE TABLE orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    type ENUM('buy', 'sell') NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    status ENUM('open', 'filled', 'cancelled') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### Positions Table
```sql
CREATE TABLE positions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    quantity INT NOT NULL,
    average_price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

## Sample Data

The database is initialized with the following sample data:

### Users
- Email: test@example.com
- Password: password123

### Holdings
- AAPL: 10 shares at $150.50
- GOOGL: 5 shares at $2800.00

### Orders
- Buy AAPL: 10 shares at $150.50
- Sell GOOGL: 5 shares at $2800.00

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 