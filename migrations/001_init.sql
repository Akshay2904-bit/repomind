--Enable vector extension
CREATE EXTENSION IF NOT EXISTS vector;

--Create repositories table
CREATE TABLE IF NOT EXISTS repositories (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    url TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
    );

--table to store chunks and create their vector embeddings
CREATE TABLE code_chunks (
    id SERIAL PRIMARY KEY,
    repo_id INT REFERENCES repositories(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    start_line INT,
    end_line INT,
    embedding VECTOR(1536), -- Assuming 1536 dimensions for the embedding
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- HNSW index enables fast approximate nearest-neighbour search
-- This is what makes vector search sub-50ms even with millions of rows
CREATE INDEX ON code_chunks USING hnsw (embedding vector_cosine_ops);