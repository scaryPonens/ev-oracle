-- Update embedding column to support 768 dimensions (for Ollama nomic-embed-text)
-- This migration changes the vector dimension from 1536 to 768

-- Drop the existing index
DROP INDEX IF EXISTS ev_specs_embedding_idx;

-- Alter the embedding column to 768 dimensions
-- Note: This will fail if there are existing rows with 1536-dimensional vectors
-- You may need to clear existing data first or create a new column and migrate
ALTER TABLE ev_specs ALTER COLUMN embedding TYPE vector(768);

-- Recreate the index with the new dimension
CREATE INDEX ev_specs_embedding_idx ON ev_specs 
 USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

