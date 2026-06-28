CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE classification_params (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE classifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL,
  created_by INTEGER NOT NULL,
  is_public BOOLEAN NOT NULL DEFAULT FALSE,
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TYPE LOGIC_OPERATOR AS ENUM ('OR', 'AND');

CREATE TABLE fractions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL,
  classification_id UUID NOT NULL REFERENCES classifications(id) ON DELETE CASCADE,
  order_index INTEGER NOT NULL CHECK (order_index >= 0),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE conditions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fraction_id UUID NOT NULL REFERENCES fractions(id) ON DELETE CASCADE,
  name VARCHAR NOT NULL,
  operator LOGIC_OPERATOR NOT NULL,
  connection LOGIC_OPERATOR NOT NULL,
  order_index INTEGER NOT NULL CHECK (order_index >= 0),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TYPE param_operator AS ENUM ('<', '<=', '==', '>=', '>', '!=');

CREATE TABLE params (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  condition_id UUID NOT NULL REFERENCES conditions(id) ON DELETE CASCADE,
  name VARCHAR NOT NULL,
  operator param_operator NOT NULL,
  value NUMERIC(10, 2) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE user_active_classification (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id INTEGER NOT NULL,
  classification_id UUID NOT NULL REFERENCES classifications(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  UNIQUE(user_id)
);

-- Classifications table indexes
CREATE INDEX idx_classifications_created_by ON classifications(created_by);

CREATE INDEX idx_classifications_product_id ON classifications(product_id);

CREATE INDEX idx_classifications_is_public ON classifications(is_public);

-- Fractions table indexes
CREATE INDEX idx_fractions_classification_id ON fractions(classification_id);

CREATE INDEX idx_fractions_order_index ON fractions(order_index);

CREATE INDEX idx_fractions_classification_order ON fractions(classification_id, order_index);

-- Conditions table indexes
CREATE INDEX idx_conditions_fraction_id ON conditions(fraction_id);

CREATE INDEX idx_conditions_order_index ON conditions(order_index);

CREATE INDEX idx_conditions_fraction_order ON conditions(fraction_id, order_index);

-- Params table indexes
CREATE INDEX idx_params_condition_id ON params(condition_id);

-- User active classification indexes
CREATE INDEX idx_user_active_classification_user_id ON user_active_classification(user_id);

CREATE INDEX idx_user_active_classification_classification_id ON user_active_classification(classification_id);

CREATE INDEX idx_user_active_classification_created_at ON user_active_classification(created_at);

-- Markup
CREATE TABLE markups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL,
  created_by INTEGER NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE markup_fractions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  markup_id UUID NOT NULL REFERENCES markups(id) ON DELETE CASCADE,
  name VARCHAR NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

-- Fraction Objects (association table for markup fractions and objects)
CREATE TABLE markup_fraction_objects (
  markup_fraction_id UUID NOT NULL REFERENCES markup_fractions(id) ON DELETE CASCADE,
  object_id BIGINT NOT NULL,
  PRIMARY KEY (markup_fraction_id, object_id)
);

-- Таблица для связи markup <-> analyses (analyses_ids)
CREATE TABLE markup_analyses (
  markup_id UUID NOT NULL REFERENCES markups(id) ON DELETE CASCADE,
  analysis_id BIGINT NOT NULL,
  PRIMARY KEY (markup_id, analysis_id)
);

-- Function to update updated_at timestamp
CREATE
OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();

RETURN NEW;

END;

$$ language 'plpgsql';

-- Add triggers to tables with updated_at columns
CREATE TRIGGER update_classifications_updated_at BEFORE
UPDATE
  ON classifications FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_fractions_updated_at BEFORE
UPDATE
  ON fractions FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_conditions_updated_at BEFORE
UPDATE
  ON conditions FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_params_updated_at BEFORE
UPDATE
  ON params FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_user_active_classification_updated_at BEFORE
UPDATE
  ON user_active_classification FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_markups_updated_at BEFORE
UPDATE
  ON markups FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_products_updated_at BEFORE
UPDATE
  ON products FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();