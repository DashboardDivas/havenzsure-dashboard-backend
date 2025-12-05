-- +goose Up
-- +goose StatementBegin
-- 1) work_order_image: original images for work orders + soft delete + status
CREATE TABLE app.work_order_image (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    work_order_id uuid NOT NULL
        REFERENCES app.work_orders(id),

    storage_path text NOT NULL,          -- GCS object path
    public_url text,                     -- URL for frontend access (e.g., signed URL or CDN)
    thumbnail_url text,                  -- Thumbnail (optional)

    original_filename text,              -- Filename when uploaded by user
    mime_type text,
    file_size_bytes bigint,
    width_px integer,
    height_px integer,

    -- view_angle text,                     --not implemented, but might be useful for future use: front / rear / left / right / roof / detail
    -- sort_order integer,

    status text NOT NULL DEFAULT 'draft', -- draft / ready_for_scan / scanning / scan_completed / scan_failed / archived

    created_by_user_id uuid
        REFERENCES app.users(id),
    updated_by_user_id uuid,
    deleted_at timestamptz,
    deleted_by_user_id uuid,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- 2) index: find wo images by work_order_id + status + not deleted
CREATE INDEX idx_work_order_image_work_order_id_not_deleted
    ON app.work_order_image(work_order_id)
    WHERE deleted_at IS NULL;

-- 3) trigger: auto-update updated_at on row update
CREATE TRIGGER trg_set_updated_at_work_order_image
BEFORE UPDATE ON app.work_order_image
FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_set_updated_at_work_order_image
    ON app.work_order_image;

DROP TABLE IF EXISTS app.work_order_image;
-- +goose StatementEnd
