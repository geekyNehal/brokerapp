-- Create database if not exists
CREATE DATABASE IF NOT EXISTS brokerapp;

-- Use the database
USE brokerapp;

-- Drop existing root user if exists
DROP USER IF EXISTS 'root'@'%';

-- Create root user with proper permissions
CREATE USER 'root'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES; 