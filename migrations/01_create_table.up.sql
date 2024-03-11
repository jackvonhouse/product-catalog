BEGIN;

DROP TABLE IF EXISTS category CASCADE;
CREATE TABLE IF NOT EXISTS category (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,

    CONSTRAINT category_unique UNIQUE (name)
);

DROP TABLE IF EXISTS product CASCADE;
CREATE TABLE IF NOT EXISTS product (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,

    CONSTRAINT product_unique UNIQUE (name)
);

DROP TABLE IF EXISTS product_of_category CASCADE;
CREATE TABLE IF NOT EXISTS product_of_category (
    product_id INTEGER NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES category(id) ON DELETE CASCADE,

    CONSTRAINT unique_product_of_category UNIQUE (product_id, category_id)
);

COMMIT;
