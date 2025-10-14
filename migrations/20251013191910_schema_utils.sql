-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.set_updated_at() RETURNS trigger
AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql
  SET search_path = app, pg_temp;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS app.set_updated_at();
-- +goose StatementEnd