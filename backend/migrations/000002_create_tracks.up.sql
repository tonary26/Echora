CREATE TABLE tracks (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    cover_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX tracks_user_id_idx ON tracks(user_id);