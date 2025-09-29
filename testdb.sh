#!/bin/bash
echo "Testing database connection..."
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "Database: $DB_NAME"
echo "User: $DB_USER"

# Test basic connectivity
echo "Testing ping..."
ping -c 3 $DB_HOST

# Test port accessibility  
echo "Testing port accessibility..."
timeout 5 bash -c "</dev/tcp/$DB_HOST/$DB_PORT" && echo "Port $DB_PORT is open" || echo "Port $DB_PORT is closed"

# Test PostgreSQL connection
echo "Testing PostgreSQL connection..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();"
