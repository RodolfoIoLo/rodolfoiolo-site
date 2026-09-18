CREATE TABLE comments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  parent_id BIGINT REFERENCES comments(id) ON DELETE CASCADE,
  author_name VARCHAR(100) NOT NULL,
  author_email_hash CHAR(64),
  content TEXT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  visitor_id CHAR(64) NOT NULL,
  ip_hash CHAR(64) NOT NULL,
  user_agent VARCHAR(500),
  reviewed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT comments_status_check CHECK (status IN ('pending', 'approved', 'rejected')),
  CONSTRAINT comments_not_self_parent CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE INDEX comments_article_status_created_idx
  ON comments (article_id, status, created_at);
CREATE INDEX comments_pending_idx
  ON comments (created_at)
  WHERE status = 'pending';

CREATE TABLE likes (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  visitor_id CHAR(64) NOT NULL,
  ip_hash CHAR(64) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT likes_article_visitor_unique UNIQUE (article_id, visitor_id)
);
