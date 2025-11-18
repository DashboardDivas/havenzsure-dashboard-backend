#!/bin/bash
# Execute database migrations using goose
# Citation: Generated with assistance from Claude (Anthropic), reviewed and modified by AN-NI HUANG
# Date: 2025-11-13
#
# Supports special characters in passwords (including @, !, $, %, etc.)
# Usage:
#   bash scripts/migrate.sh       # Run migrations on development database
#   bash scripts/migrate.sh test  # Run migrations on test database
#
# Requirements:
#   - goose installed: go install github.com/pressly/goose/v3/cmd/goose@latest
#   - .env file with required variables (see .env.example)
#   - PostgreSQL container running (docker compose up -d)
#   - Database roles and database already created (bash scripts/setup-db.sh)

set -e

# ============================================================================
# ENVIRONMENT SETUP
# ============================================================================

# Check for .env file
if [ ! -f ".env" ]; then
  echo "Error: .env file not found"
  echo "Please create .env by copying .env.example and filling in your values"
  exit 1
fi

# Load environment variables from .env
set -a
source .env
set +a

# ============================================================================
# VALIDATION
# ============================================================================

# Check for required environment variables
REQUIRED_VARS=("DB_OWNER_USER" "DB_OWNER_PASSWORD" "DB_NAME")
MISSING_VARS=""

for var in "${REQUIRED_VARS[@]}"; do
  if [ -z "${!var}" ]; then
    MISSING_VARS="$MISSING_VARS\n  - $var"
  fi
done

if [ ! -z "$MISSING_VARS" ]; then
  echo "Error: Required variables not set in .env"
  echo -e "Missing variables:$MISSING_VARS"
  exit 1
fi

# ============================================================================
# PORT DETERMINATION
# ============================================================================

# Determine port based on environment
# Use DB_TEST_PORT for test (default 5434), DB_PORT for development (default 5433)
if [ "$1" = "test" ]; then
  PORT="${DB_TEST_PORT:-5434}"
  ENV_NAME="test"
else
  PORT="${DB_PORT:-5433}"
  ENV_NAME="development"
fi

# ============================================================================
# MIGRATION EXECUTION
# ============================================================================

# Pass password via PGPASSWORD environment variable instead of URL
# This approach safely handles special characters like @ ! $ % etc.
# without URL encoding issues
export PGPASSWORD="$DB_OWNER_PASSWORD"

# Build database connection string without password in URL
DATABASE_URL="postgres://${DB_OWNER_USER}@localhost:${PORT}/${DB_NAME}?sslmode=disable"

echo "=================================================="
echo "Database Migration"
echo "=================================================="
echo "Environment: $ENV_NAME"
echo "Host: localhost"
echo "Port: $PORT"
echo "Database: $DB_NAME"
echo "User: $DB_OWNER_USER"
echo "=================================================="
echo ""

# Execute goose migrations
# Password is passed via PGPASSWORD environment variable for security
if goose -dir ./migrations postgres "$DATABASE_URL" up; then
  echo ""
  echo "✓ Migrations completed successfully"
else
  echo ""
  echo "✗ Migration failed"
  unset PGPASSWORD
  exit 1
fi

# Clean up environment variable
unset PGPASSWORD

echo "Done!"