-- +goose Up
-- +goose StatementBegin
CREATE TABLE events_api.events (
    "type" TEXT NOT NULL,
    source TEXT NOT NULL,
    subject TEXT NOT NULL,
    id TEXT PRIMARY KEY,
    time timestamptz NOT NULL,
    data jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE events_api.events;
-- +goose StatementEnd
