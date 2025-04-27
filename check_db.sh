#!/bin/bash

# Check if MySQL container is running
if ! docker ps | grep -q brokerapp-mysql; then
    echo "MySQL container is not running. Please start it first."
    exit 1
fi

# Check user
echo "Checking user..."
docker exec -it brokerapp-mysql mysql -ubrokerapp -pbrokerapp brokerapp -e "
SELECT id, email, created_at 
FROM users 
WHERE email='user@example.com';"

# Check holdings
echo -e "\nChecking holdings..."
docker exec -it brokerapp-mysql mysql -ubrokerapp -pbrokerapp brokerapp -e "
SELECT h.* 
FROM holdings h 
JOIN users u ON h.user_id = u.id 
WHERE u.email='user@example.com';"

# Check refresh tokens
echo -e "\nChecking refresh tokens..."
docker exec -it brokerapp-mysql mysql -ubrokerapp -pbrokerapp brokerapp -e "
SELECT rt.* 
FROM refresh_tokens rt 
JOIN users u ON rt.user_id = u.id 
WHERE u.email='user@example.com';" 