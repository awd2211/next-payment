#!/bin/bash

# Merchant Services Database Initialization Script
# Creates databases for merchant-policy-service and merchant-quota-service

set -e

echo "=================================================="
echo "Merchant Services Database Initialization"
echo "=================================================="
echo ""

# Database connection settings
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-40432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}

echo "Database Host: $DB_HOST:$DB_PORT"
echo "Database User: $DB_USER"
echo ""

# Function to create database if not exists
create_database() {
    local db_name=$1
    echo "Creating database: $db_name"

    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$db_name'" | grep -q 1 || \
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $db_name"

    if [ $? -eq 0 ]; then
        echo "✅ Database $db_name created/verified"
    else
        echo "❌ Failed to create database $db_name"
        exit 1
    fi
}

# Create databases
echo "Step 1: Creating databases..."
echo "-----------------------------------"
create_database "payment_merchant_policy"
create_database "payment_merchant_quota"
echo ""

# Grant privileges
echo "Step 2: Granting privileges..."
echo "-----------------------------------"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "GRANT ALL PRIVILEGES ON DATABASE payment_merchant_policy TO $DB_USER"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "GRANT ALL PRIVILEGES ON DATABASE payment_merchant_quota TO $DB_USER"
echo "✅ Privileges granted"
echo ""

# Verify databases
echo "Step 3: Verifying databases..."
echo "-----------------------------------"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -l | grep payment_merchant_policy
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -l | grep payment_merchant_quota
echo ""

echo "=================================================="
echo "✅ Database initialization complete!"
echo "=================================================="
echo ""
echo "Databases created:"
echo "  1. payment_merchant_policy (port: 40012)"
echo "  2. payment_merchant_quota  (port: 40022)"
echo ""
echo "Next steps:"
echo "  1. Start merchant-policy-service  → cd backend/services/merchant-policy-service && go run cmd/main.go"
echo "  2. Start merchant-quota-service   → cd backend/services/merchant-quota-service && go run cmd/main.go"
echo "  3. Auto-migrate will create tables on first startup"
echo ""
