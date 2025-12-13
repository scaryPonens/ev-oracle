-- Drop index
DROP INDEX IF EXISTS ev_specs_embedding_idx;

-- Drop table
DROP TABLE IF EXISTS ev_specs;

-- Note: We don't drop the vector extension as it might be used by other tables

