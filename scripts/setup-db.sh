#!/bin/bash
# Setup database roles and database for development and testing
# Citation: Generated with assistance from Claude (Anthropic), reviewed and modified by AN-NI HUANG
# Date: 2025-11-13
#
# Usage:
#   bash scripts/setup-db.sh       # Setup development database
#   bash scripts/setup-db.sh dev   # Setup development database (explicit)
#   bash scripts/setup-db.sh test  # Setup test database
#
# Requirements:
#   - .env file with required variables 
#   - PostgreSQL container running (docker compose up -d or docker compose -f compose.test.yml up -d)
#   - Docker must be available
#
# Environment Variables (from .env):
#   Required:
#     - POSTGRES_USER: Superuser username (default: postgres)
#     - POSTGRES_PASSWORD: Superuser password
#     - DB_NAME: Database name to create
#     - DB_OWNER_USER: Owner user name
#     - DB_OWNER_PASSWORD: Password for owner user
#     - DB_APP_USER: App user name
#     - DB_APP_PASSWORD: Password for app user
#   Optional:
#     - POSTGRES_DB: Superuser's default database (default: postgres)
#     - DEV_CONTAINER_NAME: Development container name (default: havenzsure-postgres)
#     - TEST_CONTAINER_NAME: Test container name (default: havenzsure-postgres-test)

set -e 

# ============================================================================
# ARGUMENT PARSING
# ============================================================================

ENV=${1:-dev}

# Validate environment argument
if [ "$ENV" != "dev" ] && [ "$ENV" != "test" ]; then
  echo "Error: Invalid environment '$ENV'"
  echo "Usage: bash scripts/setup-db.sh [dev|test]"
  exit 1
fi

# ============================================================================
# CONTAINER SELECTION
# ============================================================================

if [ "$ENV" = "test" ]; then
  CONTAINER="${TEST_CONTAINER_NAME:-havenzsure-postgres-test}"
  echo "Setting up TEST database..."
else
  CONTAINER="${DEV_CONTAINER_NAME:-havenzsure-postgres}"
  echo "Setting up DEVELOPMENT database..."
fi

# ============================================================================
# ENVIRONMENT FILE VALIDATION
# ============================================================================

if [ ! -f ".env" ]; then
  echo "Error: .env not found"
  exit 1
fi

# Load environment variables from .env
set -a
source .env
set +a

# ============================================================================
# REQUIRED VARIABLE VALIDATION
# ============================================================================

# Check for all required environment variables
REQUIRED_VARS=("DB_OWNER_USER" "DB_OWNER_PASSWORD" "DB_APP_USER" "DB_APP_PASSWORD" "POSTGRES_PASSWORD" "DB_NAME")
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
# OPTIONAL VARIABLE DEFAULTS
# ============================================================================

# Default PostgreSQL superuser and database
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-postgres}"

# ============================================================================
# CONTAINER AVAILABILITY CHECK
# ============================================================================

# Check if container exists and is running 
CONTAINER_ID=$(docker ps -q -f "name=^${CONTAINER}$" 2>/dev/null || echo "")

if [ -z "$CONTAINER_ID" ]; then
  echo "Error: Container '$CONTAINER' is not running"
  echo ""
  if [ "$ENV" = "test" ]; then
    echo "To start the test database container, run:"
    echo "  docker compose -f compose.test.yml up -d"
  else
    echo "To start the development database container, run:"
    echo "  docker compose up -d"
  fi
  exit 1
fi

# ============================================================================
# PASSWORD ESCAPING FOR SQL
# ============================================================================

# Escape single quotes in passwords for safe SQL execution
# Converts ' to '' (SQL standard for escaping single quotes)
OWNER_PWD_ESCAPED="${DB_OWNER_PASSWORD//\'/\'\'}"
APP_PWD_ESCAPED="${DB_APP_PASSWORD//\'/\'\'}"
DB_NAME_ESCAPED="${DB_NAME//\'/\'\'}"

# ============================================================================
# DISPLAY SETUP INFORMATION
# ============================================================================

echo ""
echo "=================================================="
echo "Database Setup - $ENV"
echo "=================================================="
echo "Container: $CONTAINER"
echo "PostgreSQL User: $POSTGRES_USER"
echo "Database: $DB_NAME"
echo "Owner User: $DB_OWNER_USER"
echo "App User: $DB_APP_USER"
echo "=================================================="
echo ""

# ============================================================================
# DROP EXISTING ROLES AND DATABASE (CLEAN SLATE)
# ============================================================================

echo "Step 1: Dropping existing roles and database (if they exist)..."

# Drop app user first (has no dependencies)
echo "  - Dropping user: $DB_APP_USER"
echo "DROP ROLE IF EXISTS $DB_APP_USER;" | \
docker exec -i -e PGPASSWORD="$POSTGRES_PASSWORD" "$CONTAINER" \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1 || true

# Drop database
echo "  - Dropping database: $DB_NAME"
echo "DROP DATABASE IF EXISTS $DB_NAME_ESCAPED;" | \
docker exec -i -e PGPASSWORD="$POSTGRES_PASSWORD" "$CONTAINER" \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1 || true

# Drop owner user (drop database first to avoid ownership issues)
echo "  - Dropping user: $DB_OWNER_USER"
echo "DROP ROLE IF EXISTS $DB_OWNER_USER;" | \
docker exec -i -e PGPASSWORD="$POSTGRES_PASSWORD" "$CONTAINER" \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1 || true
# ============================================================================
# CREATE NEW USERS
# ============================================================================

echo ""
echo "Step 2: Creating new users..."

# Create owner user with LOGIN and password
echo "  - Creating user: $DB_OWNER_USER "
echo "CREATE ROLE $DB_OWNER_USER LOGIN PASSWORD '$OWNER_PWD_ESCAPED';" | \
docker exec -i -e PGPASSWORD="$POSTGRES_PASSWORD" "$CONTAINER" \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"

# Create app user with LOGIN and password
echo "  - Creating user: $DB_APP_USER "
echo "CREATE ROLE $DB_APP_USER LOGIN PASSWORD '$APP_PWD_ESCAPED';" | \
docker exec -i -e PGPASSWORD="$POSTGRES_PASSWORD" "$CONTAINER" \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"

# ============================================================================
# CREATE DATABASE
# ============================================================================

echo ""
echo "Step 3: Creating database..."

echo "  - Creating database: $DB_NAME with owner $DB_OWNER_USER"
echo "CREATE DATABASE $DB_NAME_ESCAPED OWNER $DB_OWNER_USER;" | \
docker exec -i -e PGPASSWORD="$POSTGRES_PASSWORD" "$CONTAINER" \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"

# ============================================================================
# VERIFICATION
# ============================================================================

echo ""
echo "Step 4: Verifying setup..."

# Verify users exist
echo "  - Verifying users..."
echo "\du" | \
docker exec -i -e PGPASSWORD="$POSTGRES_PASSWORD" "$CONTAINER" \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" | grep -E "($DB_OWNER_USER|$DB_APP_USER)" || true

# Verify database exists and owner
echo "  - Verifying database..."
echo "\l $DB_NAME_ESCAPED" | \
docker exec -i -e PGPASSWORD="$POSTGRES_PASSWORD" "$CONTAINER" \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" | grep "$DB_NAME_ESCAPED" || true

# ============================================================================
# COMPLETION
# ============================================================================

echo ""
echo "=================================================="
echo "✓ Setup completed successfully!"
echo "=================================================="
echo ""
echo "Next steps:"
echo "1. Run migrations: bash scripts/migrate.sh${ENV:+ $ENV}"
echo "2. Connect to database: psql -U $DB_OWNER_USER -d $DB_NAME_ESCAPED"
echo ""