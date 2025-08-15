CREATE TABLE IF NOT EXISTS orders (
     id SERIAL PRIMARY KEY,
     user_id INTEGER NOT NULL REFERENCES users(id),
     total NUMERIC(10, 2) NOT NULL,
     status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'cancelled')),
     address TEXT NOT NULL,
     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
