# Database Setup Guide

This guide explains how to start the local Postgres container, create roles and the project database, run migrations for the schema, and verify the setup.

---

## 1. Start Postgres with Docker Compose

From the project root, run:

```bash
docker compose up -d
```

This starts a Postgres container (`havenzsure-postgres`).

Confirm it’s running:

```bash
docker ps
```

You should see something like:

```
0.0.0.0:5433->5432/tcp
```

---

## 2. Connect with pgAdmin (superuser)

1. Open **pgAdmin**
2. Right-click **Servers → Register → Server**
3. Fill in the details:

```
Name: Local Container
Host: <DB_HOST>
Port: <DB_PORT>
Maintenance DB: <POSTGRES_DB>
Username: <POSTGRES_USER>
Password: <POSTGRES_PASSWORD>
```

At this point the **database server is already running** inside the container. You are now connected as the `postgres` superuser.

---

## 3. Create Roles and Database

With the superuser connection open (database = `postgres`), open **Query Tool** and run:

```sql
CREATE ROLE app_owner LOGIN PASSWORD '<DB_OWNER_PASSWORD>';
CREATE ROLE app_user  LOGIN PASSWORD '<DB_APP_PASSWORD>';
CREATE DATABASE havenzsure OWNER app_owner;
```

> Replace `<owner_password>` and `<user_password>` with secure values.

---

## 4. Run Migrations (Schema Changes)

Install migration tool Goose

```go
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Apply migrations with Goose:

```bash
goose -dir ./migrations postgres "host=<DB_HOST> port=<DB_PORT> user=<DB_OWNER_USER> password=<DB_OWNER_PASSWORD> dbname=<DB_NAME> sslmode=disable" up
```

This will create tables and other objects defined in migrations folder.

---

## 5. Verify with PgAdmin

Connect again in pgAdmin, this time as `app_owner`:

```
Name: Havenzsure-AppOwner
Host: <DB_HOST>
Port: <DB_PORT>
Maintenance DB: <DB_NAME>
Username: <DB_OWNER_USER>
Password: <DB_OWNER_PASSWORD>
```

Expand the database tree — you should see the `app` schema in the havanzsure db.

Check search path:

```sql
SHOW search_path;
```

Expected result:

```
app, public
```

---

## Summary

- **Run container** → `docker compose up -d`
- **Create roles & DB** → simple SQL (once, with placeholders for passwords)
- **Run migrations** → Goose manages schema changes
- **Verify with PgAdmin** → confirm `app` schema

---

## Resetting the Database (if needed)

```bash
docker compose down -v
docker compose up -d
# Recreate roles & database manually (step 3)
# Then reapply schema migrations
goose -dir ./migrations postgres "host=<DB_HOST> port=<DB_PORT> user=<DB_OWNER_USER> password=<DB_OWNER_PASSWORD> dbname=<DB_NAME> sslmode=disable" up
```
