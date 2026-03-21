ALTER TABLE categories ADD COLUMN account_id INTEGER NOT NULL REFERENCES accounts(id);
CREATE UNIQUE INDEX idx_categories_account_name ON categories(account_id, name);
CREATE INDEX idx_categories_account_id ON categories(account_id);
