CREATE TABLE IF NOT EXISTS products (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    parent_id UUID NULL REFERENCES products(id),
    category_slug VARCHAR(255) NOT NULL REFERENCES categories(slug) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    image VARCHAR(255) NOT NULL,
    description TEXT NULL,
    reguler_price BIGINT DEFAULT 0,
    sale_price BIGINT DEFAULT 0,
    unit VARCHAR(255) DEFAULT 'gram',
    weight BIGINT DEFAULT 0,
    variant INT DEFAULT 1,
    status VARCHAR(255) DEFAULT 'DRAFT',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL
);

CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_category_slug ON products(category_slug);