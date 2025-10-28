-- +goose Up
-- +goose StatementBegin

-- This table is for Work Order Intake Form.

-- Purpose: Create normalized tables for Work Order intake flow.
-- Scope: structure only (types, defaults, NOT NULL). No validation/normalization here.
-- Notes: Constraints (regex, ranges, UNIQUE, FKs, indexes) live in align_*_constraints.sql

-- in this version, work orders are linked to shops (shop_id), customers (customer_id), vehicles (vin), users (created_by_user_id), insurance (insurance_id)
-- insurance record is nullable for now as not all work orders will have insurance associated
CREATE EXTENSION IF NOT EXISTS PGCRYPTO;

CREATE EXTENSION IF NOT EXISTS CITEXT;

-- 0) Enum type
CREATE TYPE app.work_order_status AS ENUM ('waiting_for_inspection','completed','in_progress','follow_up_needed','awaiting_info');

-- 1) Main table for work orders
CREATE TABLE app.work_orders (
-- 1.1: work order identifier
  work_order_id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  work_order_code              TEXT NOT NULL, -- human-readable code for work order (e.g., WO-0001)
                                              -- should be unique but needs final decision on how it's uniqueness is enforced (per shop? global?)

-- 1.2: tables associations
  customer_id       UUID NOT NULL, -- FK → app.customers(customer_id)
  shop_id           UUID NOT NULL, -- FK → app.shops(shop_id)
  vehicle_id        UUID NOT NULL, -- FK → app.vehicles(vehicle_id)
--     TODO: apply user
  created_by_user_id UUID, -- FK → app.users(user_id)
  --   TODO: Needs final approval on the situation where assignee creates work order on their own
  --   TODO: As we allow reassignment, should we track assignment, assignment history as seperate table? 
  --   assignment_id UUID, -- FK to users table; should not be nullable

-- 1.3: work order details
  status            app.work_order_status NOT NULL DEFAULT 'waiting_for_inspection',
  damage_date       TIMESTAMPTZ,

-- 1.4: timestamps
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2) Customer
Create table app.customers (
    -- 2.1: customer identifier
    customer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 2.2: customer details
    first_name   TEXT NOT NULL,
    last_name    TEXT NOT NULL,
    address      TEXT NOT NULL,
    city         TEXT NOT NULL,
    postal_code  TEXT NOT NULL,
    province     TEXT NOT NULL, --how come in frontend it's state? we are not US based
    zip          TEXT NOT NULL,
    email        CITEXT,
    phone        TEXT
);

-- 3) Vehicle
Create table app.vehicles (
    -- 3.1: vehicle identifier
    vehicle_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 3.2: vehicle details
    make         TEXT NOT NULL,
    model        TEXT NOT NULL,
    body_style   TEXT,
    model_year   SMALLINT NOT NULL,
    vin          CITEXT NOT NULL,
    color        TEXT,
    plate_number TEXT --for some reason this is nullable in frontend
);

-- 4) Insurance info
Create table app.insurance_info (
    work_order_id UUID PRIMARY KEY,
    insurance_company TEXT NOT NULL,
    agent_first_name TEXT NOT NULL,
    agent_last_name TEXT NOT NULL,
    agent_phone TEXT NOT NULL,
    policy_number TEXT NOT NULL,
    claim_number TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop tables in dependency-safe order
-- 1) insurance_info depends on work_orders
DROP TABLE IF EXISTS app.insurance_info CASCADE;

-- 2) work_orders depends on customers, vehicles, shops, users
DROP TABLE IF EXISTS app.work_orders CASCADE;

-- 3) vehicles and customers are independent
DROP TABLE IF EXISTS app.vehicles CASCADE;
DROP TABLE IF EXISTS app.customers CASCADE;

-- 4) enum type
DROP TYPE IF EXISTS app.work_order_status;
-- +goose StatementEnd
