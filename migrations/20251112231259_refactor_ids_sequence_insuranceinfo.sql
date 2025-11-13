-- +goose Up
-- +goose StatementBegin

-- 3 things to do in this migration:
-- 1)Change the column name for workorder customer and vehicle (*_id -> id, *_code -> code, )
--      1.1)Update all foreign keys (ids) and indexes to match new column names
        -- Old FK reference:
            -- ALTER TABLE app.insurance_info
            -- ADD CONSTRAINT fk_insurance_work_order
            -- FOREIGN KEY (work_order_id) REFERENCES app.work_orders(work_order_id)
            -- ON DELETE CASCADE;
            -- ALTER TABLE app.work_orders
            -- ADD CONSTRAINT fk_wo_customer FOREIGN KEY (customer_id) 
            --     REFERENCES app.customers(customer_id) ON DELETE CASCADE,
            -- ADD CONSTRAINT fk_wo_shop     FOREIGN KEY (shop_id)     
            --     REFERENCES app.shop(id) ON DELETE CASCADE,
            -- ADD CONSTRAINT fk_wo_vehicle  FOREIGN KEY (vehicle_id)  
            --     REFERENCES app.vehicles(vehicle_id) ON DELETE CASCADE;
-- 2)Ad id to insurance_info table as primary key instead of work_order_id
--      2.1)Update the table name to insurance from insurance_info as it seems meaningless to have "info" semantically

--Drop FKs depending on columns to be renamed
ALTER TABLE app.insurance_info
  DROP CONSTRAINT IF EXISTS fk_insurance_work_order;

ALTER TABLE app.work_orders
    DROP CONSTRAINT IF EXISTS fk_wo_customer,
    DROP CONSTRAINT IF EXISTS fk_wo_vehicle;

--Rename columns in work_orders table
ALTER TABLE app.customers
    RENAME COLUMN customer_id TO id;

ALTER TABLE app.vehicles
    RENAME COLUMN vehicle_id TO id;

ALTER TABLE app.work_orders
    RENAME COLUMN work_order_id TO id;
ALTER TABLE app.work_orders
    RENAME COLUMN work_order_code TO code;

ALTER TABLE app.insurance_info
    Drop CONSTRAINT IF EXISTS insurance_info_pkey,
    ADD COLUMN id UUID DEFAULT gen_random_uuid();

ALTER TABLE app.insurance_info
    ALTER COLUMN id SET NOT NULL;

ALTER TABLE app.insurance_info
    ADD CONSTRAINT pk_insurance PRIMARY KEY (id);

ALTER TABLE app.insurance_info
    ADD CONSTRAINT fk_insurance_work_order
        FOREIGN KEY (work_order_id) 
        REFERENCES app.work_orders(id)
        ON DELETE CASCADE;

ALTER TABLE app.work_orders
  ADD CONSTRAINT fk_wo_customer
    FOREIGN KEY (customer_id)
    REFERENCES app.customers(id)
    ON DELETE CASCADE,
  ADD CONSTRAINT fk_wo_vehicle
    FOREIGN KEY (vehicle_id)
    REFERENCES app.vehicles(id)
    ON DELETE CASCADE;

ALTER TABLE app.insurance_info
  RENAME TO insurance;

CREATE OR REPLACE FUNCTION app.normalize_work_orders_row()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.code := trim(NEW.code);
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION app.work_order_set_code()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF coalesce(btrim(NEW.code), '') = '' THEN
    NEW.code := app.generate_work_order_code();
  END IF;
  RETURN NEW;
END;
$$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- use NEW.work_order_code in normalize_work_orders_row
CREATE OR REPLACE FUNCTION app.normalize_work_orders_row()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.work_order_code := trim(NEW.work_order_code);
  RETURN NEW;
END;
$$;

--use NEW.work_order_code in work_order_set_code
CREATE OR REPLACE FUNCTION app.work_order_set_code()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF coalesce(btrim(NEW.work_order_code), '') = '' THEN
    NEW.work_order_code := app.generate_work_order_code();
  END IF;
  RETURN NEW;
END;
$$;

-- Rename table back
ALTER TABLE app.insurance
  RENAME TO insurance_info;

-- Drop new FKs on work_orders
ALTER TABLE app.work_orders
  DROP CONSTRAINT IF EXISTS fk_wo_customer,
  DROP CONSTRAINT IF EXISTS fk_wo_vehicle;

-- Drop FK + PK on insurance_info, remove id column
ALTER TABLE app.insurance_info
  DROP CONSTRAINT IF EXISTS fk_insurance_work_order,
  DROP CONSTRAINT IF EXISTS pk_insurance;

ALTER TABLE app.insurance_info
  DROP COLUMN IF EXISTS id;

-- Restore original PK on work_order_id
ALTER TABLE app.insurance_info
  ADD CONSTRAINT insurance_info_pkey PRIMARY KEY (work_order_id);

-- Rename columns back
ALTER TABLE app.work_orders
  RENAME COLUMN id TO work_order_id;

ALTER TABLE app.work_orders
  RENAME COLUMN code TO work_order_code;

ALTER TABLE app.customers
  RENAME COLUMN id TO customer_id;

ALTER TABLE app.vehicles
  RENAME COLUMN id TO vehicle_id;

-- Restore original FKs
ALTER TABLE app.insurance_info
  ADD CONSTRAINT fk_insurance_work_order
    FOREIGN KEY (work_order_id)
    REFERENCES app.work_orders(work_order_id)
    ON DELETE CASCADE;

ALTER TABLE app.work_orders
  ADD CONSTRAINT fk_wo_customer
    FOREIGN KEY (customer_id)
    REFERENCES app.customers(customer_id)
    ON DELETE CASCADE,
  ADD CONSTRAINT fk_wo_vehicle
    FOREIGN KEY (vehicle_id)
    REFERENCES app.vehicles(vehicle_id)
    ON DELETE CASCADE;

-- +goose StatementEnd
