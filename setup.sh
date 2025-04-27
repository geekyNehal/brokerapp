#!/bin/bash

# Function to check if a service is ready
wait_for_service() {
    local service=$1
    local port=$2
    local max_attempts=30
    local attempt=1

    echo "Waiting for $service to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if nc -z localhost $port; then
            echo "$service is ready!"
            return 0
        fi
        echo "Attempt $attempt: $service not ready yet, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    echo "Error: $service failed to start within the expected time"
    return 1
}

# Start the services
echo "Starting services..."
docker-compose up -d

# Wait for MySQL to be ready
wait_for_service "MySQL" 3306 || exit 1

# Wait for the app to be ready
wait_for_service "App" 8080 || exit 1

# Function to create a user and get tokens
create_user() {
    local email=$1
    local password=$2

    echo "Creating user: $email"
    response=$(curl -s -X POST http://localhost:8080/api/signup \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"$password\"}")

    if [ $? -ne 0 ]; then
        echo "Error creating user"
        return 1
    fi

    # Extract tokens from response using grep and sed
    access_token=$(echo $response | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')
    refresh_token=$(echo $response | grep -o '"refresh_token":"[^"]*"' | sed 's/"refresh_token":"//;s/"//')

    if [ -z "$access_token" ]; then
        echo "Error: Failed to get access token"
        return 1
    fi

    echo "User created successfully!"
    echo "Access Token: $access_token"
    echo "Refresh Token: $refresh_token"
    return 0
}

# Function to seed holdings data
seed_holdings() {
    local access_token=$1
    local user_id=$2

    echo "Seeding holdings data..."
    holdings=(
        '{"symbol":"AAPL","quantity":100,"price":150.25}'
        '{"symbol":"GOOGL","quantity":50,"price":2800.75}'
        '{"symbol":"MSFT","quantity":75,"price":300.50}'
    )

    for holding in "${holdings[@]}"; do
        curl -s -X POST http://localhost:8080/api/holdings \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $access_token" \
            -d "$holding"
    done
}

# Function to seed orders data
seed_orders() {
    local access_token=$1
    local user_id=$2

    echo "Seeding orders data..."
    orders=(
        '{"symbol":"AAPL","side":"buy","price":150.25,"quantity":100}'
        '{"symbol":"GOOGL","side":"buy","price":2800.75,"quantity":50}'
        '{"symbol":"MSFT","side":"buy","price":300.50,"quantity":75}'
    )

    for order in "${orders[@]}"; do
        curl -s -X POST http://localhost:8080/api/orders \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $access_token" \
            -d "$order"
    done
}

# Function to seed positions data
seed_positions() {
    local access_token=$1
    local user_id=$2

    echo "Seeding positions data..."
    positions=(
        '{"symbol":"AAPL","quantity":100,"entry_price":150.25,"current_price":155.00}'
        '{"symbol":"GOOGL","quantity":50,"entry_price":2800.75,"current_price":2850.00}'
        '{"symbol":"MSFT","quantity":75,"entry_price":300.50,"current_price":305.00}'
    )

    for position in "${positions[@]}"; do
        curl -s -X POST http://localhost:8080/api/positions \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $access_token" \
            -d "$position"
    done
}

# Main setup process
echo "Welcome to BrokerApp Setup!"
echo "Please enter the following details to create a new user:"

read -p "Email: " email
read -s -p "Password: " password
echo

# Create user and get tokens
create_user "$email" "$password" || exit 1

# Get user ID from the access token (you might need to adjust this based on your JWT structure)
user_id=$(echo $access_token | cut -d'.' -f2 | base64 -d | jq -r '.user_id')

# Seed data for the new user
seed_holdings "$access_token" "$user_id"
seed_orders "$access_token" "$user_id"
seed_positions "$access_token" "$user_id"

echo "Setup completed successfully!"
echo "You can now log in with:"
echo "Email: $email"
echo "Password: $password" 