-- Enable the vector extension (required for pgvector)
CREATE EXTENSION IF NOT EXISTS vector;
 
-- Table to track which repositories you have indexed
CREATE TABLE repositories (
    id         SERIAL PRIMARY KEY,
    name       TEXT UNIQUE NOT NULL,
    url        TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
 
-- Table to store code chunks and their vector embeddings
CREATE TABLE code_chunks (
    id         SERIAL PRIMARY KEY,
    repo_id    INT REFERENCES repositories(id) ON DELETE CASCADE,
    file_path  TEXT NOT NULL,
    content    TEXT NOT NULL,
    start_line INT,
    end_line   INT,
    embedding  vector(768),   -- 768 = nomic-embed-text output dimension
    created_at TIMESTAMPTZ DEFAULT NOW()
);
 
-- HNSW index enables fast approximate nearest-neighbour search
-- This is what makes vector search sub-50ms even with millions of rows
CREATE INDEX ON code_chunks USING hnsw (embedding vector_cosine_ops);
