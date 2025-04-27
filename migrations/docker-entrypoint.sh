#!/bin/bash
set -e

# Start MySQL server temporarily
mysqld --skip-networking --socket=/var/run/mysqld/mysqld.sock &
MYSQL_PID=$!

# Wait for MySQL to start
until mysqladmin ping -h localhost --silent; do
    sleep 1
done

# Create database and user
mysql -u root -ppassword << EOF
CREATE DATABASE IF NOT EXISTS brokerapp;
CREATE USER IF NOT EXISTS 'root'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' WITH GRANT OPTION;
ALTER USER 'root'@'%' IDENTIFIED WITH mysql_native_password BY 'password';
FLUSH PRIVILEGES;
EOF

# Stop the temporary MySQL server
kill $MYSQL_PID
wait $MYSQL_PID

# Start MySQL normally
exec "$@" 