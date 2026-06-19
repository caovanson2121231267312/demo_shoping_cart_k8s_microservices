CREATE TABLE IF NOT EXISTS articles (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title        VARCHAR(500) NOT NULL,
    slug         VARCHAR(500) UNIQUE NOT NULL,
    excerpt      TEXT,
    content      TEXT NOT NULL DEFAULT '',
    cover_image  TEXT,
    category     VARCHAR(100) NOT NULL DEFAULT 'tin-tuc',
    tags         TEXT[] NOT NULL DEFAULT '{}',
    author_name  VARCHAR(255) NOT NULL DEFAULT 'Shop Cao Văn Sơn',
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    is_featured  BOOLEAN NOT NULL DEFAULT FALSE,
    view_count   INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
CREATE INDEX IF NOT EXISTS idx_articles_published_created ON articles(is_published, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category);
CREATE INDEX IF NOT EXISTS idx_articles_featured ON articles(is_featured) WHERE is_featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_articles_title_trgm ON articles USING gin (title gin_trgm_ops);
