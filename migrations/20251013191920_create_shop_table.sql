-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS PGCRYPTO;

CREATE EXTENSION IF NOT EXISTS CITEXT;

-- 1) Enum type
CREATE TYPE app.shop_status AS ENUM ('active','inactive');

-- 2) Main table 
CREATE TABLE app.shop (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code         VARCHAR(10)  NOT NULL CONSTRAINT uq_shop_code UNIQUE, 
  shop_name    TEXT NOT NULL,
  status       app.shop_status NOT NULL DEFAULT 'active',
  address      TEXT NOT NULL,        -- street address
  city         TEXT NOT NULL,
  province     CHAR(2)      NOT NULL,     -- AB/BC/ON...
  postal_code  TEXT      NOT NULL,        -- T2P2B5
  contact_name TEXT NOT NULL,
  phone        TEXT NOT NULL,        -- 403-555-1234 
  email        CITEXT NOT NULL,      -- regular email format (case-insensitive)

  created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT ck_shop_province CHECK (province IN
    ('AB','BC','MB','NB','NL','NT','NS','NU','ON','PE','QC','SK','YT')),
  
  CONSTRAINT ck_shop_postal CHECK (postal_code ~ '^[A-Z][0-9][A-Z][0-9][A-Z][0-9]$'),

  CONSTRAINT ck_shop_phone
    CHECK (phone ~ '^[0-9]{3}-[0-9]{3}-[0-9]{4}$'),  

  CONSTRAINT ck_shop_email
    CHECK (
      email ~* '^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$' 
    )
);

-- 3) Trigger to auto-update updated_at on row modification
CREATE OR REPLACE TRIGGER trg_shop_updated_at
BEFORE UPDATE ON app.shop
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

-- 4) Normalize 
CREATE OR REPLACE FUNCTION app.shop_normalize()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.province := UPPER(NEW.province);
  NEW.postal_code := UPPER(REPLACE(NEW.postal_code, ' ', ''));
  NEW.code := UPPER(NEW.code);
  NEW.email := TRIM(NEW.email);
  RETURN NEW;
END;
$$;

CREATE OR REPLACE TRIGGER trg_shop_normalize
BEFORE INSERT OR UPDATE ON app.shop
FOR EACH ROW
EXECUTE FUNCTION app.shop_normalize();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove trigger and function
DROP TRIGGER IF EXISTS trg_shop_normalize ON app.shop;
DROP FUNCTION IF EXISTS app.shop_normalize;

DROP TRIGGER IF EXISTS trg_shop_updated_at ON app.shop;

--Remove table and type
DROP TABLE IF EXISTS app.shop;
DROP TYPE IF EXISTS app.shop_status;

-- +goose StatementEnd