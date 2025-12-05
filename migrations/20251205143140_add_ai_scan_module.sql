-- +goose Up
-- +goose StatementBegin
------------------------------------------------------------
-- AI Scan Module
-- - ai_scan_job: scan jobs per work order
-- - ai_scan_job_image: per-image status within a job
-- - ai_detection_raw: raw JSON payloads
-- - ai_detection: structured per-detection data
------------------------------------------------------------

------------------------------------------------------------
-- 1) ai_scan_job: one scan request per work order
------------------------------------------------------------
CREATE TABLE app.ai_scan_job (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    work_order_id uuid NOT NULL
        REFERENCES app.work_orders(id),

    requested_by_user_id uuid NOT NULL
        REFERENCES app.users(id),

    status text NOT NULL CHECK (
        status IN ('pending', 'queued', 'running', 'completed', 'failed', 'canceled')
    ),

    model_name text NOT NULL,
    model_version text,

    total_images integer NOT NULL DEFAULT 0,
    success_images integer NOT NULL DEFAULT 0,
    failed_images integer NOT NULL DEFAULT 0,
    total_detections integer,

    external_job_id text,

    requested_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    completed_at timestamptz,
    canceled_at timestamptz,

    error_code text,
    error_message text,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_scan_job_work_order_id
    ON app.ai_scan_job(work_order_id);

CREATE INDEX idx_ai_scan_job_status
    ON app.ai_scan_job(status);


------------------------------------------------------------
-- 2) ai_scan_job_image: job × image with per-image status
------------------------------------------------------------
CREATE TABLE app.ai_scan_job_image (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    ai_scan_job_id uuid NOT NULL
        REFERENCES app.ai_scan_job(id),

    work_order_image_id uuid NOT NULL
        REFERENCES app.work_order_image(id),

    status text NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'queued', 'running', 'success', 'failed', 'skipped', 'canceled')
    ),

    error_code text,
    error_message text,
    request_payload jsonb,

    processed_at timestamptz,
    detection_count integer,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_scan_job_image_job_id
    ON app.ai_scan_job_image(ai_scan_job_id);

CREATE INDEX idx_ai_scan_job_image_image_id
    ON app.ai_scan_job_image(work_order_image_id);


------------------------------------------------------------
-- 3) ai_detection_raw: raw JSON per processed image
------------------------------------------------------------
CREATE TABLE app.ai_detection_raw (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    ai_scan_job_image_id uuid NOT NULL
        REFERENCES app.ai_scan_job_image(id),

    raw_payload jsonb NOT NULL,

    provider text,
    schema_version text,

    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_detection_raw_job_image_id
    ON app.ai_detection_raw(ai_scan_job_image_id);


------------------------------------------------------------
-- 4) ai_detection: structured single detection result
------------------------------------------------------------
CREATE TABLE app.ai_detection (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    ai_scan_job_image_id uuid NOT NULL
        REFERENCES app.ai_scan_job_image(id),

    work_order_image_id uuid NOT NULL
        REFERENCES app.work_order_image(id),

    work_order_id uuid NOT NULL
        REFERENCES app.work_orders(id),

    model_category text,
    mapped_category text,

    confidence numeric(4,3),
    severity text,

    bbox jsonb,
    polygon jsonb,
    area numeric,

    is_false_positive boolean NOT NULL DEFAULT false,

    status text NOT NULL DEFAULT 'proposed' CHECK (
        status IN ('proposed', 'accepted', 'rejected', 'hidden')
    ),

    notes text,

    created_by_user_id uuid NOT NULL
        REFERENCES app.users(id),
    updated_by_user_id uuid,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX idx_ai_detection_work_order_id_not_deleted
    ON app.ai_detection(work_order_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_ai_detection_image_id_not_deleted
    ON app.ai_detection(work_order_image_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_ai_detection_status_not_deleted
    ON app.ai_detection(status)
    WHERE deleted_at IS NULL;


------------------------------------------------------------
-- 5) updated_at triggers (using app.set_updated_at)
------------------------------------------------------------
CREATE TRIGGER trg_set_updated_at_ai_scan_job
BEFORE UPDATE ON app.ai_scan_job
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

CREATE TRIGGER trg_set_updated_at_ai_scan_job_image
BEFORE UPDATE ON app.ai_scan_job_image
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();

CREATE TRIGGER trg_set_updated_at_ai_detection
BEFORE UPDATE ON app.ai_detection
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
------------------------------------------------------------
-- Rollback AI Scan Module
------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_set_updated_at_ai_detection
    ON app.ai_detection;

DROP TRIGGER IF EXISTS trg_set_updated_at_ai_scan_job_image
    ON app.ai_scan_job_image;

DROP TRIGGER IF EXISTS trg_set_updated_at_ai_scan_job
    ON app.ai_scan_job;

DROP INDEX IF EXISTS idx_ai_detection_status_not_deleted;
DROP INDEX IF EXISTS idx_ai_detection_image_id_not_deleted;
DROP INDEX IF EXISTS idx_ai_detection_work_order_id_not_deleted;
DROP INDEX IF EXISTS idx_ai_detection_raw_job_image_id;
DROP INDEX IF EXISTS idx_ai_scan_job_image_image_id;
DROP INDEX IF EXISTS idx_ai_scan_job_image_job_id;
DROP INDEX IF EXISTS idx_ai_scan_job_status;
DROP INDEX IF EXISTS idx_ai_scan_job_work_order_id;

DROP TABLE IF EXISTS app.ai_detection;
DROP TABLE IF EXISTS app.ai_detection_raw;
DROP TABLE IF EXISTS app.ai_scan_job_image;
DROP TABLE IF EXISTS app.ai_scan_job;
-- +goose StatementEnd