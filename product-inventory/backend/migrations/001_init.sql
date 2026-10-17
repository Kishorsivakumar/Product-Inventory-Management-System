CREATE TABLE IF NOT EXISTS categories (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    sku         TEXT NOT NULL UNIQUE,
    category_id BIGINT NOT NULL REFERENCES categories(id),
    price       NUMERIC(12,2) NOT NULL CHECK (price > 0),
    quantity    INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive','discontinued')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);

INSERT INTO categories (name) VALUES ('Electronics'), ('Furniture'), ('Stationery'), ('Accessories')
ON CONFLICT (name) DO NOTHING;
