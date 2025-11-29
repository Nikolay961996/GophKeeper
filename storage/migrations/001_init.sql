-- пользователи
CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY,
                                     login TEXT UNIQUE NOT NULL,
                                     password_hash TEXT NOT NULL,
                                     created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- секретные данные
CREATE TABLE IF NOT EXISTS secrets (
                                       id UUID PRIMARY KEY,
                                       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    data BYTEA NOT NULL,
    metadata TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                                                                      UNIQUE(user_id, id)
    );

CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id);
CREATE INDEX IF NOT EXISTS idx_secrets_updated_at ON secrets(updated_at);
CREATE INDEX IF NOT EXISTS idx_secrets_user_updated ON secrets(user_id, updated_at);