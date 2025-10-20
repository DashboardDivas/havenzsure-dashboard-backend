-- +goose Up
-- +goose StatementBegin

-- 1) Ensure pgcrypto extension
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 2) Required ENUM types
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role_enum') THEN
    CREATE TYPE app.role_enum AS ENUM ('superadmin','admin','adjuster','bodyman');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'work_order_status') THEN
    CREATE TYPE app.work_order_status AS ENUM (
      'waiting_for_inspection',
      'in_process',
      'follow_up_required',
      'completed',
      'waiting_for_information'
    );
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'image_source') THEN
    CREATE TYPE app.image_source AS ENUM ('upload','scanner');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'service_location') THEN
    CREATE TYPE app.service_location AS ENUM ('MOBILE','IN_SHOP');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'quote_item_mode') THEN
    CREATE TYPE app.quote_item_mode AS ENUM ('FIXED','SKIP');
  END IF;
END$$;

-- 3) Trigger function for updated_at
CREATE OR REPLACE FUNCTION app.set_updated_at() RETURNS trigger
AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql
  SET search_path = app, pg_temp;

-- ===== MASTER / REFERENCE TABLES =====

CREATE TABLE IF NOT EXISTS app.users (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email         text UNIQUE NOT NULL,
  full_name     text NOT NULL,
  phone         text,
  password_hash text NOT NULL,
  role          app.role_enum NOT NULL,
  is_active     boolean NOT NULL DEFAULT true,
  shop_id       uuid,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_users_shop FOREIGN KEY (shop_id)
    REFERENCES app.shops(id) ON DELETE SET NULL,
  CONSTRAINT ck_users_email CHECK (email ~* '^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}$'),
  CONSTRAINT ck_users_phone CHECK (phone IS NULL OR phone ~* '^[0-9]{3}-[0-9]{3}-[0-9]{4}$')
);

CREATE OR REPLACE TRIGGER trg_users_updated_at
BEFORE UPDATE ON app.users
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

CREATE TABLE IF NOT EXISTS app.customers (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  first_name  text NOT NULL,
  last_name   text,
  email       text,
  phone       text,
  address     text,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_customers_email CHECK (email IS NULL OR email ~* '^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}$'),
  CONSTRAINT ck_customers_phone CHECK (phone IS NULL OR phone ~* '^[0-9]{3}-[0-9]{3}-[0-9]{4}$')
);

CREATE OR REPLACE TRIGGER trg_customers_updated_at
BEFORE UPDATE ON app.customers
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

CREATE TABLE IF NOT EXISTS app.vehicles (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  plate          text,
  make           text,
  model          text,
  year           int,
  vin            text,
  color          text,
  body_style     text,
  owner_id       uuid,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_vehicle_owner FOREIGN KEY (owner_id)
    REFERENCES app.customers(id) ON DELETE SET NULL,
  CONSTRAINT ck_vehicle_vin CHECK (vin IS NULL OR vin ~* '^[A-HJ-NPR-Z0-9]{17}$'),
  CONSTRAINT ck_vehicle_year CHECK (year BETWEEN 1900 AND EXTRACT(YEAR FROM NOW()) + 1)
);

CREATE OR REPLACE TRIGGER trg_vehicles_updated_at
BEFORE UPDATE ON app.vehicles
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

-- ===== WORK ORDER CORE =====
CREATE TABLE IF NOT EXISTS app.work_orders (
  id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code                 text UNIQUE NOT NULL,
  status               app.work_order_status NOT NULL DEFAULT 'waiting_for_inspection',
  date_received        timestamptz NOT NULL,
  date_updated         timestamptz NOT NULL DEFAULT now(),
  shop_id              uuid,
  customer_id          uuid,
  vehicle_id           uuid,
  created_by           uuid,
  technician_id        uuid,
  pre_authorized_dispatch boolean NOT NULL DEFAULT false,
  note                 text,
  created_at           timestamptz NOT NULL DEFAULT now(),
  updated_at           timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_work_orders_shop FOREIGN KEY (shop_id)
    REFERENCES app.shops(id) ON DELETE SET NULL,
  CONSTRAINT fk_work_orders_customer FOREIGN KEY (customer_id)
    REFERENCES app.customers(id) ON DELETE SET NULL,
  CONSTRAINT fk_work_orders_vehicle FOREIGN KEY (vehicle_id)
    REFERENCES app.vehicles(id) ON DELETE SET NULL,
  CONSTRAINT fk_work_orders_created_by FOREIGN KEY (created_by)
    REFERENCES app.users(id) ON DELETE SET NULL,
  CONSTRAINT fk_work_orders_technician FOREIGN KEY (technician_id)
    REFERENCES app.users(id) ON DELETE SET NULL,
  CONSTRAINT ck_work_orders_date CHECK (updated_at >= created_at)
);
CREATE INDEX IF NOT EXISTS idx_work_orders_status ON app.work_orders(status);
CREATE INDEX IF NOT EXISTS idx_work_orders_customer ON app.work_orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_vehicle ON app.work_orders(vehicle_id);

CREATE OR REPLACE TRIGGER trg_work_orders_updated_at
BEFORE UPDATE ON app.work_orders
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

-- ===== MEDIA =====
CREATE TABLE IF NOT EXISTS app.work_order_media (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  work_order_id    uuid NOT NULL,
  source           app.image_source NOT NULL,
  original_name    text,
  mime_type        text,
  url              text,
  storage_path     text,
  width_px         int,
  height_px        int,
  size_bytes       bigint,
  sort_order       int DEFAULT 0,
  uploaded_by      uuid,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_media_work_order FOREIGN KEY (work_order_id)
    REFERENCES app.work_orders(id) ON DELETE CASCADE,
  CONSTRAINT fk_media_uploaded_by FOREIGN KEY (uploaded_by)
    REFERENCES app.users(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_media_wo ON app.work_order_media(work_order_id, sort_order);

CREATE OR REPLACE TRIGGER trg_work_order_media_updated_at
BEFORE UPDATE ON app.work_order_media
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

-- ===== QUOTE =====
CREATE TABLE IF NOT EXISTS app.quotes (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  work_order_id    uuid UNIQUE NOT NULL,
  technician_id    uuid,
  shop_id          uuid,
  total_estimated  numeric(12,2),
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_quotes_work_order FOREIGN KEY (work_order_id)
    REFERENCES app.work_orders(id) ON DELETE CASCADE,
  CONSTRAINT fk_quotes_technician FOREIGN KEY (technician_id)
    REFERENCES app.users(id) ON DELETE SET NULL,
  CONSTRAINT fk_quotes_shop FOREIGN KEY (shop_id)
    REFERENCES app.shops(id) ON DELETE SET NULL,
  CONSTRAINT ck_quotes_total_nonnegative CHECK (total_estimated IS NULL OR total_estimated >= 0)
);

CREATE OR REPLACE TRIGGER trg_quotes_updated_at
BEFORE UPDATE ON app.quotes
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

CREATE TABLE IF NOT EXISTS app.quote_items (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  quote_id         uuid NOT NULL,
  media_id         uuid,
  image_label      text,
  size_text        text,
  mode             app.quote_item_mode NOT NULL,
  estimated_charge numeric(12,2) NOT NULL DEFAULT 0,
  sort_order       int DEFAULT 0,
  CONSTRAINT fk_quote_items_quote FOREIGN KEY (quote_id)
    REFERENCES app.quotes(id) ON DELETE CASCADE,
  CONSTRAINT fk_quote_items_media FOREIGN KEY (media_id)
    REFERENCES app.work_order_media(id) ON DELETE RESTRICT,
  CONSTRAINT ck_quote_items_charge_nonnegative CHECK (estimated_charge >= 0)
);
CREATE INDEX IF NOT EXISTS idx_quote_items_quote ON app.quote_items(quote_id, sort_order);

-- ===== CLAIM =====
CREATE TABLE IF NOT EXISTS app.claims (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  work_order_id      uuid UNIQUE NOT NULL,
  insurance_claimed  boolean NOT NULL DEFAULT false,
  claim_approved     boolean,
  claim_number       text,
  note               text,
  updated_by         uuid,
  updated_at         timestamptz NOT NULL DEFAULT now(),
  created_at         timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_claims_work_order FOREIGN KEY (work_order_id)
    REFERENCES app.work_orders(id) ON DELETE CASCADE,
  CONSTRAINT fk_claims_updated_by FOREIGN KEY (updated_by)
    REFERENCES app.users(id) ON DELETE SET NULL
);

-- ===== PAYMENT & DISPATCH =====
CREATE TABLE IF NOT EXISTS app.work_order_payments (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  work_order_id      uuid NOT NULL,
  is_completed       boolean NOT NULL DEFAULT false,
  note               text,
  amount             numeric(12,2),
  created_by         uuid,
  created_at         timestamptz NOT NULL DEFAULT now(),
  updated_at         timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_payments_work_order FOREIGN KEY (work_order_id)
    REFERENCES app.work_orders(id) ON DELETE CASCADE,
  CONSTRAINT fk_payments_created_by FOREIGN KEY (created_by)
    REFERENCES app.users(id) ON DELETE SET NULL,
  CONSTRAINT ck_payments_amount_nonnegative CHECK (amount IS NULL OR amount >= 0)
);
CREATE INDEX IF NOT EXISTS idx_payments_wo ON app.work_order_payments(work_order_id);

CREATE OR REPLACE TRIGGER trg_work_order_payments_updated_at
BEFORE UPDATE ON app.work_order_payments
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

CREATE TABLE IF NOT EXISTS app.work_order_dispatch (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  work_order_id      uuid UNIQUE NOT NULL,
  scheduled_at       timestamptz,
  service_location   app.service_location,
  note               text,
  updated_by         uuid,
  updated_at         timestamptz NOT NULL DEFAULT now(),
  created_at         timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_dispatch_work_order FOREIGN KEY (work_order_id)
    REFERENCES app.work_orders(id) ON DELETE CASCADE,
  CONSTRAINT fk_dispatch_updated_by FOREIGN KEY (updated_by)
    REFERENCES app.users(id) ON DELETE SET NULL
);

-- ===== ACTIVITY LOG =====
CREATE TABLE IF NOT EXISTS app.activity_logs (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  work_order_id  uuid NOT NULL,
  occurred_at    timestamptz NOT NULL DEFAULT now(),
  user_id        uuid,
  action         text NOT NULL,
  details        text,
  CONSTRAINT fk_logs_work_order FOREIGN KEY (work_order_id)
    REFERENCES app.work_orders(id) ON DELETE CASCADE,
  CONSTRAINT fk_logs_user FOREIGN KEY (user_id)
    REFERENCES app.users(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_logs_wo_time ON app.activity_logs(work_order_id, occurred_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS app.activity_logs;
DROP TABLE IF EXISTS app.work_order_dispatch;
DROP TABLE IF EXISTS app.work_order_payments;
DROP TABLE IF EXISTS app.claims;
DROP TABLE IF EXISTS app.quote_items;
DROP TABLE IF EXISTS app.quotes;
DROP TABLE IF EXISTS app.work_order_media;
DROP TABLE IF EXISTS app.work_orders;
DROP TABLE IF EXISTS app.vehicles;
DROP TABLE IF EXISTS app.customers;
DROP TABLE IF EXISTS app.users;

DROP FUNCTION IF EXISTS app.set_updated_at();
DROP TYPE IF EXISTS app.quote_item_mode;
DROP TYPE IF EXISTS app.service_location;
DROP TYPE IF EXISTS app.image_source;
DROP TYPE IF EXISTS app.work_order_status;
DROP TYPE IF EXISTS app.role_enum;

-- +goose StatementEnd