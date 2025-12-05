-- +goose Up
-- +goose StatementBegin
ALTER TABLE app.work_orders
    ADD COLUMN latest_ai_scan_job_id uuid;

ALTER TABLE app.work_orders
    ADD CONSTRAINT fk_work_order_latest_ai_scan_job
    FOREIGN KEY (latest_ai_scan_job_id)
    REFERENCES app.ai_scan_job(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app.work_orders
    DROP CONSTRAINT IF EXISTS fk_work_order_latest_ai_scan_job;

ALTER TABLE app.work_orders
    DROP COLUMN IF EXISTS latest_ai_scan_job_id;
-- +goose StatementEnd
