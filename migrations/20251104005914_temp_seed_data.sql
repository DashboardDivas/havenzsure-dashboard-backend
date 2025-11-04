-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
  -- shops
  v_shop1 uuid;
  v_shop2 uuid;

  -- customers
  v_cust1 uuid;
  v_cust2 uuid;

  -- vehicles
  v_veh1 uuid;
  v_veh2 uuid;
BEGIN
  /* ---------- 1) app.shop ---------- */
  INSERT INTO app.shop
    (code, shop_name, status, address, city, province, postal_code, contact_name, phone, email)
  VALUES
    ('DEMO1', 'Demo Auto Body North', 'active',
     '123 Demo St', 'Calgary', 'AB', 'T2P 2B5', 'Alice Demo', '403-555-1234', 'shop-north@hvzsure.local'),
    ('DEMO2', 'Demo Auto Body West',  'active',
     '456 Sample Ave', 'Vancouver', 'BC', 'V5K 0A1', 'Bob Demo',   '604-555-6789', 'shop-west@hvzsure.local')
  ON CONFLICT (code) DO NOTHING;

  SELECT id INTO v_shop1 FROM app.shop WHERE code='DEMO1' LIMIT 1;
  SELECT id INTO v_shop2 FROM app.shop WHERE code='DEMO2' LIMIT 1;

  RAISE NOTICE 'shop DEMO1 id=%; DEMO2 id=%', v_shop1, v_shop2;

  IF v_shop1 IS NULL OR v_shop2 IS NULL THEN
    RAISE EXCEPTION 'Shop rows missing; check privileges/constraints on app.shop';
  END IF;

  /* ---------- 2) app.customers ---------- */
  INSERT INTO app.customers
    (first_name, last_name, address, city, postal_code, province, email, phone)
  VALUES
    ('Diana','Customer','789 River Rd','Calgary','T2P 2B5','AB','demo.cust1@hvzsure.local','4035551234'),
    ('Evan','Customer','321 Ocean Dr','Vancouver','V5K 0A1','BC','demo.cust2@hvzsure.local','6045557890')
  ON CONFLICT DO NOTHING;

  SELECT customer_id INTO v_cust1 FROM app.customers WHERE email='demo.cust1@hvzsure.local' LIMIT 1;
  SELECT customer_id INTO v_cust2 FROM app.customers WHERE email='demo.cust2@hvzsure.local' LIMIT 1;

  RAISE NOTICE 'cust1=%; cust2=%', v_cust1, v_cust2;

  IF v_cust1 IS NULL OR v_cust2 IS NULL THEN
    RAISE EXCEPTION 'Customer rows missing; check constraints on app.customers';
  END IF;

  /* ---------- 3) app.vehicles ---------- */
  INSERT INTO app.vehicles
    (make, model, body_style, model_year, vin, color, plate_number)
  VALUES
    ('Toyota','Corolla','Sedan', 2018, '1HGCM82633A004352', 'Blue', 'ABC123'),
    ('Honda','CR-V',   'SUV',   2022, 'JH4KA4650MC000123', 'Red',  'XYZ789')
  ON CONFLICT DO NOTHING;

  SELECT vehicle_id INTO v_veh1 FROM app.vehicles WHERE vin='1HGCM82633A004352' LIMIT 1;
  SELECT vehicle_id INTO v_veh2 FROM app.vehicles WHERE vin='JH4KA4650MC000123' LIMIT 1;

  RAISE NOTICE 'veh1=%; veh2=%', v_veh1, v_veh2;

  IF v_veh1 IS NULL OR v_veh2 IS NULL THEN
    RAISE EXCEPTION 'Vehicle rows missing; check constraints on app.vehicles';
  END IF;

  INSERT INTO app.work_orders
    (customer_id, shop_id, vehicle_id, status, damage_date, created_by_user_id)
  VALUES
    (v_cust1, v_shop1, v_veh1, 'in_progress',            CURRENT_DATE - INTERVAL '3 days', NULL),
    (v_cust2, v_shop2, v_veh2, 'waiting_for_inspection',  CURRENT_DATE - INTERVAL '1 days', NULL);

  RAISE NOTICE 'counts -> shops:% customers:% vehicles:% work_orders:%',
    (SELECT count(*) FROM app.shop       WHERE code IN ('DEMO1','DEMO2')),
    (SELECT count(*) FROM app.customers  WHERE email IN ('demo.cust1@hvzsure.local','demo.cust2@hvzsure.local')),
    (SELECT count(*) FROM app.vehicles   WHERE vin   IN ('1HGCM82633A004352','JH4KA4650MC000123')),
    (SELECT count(*) FROM app.work_orders WHERE customer_id IN (v_cust1, v_cust2));
END $$;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
  v_cust1 uuid;
  v_cust2 uuid;
  v_veh1  uuid;
  v_veh2  uuid;
  v_shop1 uuid;
  v_shop2 uuid;
BEGIN
  SELECT customer_id INTO v_cust1 FROM app.customers WHERE email='demo.cust1@hvzsure.local' LIMIT 1;
  SELECT customer_id INTO v_cust2 FROM app.customers WHERE email='demo.cust2@hvzsure.local' LIMIT 1;
  SELECT vehicle_id  INTO v_veh1  FROM app.vehicles  WHERE vin='1HGCM82633A004352' LIMIT 1;
  SELECT vehicle_id  INTO v_veh2  FROM app.vehicles  WHERE vin='JH4KA4650MC000123' LIMIT 1;
  SELECT id     INTO v_shop1 FROM app.shop      WHERE code='DEMO1' LIMIT 1;
  SELECT id     INTO v_shop2 FROM app.shop      WHERE code='DEMO2' LIMIT 1;

  DELETE FROM app.work_orders
   WHERE (customer_id IN (v_cust1, v_cust2))
      OR (vehicle_id  IN (v_veh1, v_veh2))
      OR (shop_id     IN (v_shop1, v_shop2));

  DELETE FROM app.vehicles
   WHERE vin IN ('1HGCM82633A004352','JH4KA4650MC000123');

  DELETE FROM app.customers
   WHERE email IN ('demo.cust1@hvzsure.local','demo.cust2@hvzsure.local');

  DELETE FROM app.shop
   WHERE code IN ('DEMO1','DEMO2');
END $$;
-- +goose StatementEnd
