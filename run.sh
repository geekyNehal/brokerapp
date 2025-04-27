#!/bin/bash

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Load environment variables
if [ -f "$SCRIPT_DIR/.env" ]; then
    echo "Loading environment variables from $SCRIPT_DIR/.env"
    export $(cat "$SCRIPT_DIR/.env" | grep -v '^#' | xargs)
    echo "Environment variables loaded:"
    echo "DB_HOST=$DB_HOST"
    echo "DB_PORT=$DB_PORT"
    echo "DB_USER=$DB_USER"
    echo "DB_NAME=$DB_NAME"
    echo "JWT_SECRET=${JWT_SECRET:0:5}..." # Only show first 5 chars of secret
else
    echo "Error: .env file not found in $SCRIPT_DIR"
    exit 1
fi

# Function to start everything
start_all() {
    echo "Starting MySQL database..."
    docker-compose up -d mysql
    echo "Waiting for MySQL to be ready..."
    sleep 10
    echo "MySQL is ready!"

    echo "Starting the service..."
    cd "$SCRIPT_DIR" && go run cmd/brokerapp/main.go
}

# Function to stop everything
stop_all() {
    echo "Stopping the service..."
    pkill -f "go run cmd/brokerapp/main.go"
    
    echo "Stopping MySQL database..."
    docker-compose down
}

# Function to reset everything
reset_all() {
    echo "Resetting everything..."
    stop_all
    docker-compose down -v
    docker-compose up -d mysql
    echo "Waiting for MySQL to be ready..."
    sleep 10
    echo "MySQL is ready!"
}

# Function to view logs
view_logs() {
    docker-compose logs -f mysql
}

# Main script
case "$1" in
    start)
        start_all
        ;;
    stop)
        stop_all
        ;;
    reset)
        reset_all
        ;;
    logs)
        view_logs
        ;;
    *)
        echo "Usage: $0 {start|stop|reset|logs}"
        exit 1
        ;;
esac 