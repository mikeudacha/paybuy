CREATE TABLE IF NOT EXISTS order_items(
    id         SERIAL PRIMARY KEY,
    order_id   INTEGER        NOT NULL REFERENCES orders (id),
    product_id INTEGER        NOT NULL REFERENCES products (id),
    quantity   INTEGER        NOT NULL CHECK (quantity > 0),
    price      NUMERIC(10, 2) NOT NULL
);