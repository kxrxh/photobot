-- Trigram index for ILIKE '%name%' filters on weeds.name (ListWeeds)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_weeds_name_trgm ON weeds USING gin (name gin_trgm_ops);
