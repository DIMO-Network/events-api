-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS events_user_id_index
    on events_api.events (user_id);

CREATE INDEX IF NOT EXISTS events_user_id_timestamp_index
    on events_api.events (user_id, timestamp);

CREATE INDEX IF NOT EXISTS events_user_id_device_id_timestamp_index
    on events_api.events (user_id, device_id, timestamp);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX events_api.events_user_id_index;
DROP INDEX events_api.events_user_id_timestamp_index;
DROP INDEX events_api.events_user_id_device_id_timestamp_index;
-- +goose StatementEnd
