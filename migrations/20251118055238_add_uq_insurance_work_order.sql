-- +goose Up
ALTER TABLE app.insurance
  ADD CONSTRAINT uq_insurance_work_order UNIQUE (work_order_id);

ALTER TABLE app.insurance
  ALTER COLUMN agent_first_name DROP NOT NULL,
  ALTER COLUMN agent_last_name DROP NOT NULL,
  ALTER COLUMN agent_phone DROP NOT NULL,
  ALTER COLUMN policy_number DROP NOT NULL,
  ALTER COLUMN claim_number DROP NOT NULL;

-- +goose Down
ALTER TABLE app.insurance
  DROP CONSTRAINT IF EXISTS uq_insurance_work_order;

ALTER TABLE app.insurance
  ALTER COLUMN agent_first_name SET NOT NULL,
  ALTER COLUMN agent_last_name SET NOT NULL,
  ALTER COLUMN agent_phone SET NOT NULL,
  ALTER COLUMN policy_number SET NOT NULL,
  ALTER COLUMN claim_number SET NOT NULL;
