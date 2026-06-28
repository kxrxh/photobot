-- Migration: Add classification_params table and indexes
-- Safe to run multiple times

-- Ensure helper trigger function exists
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_proc p
    JOIN pg_namespace n ON n.oid = p.pronamespace
    WHERE p.proname = 'update_updated_at_column' AND n.nspname = 'public'
  ) THEN
    EXECUTE $$
      CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
      BEGIN
        NEW.updated_at = NOW();
        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql;
    $$;
  END IF;
END$$;

-- Create table if it doesn't exist
CREATE TABLE IF NOT EXISTS classification_params (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_classification_params_name ON classification_params(name);
CREATE INDEX IF NOT EXISTS idx_classification_params_created_at ON classification_params(created_at);

-- updated_at trigger (create only if missing)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'update_classification_params_updated_at'
  ) THEN
    EXECUTE 'CREATE TRIGGER update_classification_params_updated_at
             BEFORE UPDATE ON classification_params
             FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();';
  END IF;
END$$;


