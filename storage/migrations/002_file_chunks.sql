-- метаданные файлов
CREATE TABLE IF NOT EXISTS file_metadata (
                                             id UUID PRIMARY KEY,
                                             user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    total_chunks INTEGER NOT NULL,
    chunk_size INTEGER NOT NULL,
    checksum TEXT,
    mime_type TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
                                                                                            );

-- чанки файлов
CREATE TABLE IF NOT EXISTS file_chunks (
                                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES file_metadata(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    chunk_data BYTEA NOT NULL,
    chunk_checksum TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(file_id, chunk_index)
    );

CREATE INDEX IF NOT EXISTS idx_file_chunks_file_id ON file_chunks(file_id);
CREATE INDEX IF NOT EXISTS idx_file_chunks_file_id_index ON file_chunks(file_id, chunk_index);
CREATE INDEX IF NOT EXISTS idx_file_metadata_user_id ON file_metadata(user_id);

ALTER TABLE secrets
    ADD COLUMN IF NOT EXISTS file_id UUID REFERENCES file_metadata(id) ON DELETE SET NULL;