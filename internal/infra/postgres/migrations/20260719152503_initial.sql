CREATE TABLE messages (
    id         BIGSERIAL PRIMARY KEY,
    user_id    INTEGER   NOT NULL,
    chat_id    TEXT      NOT NULL,
    text       TEXT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
