-- Rollback: Change embedding column back to 1536 dimensions (for OpenAI)

-- Drop the existing index
DROP INDEX IF EXISTS ev_specs_embedding_idx;

-- Alter the embedding column back to 1536 dimensions
-- Note: This will fail if there are existing rows with 768-dimensional vectors
ALTER TABLE ev_specs ALTER COLUMN embedding TYPE vector(1536);

-- Recreate the index with the original dimension
CREATE INDEX ev_specs_embedding_idx ON ev_specs 
 USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

