-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create ev_specs table
CREATE TABLE IF NOT EXISTS ev_specs (
    id SERIAL PRIMARY KEY,
    make VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL,
    capacity_kwh FLOAT NOT NULL,
    power_kw FLOAT NOT NULL,
    chemistry VARCHAR(100) NOT NULL,
    embedding vector(1536),  -- Dimension for OpenAI text-embedding-3-small model
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(make, model, year)
);

-- Create IVFFlat index for vector similarity search
-- lists = 100 is a good default for small to medium datasets (up to 10,000 rows)
-- For larger datasets, consider using sqrt(rows) for optimal performance
CREATE INDEX IF NOT EXISTS ev_specs_embedding_idx ON ev_specs 
 USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

