CREATE TABLE categories (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(120) NOT NULL UNIQUE,
  description VARCHAR(500),
  icon VARCHAR(100),
  parent_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT categories_not_self_parent CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE TABLE tags (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(120) NOT NULL UNIQUE,
  color CHAR(7),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT tags_color_check CHECK (color IS NULL OR color ~ '^#[0-9A-Fa-f]{6}$')
);

CREATE TABLE articles (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug VARCHAR(200) NOT NULL UNIQUE,
  title VARCHAR(300) NOT NULL,
  summary VARCHAR(500) NOT NULL,
  content TEXT NOT NULL,
  cover_image VARCHAR(500),
  category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
  author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  status VARCHAR(20) NOT NULL DEFAULT 'draft',
  is_top BOOLEAN NOT NULL DEFAULT false,
  view_count BIGINT NOT NULL DEFAULT 0,
  like_count BIGINT NOT NULL DEFAULT 0,
  comment_count BIGINT NOT NULL DEFAULT 0,
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT articles_status_check CHECK (status IN ('draft', 'published', 'archived')),
  CONSTRAINT articles_publish_time_check CHECK (status <> 'published' OR published_at IS NOT NULL),
  CONSTRAINT articles_nonnegative_counts CHECK (
    view_count >= 0 AND like_count >= 0 AND comment_count >= 0
  )
);

CREATE INDEX articles_public_idx
  ON articles (is_top DESC, published_at DESC)
  WHERE status = 'published';
CREATE INDEX articles_category_idx ON articles (category_id);
CREATE INDEX articles_title_lower_idx ON articles (lower(title));

CREATE TABLE article_tags (
  article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (article_id, tag_id)
);

CREATE INDEX article_tags_tag_idx ON article_tags (tag_id, article_id);

CREATE TABLE projects (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug VARCHAR(200) NOT NULL UNIQUE,
  name VARCHAR(200) NOT NULL,
  summary VARCHAR(500) NOT NULL,
  content TEXT NOT NULL,
  cover_image VARCHAR(500),
  technologies JSONB NOT NULL DEFAULT '[]'::jsonb,
  repository_url VARCHAR(500),
  demo_url VARCHAR(500),
  featured BOOLEAN NOT NULL DEFAULT false,
  sort_order INTEGER NOT NULL DEFAULT 0,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  started_at DATE,
  completed_at DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT projects_status_check CHECK (status IN ('active', 'archived')),
  CONSTRAINT projects_technologies_array_check CHECK (jsonb_typeof(technologies) = 'array'),
  CONSTRAINT projects_date_order_check CHECK (
    completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at
  )
);

CREATE INDEX projects_public_sort_idx
  ON projects (featured DESC, sort_order, updated_at DESC)
  WHERE status = 'active';
CREATE INDEX projects_technologies_gin_idx ON projects USING GIN (technologies);
