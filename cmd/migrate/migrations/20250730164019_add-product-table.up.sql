CREATE TABLE IF NOT EXISTS products (
     id SERIAL PRIMARY KEY,
     name VARCHAR(255) NOT NULL,
     description TEXT NOT NULL,
     image VARCHAR(255) NOT NULL,
     price NUMERIC(10, 2) NOT NULL,
     quantity INTEGER NOT NULL CHECK (quantity >= 0),
     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
