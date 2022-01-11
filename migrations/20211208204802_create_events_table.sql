-- +goose Up
-- +goose StatementBegin
CREATE TABLE events_api.events (
    id CHAR(27) PRIMARY KEY, -- KSUID
    type TEXT NOT NULL,
    sub_type TEXT NOT NULL,
    user_id TEXT NOT NULL,
    device_id CHAR(27), -- KSUID
    timestamp timestamptz NOT NULL,
    data jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE events_api.events;
-- +goose StatementEnd
