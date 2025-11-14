# Database Setup Guide

This guide explains how to set up local Postgres containers for both development and testing, create roles and databases, run migrations, and verify the setup.

---

## Quick Start (Recommended)

```bash
# 1. Start both containers
docker compose up -d
docker compose -f compose.test.yml up -d

# 2. Setup and migrate both databases
bash scripts/setup-db.sh dev && bash scripts/migrate.sh
bash scripts/setup-db.sh test && bash scripts/migrate.sh test
```

Done! Your databases are ready. ✓

---

## Development Database Setup

### 1. Start Container

```bash
docker compose up -d
```

Verify it's running:

```bash
docker ps | grep havenzsure-postgres
```

### 2. Create Roles and Database

**Using automated script (Recommended):**

```bash
bash scripts/setup-db.sh dev
```

**Or manually via pgAdmin:**

1. Connect as superuser (`postgres`)
2. Run in Query Tool:
   ```sql
   CREATE ROLE <DB_OWNER_USER> LOGIN PASSWORD '<DB_OWNER_PASSWORD>';
   CREATE ROLE <DB_APP_USER>  LOGIN PASSWORD '<DB_APP_PASSWORD>';
   CREATE DATABASE <DB_NAME> OWNER <DB_OWNER_USER>;
   ```

### 3. Run Migrations

```bash
bash scripts/migrate.sh
```

### 4. Verify with pgAdmin

Connect as `<DB_OWNER_USER>` in pgAdmin:

```
Name: Havenzsure-AppOwner
Host: <DB_HOST>
Port: <DB_PORT>
Database: <DB_NAME>
Username: <DB_OWNER_USER>
Password: <DB_OWNER_PASSWORD>
```

Check schema exists:

```sql
SHOW search_path;
-- Expected: <DB_SCHEMA>, public
```

---

## Test Database Setup

The test database is isolated and runs on a separate port (5434 by default).

### 1. Start Container

```bash
docker compose -f compose.test.yml up -d
```

Verify it's running:

```bash
docker ps | grep havenzsure-postgres-test
```

### 2. Create Roles and Database

```bash
bash scripts/setup-db.sh test
```

### 3. Run Migrations

```bash
bash scripts/migrate.sh test
```

---

## Application Connection

The application and all tests connect using `<DB_APP_USER>` to enforce the principle of least privilege:

- **Development & Testing**: Application code uses `<DB_APP_USER>` (limited permissions)
- **Migrations**: Only migrations use `<DB_OWNER_USER>` (schema changes)
- **Production**: Only `<DB_APP_USER>` credentials are used

This ensures the application works correctly with production-level permissions.

---

## Running Migrations After Schema Changes

When new migrations are added to the repository:

```bash
# Development database
bash scripts/migrate.sh

# Test database
bash scripts/migrate.sh test
```

---

## Verification

### Quick Verification - Development Database

Check development database is set up correctly:

```bash
# Check roles exist in development
docker exec -e PGPASSWORD="<POSTGRES_PASSWORD>" havenzsure-postgres psql -U <POSTGRES_USER> -d postgres -c "\du"

# Check tables were created
docker exec -e PGPASSWORD="<DB_OWNER_PASSWORD>" havenzsure-postgres psql -U <DB_OWNER_USER> -d <DB_NAME> -c "\dt <DB_SCHEMA>.*"

# Verify app_user can read tables
docker exec -e PGPASSWORD="<DB_APP_PASSWORD>" havenzsure-postgres psql -U <DB_APP_USER> -d <DB_NAME> -c "SELECT COUNT(*) FROM <DB_SCHEMA>.shop;"
```

### Quick Verification - Test Database

Check test database is set up correctly:

```bash
# Check tables were created in test database
docker exec -e PGPASSWORD="<DB_OWNER_PASSWORD>" havenzsure-postgres-test psql -U <DB_OWNER_USER> -d <DB_NAME> -c "\dt <DB_SCHEMA>.*"

# Verify app_user can read tables (for integration tests)
docker exec -e PGPASSWORD="<DB_APP_PASSWORD>" havenzsure-postgres-test psql -U <DB_APP_USER> -d <DB_NAME> -c "SELECT COUNT(*) FROM <DB_SCHEMA>.shop;"
```

### Verify Database Owner

```bash
# Development
docker exec -e PGPASSWORD="<POSTGRES_PASSWORD>" havenzsure-postgres psql -U <POSTGRES_USER> -d postgres -c "\l <DB_NAME>"

# Test
docker exec -e PGPASSWORD="<POSTGRES_PASSWORD>" havenzsure-postgres-test psql -U <POSTGRES_USER> -d postgres -c "\l <DB_NAME>"
```

Both should show `<DB_OWNER_USER>` as the owner.

---

## Resetting Databases

### Reset Development Database

```bash
docker compose down -v
docker compose up -d
bash scripts/setup-db.sh dev
bash scripts/migrate.sh
```

### Reset Test Database

```bash
docker compose -f compose.test.yml down -v
docker compose -f compose.test.yml up -d
bash scripts/setup-db.sh test
bash scripts/migrate.sh test
```

### Reset Development Database

If you need to reset the development database (including volumes):

```bash
docker compose down -v
docker compose up -d
bash scripts/setup-db.sh dev
bash scripts/migrate.sh
```

### Stop Test Database (Keep Development Running)

If you want to stop only the test database:

```bash
docker compose -f compose.test.yml down -v
```

The development database will continue running.

---

## Environment Variables (.env)

Required variables in `.env`:

```env
# PostgreSQL Superuser
POSTGRES_USER=postgres
POSTGRES_PASSWORD=<superuser_password>
POSTGRES_DB=postgres

# Database
DB_HOST=localhost
DB_PORT=5433
DB_TEST_PORT=5434
DB_NAME=<database_name>
DB_SCHEMA=<schema_name>

# Database Roles
DB_OWNER_USER=<owner_username>
DB_OWNER_PASSWORD=<owner_password>
DB_APP_USER=<app_username>
DB_APP_PASSWORD=<app_password>

# Containers (optional - defaults shown)
DEV_CONTAINER_NAME=havenzsure-postgres
TEST_CONTAINER_NAME=havenzsure-postgres-test
```

---

## Troubleshooting

### Container Not Running

```bash
# Check running containers
docker ps | grep havenzsure

# Start development container
docker compose up -d

# Start test container
docker compose -f compose.test.yml up -d
```

### Setup Failed

If `setup-db.sh` fails:

1. Ensure container is running: `docker ps`
2. Verify .env file exists and is correct: `cat .env`
3. Check required variables: `grep -E "(DB_|POSTGRES_)" .env`
4. Verify database doesn't already exist: `docker exec ... psql -l | grep <DB_NAME>`

### Migration Failed

If `migrate.sh` fails:

1. Ensure database was created: `bash scripts/setup-db.sh [dev|test]`
2. Check goose is installed: `goose --version`
3. Verify connection to database: `docker exec -e PGPASSWORD="..." havenzsure-postgres psql -U <user> -d <db> -c "SELECT 1"`
4. Check migration files exist: `ls -la migrations/`

---

## Command Reference

| Task                     | Command                                                                                                                                                   |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Development Database** |                                                                                                                                                           |
| Start dev DB             | `docker compose up -d`                                                                                                                                    |
| Setup dev                | `bash scripts/setup-db.sh dev`                                                                                                                            |
| Migrate dev              | `bash scripts/migrate.sh`                                                                                                                                 |
| Reset dev                | `docker compose down -v && docker compose up -d && bash scripts/setup-db.sh dev && bash scripts/migrate.sh`                                               |
| **Test Database**        |                                                                                                                                                           |
| Start test DB            | `docker compose -f compose.test.yml up -d`                                                                                                                |
| Setup test               | `bash scripts/setup-db.sh test`                                                                                                                           |
| Migrate test             | `bash scripts/migrate.sh test`                                                                                                                            |
| Stop test                | `docker compose -f compose.test.yml down -v`                                                                                                              |
| Reset test               | `docker compose -f compose.test.yml down -v && docker compose -f compose.test.yml up -d && bash scripts/setup-db.sh test && bash scripts/migrate.sh test` |

---
