CREATE TABLE site_settings (
  key VARCHAR(100) PRIMARY KEY,
  value JSONB NOT NULL,
  is_public BOOLEAN NOT NULL DEFAULT false,
  updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE operation_logs (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  actor_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
  action VARCHAR(100) NOT NULL,
  resource_type VARCHAR(50) NOT NULL,
  resource_id VARCHAR(100) NOT NULL,
  request_id VARCHAR(100),
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  ip_hash CHAR(64),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX operation_logs_resource_idx
  ON operation_logs (resource_type, resource_id, created_at DESC);
CREATE INDEX operation_logs_actor_idx
  ON operation_logs (actor_id, created_at DESC);

CREATE TABLE media_assets (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  object_key VARCHAR(500) NOT NULL UNIQUE,
  original_name VARCHAR(255) NOT NULL,
  mime_type VARCHAR(100) NOT NULL,
  byte_size BIGINT NOT NULL,
  width INTEGER NOT NULL,
  height INTEGER NOT NULL,
  sha256 CHAR(64) NOT NULL,
  uploaded_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT media_assets_size_check CHECK (byte_size > 0),
  CONSTRAINT media_assets_dimensions_check CHECK (width > 0 AND height > 0)
);

CREATE INDEX media_assets_created_idx ON media_assets (created_at DESC);

INSERT INTO site_settings (key, value, is_public)
VALUES
  ('profile', '{"name":"Rodolfo Iolo","headline":"Computer Science Student","bio":"Learning in public through software projects."}', true),
  ('social_links', '[]', true),
  ('site', '{"title":"Rodolfo Iolo","description":"Articles, projects, and engineering notes."}', true);
