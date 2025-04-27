#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 {build|up|down|logs|clean}"
    exit 1
}

# Check if at least one argument is provided
if [ $# -lt 1 ]; then
    usage
fi

case "$1" in
    build)
        echo "Building Docker images..."
        docker-compose build
        ;;
    up)
        echo "Starting services..."
        docker-compose up -d
        ;;
    down)
        echo "Stopping services..."
        docker-compose down
        ;;
    logs)
        echo "Showing logs..."
        docker-compose logs -f
        ;;
    clean)
        echo "Cleaning up Docker resources..."
        docker-compose down -v
        docker system prune -f
        ;;
    *)
        usage
        ;;
esac 